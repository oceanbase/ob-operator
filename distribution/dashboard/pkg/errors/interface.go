package errors

type ObError interface {
	error
	IsType(errType ErrorType) bool
	Contains(errType ErrorType) bool
	Wrap(err error) ObError
	Type() ErrorType
	Status() int
}
