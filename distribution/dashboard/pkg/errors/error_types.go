package errors

type ErrorType string

var (
	ErrNotFound           ErrorType = "NotFound"
	ErrUnauthorized       ErrorType = "Unauthorized"
	ErrForbidden          ErrorType = "Forbidden"
	ErrConflict           ErrorType = "Conflict"
	ErrInternal           ErrorType = "Internal"
	ErrTimeout            ErrorType = "Timeout"
	ErrTooManyRequests    ErrorType = "TooManyRequests"
	ErrBadRequest         ErrorType = "BadRequest"
	ErrInvalid            ErrorType = "Invalid"
	ErrNotSupported       ErrorType = "NotSupported"
	ErrAlreadyExists      ErrorType = "AlreadyExists"
	ErrNotReady           ErrorType = "NotReady"
	ErrNotImplemented     ErrorType = "NotImplemented"
	ErrServiceUnavailable ErrorType = "ServiceUnavailable"
)
