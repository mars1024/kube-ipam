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

package utils

import (
	"regexp"
	"strings"
)

const (
	DNSLabelRFC1123 = `^[a-zA-Z0-9][-a-zA-Z0-9]{0,62}$`
)

func ToKubeName(IP string) string {
	return strings.Replace(IP, ".", "-", -1)
}

func ToIP(kubeName string) string {
	return strings.Replace(kubeName, "-", ".", -1)
}

func IsKubeName(str string) bool {
	matched, err := regexp.MatchString(DNSLabelRFC1123, str)
	if err != nil {
		return false
	}
	return matched
}
