package errors

func NewNotImplemented(msg string) ObError {
	return &oberror{
		errorType: ErrNotImplemented,
		message:   msg,
	}
}

func NewBadRequest(msg string) ObError {
	return &oberror{
		errorType: ErrBadRequest,
		message:   msg,
	}
}

func NewUnauthorized(msg string) ObError {
	return &oberror{
		errorType: ErrUnauthorized,
		message:   msg,
	}
}

func NewNotFound(msg string) ObError {
	return &oberror{
		errorType: ErrNotFound,
		message:   msg,
	}
}

func NewInternal(msg string) ObError {
	return &oberror{
		errorType: ErrInternal,
		message:   msg,
	}
}
