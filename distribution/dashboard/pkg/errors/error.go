package errors

import (
	"fmt"
	"net/http"
)

type oberror struct {
	errorType ErrorType
	message   string
	children  []*oberror
}

func (e *oberror) Error() string {
	return fmt.Sprintf("Error %s: %s", e.errorType, e.message)
}

func (e *oberror) IsType(errType ErrorType) bool {
	return e.errorType == errType
}

func (e *oberror) Contains(errType ErrorType) bool {
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

func (e *oberror) Wrap(err error) ObError {
	if err == nil {
		return e
	}
	obErr, ok := err.(*oberror)
	if !ok {
		obErr = &oberror{
			errorType: e.errorType,
			message:   err.Error(),
		}
	}
	e.children = append(e.children, obErr)
	return e
}

func (e *oberror) Type() ErrorType {
	return e.errorType
}

func (e *oberror) Status() int {
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
