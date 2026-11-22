/*
Copyright (c) 2023 OceanBase
ob-operator is licensed under Mulan PSL v2.
You can use this software according to the terms and conditions of the Mulan PSL v2.
You may obtain a copy of Mulan PSL v2 at:
         http://license.coscl.org.cn/MulanPSL2
THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
See the Mulan PSL v2 for more details.
*/

package observer

import (
	"testing"

	apitypes "github.com/oceanbase/ob-operator/api/types"
	"github.com/oceanbase/ob-operator/api/v1alpha1"
	oceanbaseconst "github.com/oceanbase/ob-operator/internal/const/oceanbase"
)

func TestObserverContainerTimeZone(t *testing.T) {
	manager := &OBServerManager{OBServer: &v1alpha1.OBServer{}}
	manager.OBServer.Spec.OBServerTemplate = &apitypes.OBServerTemplate{
		Resource: &apitypes.ResourceSpec{},
		Storage: &apitypes.OceanbaseStorageSpec{
			DataStorage:    &apitypes.StorageSpec{},
			RedoLogStorage: &apitypes.StorageSpec{},
		},
	}
	for _, tc := range []struct{ name, mode, want string }{
		{"classic inherits the image time zone", "", ""},
		{"shared storage uses the oceanseek time zone", oceanbaseconst.DeploymentModeSharedStorage, "Asia/Shanghai"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			cluster := &v1alpha1.OBCluster{}
			cluster.Spec.DeploymentMode = tc.mode
			container := manager.createOBServerContainer(cluster)
			for _, env := range container.Env {
				if env.Name == "TZ" {
					if tc.want == "" || env.Value != tc.want {
						t.Fatalf("unexpected TZ override %q for mode %q", env.Value, tc.mode)
					}
					return
				}
			}
			if tc.want != "" {
				t.Fatalf("expected TZ=%s", tc.want)
			}
		})
	}
}
