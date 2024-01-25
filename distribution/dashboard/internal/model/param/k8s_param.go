package param

type CreateNamespaceParam struct {
	Namespace string `json:"namespace"`
}

type QueryEventParam struct {
	ObjectType string `json:"objectType" query:"objectType" binding:"omitempty"`
	Type       string `json:"type" query:"type" binding:"omitempty"`
	Name       string `json:"name" query:"name" binding:"omitempty"`
	Namespace  string `json:"namespace" query:"namespace" binding:"omitempty"`
}
