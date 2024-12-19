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
package backup

import "k8s.io/kubectl/pkg/util/templates"

var createExample = templates.Examples(`
	# Create a backup policy by OSS
	okctl backup create <tenant_name> --archive-path=oss://<bucket_name>/<path> --bak-data-path=oss://<bucket_name>/<path>  --oss-access-id=<access_id> --oss-access-key=<access_key> --inc="0 0 * * 1,2,3," --full="0 0 * * 4,5"
	
	# Create a backup policy by NFS
	okctl backup create <tenant_name> --archive-path=<path> --bak-data-path=<path> --bak-encryption-password=<password> --inc="0 0 * * 1,2,3," --full="0 0 * * 4,5"
`)

// TODO: add more examples
