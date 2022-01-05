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

package provider

import (
	"os/exec"
)

func DirInit() {
	cmd := exec.Command("rm", "-rf", "/home/admin/oceanbase/log")
	cmd.Run()
	cmd = exec.Command("mkdir", "-p", "/home/admin/log")
	cmd.Run()
	cmd = exec.Command("ln", "-sf", "/home/admin/log", "/home/admin/oceanbase/log")
	cmd.Run()
	cmd = exec.Command("rm", "-rf", "/home/admin/oceanbase/store")
	cmd.Run()
	cmd = exec.Command("mkdir", "-p", "/home/admin/oceanbase/store")
	cmd.Run()
	cmd = exec.Command("mkdir", "-p", "/home/admin/data_log/clog")
	cmd.Run()
	cmd = exec.Command("ln", "-sf", "/home/admin/data_log/clog", "/home/admin/oceanbase/store/clog")
	cmd.Run()
	cmd = exec.Command("mkdir", "-p", "/home/admin/data_log/ilog")
	cmd.Run()
	cmd = exec.Command("ln", "-sf", "/home/admin/data_log/ilog", "/home/admin/oceanbase/store/ilog")
	cmd.Run()
	cmd = exec.Command("mkdir", "-p", "/home/admin/data_log/slog")
	cmd.Run()
	cmd = exec.Command("ln", "-sf", "/home/admin/data_log/slog", "/home/admin/oceanbase/store/slog")
	cmd.Run()
	cmd = exec.Command("mkdir", "-p", "/home/admin/data_file/sort_dir")
	cmd.Run()
	cmd = exec.Command("ln", "-sf", "/home/admin/data_file/sort_dir", "/home/admin/oceanbase/store/sort_dir")
	cmd.Run()
	cmd = exec.Command("mkdir", "-p", "/home/admin/data_file/sstable")
	cmd.Run()
	cmd = exec.Command("ln", "-sf", "/home/admin/data_file/sstable", "/home/admin/oceanbase/store/sstable")
	cmd.Run()
	cmd = exec.Command("chown", "-R", "admin:admin", "/home/admin/")
	cmd.Run()
}
