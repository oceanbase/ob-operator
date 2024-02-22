package response

type APIResponse struct {
	Data       interface{} `json:"data"`
	Message    string      `json:"message"`
	Successful bool        `json:"successful"`
}
