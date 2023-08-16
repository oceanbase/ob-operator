package v1alpha1

type ResourcePoolSpec struct {
	ZoneList   string `json:"zone"`  // zone1
	UnitNumber int    `json:"unitNum"`
	Priority   int          `json:"priority,omitempty"`
	Type       LocalityType `json:"type"`
	UnitConfig UnitConfig   `json:"resource"`
}
