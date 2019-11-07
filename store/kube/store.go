/*
 Copyright 2019 Bruce Ma

 Licensed under the Apache License, Version 2.0 (the "License");
 you may not use this file except in compliance with the License.
 You may obtain a copy of the License at

     http://www.apache.org/licenses/LICENSE-2.0

 Unless required by applicable law or agreed to in writing, software
 distributed under the License is distributed on an "AS IS" BASIS,
 WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 See the License for the specific language governing permissions and
 limitations under the License.
*/

package kube

import (
	"fmt"
	"net"
	"sync"
	"time"

	resourcev1 "github.com/mars1024/kube-ipam/pkg/apis/resource/v1"
	"github.com/mars1024/kube-ipam/pkg/client/clientset/versioned"
	"github.com/mars1024/kube-ipam/pkg/client/informers/externalversions"
	"github.com/mars1024/kube-ipam/pkg/utils"
	"github.com/mars1024/kube-ipam/store"
	"github.com/mars1024/kube-ipam/types"
	"github.com/sirupsen/logrus"

	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
)

// check if Store overrides all interfaces of IPAMStore
var _ store.IPAMStore = &Store{}

var LoggerStore = logrus.WithFields(logrus.Fields{"component": "store/kube"})

type Store struct {
	*sync.RWMutex

	resourceClient          versioned.Interface
	resourceInformerFactory externalversions.SharedInformerFactory
	resourceSynced          []cache.InformerSynced

	stopEverything <-chan struct{}

	cache *Cache
}

func NewStore(masterURL, kubeConfig string, stopCh <-chan struct{}) (*Store, error) {
	cfg, err := clientcmd.BuildConfigFromFlags(masterURL, kubeConfig)
	if err != nil {
		return nil, fmt.Errorf("fail to build kubernetes config: %v", err)
	}

	// create resource client
	resourceClient, err := versioned.NewForConfig(cfg)
	if err != nil {
		return nil, fmt.Errorf("fail to new resource client: %v", err)
	}

	// create informer factory
	resourceInformerFactory := externalversions.NewSharedInformerFactory(resourceClient, time.Second*30)

	// create informers
	networkInformer := resourceInformerFactory.Resource().V1().Networks()
	lastReservedIPInformer := resourceInformerFactory.Resource().V1().LastReservedIPs()
	usingIPInformer := resourceInformerFactory.Resource().V1().UsingIPs()

	s := &Store{
		RWMutex:                 new(sync.RWMutex),
		resourceClient:          resourceClient,
		resourceInformerFactory: resourceInformerFactory,
		resourceSynced: []cache.InformerSynced{
			networkInformer.Informer().HasSynced,
			lastReservedIPInformer.Informer().HasSynced,
			usingIPInformer.Informer().HasSynced,
		},
		stopEverything: stopCh,
		cache:          NewCache(),
	}

	// add handlers
	LoggerStore.Info("Setting up event handlers")
	networkInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    s.addNetworkToCache,
		UpdateFunc: s.updateNetworkInCache,
		DeleteFunc: s.deleteNetworkFromCache,
	})

	lastReservedIPInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    s.addLastReservedIPToCache,
		UpdateFunc: s.updateLastReservedIPInCache,
		DeleteFunc: s.deleteLastReservedIPFromCache,
	})

	usingIPInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    s.addUsingIPToCache,
		UpdateFunc: s.updateUsingIPInCache,
		DeleteFunc: s.deleteUsingIPFromCache,
	})

	return s, nil
}

func (s *Store) Run() error {
	LoggerStore.Debug("starting resource informer factory")
	go s.resourceInformerFactory.Start(s.stopEverything)

	LoggerStore.Info("waiting for caches to sync")
	if ok := cache.WaitForCacheSync(s.stopEverything, s.resourceSynced...); !ok {
		return fmt.Errorf("fail to sync caches")
	}

	// non-blocking
	go func() {
		<-s.stopEverything
		LoggerStore.Info("kube store shutting down...")
	}()
	return nil
}

func (s *Store) CreateNetwork(name string) error {
	s.Lock()
	defer s.Unlock()

	if networkCache := s.cache.GetNetwork(name); networkCache != nil {
		return fmt.Errorf("network %s already exists", name)
	}

	// create empty network
	network := &resourcev1.Network{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
	}
	if _, err := s.resourceClient.ResourceV1().Networks().Create(network); err != nil {
		return err
	}

	return nil
}

func (s *Store) DeleteNetwork(name string) error {
	s.Lock()
	defer s.Unlock()

	networkCache := s.cache.GetNetwork(name)
	if networkCache == nil {
		return fmt.Errorf("network %s is not in cache", name)
	}
	if len(networkCache.Pools) > 0 {
		return fmt.Errorf("network with %s pools is not allowed to be deleted", len(networkCache.Pools))
	}

	if err := s.resourceClient.ResourceV1().Networks().Delete(name, nil); err != nil {
		if errors.IsNotFound(err) {
			return nil
		}
		return err
	}

	return nil
}

func (s *Store) GetNetwork(name string) (*types.Network, error) {
	s.RLock()
	defer s.RUnlock()

	networkCache := s.cache.GetNetwork(name)
	if networkCache == nil {
		return nil, fmt.Errorf("network %s is not in cache", name)
	}

	return networkCache, nil
}

func (s *Store) GetLastReservedIP(name string) (*types.LastReservedIP, error) {
	s.RLock()
	defer s.RUnlock()

	lriCache := s.cache.GetLastReservedIP(name)
	if lriCache == nil {
		return nil, fmt.Errorf("last reserved ip %s is not in cache", name)
	}

	return lriCache, nil
}

func (s *Store) AddPool(name string, pool *types.Pool) error {
	s.Lock()
	defer s.Unlock()

	// check existing and overlap for network
	networkCache := s.cache.GetNetwork(name)
	if networkCache == nil {
		return fmt.Errorf("network %s is not in cache", name)
	}
	for _, p := range networkCache.Pools {
		switch {
		case pool.Name == p.Name:
			return fmt.Errorf("network %s already has pool %s", name, pool.Name)
		case pool.Overlaps(p):
			return fmt.Errorf("new pool %+v overlaps old pool %+v in network %s", pool, p, name)
		}
	}

	// check and canonicalize pool
	if err := pool.Canonicalize(); err != nil {
		return err
	}

	// append pool to network
	network, err := s.resourceClient.ResourceV1().Networks().Get(name, metav1.GetOptions{})
	if err != nil {
		return err
	}
	networkClone := network.DeepCopy()
	networkClone.Spec.Pools = append(networkClone.Spec.Pools, resourcev1.Pool{
		Name:      pool.Name,
		PoolStart: pool.PoolStart.String(),
		PoolEnd:   pool.PoolEnd.String(),
		Gateway:   pool.Gateway.String(),
		Subnet:    pool.Subnet.String(),
		VlanId:    pool.VlanID,
	})
	if _, err = s.resourceClient.ResourceV1().Networks().Create(networkClone); err != nil {
		return err
	}

	return nil
}

func (s *Store) DelPool(networkName, poolName string) error {
	s.Lock()
	defer s.Unlock()

	// get network from kubernetes
	network, err := s.resourceClient.ResourceV1().Networks().Get(networkName, metav1.GetOptions{})
	if err != nil {
		return err
	}

	networkClone := network.DeepCopy()

	// get pool index, judge if pool is empty
	poolIndex := -1
	for index, pool := range networkClone.Spec.Pools {
		if pool.Name == poolName {
			poolIndex = index
			//TODO: check pool count
		}
	}
	if poolIndex < 0 {
		return fmt.Errorf("network %s does not have pool %s", networkName, poolName)
	}

	// remove pool from network
	networkClone.Spec.Pools = append(networkClone.Spec.Pools[:poolIndex], networkClone.Spec.Pools[poolIndex+1:]...)
	if _, err = s.resourceClient.ResourceV1().Networks().Update(networkClone); err != nil {
		return err
	}

	return nil
}

func (*Store) CountPool(network, pool string) (total, used int, err error) {
	panic("implement me")
}

func (s *Store) Reserve(network, pool, namespace, name string, ip net.IP) (bool, error) {
	s.Lock()
	defer s.Unlock()

	if s.cache.IsIPUsing(utils.ToKubeName(ip.String())) {
		return false, nil
	}

	reserved, err := s.createUsingIP(network, pool, namespace, name, ip.String())
	if reserved {
		// fail safe
		_ = s.updateLastReservedIP(network, pool, ip.String())
	}

	return reserved, err
}

func (s *Store) Release(ip net.IP) error {
	s.Lock()
	defer s.Unlock()

	return s.deleteUsingIP(ip.String())
}

func (*Store) ReleaseByName(network, pool, namespace, name string) error {
	panic("implement me")
}

func (s *Store) addNetworkToCache(obj interface{}) {
	network, ok := obj.(*resourcev1.Network)
	if !ok {
		return
	}

	s.cache.addNetwork(network)
}

func (s *Store) updateNetworkInCache(oldObj, newObj interface{}) {
	oldNetwork, ok := oldObj.(*resourcev1.Network)
	if !ok {
		return
	}
	newNetwork, ok := newObj.(*resourcev1.Network)
	if !ok {
		return
	}
	if oldNetwork.ResourceVersion == newNetwork.ResourceVersion {
		return
	}

	s.cache.updateNetwork(newNetwork)
}

func (s *Store) deleteNetworkFromCache(obj interface{}) {
	var network *resourcev1.Network
	switch t := obj.(type) {
	case *resourcev1.Network:
		network = t
	case cache.DeletedFinalStateUnknown:
		var ok bool
		network, ok = t.Obj.(*resourcev1.Network)
		if !ok {
			return
		}
	default:
		return
	}

	s.cache.deleteNetwork(network)
}

func (s *Store) addLastReservedIPToCache(obj interface{}) {
	lastReservedIP, ok := obj.(*resourcev1.LastReservedIP)
	if !ok {
		return
	}

	s.cache.addLastReservedIP(lastReservedIP)
}

func (s *Store) updateLastReservedIPInCache(oldObj, newObj interface{}) {
	oldLastReservedIP, ok := oldObj.(*resourcev1.LastReservedIP)
	if !ok {
		return
	}
	newLastReservedIP, ok := newObj.(*resourcev1.LastReservedIP)
	if !ok {
		return
	}
	if oldLastReservedIP.ResourceVersion == newLastReservedIP.ResourceVersion {
		return
	}

	s.cache.updateLastReservedIP(newLastReservedIP)
}

func (s *Store) deleteLastReservedIPFromCache(obj interface{}) {
	var lastReservedIP *resourcev1.LastReservedIP
	switch t := obj.(type) {
	case *resourcev1.LastReservedIP:
		lastReservedIP = t
	case cache.DeletedFinalStateUnknown:
		var ok bool
		lastReservedIP, ok = t.Obj.(*resourcev1.LastReservedIP)
		if !ok {
			return
		}
	default:
		return
	}

	s.cache.deleteLastReservedIP(lastReservedIP)
}

func (s *Store) addUsingIPToCache(obj interface{}) {
	usingIP, ok := obj.(*resourcev1.UsingIP)
	if !ok {
		return
	}

	s.cache.addUsingIP(usingIP)
}

func (s *Store) updateUsingIPInCache(oldObj, newObj interface{}) {
	oldUsingIP, ok := oldObj.(*resourcev1.UsingIP)
	if !ok {
		return
	}
	newUsingIP, ok := newObj.(*resourcev1.UsingIP)
	if !ok {
		return
	}
	if oldUsingIP.ResourceVersion == newUsingIP.ResourceVersion {
		return
	}

	s.cache.updateUsingIP(newUsingIP)
}

func (s *Store) deleteUsingIPFromCache(obj interface{}) {
	var usingIP *resourcev1.UsingIP
	switch t := obj.(type) {
	case *resourcev1.UsingIP:
		usingIP = t
	case cache.DeletedFinalStateUnknown:
		var ok bool
		usingIP, ok = t.Obj.(*resourcev1.UsingIP)
		if !ok {
			return
		}
	default:
		return
	}

	s.cache.deleteUsingIP(usingIP)
}

func (s *Store) createUsingIP(network, pool, namespace, name, ip string) (bool, error) {
	usingIP := &resourcev1.UsingIP{
		ObjectMeta: metav1.ObjectMeta{
			Name: utils.ToKubeName(ip),
		},
		Spec: resourcev1.UsingIPSpec{
			PodName:      name,
			PodNamespace: namespace,
			Network:      network,
			Pool:         pool,
		},
	}

	_, err := s.resourceClient.ResourceV1().UsingIPs().Create(usingIP)
	if err != nil && errors.IsAlreadyExists(err) {
		return false, nil
	}
	if err != nil {
		return false, err
	}

	return true, nil
}

func (s *Store) deleteUsingIP(ip string) error {
	return s.resourceClient.ResourceV1().UsingIPs().Delete(utils.ToKubeName(ip), nil)
}

func (s *Store) createLastReservedIP(networkName, poolName, ip string) error {
	lri := &resourcev1.LastReservedIP{
		ObjectMeta: metav1.ObjectMeta{
			Name: networkName,
		},
		Spec: resourcev1.LastReservedIPSpec{
			IP:       ip,
			PoolName: poolName,
		},
	}

	if _, err := s.resourceClient.ResourceV1().LastReservedIPs().Create(lri); err != nil {
		return err
	}
	return nil
}

func (s *Store) updateLastReservedIP(networkName, poolName, ip string) error {
	odlLri, err := s.resourceClient.ResourceV1().LastReservedIPs().Get(networkName, metav1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			return s.createLastReservedIP(networkName, poolName, ip)
		}
		return err
	}

	newLri := odlLri.DeepCopy()
	newLri.Spec.IP = ip
	newLri.Spec.PoolName = poolName

	if _, err := s.resourceClient.ResourceV1().LastReservedIPs().Update(newLri); err != nil {
		return err
	}
	return nil
}

func (s *Store) deleteLastReservedIP(networkName string) error {
	return s.resourceClient.ResourceV1().LastReservedIPs().Delete(networkName, nil)
}
