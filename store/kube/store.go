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
	resourcev1 "github.com/mars1024/kube-ipam/pkg/apis/resource/v1"
	"github.com/mars1024/kube-ipam/pkg/client/clientset/versioned"
	"github.com/mars1024/kube-ipam/pkg/client/informers/externalversions"
	"github.com/sirupsen/logrus"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
	"net"
	"sync"
	"time"

	"github.com/mars1024/kube-ipam/store"
	"github.com/mars1024/kube-ipam/types"
)

// check if Store overrides all interfaces of IPAMStore
var _ store.IPAMStore = &Store{}

var LoggerStore = logrus.WithFields(logrus.Fields{"component": "store/kube"})

type Store struct {
	*sync.Mutex

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

	store := &Store{
		Mutex:                   new(sync.Mutex),
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
		AddFunc:    store.addNetworkToCache,
		UpdateFunc: store.updateNetworkInCache,
		DeleteFunc: store.deleteNetworkFromCache,
	})

	lastReservedIPInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    store.addLastReservedIPToCache,
		UpdateFunc: store.updateLastReservedIPInCache,
		DeleteFunc: store.deleteLastReservedIPFromCache,
	})

	usingIPInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    store.addUsingIPToCache,
		UpdateFunc: store.updateUsingIPInCache,
		DeleteFunc: store.deleteUsingIPFromCache,
	})

	return store, nil
}

func (*Store) CreateNetwork(name string) error {
	panic("implement me")
}

func (*Store) DeleteNetwork(name string) error {
	panic("implement me")
}

func (*Store) GetNetwork(name string) (*types.Network, error) {
	panic("implement me")
}

func (*Store) GetLastReservedIP(name string) (*types.LastReservedIP, error) {
	panic("implement me")
}

func (*Store) AddPool(network string, pool *types.Pool) error {
	panic("implement me")
}

func (*Store) DelPool(network, pool string) error {
	panic("implement me")
}

func (*Store) CountPool(network, pool string) (total, used int, err error) {
	panic("implement me")
}

func (*Store) Reserve(network, pool, namespace, name string, ip net.IP) (bool, error) {
	panic("implement me")
}

func (*Store) Release(network, pool string, ip net.IP) error {
	panic("implement me")
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
