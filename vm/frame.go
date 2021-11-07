package vm

import (
	"github.com/yourfavoritedev/golang-interpreter/code"
	"github.com/yourfavoritedev/golang-interpreter/object"
)

// Frame is the struct that holds the execution-relevant information for a function.
// It is effectively like the inner-environment of a function, allowing the VM
// to execute its instructions and update the ip (instruction-pointer) without entangling them with the outer scopes.
// fn points to the compiled function referenced by the frame.
// ip is the instruction pointer in this frame for this function.
// basePointer is a stack pointer value to indicate where in the stack to start allocating memory when executing the function,
// it is used to create a "hole" on the stack to store all the local bindings of the function.
// Below the "hole" is the lower boundary which contains all the values on the stack before calling the function.
// If sp is 3 before calling the function, the lower boundary contains indices [0, 1, 2].
// The "hole" will be n-length deep where n is the number of local-bindings. Above the "hole" is the function's workspace,
// where it will push and pop values, if sp is 3 it should use the indices that are greater than 3+n.
// When the function exits, we can restore the stack, removing all values after the initial basePointer, thus giving us
// the stack before the function was called.
type Frame struct {
	fn          *object.CompiledFunction
	ip          int
	basePointer int
}

// NewFrame creates a new frame for the given compiled function
func NewFrame(fn *object.CompiledFunction, basePointer int) *Frame {
	return &Frame{
		fn: fn,
		// ip is initialized with -1, because we increment ip immediately when we start executing
		// the frame's instructions, thus giving us the first instruction at position 0.
		ip:          -1,
		basePointer: basePointer,
	}
}

// Instructions simply returns the instructions of the compiled function
func (f *Frame) Instructions() code.Instructions {
	return f.fn.Instructions
}
