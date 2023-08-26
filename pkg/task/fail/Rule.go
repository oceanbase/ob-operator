package fail

type FailureRule struct {
	Strategy string `json:"failureStrategy"`
	FailureStatus string `json:"failureStatus"`
}
