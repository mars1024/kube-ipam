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
	v1 "github.com/mars1024/kube-ipam/pkg/apis/resource/v1"
	"github.com/mars1024/kube-ipam/types"
	"github.com/sirupsen/logrus"
	"sync"
)

var LOGGER = logrus.WithFields(logrus.Fields{"component": "cache"})

type cache struct {
	*sync.RWMutex

	networks        map[string]*types.Network
	usingIPs        map[string]string
	lastReservedIPs map[string]*types.LastReservedIP
}

func NewCache() *cache {
	return &cache{
		RWMutex:         new(sync.RWMutex),
		networks:        make(map[string]*types.Network),
		usingIPs:        make(map[string]string),
		lastReservedIPs: make(map[string]*types.LastReservedIP),
	}
}

func (c *cache) addNetwork(network *v1.Network) {
	c.Lock()
	defer c.Unlock()

	net, err := types.GetNetworkFromCRD(network)
	if err != nil {
		LOGGER.Errorf("fail to add network %+v to cache : %s", network, err)
	}

	c.networks[network.Name] = net
	LOGGER.Debugf("add network %s %+v to cache", network.Name, network.Spec)
}

func (c *cache) updateNetwork(network *v1.Network) {
	c.Lock()
	defer c.Unlock()

	if network.DeletionTimestamp != nil {
		delete(c.networks, network.Name)
		return
	}

	net, err := types.GetNetworkFromCRD(network)
	if err != nil {
		LOGGER.Errorf("fail to update network %+v to cache : %s", network, err)
	}

	c.networks[network.Name] = net
	LOGGER.Debugf("update network %s %+v to cache", network.Name, network.Spec)
}

func (c *cache) deleteNetwork(network *v1.Network) {
	c.Lock()
	defer c.Unlock()

	delete(c.networks, network.Name)
	LOGGER.Debugf("delete network %s %+v from cache", network.Name, network.Spec)
}

func (c *cache) addUsingIP(usingIP *v1.UsingIP) {
	c.Lock()
	defer c.Unlock()

	c.usingIPs[usingIP.Name] = usingIP.Spec.PodName
	LOGGER.Debugf("add using ip %s %+v to cache", usingIP.Name, usingIP.Spec)
}

func (c *cache) updateUsingIP(usingIP *v1.UsingIP) {
	c.Lock()
	defer c.Unlock()

	if usingIP.DeletionTimestamp != nil {
		delete(c.usingIPs, usingIP.Name)
	}

	c.usingIPs[usingIP.Name] = usingIP.Spec.PodName
	LOGGER.Debugf("update using ip %s %+v to cache", usingIP.Name, usingIP.Spec)
}

func (c *cache) deleteUsingIP(usingIP *v1.UsingIP) {
	c.Lock()
	defer c.Unlock()

	delete(c.usingIPs, usingIP.Name)
	LOGGER.Debugf("delete using ip %s %+v from cache", usingIP.Name, usingIP.Spec)
}

func (c *cache) addLastReservedIP(lastReservedIP *v1.LastReservedIP) {
	c.Lock()
	defer c.Unlock()

	c.lastReservedIPs[lastReservedIP.Name] = types.GetLastReservedIPFromCRD(lastReservedIP)
	LOGGER.Debugf("add last reserved ip %s %+v to cache", lastReservedIP.Name, lastReservedIP.Spec)
}

func (c *cache) updateLastReservedIP(lastReservedIP *v1.LastReservedIP) {
	c.Lock()
	defer c.Unlock()

	if lastReservedIP.DeletionTimestamp != nil {
		delete(c.lastReservedIPs, lastReservedIP.Name)
		return
	}

	c.lastReservedIPs[lastReservedIP.Name] = types.GetLastReservedIPFromCRD(lastReservedIP)
	LOGGER.Debugf("update last reserved ip %s %+v to cache", lastReservedIP.Name, lastReservedIP.Spec)
}

func (c *cache) deleteLastReservedIP(lastReservedIP *v1.LastReservedIP) {
	c.Lock()
	defer c.Unlock()

	delete(c.lastReservedIPs, lastReservedIP.Name)
	LOGGER.Debugf("delete last reserved ip %s %+v from cache", lastReservedIP.Name, lastReservedIP.Spec)
}

func (c *cache) GetNetwork(networkName string) *types.Network {
	c.RLock()
	defer c.RUnlock()

	if network, exists := c.networks[networkName]; exists {
		// TODO: DeepCopy
		return network
	}
	return nil
}

func (c *cache) GetLastReservedIP(networkName string) *types.LastReservedIP {
	c.RLock()
	defer c.RUnlock()

	if lastReservedIP, exists := c.lastReservedIPs[networkName]; exists {
		// TODO: DeepCopy
		return lastReservedIP
	}
	return nil
}

func (c *cache) IsIPUsing(ip string) bool {
	c.RLock()
	defer c.RUnlock()

	if _, exists := c.usingIPs[ip]; exists {
		return true
	}
	return false
}
