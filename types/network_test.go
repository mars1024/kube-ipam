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
	"net"
	"testing"
)

func TestLastReservedIP_Index(t *testing.T) {
	subnet := &net.IPNet{
		IP:   []byte{192, 168, 0, 0},
		Mask: []byte{255, 255, 255, 0},
	}

	tests := []struct {
		network *Network
		lrip    *LastReservedIP
		idx     int
	}{
		{
			network: &Network{
				Pools: []*Pool{
					{
						Name: "test",
					},
				},
			},
			lrip: &LastReservedIP{
				IP:       nil,
				PoolName: "test1",
			},
			idx: -1,
		},
		{
			network: &Network{
				Pools: []*Pool{
					{
						Name:   "test1",
						Subnet: subnet,
					},
				},
			},
			lrip: &LastReservedIP{
				IP:       net.IP([]byte{192, 168, 0, 100}),
				PoolName: "test1",
			},
			idx: 0,
		},
		{
			network: &Network{
				Pools: []*Pool{
					{
						Name:   "test1",
						Subnet: subnet,
					},
				},
			},
			lrip: &LastReservedIP{
				IP:       net.IP([]byte{192, 168, 1, 100}),
				PoolName: "test1",
			},
			idx: -1,
		},
	}

	for _, test := range tests {
		if idx, _ := test.lrip.Index(test.network); idx != test.idx {
			t.Errorf("fails %d %d", idx, test.idx)
			return
		}
	}
}
