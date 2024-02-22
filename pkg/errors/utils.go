package errors

func NewNotImplemented(msg string) ObError {
	return &httpErr{
		errorType: ErrNotImplemented,
		message:   msg,
	}
}

func NewBadRequest(msg string) ObError {
	return &httpErr{
		errorType: ErrBadRequest,
		message:   msg,
	}
}

func NewUnauthorized(msg string) ObError {
	return &httpErr{
		errorType: ErrUnauthorized,
		message:   msg,
	}
}

func NewNotFound(msg string) ObError {
	return &httpErr{
		errorType: ErrNotFound,
		message:   msg,
	}
}

func NewInternal(msg string) ObError {
	return &httpErr{
		errorType: ErrInternal,
		message:   msg,
	}
}
