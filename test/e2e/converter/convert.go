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

package converter

import (
	"encoding/json"
	"log"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
)

func ConvertObjToUnstructured(k8sObj interface{}, u *unstructured.Unstructured) {
	tmp, err := json.Marshal(k8sObj)
	if err != nil {
		log.Println(err)
	}
	_ = u.UnmarshalJSON(tmp)
}

func ConvertUnstructuredToObj(u *unstructured.Unstructured, k8sObj interface{}) {
	_ = runtime.DefaultUnstructuredConverter.FromUnstructured(u.UnstructuredContent(), k8sObj)
}
