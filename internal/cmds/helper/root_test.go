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
package helper

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestHelperStartsWithoutKubeconfig(t *testing.T) {
	const child = "OB_HELPER_TEST_WITHOUT_KUBECONFIG"
	if os.Getenv(child) == "1" {
		rootCmd.SetArgs([]string{"help"})
		if err := rootCmd.Execute(); err != nil {
			t.Fatal(err)
		}
		return
	}
	cmd := exec.Command(os.Args[0], "-test.run=^TestHelperStartsWithoutKubeconfig$")
	cmd.Env = []string{child + "=1", "HOME=" + os.Getenv("HOME"), "KUBECONFIG=" + filepath.Join(t.TempDir(), "missing")}
	output, err := cmd.CombinedOutput()
	if err != nil || !strings.Contains(string(output), "oceanbase-helper") {
		t.Fatalf("helper help failed without cluster configuration: %v\n%s", err, output)
	}
}
