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
package demo

import (
	"errors"
	"fmt"

	"github.com/manifoldco/promptui"

	"github.com/oceanbase/ob-operator/internal/cli/cluster"
	"github.com/oceanbase/ob-operator/internal/cli/tenant"
	"github.com/oceanbase/ob-operator/internal/cli/utils"
)

var (
	promptTepl = &promptui.PromptTemplates{
		Prompt:  "{{ . }} ",
		Valid:   "{{ . | green }} ",
		Invalid: "{{ . | red }} ",
		Success: "{{ . | bold }} ",
	}
	selectTepl = &promptui.SelectTemplates{
		Label:    "{{ . }} ",
		Active:   "\U0001F336 {{ . | cyan }}",
		Inactive: "  {{ . | cyan }}",
		Selected: "\U0001F336 {{ . | green | cyan }}",
	}
)

type PromptFactory struct {
	promptTepl *promptui.PromptTemplates
	selectTepl *promptui.SelectTemplates
}

// NewPromptFactory creates a new prompt factory
func NewPromptFactory() *PromptFactory {
	return &PromptFactory{
		promptTepl: promptTepl,
		selectTepl: selectTepl,
	}
}

// RunPromptE runs the prompt and returns the result or error
func (pf *PromptFactory) RunPromptE(p any) (result string, err error) {
	switch v := p.(type) {
	case *promptui.Prompt:
		if result, err = v.Run(); err != nil {
			if err == promptui.ErrInterrupt {
				return "", errors.New("interrupted by user")
			}
			return "", fmt.Errorf("failed to create cluster: %v", err)
		}
		return result, nil
	case *promptui.Select:
		if _, result, err = v.Run(); err != nil {
			if err == promptui.ErrInterrupt {
				return "", errors.New("interrupted by user")
			}
			return "", fmt.Errorf("failed to create cluster: %v", err)
		}
		return result, nil
	default:
		return "", errors.New("invalid prompt type")
	}
}

// CreatePrompt creates a prompt by prompt factory, based on the prompt type
func (pf *PromptFactory) CreatePrompt(promptType string) any {
	switch promptType {
	case cluster.FLAG_NAME:
		return &promptui.Prompt{
			Label:     "Please input the cluster name, press `enter` to use the default name `test`: ",
			Templates: pf.promptTepl,
			Validate: func(input string) error {
				if !utils.CheckResourceName(input) {
					return errors.New("invalid resource name in k8s")
				}
				return nil
			},
			AllowEdit: true,
			Default:   cluster.DEFAULT_NAME,
		}
	case cluster.FLAG_NAMESPACE:
		return &promptui.Prompt{
			Label:     "Please input the namespace, press `enter` to use default namespace: ",
			Templates: pf.promptTepl,
			Validate: func(input string) error {
				if input == "" {
					return errors.New("namespace can not be empty")
				}
				return nil
			},
			AllowEdit: true,
			Default:   cluster.DEFAULT_NAMESPACE,
		}
	case cluster.FLAG_ROOT_PASSWORD:
		return &promptui.Prompt{
			Label:     "Please input the root password, press `enter` to generate a random password: ",
			Templates: pf.promptTepl,
			Mask:      '*', // mask the input
			Validate: func(input string) error {
				if input == "" {
					return nil
				}
				if utils.CheckPassword(input) {
					return errors.New("invalid password")
				}
				return nil
			},
			AllowEdit: true,
		}
	case cluster.CLUSTER_TYPE:
		return &promptui.Select{
			Label:        "Please select the cluster type: ",
			Items:        []string{cluster.SINGLE_NODE, cluster.THREE_NODE},
			Templates:    pf.selectTepl,
			HideSelected: true,
		}
	case tenant.FLAG_TENANT_NAME_IN_K8S:
		return &promptui.Prompt{
			Label:     "Please input the tenant resource name, press `enter` to use the default name `t1`: ",
			Templates: pf.promptTepl,
			Validate: func(input string) error {
				if !utils.CheckResourceName(input) {
					return errors.New("invalid resource name in k8s")
				}
				return nil
			},
			AllowEdit: true,
			Default:   tenant.DEFAULT_TENANT_NAME_IN_K8S,
		}
	case tenant.FLAG_TENANT_NAME:
		return &promptui.Prompt{
			Label:     "Please input the tenant name, press `enter` to use the default name `t1`: ",
			Templates: pf.promptTepl,
			Validate: func(input string) error {
				if !utils.CheckTenantName(input) {
					return errors.New("invalid tenant name")
				}
				return nil
			},
			AllowEdit: true,
			Default:   tenant.DEFAULT_TENANT_NAME,
		}
	default:
		return nil
	}
}
