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

package types

import (
	"fmt"
	v1 "github.com/mars1024/kube-ipam/pkg/apis/resource/v1"
	"net"
)

type Network struct {
	Name  string  `json:"name"`
	Pools []*Pool `json:"pools"`
}

type LastReservedIP struct {
	IP       net.IP `json:"ip"`
	PoolName string `json:"pool"`
}

func (l *LastReservedIP) Index(n *Network) (int, error) {
	poolIndex := -1
	for idx, pool := range n.Pools {
		if pool.Name == l.PoolName {
			poolIndex = idx
			break
		}
	}

	switch {
	case poolIndex < 0:
		return -1, fmt.Errorf("last reserved ip's pool is not in network")
	case !n.Pools[poolIndex].Contains(l.IP):
		return -1, fmt.Errorf("last reserved ip is not in pool %s", l.PoolName)
	}
	return poolIndex, nil
}

// GetNetworkFromCRD can help get typed network from network CRD
func GetNetworkFromCRD(n *v1.Network) (*Network, error) {
	network := &Network{
		Name:  n.Name,
		Pools: make([]*Pool, 0),
	}

	for _, pool := range n.Spec.Pools {
		pl, err := GetPoolFromCRD(&pool)
		if err != nil {
			return nil, err
		}
		network.Pools = append(network.Pools, pl)
	}

	return network, nil
}

// GetLastReservedIPFromCRD can help get typed lastReservedIP from lastReservedIP CRD
func GetLastReservedIPFromCRD(ip *v1.LastReservedIP) *LastReservedIP {
	return &LastReservedIP{
		IP:       net.ParseIP(ip.Spec.IP),
		PoolName: ip.Spec.PoolName,
	}
}
