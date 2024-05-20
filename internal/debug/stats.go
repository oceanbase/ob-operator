/*
Copyright (c) 2024 OceanBase
ob-operator is licensed under Mulan PSL v2.
You can use this software according to the terms and conditions of the Mulan PSL v2.
You may obtain a copy of Mulan PSL v2 at:
         http://license.coscl.org.cn/MulanPSL2
THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
See the Mulan PSL v2 for more details.
*/

package debug

import (
	"runtime"
	"time"

	"sigs.k8s.io/controller-runtime/pkg/log"
)

func PollingRuntimeStats(cancelCh <-chan struct{}) {
	tk := time.NewTicker(10 * time.Second)
	defer tk.Stop()
	logger := log.Log.WithName("Debug - RuntimeStats")
	for {
		select {
		case <-tk.C:
			// do something
			var ms runtime.MemStats
			runtime.ReadMemStats(&ms)
			logger.Info("PollingRuntimeStats",
				"Alloc (MiB)", ms.Alloc>>20,
				"Sys (MiB)", ms.Sys>>20,
				"NumGC", ms.NumGC,
				"Go Routine", runtime.NumGoroutine(),
			)
		case <-cancelCh:
			return
		}
	}
}
