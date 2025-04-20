package common

import (
	"fmt"
)

type RuntimeError struct {
	FunctionName string
	Message      string
	Err          error
}

// All structs that have this method can be "error" type.
func (e *RuntimeError) Error() string {
	return fmt.Sprintf("%s: %s", e.FunctionName, e.Message)
}

// WrapErrorWithContext adds context information to the error.
// functionName is the current function name.
// This function maintains the error chain while only connecting function names.
func WrapErrorWithContext(functionName string, err error) error {
	if err == nil {
		return nil
	}
	// If it's already a RuntimeError, just connect the function names
	if runtimeErr, ok := err.(*RuntimeError); ok {
		runtimeErr.FunctionName = functionName + " > " + runtimeErr.FunctionName
		return runtimeErr
	}

	// If it's a normal error, convert it to RuntimeError
	return &RuntimeError{
		FunctionName: functionName,
		Message:      err.Error(),
		Err:          err,
	}
}
