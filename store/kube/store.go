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
	"net"

	"github.com/mars1024/kube-ipam/store"
	"github.com/mars1024/kube-ipam/types"
)

// check if Store overrides all interfaces of IPAMStore
var _ store.IPAMStore = &Store{}

type Store struct {
}

func (*Store) Lock() {
	panic("implement me")
}

func (*Store) Unlock() {
	panic("implement me")
}

func (*Store) Close() error {
	panic("implement me")
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
