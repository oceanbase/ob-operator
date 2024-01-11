package oceanbase

type ErrorType string

const (
	ErrorTypeBadRequest ErrorType = "BadRequest"
	ErrorTypeNotFound   ErrorType = "NotFound"
	ErrorTypeInternal   ErrorType = "Internal"
)

type OBError struct {
	Message string    `json:"message"`
	Type    ErrorType `json:"type"`
}

func (e *OBError) Error() string {
	return e.Message
}

func Is(err error, et ErrorType) bool {
	if err == nil {
		return false
	}
	obErr, ok := err.(*OBError)
	if !ok {
		return false
	}
	return obErr.Type == et
}

func NewOBError(errorType ErrorType, message string) *OBError {
	return &OBError{
		Message: message,
		Type:    errorType,
	}
}
