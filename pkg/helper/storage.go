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
	"errors"
	"os"
	"path"
	"syscall"
)

const (
	FALLOC_FL_KEEP_SIZE      = 0x01
	FALLOC_FL_PUNCH_HOLE     = 0x02
	FALLOC_FL_COLLAPSE_RANGE = 0x08
	FALLOC_FL_ZERO_RANGE     = 0x10
)

// TryFallocate tries to fallocate a file in the given directory.
// It uses syscall.Fallocate to allocate space for a file.
// Flags 0, 0x01|0x02, 0x10 are used to test different modes of fallocate.
func TryFallocate(dirPath string) error {
	err := os.MkdirAll(dirPath, 0755)
	if err != nil {
		return err
	}
	file1, err := os.OpenFile(path.Join(dirPath, "testfile1"), os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		return err
	}
	defer os.Remove(path.Join(dirPath, "testfile1"))
	defer file1.Close()
	err = errors.Join(
		syscall.Fallocate(int(file1.Fd()), 0, 0, 1<<20),
		syscall.Fallocate(int(file1.Fd()), FALLOC_FL_KEEP_SIZE|FALLOC_FL_PUNCH_HOLE, 1<<20, 1<<21),
		syscall.Fallocate(int(file1.Fd()), FALLOC_FL_ZERO_RANGE, 1<<21, 1<<22),
	)
	return err
}
