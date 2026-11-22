/*
Copyright (c) 2026 OceanBase
ob-operator is licensed under Mulan PSL v2.
You can use this software according to the terms and conditions of the Mulan PSL v2.
You may obtain a copy of Mulan PSL v2 at:
         http://license.coscl.org.cn/MulanPSL2
THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
See the Mulan PSL v2 for more details.
*/

package oceanseek

import (
	"context"
	"reflect"
	"testing"
)

func TestThreeNodesInheritPoolWithoutMutatingInput(t *testing.T) {
	fixture := newThreeNodeRunFixture(t)
	values := mergeValues(fixture.defaults, nil)
	setValue(values, "oceanseek", "scheduling", "nodeSelector", "example.com/pool")
	before := mergeValues(values, nil)
	got, err := configureThreeNodes(context.Background(), execute, values, fixture.defaults, nil, []string{"worker-1", "worker-2", "worker-3"}, false)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(values, before) {
		t.Fatal("topology generation mutated its input (retained values must be immutable)")
	}
	if pfGet(got, "dashboard", "nodeSelector", "example.com/pool") != "oceanseek" {
		t.Fatal("Dashboard lost the inherited pool restriction")
	}
	if pfGet(got, "dashboard", "nodeSelector", "kubernetes.io/hostname") != "host-1" {
		t.Fatal("Dashboard did not use the first node's actual hostname")
	}
	setValue(values, "ui", "dashboard", "nodeSelector", "example.com/pool")
	got, err = configureThreeNodes(context.Background(), execute, values, fixture.defaults, nil, []string{"worker-1", "worker-2", "worker-3"}, false)
	if err != nil || pfGet(got, "dashboard", "nodeSelector", "example.com/pool") != "ui" {
		t.Fatalf("explicit Dashboard pool was not retained for preflight validation: %v", err)
	}
}
