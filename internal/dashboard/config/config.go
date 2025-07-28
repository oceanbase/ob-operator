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

type ImageConfig struct {
	Repository string `yaml:"repository"`
	Tag        string `yaml:"tag"`
}

type JobConfig struct {
	TTLSecondsAfterFinished int32 `yaml:"ttlSecondsAfterFinished"`
}

type Config struct {
	OBDiag      ImageConfig `yaml:"obdiag"`
	OBHelper    ImageConfig `yaml:"ob-helper"`
	Inspection  JobConfig   `yaml:"inspection"`
	Diagnose    JobConfig   `yaml:"diagnose"`
	DownloadLog JobConfig   `yaml:"download-log"`
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
	if err = yaml.Unmarshal(data, &dashboardConfig); err != nil {
		return errors.Wrap(err, "unmarshal config file failed")
	}
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
