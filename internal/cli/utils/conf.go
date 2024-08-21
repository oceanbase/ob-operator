package utils

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v2"
)

type ComponentVersions struct {
	Components map[string]string `yaml:"components"`
}

// filePath for test
var filePath = "internal/cli/LATEST_VERSION.yaml"

func GetComponentsConf() map[string]string {
	var Components ComponentVersions
	data, err := os.ReadFile(filePath)
	if err != nil {
		panic(fmt.Errorf("Error reading LATEST_VERSION file: %v", err))
	}

	err = yaml.Unmarshal(data, &Components)
	// panic if file not exists
	if err != nil {
		panic(fmt.Errorf("Error decoding LATEST_VERSION file: %v", err))
	}
	return Components.Components
}
