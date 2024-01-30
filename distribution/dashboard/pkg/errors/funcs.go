package errors

import "fmt"

func New(errType ErrorType, message string) ObError {
	return &httpErr{
		errorType: errType,
		message:   message,
	}
}

func Newf(errType ErrorType, format string, args ...interface{}) ObError {
	return New(errType, fmt.Sprintf(format, args...))
}

func Wrap(err error, errType ErrorType, message string) ObError {
	return New(errType, message).Wrap(err)
}

func Wrapf(err error, errType ErrorType, format string, args ...interface{}) ObError {
	return Wrap(err, errType, fmt.Sprintf(format, args...))
}
