package v1alpha1

type ResourcePoolSpec struct {
	ZoneList   string       `json:"zoneList"`
	Priority   int          `json:"priority,omitempty"`
	Type       LocalityType `json:"type"`
	UnitConfig UnitConfig   `json:"resource"`
}
