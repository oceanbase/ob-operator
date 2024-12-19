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
package config

import (
	"errors"
	"io"
	"net/http"

	"gopkg.in/yaml.v2"

	"github.com/oceanbase/ob-operator/internal/cli/utils"
)

// Chart struct to parse the chart.yaml
type Chart struct {
	apiVersion  string `yaml:"apiVersion"`
	AppVersion  string `yaml:"appVersion"`
	description string `yaml:"description"`
	name        string `yaml:"name"`
	chartType   string `yaml:"type"`
	version     string `yaml:"version"`
}

const (
	stableVersion = "stable"
	devVersion    = "master"
)

// ComponentList is the list of components that can be installed
var ComponentList = []string{
	"cert-manager", "ob-operator", "ob-dashboard", "local-path-provisioner", "ob-operator-dev",
}

// ComponentUpdateList is the list of components that can be updated
var ComponentUpdateList = []string{
	"cert-manager", "ob-operator", "ob-dashboard", "local-path-provisioner",
}

var versionURLs = map[string]string{
	"ob-dashboard":           "https://raw.githubusercontent.com/oceanbase/ob-operator/refs/heads/stable/charts/oceanbase-dashboard/Chart.yaml",
	"local-path-provisioner": "https://raw.githubusercontent.com/rancher/local-path-provisioner/refs/tags/v0.0.30/deploy/chart/local-path-provisioner/Chart.yaml",
}

func getVersionFromChart(component string) (string, error) {
	url, ok := versionURLs[component]
	if !ok {
		return "", errors.New("url not found for the component")
	}

	// get the yaml
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", errors.New("failed to fetch chart: " + resp.Status)
	}

	// parse the yaml
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	var chart Chart
	err = yaml.Unmarshal(body, &chart)
	if err != nil {
		return "", err
	}

	return chart.AppVersion, nil
}

func getVersion(component string) (string, error) {
	switch component {
	case "ob-dashboard", "local-path-provisioner":
		return getVersionFromChart(component)
	case "cert-manager", "ob-operator":
		return stableVersion, nil
	case "ob-operator-dev":
		return devVersion, nil
	default:
		return "", errors.New("version not found for the component")
	}
}

// GetAllComponents returns all the components
func GetAllComponents() (map[string]string, error) {
	components := make(map[string]string)
	for _, component := range ComponentList {
		version, err := getVersion(component)
		if err != nil {
			return nil, err
		}
		components[component] = version
	}
	return components, nil
}

// GetDefaultComponents returns the default components to be installed if not specified
func GetDefaultComponents() (map[string]string, error) {
	var installList []string
	components, err := GetAllComponents()
	if err != nil {
		return nil, err
	}
	defaultComponents := make(map[string]string)
	if !utils.CheckIfComponentExists("cert-manager") {
		installList = []string{"cert-manager", "ob-operator", "ob-dashboard"}
	} else {
		installList = []string{"ob-operator", "ob-dashboard"}
	}
	for _, component := range installList {
		defaultComponents[component] = components[component]
	}
	return defaultComponents, nil
}
