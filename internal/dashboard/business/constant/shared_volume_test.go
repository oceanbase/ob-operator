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

package constant

import "testing"

func TestSharedVolumeSecurityContext(t *testing.T) {
	for _, level := range []string{"", " ", "s0", "s0:c123,c456"} {
		t.Run(level, func(t *testing.T) {
			t.Setenv("SHARED_VOLUME_SELINUX_LEVEL", level)
			got := SharedVolumeSecurityContext()
			if level == "" || level == " " {
				if got != nil {
					t.Fatal("legacy deployment must retain admission defaults")
				}
				return
			}
			if got == nil || got.SELinuxOptions == nil || got.SELinuxOptions.Level != level {
				t.Fatalf("shared PVC level lost: %+v", got)
			}
			if got.SELinuxOptions.Type != "" || got.SELinuxOptions.User != "" || got.SELinuxOptions.Role != "" {
				t.Fatal("must retain runtime type, user and role")
			}
			got.SELinuxOptions.Level = "changed"
			if SharedVolumeSecurityContext().SELinuxOptions.Level != level {
				t.Fatal("mutable context shared between jobs")
			}
		})
	}
}
