package obzone

import (
	"testing"

	"k8s.io/apimachinery/pkg/api/resource"

	apitypes "github.com/oceanbase/ob-operator/api/types"
	"github.com/oceanbase/ob-operator/api/v1alpha1"
)

func TestMonitorRolloutPatchesChangedLimitsOnce(t *testing.T) {
	m, _ := rollingFixture(t, []string{"observer"}, false, false)
	if !m.CanReplayTask(tModifyPodTemplate) || !m.CanReplayTask(tWaitForOBServerTemplateModifying) || !m.CanReplayTask(tWaitOBServerRunning) {
		t.Fatal("lost zone rollout must resume its idempotent patch and readiness checks")
	}
	read := func() *v1alpha1.OBServer {
		t.Helper()
		list, err := m.listOBServers()
		if err != nil || len(list.Items) != 1 {
			t.Fatalf("read Observer: %v, %v", list, err)
		}
		return &list.Items[0]
	}
	template := &apitypes.MonitorTemplate{Image: "agent", Resource: &apitypes.ResourceSpec{
		Cpu: resource.MustParse("0.5"), Memory: resource.MustParse("1Gi"),
	}}
	observer := read()
	observer.Spec.MonitorTemplate = template.DeepCopy()
	if err := m.Client.Update(m.Ctx, observer); err != nil {
		t.Fatal(err)
	}
	m.OBZone.Spec.Topology.MonitorTemplate = template.DeepCopy()
	m.OBZone.Spec.Topology.MonitorTemplate.Resource.Memory = resource.MustParse("1024Mi")
	if m.checkIfMonitorMutated(read()) {
		t.Fatal("equivalent override triggered a rollout")
	}
	desired := m.OBZone.Spec.Topology.MonitorTemplate
	desired.Resource.Cpu = resource.MustParse("0")
	if !m.checkIfMonitorMutated(read()) {
		t.Fatal("removal of explicit CPU limit did not trigger a rollout")
	}
	before := read().ResourceVersion
	if err := ModifyPodTemplate(m); err != nil {
		t.Fatal(err)
	}
	got := read()
	if got.Spec.MonitorTemplate == nil || !got.Spec.MonitorTemplate.Resource.Cpu.IsZero() || got.Annotations["oceanbase.oceanbase.com/monitor-revision"] != before {
		t.Fatalf("updated monitor and revision: %+v, %v", got.Spec.MonitorTemplate, got.Annotations)
	}
	if err := ModifyPodTemplate(m); err != nil {
		t.Fatal(err)
	}
	if read().ResourceVersion != got.ResourceVersion {
		t.Fatal("retry issued a second monitor patch")
	}
}
