package sql

type PlanDetailParam struct {
	PlanIdentity `json:",inline"`
	Namespace    string `json:"namespace" binding:"required"`
	OBTenant     string `json:"obtenant" binding:"required"`
}
