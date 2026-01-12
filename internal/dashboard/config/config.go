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

package config

import (
	"os"
	"sync"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
)

type ToolConfig struct {
	Image string `yaml:"image"`
}

type InspectionConfig struct {
	OBDiag   ToolConfig `yaml:"obdiag"`
	OBHelper ToolConfig `yaml:"oceanbase-helper"`
}

type JobTypeConfig struct {
	TTLSecondsAfterFinished int32 `yaml:"ttlSecondsAfterFinished"`
}

type JobConfig struct {
	Inspection JobTypeConfig `yaml:"inspection"`
	Normal     JobTypeConfig `yaml:"normal"`
}

type SQLAnalyzerConfig struct {
	Image                        string `yaml:"image"`
	RetentionDays                int    `yaml:"retentionDays"`
	StorageSize                  string `yaml:"storageSize"`
	CollectionIntervalSeconds    int    `yaml:"collectionIntervalSeconds"`
	CompactionIntervalSeconds    int    `yaml:"compactionIntervalSeconds"`
	CPURequest                   string `yaml:"cpuRequest"`
	CPULimit                     string `yaml:"cpuLimit"`
	MemoryRequest                string `yaml:"memoryRequest"`
	MemoryLimit                  string `yaml:"memoryLimit"`
	SqlAuditLimit                int    `yaml:"sqlAuditLimit"`
	SlowSqlThresholdMilliSeconds int    `yaml:"slowSqlThresholdMilliSeconds"`
}

type Config struct {
	Inspection  InspectionConfig  `yaml:"inspection"`
	Job         JobConfig         `yaml:"job"`
	SQLAnalyzer SQLAnalyzerConfig `yaml:"sqlAnalyzer"`
}

var (
	dashboardConfig *Config
	once            sync.Once
)

func loadConfig(configFile string) error {
	data, err := os.ReadFile(configFile)
	if err != nil {
		return errors.Wrap(err, "read config file failed")
	}
	var newConfig Config
	if err = yaml.Unmarshal(data, &newConfig); err != nil {
		return errors.Wrap(err, "unmarshal config file failed")
	}
	dashboardConfig = &newConfig
	return nil
}

func GetConfig() *Config {
	once.Do(func() {
		if err := loadConfig("/etc/dashboard/config.yaml"); err != nil {
			panic(err)
		}
	})
	return dashboardConfig
}
