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

package store

import (
	"net"

	"github.com/mars1024/kube-ipam/types"
)

// IPAMStore is a store interface used by IPAM
type IPAMStore interface {
	// Network
	CreateNetwork(name string) error
	DeleteNetwork(name string) error
	GetNetwork(name string) (*types.Network, error)
	GetLastReservedIP(name string) (*types.LastReservedIP, error)

	// Pool
	AddPool(network string, pool *types.Pool) error
	DelPool(network, pool string) error
	CountPool(network, pool string) (total, used int, err error)

	// IP
	Reserve(network, pool, namespace, name string, ip net.IP) (bool, error)
	Release(ip net.IP) error
	ReleaseByName(network, pool, namespace, name string) error
}
