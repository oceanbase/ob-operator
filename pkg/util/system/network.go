/*
Copyright (c) 2021 OceanBase
ob-operator is licensed under Mulan PSL v2.
You can use this software according to the terms and conditions of the Mulan PSL v2.
You may obtain a copy of Mulan PSL v2 at:
         http://license.coscl.org.cn/MulanPSL2
THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
See the Mulan PSL v2 for more details.
*/

package system

import (
	"net"
)

func GetNICInfo(name string) (map[string]string, error) {
	res := make(map[string]string)
	netInterfaces, err := net.Interfaces()
	if err != nil {
		return res, err
	}
	for _, netInterface := range netInterfaces {
		if netInterface.Name == name {
			addrs, _ := netInterface.Addrs()
			res["nic"] = name
			res["ip"] = addrs[0].String()
		}
	}
	return res, nil
}
