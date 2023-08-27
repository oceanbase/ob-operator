package fail

type FailureRule struct {
	Strategy      string `json:"failureStrategy"`
	NextTryStatus string `json:"failureStatus"`
}
