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

package helper

type OBDiagConfig struct {
	OBCluster OBClusterConfig `yaml:"obcluster"`
}

type OBClusterConfig struct {
	DBHost        string          `yaml:"db_host"`
	DBPort        int             `yaml:"db_port"`
	OBClusterName string          `yaml:"ob_cluster_name"`
	TenantSys     TenantSysConfig `yaml:"tenant_sys"`
	Servers       ServerConfig    `yaml:"servers"`
}

type TenantSysConfig struct {
	User     string `yaml:"user"`
	Password string `yaml:"password"`
}

type ServerConfig struct {
	Nodes  []NodeConfig `yaml:"nodes"`
	Global GlobalConfig `yaml:"global"`
}

type NodeConfig struct {
	PodName string `yaml:"pod_name"`
	IP      string `yaml:"ip"`
}

type GlobalConfig struct {
	Namespace     string `yaml:"namespace"`
	SshType       string `yaml:"ssh_type"`
	ContainerName string `yaml:"container_name"`
	HomePath      string `yaml:"home_path"`
	DataDir       string `yaml:"data_dir"`
	RedoDir       string `yaml:"redo_dir"`
}
