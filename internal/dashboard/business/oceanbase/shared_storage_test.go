package oceanbase

import (
	"encoding/json"
	"testing"

	apitypes "github.com/oceanbase/ob-operator/api/types"
	modelcommon "github.com/oceanbase/ob-operator/internal/dashboard/model/common"
	"github.com/oceanbase/ob-operator/internal/dashboard/model/param"
	"github.com/oceanbase/ob-operator/internal/dashboard/model/response"
)

func TestGenerateSharedStorageCluster(t *testing.T) {
	p := &param.CreateOBClusterParam{
		Name: "ss", Namespace: "test", ClusterName: "ss", ClusterId: 9,
		Mode: modelcommon.ClusterModeService, DeploymentMode: "shared_storage",
		SharedStorageInfo: &modelcommon.SharedStorageSpec{BucketURL: "s3://bucket", SecretRef: modelcommon.ObjectReference{Name: "object-store"}, MaxIOPS: "1000"},
		LogServiceRef:     &modelcommon.ObjectReference{Name: "ls"},
		OBServer:          &param.OBServerSpec{Image: "ai:4.6.2.0", Storage: &param.OBServerStorageSpec{}},
		Topology:          []param.ZoneTopology{{Zone: "zone1", Replicas: 1}},
	}
	cluster := generateOBClusterInstance(p)
	if cluster.Spec.DeploymentMode != "shared_storage" || cluster.Spec.LogServiceRef.Name != "ls" || cluster.Spec.SharedStorageInfo.BucketURL != "s3://bucket" || cluster.Spec.SharedStorageInfo.MaxIOPS != "1000" {
		t.Fatalf("SS fields lost: %+v", cluster.Spec)
	}
	if cluster.Spec.OBServerTemplate.Storage.RedoLogStorage != nil || !cluster.SupportStaticIP() {
		t.Fatal("SS must omit redoLog and preserve SERVICE networking")
	}
	// Verify the detail JSON contract exposes references, not credential data.
	data, err := json.Marshal(response.OBClusterExtra{SharedStorageInfo: modelcommon.SharedStorageFromAPI(cluster.Spec.SharedStorageInfo), LogServiceRef: modelcommon.LogServiceRefFromAPI(cluster.Spec.LogServiceRef)})
	if err != nil {
		t.Fatal(err)
	}
	var decoded map[string]any
	if err = json.Unmarshal(data, &decoded); err != nil {
		t.Fatal(err)
	}
	if decoded["sharedStorageInfo"] == nil || decoded["logServiceRef"] == nil {
		t.Fatal(string(data))
	}

	p.DeploymentMode = ""
	p.SharedStorageInfo, p.LogServiceRef = nil, nil
	p.OBServer.Storage.RedoLog = &modelcommon.StorageSpec{SizeGB: 30}
	normal := generateOBClusterInstance(p)
	if normal.Spec.DeploymentMode != "" || normal.Spec.SharedStorageInfo != nil || normal.Spec.LogServiceRef != nil || normal.Spec.OBServerTemplate.Storage.RedoLogStorage == nil {
		t.Fatal("Legacy normal create changed")
	}
}

func TestSharedStorageDetailExcludesInlineCredentials(t *testing.T) {
	for _, tc := range []struct{ raw, want string }{
		{"s3://bucket/log?host=http://minio:9000&s3_region=local&max_iops=100", "s3://bucket/log?host=http://minio:9000&s3_region=local&max_iops=100"},
		{"s3://user:example@bucket/log?host=minio&access_key=example#private", "s3://bucket/log?host=minio"},
		{"oss://bucket?host=https://store.test%3Fsignature%3Dexample", "oss://bucket"},
		{"s3://bucket?host=minio%26access_key%3Dexample", "s3://bucket"},
	} {
		got := modelcommon.SharedStorageFromAPI(&apitypes.SharedStorageSpec{BucketURL: tc.raw})
		if got.BucketURL != tc.want {
			t.Fatal("storage detail returned an unexpected public location")
		}
	}
}
