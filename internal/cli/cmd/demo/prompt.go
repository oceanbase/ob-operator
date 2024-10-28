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
	"github.com/oceanbase/ob-operator/internal/cli/utils"
)

var tepl *promptui.PromptTemplates = &promptui.PromptTemplates{
	Prompt:  "{{ . }} ",
	Valid:   "{{ . | green }} ",
	Invalid: "{{ . | red }} ",
	Success: "{{ . | bold }} ",
}

type PromptFactory struct {
	template *promptui.PromptTemplates
}

func NewPromptFactory() *PromptFactory {
	return &PromptFactory{
		template: tepl,
	}
}

func RunPromptsForCluster(pf *PromptFactory, o *cluster.CreateOptions) (err error) {
	prompt := pf.CreatePrompt(cluster.FLAG_NAME)
	if o.Name, err = pf.RunPromptE(prompt); err != nil {
		return err
	}
	prompt = pf.CreatePrompt(cluster.FLAG_NAMESPACE)
	if o.Namespace, err = pf.RunPromptE(prompt); err != nil {
		return err
	}
	prompt = pf.CreatePrompt(cluster.FLAG_ROOT_PASSWORD)
	if o.RootPassword, err = pf.RunPromptE(prompt); err != nil {
		return err
	}
	prompt = pf.CreatePrompt(cluster.FLAG_BACKUP_ADDRESS)
	if o.BackupVolume.Address, err = pf.RunPromptE(prompt); err != nil {
		return err
	}
	prompt = pf.CreatePrompt(cluster.FLAG_BACKUP_PATH)
	if o.BackupVolume.Path, err = pf.RunPromptE(prompt); err != nil {
		return err
	}
	if err := o.Complete(); err != nil {
		return err
	}
	if err := o.SetDefaultConfig(cluster.SINGLE_NODE); err != nil {
		return err
	}
	return nil
}

func (pf *PromptFactory) RunPromptE(p *promptui.Prompt) (result string, err error) {
	if result, err = p.Run(); err != nil {
		if err == promptui.ErrInterrupt {
			return "", errors.New("interrupted by user")
		}
		return "", fmt.Errorf("failed to create cluster: %v", err)
	}
	return result, nil
}

func (pf *PromptFactory) CreatePrompt(promptType string) *promptui.Prompt {
	switch promptType {
	case cluster.FLAG_NAME:
		return &promptui.Prompt{
			Label:     "Please input the cluster name (Default `test`): ",
			Templates: pf.template,
			Validate: func(input string) error {
				if !utils.CheckResourceName(input) {
					return errors.New("invalid cluster name")
				}
				return nil
			},
			Default: cluster.DEFAULT_NAME,
		}
	case cluster.FLAG_NAMESPACE:
		return &promptui.Prompt{
			Label:     "Please input the namespace (Default `default`): ",
			Templates: pf.template,
			Validate: func(input string) error {
				if input == "" {
					return errors.New("namespace can not be empty")
				}
				return nil
			},
			Default: cluster.DEFAULT_NAMESPACE,
		}
	case cluster.FLAG_ROOT_PASSWORD:
		return &promptui.Prompt{
			Label:     "Please input the root password (if not used, generate random password): ",
			Templates: pf.template,
			Validate: func(input string) error {
				if input == "" {
					return nil
				}
				if utils.CheckPassword(input) {
					return errors.New("invalid password")
				}
				return nil
			},
			Default: utils.GenerateRandomPassword(8, 32),
		}
	case cluster.FLAG_BACKUP_ADDRESS:
		return &promptui.Prompt{
			Label:     "Please input the backup address (if not set, it will be empty): ",
			Templates: pf.template,
			Default:   "",
		}
	case cluster.FLAG_BACKUP_PATH:
		return &promptui.Prompt{
			Label:     "Please input the backup path (if not set, it will be empty): ",
			Templates: pf.template,
			Default:   "",
		}
	default:
		return nil
	}
}
