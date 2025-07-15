/*
Copyright (c) 2025 OceanBase
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

const (
	LABEL_MANAGED_BY        = "job.oceanbase.com/managed-by"
	LABEL_JOB_TYPE          = "job.oceanbase.com/job-type"
	LABEL_REF_NAMESPACE     = "job.oceanbase.com/ref-namespace"
	LABEL_REF_NAME          = "job.oceanbase.com/ref-name"
	LABEL_REF_OBCLUSTERNAME = "job.oceanbase.com/ref-obcluster-name"
	LABEL_ATTACHMENT_ID     = "job.oceanbase.com/attachment-id"
)

const (
	JOB_TYPE_DIAGNOSE   = "diagnose"
	JOB_TYPE_LOG        = "log"
	JOB_TYPE_INSPECTION = "inspection"
	JOB_TYPE_OBCONFIG   = "obconfig"
	INSPECTION_SCENARIO = "scenario"
)
