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

	"github.com/containernetworking/plugins/pkg/ip"
)

type Pool struct {
	Name      string     `json:"name"`
	PoolStart net.IP     `json:"poolStart"`
	PoolEnd   net.IP     `json:"poolEnd"`
	Gateway   net.IP     `json:"gateway"`
	Subnet    *net.IPNet `json:"subnet"`
	VlanID    *int32     `json:"vlanID"`
}

// Canonicalize takes a given pool and ensures that all information is consistent,
// filling out Start, End, and Gateway with sane values if missing
func (p *Pool) Canonicalize() error {
	if err := p.Validate(); err != nil {
		return err
	}

	if p.PoolStart == nil {
		p.PoolStart = ip.NextIP(p.Subnet.IP)
	}
	if p.PoolEnd == nil {
		p.PoolEnd = lastIP(p.Subnet)
	}

	return nil
}

// Validate can ensure that all necessary information are valid
func (p *Pool) Validate() error {
	// Basic validations
	switch {
	case len(p.Name) == 0:
		return fmt.Errorf("pool name %s can not be empty", p.Name)
	case p.VlanID != nil && (*p.VlanID <= 0 || (*p.VlanID > 1005 && *p.VlanID < 1025) || *p.VlanID > 4094):
		return fmt.Errorf("pool vlanID %d is invalid", *p.VlanID)
	case p.Gateway == nil:
		return fmt.Errorf("pool gateway is invalid")
	case p.Subnet == nil:
		return fmt.Errorf("pool subnet is invalid")
	}

	// Enhanced validations
	if err := canonicalizeIP(&p.Subnet.IP); err != nil {
		return err
	}

	// Can't create an allocator for a network with no addresses
	ones, masklen := p.Subnet.Mask.Size()
	if ones > masklen-2 {
		return fmt.Errorf("pool subnet %s too small to allocate from", p.Subnet.String())
	}

	if len(p.Subnet.IP) != len(p.Subnet.Mask) {
		return fmt.Errorf("pool subnet %s IP and Mask version mismatch", p.Subnet.String())
	}

	// Ensure Subnet IP is the network address, not some other address
	networkIP := p.Subnet.IP.Mask(p.Subnet.Mask)
	if !p.Subnet.IP.Equal(networkIP) {
		return fmt.Errorf("pool subnet has host bits set because a subnet mask of length %d the network address is %s", ones, networkIP.String())
	}

	// Gateway must in subnet
	if !p.Subnet.Contains(p.Gateway) {
		return fmt.Errorf("gateway %s not in subnet %s", p.Gateway.String(), p.Subnet.String())
	}

	// PoolStart must in subnet
	if p.PoolStart != nil {
		if err := canonicalizeIP(&p.PoolStart); err != nil {
			return err
		}

		if !p.Contains(p.PoolStart) {
			return fmt.Errorf("poolStart %s not in subnet %s", p.PoolStart.String(), p.Subnet.String())
		}
	}

	// PoolStart must in subnet
	if p.PoolEnd != nil {
		if err := canonicalizeIP(&p.PoolEnd); err != nil {
			return err
		}

		if !p.Contains(p.PoolEnd) {
			return fmt.Errorf("poolEnd %s not in subnet %s", p.PoolEnd.String(), p.Subnet.String())
		}
	}

	return nil
}

// Contains check if a given ip is in a pool
func (p *Pool) Contains(addr net.IP) bool {
	if err := canonicalizeIP(&addr); err != nil {
		return false
	}

	// Not in network
	if !p.Subnet.Contains(addr) {
		return false
	}

	if p.PoolStart != nil {
		// Before the range start
		if ip.Cmp(addr, p.PoolStart) < 0 {
			return false
		}
	}

	if p.PoolEnd != nil {
		if ip.Cmp(addr, p.PoolEnd) > 0 {
			return false
		}
	}

	return true
}

// canonicalizeIP makes sure a provided ip is in ipv4 standard form
func canonicalizeIP(ip *net.IP) error {
	if ip.To4() == nil {
		return fmt.Errorf("IP %s is not ipv4 standard form", *ip)
	}
	return nil
}

// Determine the last IP of a subnet, excluding the broadcast if IPv4
func lastIP(subnet *net.IPNet) net.IP {
	var end net.IP
	for i := 0; i < len(subnet.IP); i++ {
		end = append(end, subnet.IP[i]|^subnet.Mask[i])
	}

	if subnet.IP.To4() != nil {
		end[3]--
	}

	return end
}
