package response

type APIResponse struct {
	Data       any    `json:"data"`
	Message    string `json:"message"`
	Successful bool   `json:"successful"`
}
