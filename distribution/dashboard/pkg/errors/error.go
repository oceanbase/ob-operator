package errors

import (
	"fmt"
	"net/http"
)

type buzerror struct {
	errorType ErrorType
	message   string
	children  []*buzerror
}

func (e *buzerror) Error() string {
	return fmt.Sprintf("Error %s: %s", e.errorType, e.message)
}

func (e *buzerror) IsType(errType ErrorType) bool {
	return e.errorType == errType
}

func (e *buzerror) Contains(errType ErrorType) bool {
	if e.errorType == errType {
		return true
	}
	for _, child := range e.children {
		if child.Contains(errType) {
			return true
		}
	}
	return false
}

func (e *buzerror) Wrap(err error) ObError {
	if err == nil {
		return e
	}
	obErr, ok := err.(*buzerror)
	if !ok {
		obErr = &buzerror{
			errorType: e.errorType,
			message:   err.Error(),
		}
	}
	e.children = append(e.children, obErr)
	return e
}

func (e *buzerror) Type() ErrorType {
	return e.errorType
}

func (e *buzerror) Status() int {
	switch e.errorType {
	case ErrBadRequest:
		return http.StatusBadRequest
	case ErrUnauthorized:
		return http.StatusUnauthorized
	case ErrNotImplemented:
		return http.StatusNotImplemented
	case ErrInternal:
		return http.StatusInternalServerError
	case ErrNotFound:
		return http.StatusNotFound
	default:
		return http.StatusServiceUnavailable
	}
}
