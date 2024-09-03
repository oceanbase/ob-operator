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

type IndexCategory string
type IndexStatus string

const (
	IndexCategoryPrimaryKey   IndexCategory = "primaryKey"
	IndexCategoryGlobalNormal IndexCategory = "globalNormal"
	IndexCategoryGlobalUnique IndexCategory = "globalUnique"
	IndexCategoryLocalNormal  IndexCategory = "localNormal"
	IndexCategoryLocalUnique  IndexCategory = "localUnique"
)

const (
	IndexStatusCreating  IndexStatus = "creating"
	IndexStatusAvailable IndexStatus = "available"
	IndexStatusError     IndexStatus = "error"
)

type IndexInfo struct {
	TableName string        `json:"tableName" binding:"required"`
	Category  IndexCategory `json:"category" binding:"required"`
	IndexName string        `json:"indexName" binding:"required"`
	Columns   []string      `json:"columns" binding:"required"`
	Status    IndexStatus   `json:"status" binding:"required"`
}
