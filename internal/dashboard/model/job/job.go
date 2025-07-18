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

package job

type JobStatus string

const (
	JobStatusSuccessful JobStatus = "successful"
	JobStatusFailed     JobStatus = "failed"
	JobStatusRunning    JobStatus = "running"
	JobStatusPending    JobStatus = "pending"
)

type JobResult struct {
	ExitCode     int32  `json:"exitCode" binding:"required"`
	Output       string `json:"output,omitempty"`
	AttachmentId string `json:"attachmentId,omitempty"`
}

type Job struct {
	Name       string     `json:"name" binding:"required"`
	Namespace  string     `json:"namespace" binding:"required"`
	Status     JobStatus  `json:"status" binding:"required"`
	StartTime  int64      `json:"startTime,omitempty"`
	FinishTime int64      `json:"finishTime,omitempty"`
	Result     *JobResult `json:"result,omitempty"`
}
