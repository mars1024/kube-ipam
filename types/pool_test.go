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

func TestPool_Validate(t *testing.T) {
	_, subnet1, _ := net.ParseCIDR("192.168.0.0/24")
	_, subnet2, _ := net.ParseCIDR("192.168.0.0/31")
	poolStart1 := net.ParseIP("192.168.0.10")
	poolStart2 := net.ParseIP("192.168.0.100")
	poolEnd1 := net.ParseIP("192.168.0.50")
	poolEnd2 := net.ParseIP("192.168.0.250")
	poolEnd3 := net.ParseIP("192.168.1.100")
	gateway1 := net.ParseIP("192.168.0.254")
	gateway2 := net.ParseIP("192.168.2.2")
	vlanID1 := int32(-100)
	vlanID2 := int32(1010)
	vlanID3 := int32(50)

	pools := []Pool{
		{
			Name:      "",
			PoolStart: poolStart1,
			PoolEnd:   poolEnd2,
			Gateway:   gateway1,
			Subnet:    subnet1,
			VlanID:    nil,
		},
		{
			Name:      "test1",
			PoolStart: nil,
			PoolEnd:   nil,
			Gateway:   nil,
			Subnet:    subnet1,
			VlanID:    nil,
		},
		{
			Name:      "test2",
			PoolStart: nil,
			PoolEnd:   nil,
			Gateway:   gateway1,
			Subnet:    nil,
			VlanID:    nil,
		},
		{
			Name:      "test3",
			PoolStart: nil,
			PoolEnd:   nil,
			Gateway:   gateway1,
			Subnet:    subnet1,
			VlanID:    &vlanID1,
		},
		{
			Name:      "test4",
			PoolStart: nil,
			PoolEnd:   nil,
			Gateway:   gateway1,
			Subnet:    subnet1,
			VlanID:    &vlanID2,
		},
		{
			Name:      "test5",
			PoolStart: nil,
			PoolEnd:   nil,
			Gateway:   gateway1,
			Subnet:    subnet2,
			VlanID:    &vlanID3,
		},
		{
			Name:      "test6",
			PoolStart: nil,
			PoolEnd:   nil,
			Gateway:   gateway2,
			Subnet:    subnet1,
			VlanID:    &vlanID3,
		},
		{
			Name:      "test7",
			PoolStart: poolStart2,
			PoolEnd:   poolEnd1,
			Gateway:   gateway1,
			Subnet:    subnet1,
			VlanID:    &vlanID3,
		},
		{
			Name:      "test8",
			PoolStart: poolStart1,
			PoolEnd:   poolEnd3,
			Gateway:   gateway1,
			Subnet:    subnet1,
			VlanID:    nil,
		},
	}

	for _, pool := range pools {
		if err := pool.Validate(); err == nil {
			t.Errorf("invalid pool %+v pass the validation", pool)
		}
	}
}

func Test_LastIP(t *testing.T) {
	_, subnet1, _ := net.ParseCIDR("192.168.0.0/24")
	_, subnet2, _ := net.ParseCIDR("172.16.0.0/22")
	_, subnet3, _ := net.ParseCIDR("192.168.0.0/26")
	ip1 := net.ParseIP("192.168.0.254")
	ip2 := net.ParseIP("172.16.3.254")
	ip3 := net.ParseIP("192.168.0.62")

	subnetIpMap := map[*net.IPNet]net.IP{
		subnet1: ip1,
		subnet2: ip2,
		subnet3: ip3,
	}

	for subnet, ip := range subnetIpMap {
		if !ip.Equal(lastIP(subnet)) {
			t.Errorf("subnet %s 's last IP is not %s", subnet.String(), ip.String())
		}
	}
}

func TestPool_Canonicalize(t *testing.T) {
	_, subnet, _ := net.ParseCIDR("192.168.0.0/24")
	vlanID := int32(10)

	pool := Pool{
		Name:    "test",
		Subnet:  subnet,
		Gateway: net.ParseIP("192.168.0.100"),
		VlanID:  &vlanID,
	}

	if err := pool.Canonicalize(); err != nil {
		t.Errorf("fail to canonicalize pool %+v : %s", pool, err)
	}

	//t.Logf("canonicalize pool to %+v", pool)
}

func TestPool_Contains(t *testing.T) {
	_, subnet, _ := net.ParseCIDR("192.168.0.0/24")
	gateway := net.ParseIP("192.168.0.254")

	var tests = []struct {
		pool    *Pool
		ip      net.IP
		contain bool
	}{
		{
			pool: &Pool{
				Gateway: gateway,
				Subnet:  subnet,
			},
			ip:      net.ParseIP("192.168.0.100"),
			contain: true,
		},
		{
			pool: &Pool{
				Gateway: gateway,
				Subnet:  subnet,
			},
			ip:      net.ParseIP("192.168.1.100"),
			contain: false,
		},
		{
			pool: &Pool{
				PoolStart: net.ParseIP("192.168.0.150"),
				Gateway:   gateway,
				Subnet:    subnet,
			},
			ip:      net.ParseIP("192.168.0.100"),
			contain: false,
		},
		{
			pool: &Pool{
				PoolEnd: net.ParseIP("192.168.0.50"),
				Gateway: gateway,
				Subnet:  subnet,
			},
			ip:      net.ParseIP("192.168.0.100"),
			contain: false,
		},
	}

	for _, test := range tests {
		if test.pool.Contains(test.ip) != test.contain {
			t.Errorf("fails")
			return
		}
	}
}