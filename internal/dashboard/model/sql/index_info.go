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

package sql

type IndexInfo struct {
	TableName  string   `json:"tableName" binding:"required"`
	IndexType  string   `json:"indexType" binding:"required"`
	Uniqueness string   `json:"uniqueness" binding:"required"`
	IndexName  string   `json:"indexName" binding:"required"`
	Columns    []string `json:"columns" binding:"required"`
	Status     string   `json:"status" binding:"required"`
}
