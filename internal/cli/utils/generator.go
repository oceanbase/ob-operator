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
package utils

import (
	"context"
	"crypto/rand"
	"fmt"
	"strings"
	"time"

	apitypes "github.com/oceanbase/ob-operator/api/types"
	"github.com/oceanbase/ob-operator/pkg/k8s/client"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8srand "k8s.io/apimachinery/pkg/util/rand"
)

// CreatePasswordSecret creates a secret with password
func CreatePasswordSecret(ctx context.Context, namespace, name, password string) error {
	k8sclient := client.GetClient()
	_, err := k8sclient.ClientSet.CoreV1().Secrets(namespace).Create(ctx, &corev1.Secret{
		ObjectMeta: v1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		StringData: map[string]string{
			"password": password,
		},
	}, v1.CreateOptions{})
	return err
}

// GenerateUUID returns uuid
func GenerateUUID() string {
	return k8srand.String(12)
}

// GenerateUserSecrets generated user secrets
func GenerateUserSecrets(clusterName string, clusterId int64) *apitypes.OBUserSecrets {
	return &apitypes.OBUserSecrets{
		Root:     fmt.Sprintf("%s-%d-root-%s", clusterName, clusterId, GenerateUUID()),
		ProxyRO:  fmt.Sprintf("%s-%d-proxyro-%s", clusterName, clusterId, GenerateUUID()),
		Monitor:  fmt.Sprintf("%s-%d-monitor-%s", clusterName, clusterId, GenerateUUID()),
		Operator: fmt.Sprintf("%s-%d-operator-%s", clusterName, clusterId, GenerateUUID()),
	}
}

// GenerateClusterID generated random cluster ID
func GenerateClusterID() int64 {
	clusterID := time.Now().Unix() % factor
	if clusterID != 0 {
		return clusterID
	}
	return GenerateClusterID()
}

// GenerateNaivePassword generated naive password in length 16
func GenerateNaivePassword() string {
	return k8srand.String(16)
}

// GenerateRandomPassword generated random password in range [minLength,maxLength]
func GenerateRandomPassword(minLength int, maxLength int) string {
	const (
		minUppercase   = 2
		minLowercase   = 2
		minNumber      = 2
		minSpecialChar = 2
	)
	var (
		countUppercase   int
		countLowercase   int
		countNumber      int
		countSpecialChar int
	)

	var sb strings.Builder
	for countUppercase < minUppercase || countLowercase < minLowercase || countNumber < minNumber || countSpecialChar < minSpecialChar {
		b := make([]byte, 1)
		_, err := rand.Read(b)
		if err != nil {
			panic(err)
		}

		randomIndex := int(b[0]) % len(characters)
		randomChar := characters[randomIndex]
		if err := sb.WriteByte(randomChar); err != nil {
			panic(err)
		}
		switch {
		case strings.ContainsRune("ABCDEFGHIJKLMNOPQRSTUVWXYZ", rune(randomChar)):
			countUppercase++
		case strings.ContainsRune("abcdefghijklmnopqrstuvwxyz", rune(randomChar)):
			countLowercase++
		case strings.ContainsRune("0123456789", rune(randomChar)):
			countNumber++
		default:
			countSpecialChar++
		}
	}
	if len(sb.String()) < minLength || len(sb.String()) > maxLength {
		return GenerateRandomPassword(minLength, maxLength)
	}
	return sb.String()
}
