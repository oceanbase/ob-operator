/*
Copyright (c) 2026 OceanBase
ob-operator is licensed under Mulan PSL v2.
You can use this software according to the terms and conditions of the Mulan PSL v2.
You may obtain a copy of Mulan PSL v2 at:
         http://license.coscl.org.cn/MulanPSL2
THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
See the Mulan PSL v2 for more details.
*/

package types

import "testing"

func TestValidateSharedStorageBucketURL(t *testing.T) {
	for _, value := range []string{
		"s3://bucket?host=http://minio:9000",
		"s3://bucket/prefix?host=https://s3.example.com&s3_region=test",
		"oss://bucket/prefix?host=oss-cn-hangzhou.aliyuncs.com",
		"cos://bucket/prefix?host=cos.example.com",
	} {
		if err := ValidateSharedStorageBucketURL(value); err != nil {
			t.Errorf("valid storage URL rejected: %v", err)
		}
	}
	for _, value := range []string{"", "invalid-storage-url", "http://bucket", "s3:///bucket?host=http://minio", "s3://bucket", "s3://bucket?host=", "s3://bucket?host=http://", "s3://bucket?host=http://one&host=http://two", "s3://user:password@bucket?host=http://minio", "s3://bucket?host=http://minio#fragment"} {
		if ValidateSharedStorageBucketURL(value) == nil {
			t.Errorf("invalid storage URL accepted: %q", value)
		}
	}
}
