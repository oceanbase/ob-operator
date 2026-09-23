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

package common

import (
	"net/http"
	"os"
	"testing"

	"gopkg.in/yaml.v2"

	httperr "github.com/oceanbase/ob-operator/pkg/errors"
)

func TestValidateMetricColumns(t *testing.T) {
	for key := range ColumnAggregations {
		if err := ValidateMetricColumns([]string{key}); err != nil || BuildMetricExpression(key) == "" {
			t.Errorf("registered metric %q rejected: %v", key, err)
		}
	}
	for _, cols := range [][]string{nil, {}, {"elapsed_time", "elapsed_time"}} {
		if err := ValidateMetricColumns(cols); err != nil {
			t.Errorf("compatible selection rejected: %v", err)
		}
	}
	for _, col := range []string{"", "unknown", "elapsed_time;", "elapsed_time/**/", "(SELECT 1)", " elapsed_time", "elapsed_time\n"} {
		err := ValidateMetricColumns([]string{"elapsed_time", col})
		e, ok := err.(httperr.ObError)
		if !ok || e.Status() != http.StatusBadRequest {
			t.Errorf("%q: expected BadRequest, got %v", col, err)
		}
	}
}

func TestValidateSortColumn(t *testing.T) {
	columns := []string{""}
	for key := range ColumnAggregations {
		columns = append(columns, key)
	}
	for key := range FixedDimensions {
		columns = append(columns, key)
	}
	for key := range Dimensions {
		columns = append(columns, key)
	}
	for _, col := range columns {
		if err := ValidateSortColumn(col); err != nil {
			t.Errorf("%q: %v", col, err)
		}
	}
	for _, col := range []string{"user()", "1", "sql_id DESC", "elapsed_time;", "elapsed_time/**/", "(SELECT 1)", " elapsed_time", "elapsed_time\n", "unknown"} {
		err := ValidateSortColumn(col)
		e, ok := err.(httperr.ObError)
		if !ok || e.Status() != http.StatusBadRequest {
			t.Errorf("%q: expected BadRequest, got %v", col, err)
		}
	}
}

func TestDashboardLatencyCatalogRemainsSupported(t *testing.T) {
	for _, locale := range []string{"en_US", "zh_CN"} {
		data, err := os.ReadFile("../../assets/dashboard/sql_metric_" + locale + ".yaml")
		if err != nil {
			t.Fatal(err)
		}
		var categories []struct {
			Category string `yaml:"category"`
			Metrics  []struct {
				Key string `yaml:"key"`
			} `yaml:"metrics"`
		}
		if err := yaml.Unmarshal(data, &categories); err != nil {
			t.Fatal(err)
		}
		count := 0
		for _, category := range categories {
			for _, metric := range category.Metrics {
				if err := ValidateSortColumn(metric.Key); err != nil {
					t.Errorf("unsupported sort key %s/%s: %v", locale, metric.Key, err)
				}
			}
			if category.Category != "latency" {
				continue
			}
			for _, metric := range category.Metrics {
				count++
				if err := ValidateMetricColumns([]string{metric.Key}); err != nil {
					t.Errorf("%s/%s: %v", locale, metric.Key, err)
				}
			}
		}
		if count == 0 {
			t.Fatalf("no latency catalog found for %s", locale)
		}
	}
}
