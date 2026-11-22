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

package utils

import (
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestReadBackupEncryptionPassword(t *testing.T) {
	client := storageTestClient(
		&corev1.Secret{ObjectMeta: metav1.ObjectMeta{Namespace: "qa", Name: "empty"}},
		&corev1.Secret{ObjectMeta: metav1.ObjectMeta{Namespace: "qa", Name: "wrong-key"}, Data: map[string][]byte{"other": []byte("unused")}},
		&corev1.Secret{ObjectMeta: metav1.ObjectMeta{Namespace: "qa", Name: "valid"}, Data: map[string][]byte{"password": []byte(" preserve spaces ")}},
	)
	for _, name := range []string{"missing", "empty", "wrong-key"} {
		if _, err := ReadBackupEncryptionPassword(client, "qa", name); err == nil {
			t.Errorf("accepted unusable encryption Secret %q", name)
		}
	}
	got, err := ReadBackupEncryptionPassword(client, "qa", "valid")
	if err != nil || got != " preserve spaces " {
		t.Fatal("did not preserve the configured password")
	}
}
