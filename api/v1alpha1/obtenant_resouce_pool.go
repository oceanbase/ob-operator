package v1alpha1

type ResourcePoolSpec struct {
	ZoneList   string       `json:"zoneList"`
	UnitNumber int          `json:"unitNum"`
	Priority   int          `json:"priority,omitempty"`
	Type       LocalityType `json:"type"`
	UnitConfig UnitConfig   `json:"resource"`
}
