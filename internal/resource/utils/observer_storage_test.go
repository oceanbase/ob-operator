package utils

import (
	"testing"

	apitypes "github.com/oceanbase/ob-operator/api/types"
	"github.com/oceanbase/ob-operator/api/v1alpha1"
	oc "github.com/oceanbase/ob-operator/internal/const/oceanbase"
)

func TestEffectiveZoneStorageClasses(t *testing.T) {
	template := func(log, redo string) *apitypes.OBServerTemplate {
		s := &apitypes.OceanbaseStorageSpec{DataStorage: &apitypes.StorageSpec{StorageClass: "data"}, LogStorage: &apitypes.StorageSpec{StorageClass: log}}
		if redo != "" {
			s.RedoLogStorage = &apitypes.StorageSpec{StorageClass: redo}
		}
		return &apitypes.OBServerTemplate{Storage: s}
	}
	base, override := template("log", "redo"), template("custom", "archive")
	zone := &v1alpha1.OBZone{}
	zone.Spec.OBServerTemplate = base
	if EffectiveZoneTemplate(zone) != base {
		t.Fatal("missing default")
	}
	zone.Spec.Topology.OBServerTemplate = override
	if EffectiveZoneTemplate(zone) != override {
		t.Fatal("per-zone storage override lost")
	}
	for _, c := range []struct {
		name  string
		a, b  *apitypes.OBServerTemplate
		anno  map[string]string
		match bool
	}{
		{"standalone classes", base, override, nil, false},
		{"multiple claims add redo", template("log", ""), base, map[string]string{oc.AnnotationsSinglePVC: "false"}, false},
		{"multiple claims remove redo", base, template("log", ""), map[string]string{oc.AnnotationsSinglePVC: "false"}, false},
		{"old unknown layout", base, template("log", ""), nil, true},
		{"shared has no redo", base, template("log", "unused"), map[string]string{oc.AnnotationsDeploymentMode: oc.DeploymentModeSharedStorage}, true},
		{"single claim uses data class", base, override, map[string]string{oc.AnnotationsSinglePVC: "true"}, true},
	} {
		t.Run(c.name, func(t *testing.T) {
			if got := SameObserverStorageClasses(c.a, c.b, c.anno); got != c.match {
				t.Fatalf("classes match=%v, want %v", got, c.match)
			}
		})
	}
}
