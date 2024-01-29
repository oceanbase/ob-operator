package errors

func NewNotImplemented(msg string) ObError {
	return &buzerror{
		errorType: ErrNotImplemented,
		message:   msg,
	}
}

func NewBadRequest(msg string) ObError {
	return &buzerror{
		errorType: ErrBadRequest,
		message:   msg,
	}
}

func NewUnauthorized(msg string) ObError {
	return &buzerror{
		errorType: ErrUnauthorized,
		message:   msg,
	}
}

func NewNotFound(msg string) ObError {
	return &buzerror{
		errorType: ErrNotFound,
		message:   msg,
	}
}

func NewInternal(msg string) ObError {
	return &buzerror{
		errorType: ErrInternal,
		message:   msg,
	}
}
