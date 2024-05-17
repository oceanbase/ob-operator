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

package coordinator

import "time"

type config struct {
	NormalRequeueDuration    time.Duration
	ExecutionRequeueDuration time.Duration
	PausedRequeueDuration    time.Duration

	TaskMaxRetryTimes         int
	TaskRetryBackoffThreshold int

	PauseAnnotation          string
	IgnoreDeletionAnnotation string
}

var cfg = &config{
	NormalRequeueDuration:    30 * time.Second,
	ExecutionRequeueDuration: 1 * time.Second,
	PausedRequeueDuration:    5 * time.Second,

	TaskMaxRetryTimes:         99,
	TaskRetryBackoffThreshold: 16,
}

func SetMaxRetryTimes(maxRetryTimes int) {
	cfg.TaskMaxRetryTimes = maxRetryTimes
}

func SetRetryBackoffThreshold(retryBackoffThreshold int) {
	cfg.TaskRetryBackoffThreshold = retryBackoffThreshold
}

func SetPausedAnnotation(pausedAnnotation string) {
	cfg.PauseAnnotation = pausedAnnotation
}

func SetIgnoreDeletionAnnotation(ignoreDeletionAnnotation string) {
	cfg.IgnoreDeletionAnnotation = ignoreDeletionAnnotation
}
