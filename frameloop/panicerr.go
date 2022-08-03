package frameloop

import (
	"fmt"
)

// ErrPanic is used to check for a PanicErr with errors.Is, and should
// not be used for anything else.
var ErrPanic = &PanicErr{}

// PanicErr is used when a panic is turned into an error.
//
// If the panic value is an error, this error unwraps into that error.
type PanicErr struct {
	// Where is where the panic occurred, and is used in the message.
	Where string

	// Value is the recovered value from the panic.
	Value interface{}
}

// Error implements the error interface.
func (e PanicErr) Error() string {
	return fmt.Sprintf("panic in %s: %v", e.Where, e.Value)
}

// Unwrap returns the panic value if it was an error, or nil otherwise.
func (e PanicErr) Unwrap() error {
	if v, ok := e.Value.(error); ok {
		return v
	}
	return nil
}

// Is checks if the target is ErrPanic, to allow use with errors.Is
func (e PanicErr) Is(target error) bool {
	return target == ErrPanic
}

// As converts between PanicErr and *PanicErr, for use with errors.As.
func (e PanicErr) As(target interface{}) bool {
	if t, ok := target.(*PanicErr); ok {
		// Turn pointer into plain value
		*t = e
		return true
	}
	if t, ok := target.(**PanicErr); ok {
		// Turn plain value into pointer
		*t = &e
		return true
	}
	return false
}
