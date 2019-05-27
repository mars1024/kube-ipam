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
package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +genclient:noStatus
// +genclient:nonNamespaced
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +k8s:openapi-gen=true

// Network is a specification for a network resource
type Network struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec NetworkSpec `json:"spec"`
}

// NetworkSpec is the spec for a network resource
type NetworkSpec struct {
	Pools []Pool `json:"pools"`
}

// Pool is a part of network spec which includes some network-related info
type Pool struct {
	Name      string `json:"name,omitempty"`
	PoolStart string `json:"poolStart,omitempty"`
	PoolEnd   string `json:"poolEnd,omitempty"`
	Gateway   string `json:"gateway,omitempty"`
	Subnet    string `json:"subnet,omitempty"`
	VlanId    int    `json:"vlanId,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// NetworkList is a list of network resources
type NetworkList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []Network `json:"items"`
}

// +genclient
// +genclient:noStatus
// +genclient:nonNamespaced
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +k8s:openapi-gen=true

// LastReservedIP is a specification for a last-reserved-ip resource of a network
type LastReservedIP struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec LastReservedIPSpec `json:"spec"`
}

// LastReservedIPSpec is the spec for a last-reserved-ip resource
type LastReservedIPSpec struct {
	IP       string `json:"ip,omitempty"`
	PoolName string `json:"poolName,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// LastReservedIPList is a list of last-reserved-ip resources
type LastReservedIPList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []LastReservedIP `json:"items"`
}

// +genclient
// +genclient:noStatus
// +genclient:nonNamespaced
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +k8s:openapi-gen=true

// UsingIP is a specification for an IP resource which has been used
type UsingIP struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec UsingIPSpec `json:"spec"`
}

// UsingIPSpec is the spec for an using IP resource
type UsingIPSpec struct {
	PodName      string `json:"podName,omitempty"`
	PodNamespace string `json:"podNamespace,omitempty"`
	Network      string `json:"network,omitempty"`
	Pool         string `json:"pool,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// UsingIPList is a list of using IP resources
type UsingIPList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []UsingIP `json:"items"`
}
