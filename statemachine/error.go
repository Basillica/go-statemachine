package statemachine

import "fmt"

// CustomError is a type that can be checked by the state machine's error handling.
type CustomError struct {
	Name string
	Err  error
}

func (e *CustomError) Error() string {
	return e.Name + ": " + e.Err.Error()
}

// Unwrap allows errors.Is and errors.As to check for the underlying error.
func (e *CustomError) Unwrap() error {
	return e.Err
}

// Define specific custom errors
var (
	ErrAPIBadGateway = &CustomError{Name: "API_BAD_GATEWAY", Err: fmt.Errorf("api service is unavailable")}
	ErrTimeout       = &CustomError{Name: "TIMEOUT", Err: fmt.Errorf("task timed out")}
)
