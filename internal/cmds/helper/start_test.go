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
	"testing"
)

func TestObserverInheritsHardNofileLimit(t *testing.T) {
	const childEnv = "OB_HELPER_TEST_NOFILE"
	if hard := os.Getenv(childEnv); hard != "" {
		err := runObserver(`soft=$(ulimit -Sn); hard=$(ulimit -Hn)
if [ "$soft" != "$OB_HELPER_TEST_NOFILE" ] || [ "$hard" != "$OB_HELPER_TEST_NOFILE" ]; then
	printf 'Observer child got soft=%s hard=%s, want both %s\n' "$soft" "$hard" "$OB_HELPER_TEST_NOFILE" >&2
	exit 1
fi`)
		if err != nil {
			t.Fatal(err)
		}
		return
	}

	for _, tc := range []struct{ name, soft, hard string }{
		{"low_soft_limit", "1024", "4096"},
		{"already_at_hard_limit", "1024", "1024"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			// Set the limits before starting Go so the test exercises the real
			// child process inheritance, including Go's exec limit handling.
			cmd := exec.Command("bash", "-c",
				`ulimit -Sn "$1" && ulimit -Hn "$2" && exec "$3" -test.run '^TestObserverInheritsHardNofileLimit$'`,
				"nofile-test", tc.soft, tc.hard, os.Args[0])
			cmd.Env = append(os.Environ(), childEnv+"="+tc.hard)
			if output, err := cmd.CombinedOutput(); err != nil {
				t.Fatalf("Observer startup: %v\n%s", err, output)
			}
		})
	}
}
