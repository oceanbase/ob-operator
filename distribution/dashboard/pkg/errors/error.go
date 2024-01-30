package errors

import (
	"fmt"
	"net/http"
)

type httpErr struct {
	errorType ErrorType
	message   string
	children  []*httpErr
}

func (e *httpErr) Error() string {
	return fmt.Sprintf("Error %s: %s", e.errorType, e.message)
}

func (e *httpErr) IsType(errType ErrorType) bool {
	return e.errorType == errType
}

func (e *httpErr) Contains(errType ErrorType) bool {
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

func (e *httpErr) Wrap(err error) ObError {
	if err == nil {
		return e
	}
	obErr, ok := err.(*httpErr)
	if !ok {
		obErr = &httpErr{
			errorType: e.errorType,
			message:   err.Error(),
		}
	}
	e.children = append(e.children, obErr)
	return e
}

func (e *httpErr) Type() ErrorType {
	return e.errorType
}

func (e *httpErr) Status() int {
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
