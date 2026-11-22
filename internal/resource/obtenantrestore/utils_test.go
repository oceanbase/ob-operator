/*
Copyright (c) 2026 OceanBase
ob-operator is licensed under Mulan PSL v2.
You may obtain a copy of Mulan PSL v2 at:
         http://license.coscl.org.cn/MulanPSL2
THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND.
*/

package obtenantrestore

import (
	"context"
	"crypto/rand"
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	"github.com/oceanbase/ob-operator/api/constants"
	apitypes "github.com/oceanbase/ob-operator/api/types"
	"github.com/oceanbase/ob-operator/api/v1alpha1"
)

func TestRestoreSourceStorageOptions(t *testing.T) {
	const namespace = "backup-test"
	accessID, secretKey := rand.Text(), rand.Text()
	credentials := "&access_id=" + accessID + "&access_key=" + secretKey + ""
	for _, tt := range []struct {
		name                                                                    string
		provider                                                                apitypes.BackupDestType
		scheme, secretKey, dataOption, archiveOption, dataSuffix, archiveSuffix string
	}{
		{"S3 compatible explicit model", constants.BackupDestTypeS3Compatible, "s3", "addressingModel", "path", "path", "&addressing_model=path", "&addressing_model=path"},
		{"S3 compatible default", constants.BackupDestTypeS3Compatible, "s3", "", "", "", "", ""},
		{"S3 region", constants.BackupDestTypeS3, "s3", "s3Region", "us-east-1", "us-west-2", "&s3_region=us-east-1", "&s3_region=us-west-2"},
		{"COS app ID", constants.BackupDestTypeCOS, "cos", "appId", "100", "200", "&appid=100", "&appid=200"},
		{"OSS credentials", constants.BackupDestTypeOSS, "oss", "", "", "", "", ""},
	} {
		t.Run(tt.name, func(t *testing.T) {
			secret := func(name, option string) *corev1.Secret {
				data := map[string][]byte{"accessId": []byte(accessID), "accessKey": []byte(secretKey)}
				if tt.secretKey != "" {
					data[tt.secretKey] = []byte(option)
				}
				return &corev1.Secret{ObjectMeta: metav1.ObjectMeta{Namespace: namespace, Name: name}, Data: data}
			}
			dest := func(name string) *apitypes.BackupDestination {
				return &apitypes.BackupDestination{Type: tt.provider, Path: tt.scheme + "://bucket/" + name + "?host=store", OSSAccessSecret: name}
			}
			source := v1alpha1.RestoreSourceSpec{BakDataSource: dest("data"), ArchiveSource: dest("archive")}
			manager := &ObTenantRestoreManager{
				Ctx:      context.Background(),
				Client:   fake.NewClientBuilder().WithObjects(secret("data", tt.dataOption), secret("archive", tt.archiveOption)).Build(),
				Resource: &v1alpha1.OBTenantRestore{ObjectMeta: metav1.ObjectMeta{Namespace: namespace}, Spec: v1alpha1.OBTenantRestoreSpec{Source: source}},
			}
			want := source.BakDataSource.Path + credentials + tt.dataSuffix + "," + source.ArchiveSource.Path + credentials + tt.archiveSuffix
			got, err := manager.getSourceUri()
			if err != nil || got != want {
				t.Fatal("unexpected restore source URI or provider options")
			}
		})
	}
}

func TestRestoreSourceLegacyPaths(t *testing.T) {
	for _, tt := range []struct {
		name   string
		source v1alpha1.RestoreSourceSpec
		want   string
	}{
		{"deprecated URI", v1alpha1.RestoreSourceSpec{SourceUri: "file:///legacy/data,file:///legacy/archive"}, "file:///legacy/data,file:///legacy/archive"},
		{"explicit and implicit NFS", v1alpha1.RestoreSourceSpec{
			BakDataSource: &apitypes.BackupDestination{Type: constants.BackupDestTypeNFS, Path: "data"},
			ArchiveSource: &apitypes.BackupDestination{Path: "archive"},
		}, "file:///ob-backup/data,file:///ob-backup/archive"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			manager := &ObTenantRestoreManager{Resource: &v1alpha1.OBTenantRestore{Spec: v1alpha1.OBTenantRestoreSpec{Source: tt.source}}}
			got, err := manager.getSourceUri()
			if err != nil || got != tt.want {
				t.Fatalf("restore source = %q, error = %v; want %q", got, err, tt.want)
			}
		})
	}
}

func TestMissingRestoreSource(t *testing.T) {
	manager := &ObTenantRestoreManager{Resource: &v1alpha1.OBTenantRestore{}}
	if _, err := manager.getSourceUri(); err == nil {
		t.Fatal("expected an error for missing data and archive sources")
	}
}
