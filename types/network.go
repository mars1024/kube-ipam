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
	"net"
)

type Network struct {
	Name  string  `json:"name"`
	Pools []*Pool `json:"pools"`
}

type LastReservedIP struct {
	IP   net.IP `json:"ip"`
	Pool string `json:"pool"`
}

func (l *LastReservedIP) Index(n *Network) (int, error) {
	poolIndex := -1
	for idx, pool := range n.Pools {
		if pool.Name == l.Pool {
			poolIndex = idx
			break
		}
	}

	switch {
	case poolIndex < 0:
		return -1, fmt.Errorf("last reserved ip's pool is not in network")
	case !n.Pools[poolIndex].Contains(l.IP):
		return -1, fmt.Errorf("last reserved ip is not in pool %s", l.Pool)
	}
	return poolIndex, nil
}
