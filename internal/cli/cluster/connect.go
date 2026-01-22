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
package cluster

import (
	"encoding/base64"
	"fmt"
	"os/exec"
	"strings"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	kubeerrors "k8s.io/apimachinery/pkg/api/errors"

	"github.com/oceanbase/ob-operator/internal/cli/generic"
	utils "github.com/oceanbase/ob-operator/internal/cli/utils"
	"github.com/oceanbase/ob-operator/internal/clients"
)

type ConnectOptions struct {
	generic.ResourceOption
	ClusterId  int64
	ObserverIp string
	TenantName string
	Database   string
	User       string
	Password   string
	Port       string
}

func NewConnectOptions() *ConnectOptions {
	return &ConnectOptions{}
}

func (o *ConnectOptions) Validate() error {
	if o.Namespace == "" {
		return errors.New("namespace is not specified")
	}
	if o.User == "" {
		return errors.New("user is not specified")
	}
	if o.TenantName == "" {
		return errors.New("tenant name is not specified")
	}
	if o.Port == "" {
		return errors.New("port is not specified")
	}
	return nil
}

func (o *ConnectOptions) Parse(cmd *cobra.Command, args []string) error {
	// parse args
	o.Name = args[0]
	o.Cmd = cmd
	// get obcluster
	obcluster, err := clients.GetOBCluster(o.Cmd.Context(), o.Namespace, o.Name)
	if err != nil {
		if kubeerrors.IsNotFound(err) {
			return fmt.Errorf("OBCluster %s not found", o.Name)
		} else {
			return err
		}
	}
	if err := utils.CheckClusterStatus(obcluster); err != nil {
		return err
	}
	o.ClusterId = obcluster.Spec.ClusterId
	return nil
}

func (o *ConnectOptions) Complete() error {
	// Try to get password from secret if not provided
	if err := o.GetPasswordFromSecret(); err != nil {
		return fmt.Errorf("failed to get password: %v", err)
	}
	return nil
}

func (o *ConnectOptions) Run() error {
	if err := o.GetObserverIp(); err != nil {
		return err
	}

	cmd := exec.Command("mysql",
		"-h"+o.ObserverIp,
		"-p"+o.Password,
		"-u"+o.User+"@"+o.TenantName,
		o.Database,
		"-A",
		"-c",
		"-P"+o.Port)

	return utils.RunCmd(cmd)
}

// getAvailableZones returns a list of available zones for the cluster
func (o *ConnectOptions) getAvailableZones() ([]string, error) {
	// Run kubectl command to get cluster status
	cmd := exec.Command("kubectl", "get", "obcluster", o.Name, "-n", o.Namespace, "-o", "yaml")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("failed to get cluster information: %v", err)
	}

	// Parse the output to find zones
	lines := strings.Split(string(output), "\n")
	zones := make([]string, 0)
	inObzones := false

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		if strings.HasPrefix(line, "obzones:") {
			inObzones = true
			continue
		}

		if inObzones && strings.HasPrefix(line, "- ") {
			if !strings.Contains(line, "status:") && !strings.Contains(line, "zone:") {
				inObzones = false
				continue
			}
		}

		// Parse zone information
		if inObzones && strings.Contains(line, "zone:") {
			fields := strings.Split(line, "zone:")
			if len(fields) == 2 {
				zone := strings.TrimSpace(fields[1])
				zones = append(zones, zone)
			}
		}
	}

	if len(zones) == 0 {
		return nil, fmt.Errorf("no available zones found for cluster %s", o.Name)
	}

	return zones, nil
}

func (o *ConnectOptions) GetObserverIp() error {
	// First get available zones
	zones, err := o.getAvailableZones()
	if err != nil {
		return err
	}

	// Try to get observer IP from each zone until we find one
	for _, zone := range zones {
		cmd := exec.Command("kubectl", "get", "pods", "-o", "wide", "-n", o.Namespace)
		output, err := cmd.CombinedOutput()
		if err != nil {
			continue
		}

		lines := strings.Split(string(output), "\n")
		if len(lines) < 2 {
			continue
		}
		for _, line := range lines[1:] { // Skip header line
			// Use alternative space to split fields
			fields := strings.FieldsFunc(line, func(r rune) bool {
				return r == ' ' || r == '\t'
			})

			if len(fields) < 7 {
				continue
			}

			podName := fields[0]
			var podIP string
			// check every field to find ip, because the ip maybe analyzed as other fields
			for i := 5; i < len(fields); i++ {
				if strings.Contains(fields[i], ".") {
					podIP = fields[i]
					break
				}
			}

			if podIP == "" {
				continue
			}

			// check if this is an observer pod for the current zone
			if strings.Contains(podName, zone) {
				parts := strings.Split(podIP, ".")
				if len(parts) != 4 {
					continue
				}

				o.ObserverIp = podIP
				return nil
			}
		}
	}

	return fmt.Errorf("no observer pod found in any available zone in namespace %s", o.Namespace)
}

// GetPasswordFromSecret gets password from secret, it can get sys tenant root password and other tenant's password
func (o *ConnectOptions) GetPasswordFromSecret() error {
	cmd := exec.Command("kubectl", "get", "secrets", "-n", o.Namespace, "-o", "jsonpath={.items[*].metadata.name}")
	output, err := cmd.CombinedOutput()
	if err == nil && len(output) > 0 {
		// find the secret and get password
		secrets := strings.Split(string(output), " ")
		// Use the exact pattern with cluster name and ID
		expectedSecretPrefix := fmt.Sprintf("%s-%d-root-", o.Name, o.ClusterId)

		for _, secretName := range secrets {
			if strings.HasPrefix(secretName, expectedSecretPrefix) {
				cmd = exec.Command("kubectl", "get", "secret", secretName, "-n", o.Namespace, "-o", "jsonpath={.data.password}")
				output, err = cmd.CombinedOutput()
				if err == nil && len(output) > 0 {
					// decode base64 password
					decodedBytes, err := base64.StdEncoding.DecodeString(string(output))
					if err == nil {
						o.Password = string(decodedBytes)
						break
					}
				}
			}
		}
	}

	if o.Password == "" {
		return fmt.Errorf("password is not specified and not found in secrets")
	}

	if o.TenantName != DEFAULT_TENANT_NAME && o.User != DEFAULT_USER {

	}

	return nil
}

func (o *ConnectOptions) AddFlags(cmd *cobra.Command) {
	cmd.Flags().StringVarP(&o.Namespace, FLAG_NAMESPACE, SHORTHAND_NAMESPACE, DEFAULT_NAMESPACE, "The namespace of the cluster")
	cmd.Flags().StringVarP(&o.Port, FLAG_PORT, SHORTHAND_PORT, DEFAULT_PORT, "The port for connecting to the cluster")
	cmd.Flags().StringVarP(&o.Database, FLAG_DATABASE, SHORTHAND_DATABASE, DEFAULT_DATABASE, "The database name of the tenant")
	cmd.Flags().StringVarP(&o.User, FLAG_USER, SHORTHAND_USER, DEFAULT_USER, "The user name of the tenant")
	cmd.Flags().StringVarP(&o.TenantName, FLAG_TENANT_NAME, SHORTHAND_TENANT_NAME, DEFAULT_TENANT_NAME, "The tenant name of the cluster")
}
