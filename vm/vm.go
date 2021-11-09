package vm

import (
	"fmt"

	"github.com/yourfavoritedev/golang-interpreter/code"
	"github.com/yourfavoritedev/golang-interpreter/compiler"
	"github.com/yourfavoritedev/golang-interpreter/object"
)

const StackSize = 2048    // arbitrary number
const GlobalsSize = 65536 // upper limit on the number of global bindings since operands are 16 bits wide.
const MaxFrames = 1024    // arbitrary number

var True = &object.Boolean{Value: true}
var False = &object.Boolean{Value: false}
var Null = &object.Null{}

// VM is the struct for our virtual-machine. It holds the bytecode instructions and constants-pool generated by the compiler.
// A VM implements a stack, as it executes the bytecode, it organizes (push, pop, etc) the evaluated constants on the stack.
// The field sp helps keep track of the position of the next item in the stack (top to bottom).
type VM struct {
	constants []object.Object
	stack     []object.Object
	// sp always points to the next free slot in the stack. If there's one element on the stack,
	// located at index 0, the value of sp would be 1 and to access that element we'd use stack[sp-1].
	sp int
	// globals helps us store and retreive values observed by the VM as it executes the bytecode instructions.
	// specifically for identifier values, in which an index for that identifier is associated and can be used to retrieve its value.
	globals []object.Object
	// frames is the data-structure used to organize the unique frames for compiled functions as the VM executes their bytecode.
	// the bytecode instructions will be held by a single "main frame" and will be added during the initializing of the VM.
	frames []*Frame
	// frameIndex refers to the position of the current frame the VM is working in
	framesIndex int
}

// New initializes a new VM using the bytecode generated by the compiler.
// VMs are initialized with an sp of 0 (the initial top). The stack
// will have a preallocated number of elements (StackSize).
func New(bytecode *compiler.Bytecode) *VM {
	// constuct a "main frame" with the bytecode instructions
	mainFn := &object.CompiledFunction{Instructions: bytecode.Instructions}
	mainFrame := NewFrame(mainFn, 0)

	frames := make([]*Frame, MaxFrames)
	frames[0] = mainFrame

	return &VM{
		constants:   bytecode.Constants,
		stack:       make([]object.Object, StackSize),
		sp:          0,
		globals:     make([]object.Object, GlobalsSize),
		frames:      frames,
		framesIndex: 1,
	}
}

// currentFrame simply returns the current frame, the framesIndex is always prepped to allocate a new frame
// which is why we need to decrement it by 1 to get the current frame.
func (vm *VM) currentFrame() *Frame {
	return vm.frames[vm.framesIndex-1]
}

// pushFrame adds a new frame to the VM's frames and preps the VM for a future frame to be added.
func (vm *VM) pushFrame(f *Frame) {
	vm.frames[vm.framesIndex] = f
	vm.framesIndex++
}

// popFrame returns the current frame and makes its position available for a future frame to be added.
func (vm *VM) popFrame() *Frame {
	vm.framesIndex--
	return vm.frames[vm.framesIndex]
}

// Run will start the VM. The VM will execute the bytecode and handle
// the specific instructions (opcode + operands) that it was provided
// from the compiler. It executes the fetch-decode-execute cycle.
func (vm *VM) Run() error {
	var ip int
	var ins code.Instructions
	var op code.Opcode

	// iterate through all instructions in the current frame.
	for vm.currentFrame().ip < len(vm.currentFrame().Instructions())-1 {
		vm.currentFrame().ip++

		ip = vm.currentFrame().ip
		ins = vm.currentFrame().Instructions()

		// FETCH the instruction (opcode + operand) at the specific position (ip, the instruction pointer)
		// then convert the instruction's first-byte into an Opcode (which is what we expect it to be)
		op = code.Opcode(ins[ip])
		// DECODE SECTION
		switch op {
		// OpConstant has an operand to decode
		case code.OpConstant:
			// grab the two-byte operand for the OpConstant instruction (the operand starts right after the Opcode byte)
			operand := ins[ip+1:]
			// decode the operand, getting back the identifier for the constant's position in the constants pool
			constIndex := code.ReadUint16(operand)
			// increment the instruction-pointer by 2 because OpConstant has one two-byte wide operand
			vm.currentFrame().ip += 2
			// EXECUTE, grab the constant from the pool and push it on to the stack
			err := vm.push(vm.constants[constIndex])
			if err != nil {
				return err
			}

		// Execute the binary operation for the Opcode arithmetic instruction.
		case code.OpAdd, code.OpSub, code.OpMul, code.OpDiv:
			err := vm.executeBinaryOperation(op)
			if err != nil {
				return err
			}

		// Execute the comparison operation for the Opcode comparison instruction.
		case code.OpGreaterThan, code.OpEqual, code.OpNotEqual:
			err := vm.executeComparison(op)
			if err != nil {
				return err
			}

		// Execute the minus "-" operation for this Opcode instruction.
		case code.OpMinus:
			err := vm.executeMinusOperator()
			if err != nil {
				return err
			}

		// Execute the bang "!" operation for this Opcode instruction.
		case code.OpBang:
			err := vm.executeBangOperator()
			if err != nil {
				return err
			}

		// Execute the boolean Opcode instructions. Simply push the corresponding Object.Boolean to the stack.
		case code.OpTrue:
			err := vm.push(True)
			if err != nil {
				return err
			}
		case code.OpFalse:
			err := vm.push(False)
			if err != nil {
				return err
			}

		// Execute OpJump instruction to jump to the next instruction byte after compiing a truthy condition.
		case code.OpJump:
			operand := ins[ip+1:]
			// decode the operand and get back the absolute position of the byte to jump to
			pos := int(code.ReadUint16(operand))
			// since we're in a loop that increments ip with each iteration, we need to set ip
			// to the offset right before the one we want. That lets the loop do its work
			// and ip gets set to the value we want in the next cycle to process that instruction
			vm.currentFrame().ip = pos - 1

		// Execute OpJumpNotTruthy instruction to jump to the next instruction byte after compiing a falsey condition.
		case code.OpJumpNotTruthy:
			operand := ins[ip+1:]
			// decode the operand and get back the absolute position of the byte to jump to if condition is not truthy
			pos := int(code.ReadUint16(operand))
			// increment the instruction-pointer by 2 because OpJumpNotTruthy has one two-byte wide operand
			// this would prepare us for the next iteration to evaluate the OpConstant - the result of a truthy condition
			vm.currentFrame().ip += 2

			// pop the condition constant (True or False) and determine where we need to jump
			condition := vm.pop()
			if !isTruthy(condition) {
				// jump pass the consequence when the condition is falsey to process the next instruction
				vm.currentFrame().ip = pos - 1
			}

		// Execute OpSetGlobal instruction
		case code.OpSetGlobal:
			// decode the operand to get back the global index associated with that identifier
			globalIndex := code.ReadUint16(ins[ip+1:])
			vm.currentFrame().ip += 2

			// pop the top element off the stack, which should be the value bound to an identifier
			// and save that value in the vm's globals store under the specified index. Making it easy
			// to retrieve when we need to push that value on to the stack again.
			vm.globals[globalIndex] = vm.pop()

		// Execute OpGetGlobal instruction
		case code.OpGetGlobal:
			// decode the operand to get back the global index associated with that identifier
			globalIndex := code.ReadUint16(ins[ip+1:])
			vm.currentFrame().ip += 2

			// with an OpGetGlobal instruction, we can assume that vm.globals has already
			// recorded the value associated with this identifier in its store at the
			// globalIndex. We simply need to push that value back onto the stack.
			err := vm.push(vm.globals[globalIndex])
			if err != nil {
				return err
			}

		// Execute OpSetLocal instruction
		case code.OpSetLocal:
			operand := ins[ip+1]
			localIndex := int(operand)
			vm.currentFrame().ip += 1

			frame := vm.currentFrame()

			// set element in stack "hole" reserved for local binding value
			vm.stack[frame.basePointer+localIndex] = vm.pop()

		// Execute OpGetLocal instruction
		case code.OpGetLocal:
			operand := ins[ip+1]
			localIndex := int(operand)
			vm.currentFrame().ip += 1

			frame := vm.currentFrame()
			// push the value in the "hole" to the stack
			err := vm.push(vm.stack[frame.basePointer+localIndex])
			if err != nil {
				return err
			}

		// Execute OpGetBuiltin instruction
		case code.OpGetBuiltin:
			operand := ins[ip+1]
			builtinIndex := int(operand)
			vm.currentFrame().ip += 1
			// use index to grab the built-in function from the object.Builtins slice
			definition := object.Builtins[builtinIndex]
			// push the built-in function to the stack
			err := vm.push(definition.Builtin)
			if err != nil {
				return err
			}

		// Execute OpArray instruction, it should construct an array and push it on to the stack,
		// using the values (if any) that were previously loaded.
		case code.OpArray:
			// derive the number of elements to pull from the operand
			operand := ins[ip+1:]
			numElements := int(code.ReadUint16(operand))
			vm.currentFrame().ip += 2

			// construct a new array using elements on the stack, buildArray needs a starting index and non-inclusive ending index
			array := vm.buildArray(vm.sp-numElements, vm.sp)
			// sp (stack-pointer) needs to be updated after using the elements to build the new array
			vm.sp = vm.sp - numElements
			// push the new array onto the stack
			err := vm.push(array)

			if err != nil {
				return err
			}

		// Execute OpHash instruction, it should construct a new hash map and push it on to the stack,
		// using the values (if any) that were previously loaded
		case code.OpHash:
			// derive the number of elements to pull from the operand
			operand := ins[ip+1:]
			numElements := int(code.ReadUint16(operand))
			vm.currentFrame().ip += 2

			// construct a new map using elements on the stack, buildHash needs a starting index and non-inclusive ending index
			hash, err := vm.buildHash(vm.sp-numElements, vm.sp)
			if err != nil {
				return err
			}
			// sp (stack-pointer) needs to be updated after using the elements to build the new array
			vm.sp = vm.sp - numElements

			// push the new hash onto the stack
			err = vm.push(hash)
			if err != nil {
				return err
			}

		// Execute OpIndex instruction, it should pop the two elements before the sp, the index object and
		// then the expression object to be indexed. Finally it should push the result of the index operation onto the stack.
		case code.OpIndex:
			index := vm.pop()
			left := vm.pop()

			err := vm.executeIndexExpression(left, index)
			if err != nil {
				return err
			}

		// Execute OpCall instruction, it should grab the current compiled function object before the stack pointer
		// and create a new frame for it. On the next iteration, the main while loop will enter this frame and execute its instructions
		case code.OpCall:
			// get the number of arguments expected by the function. we need them to effectively find the function constant on the stack
			operand := ins[ip+1]
			numArgs := int(operand)
			vm.currentFrame().ip += 1
			// execute the function
			err := vm.executeCall(int(numArgs))
			if err != nil {
				return err
			}

		// Execute OpReturnValue instruction. It should pop the returnValue sitting before the stack pointer and exit
		// the inner-execution context accordingly.
		case code.OpReturnValue:
			// pop the return value object sitting before sp and adjust sp
			returnValue := vm.pop()
			// pop the frame so the loop can leave this inner execution context
			frame := vm.popFrame()
			// the frame.basePointer is the index where the compiledFunctions work(the "hole" and all values produced in the function) starts.
			// that means frame.basePointer - 1 should be where the compiledFunction constant is on the stack. Upon successful execution of the call-expression,
			// we need to replace the function constant with the actual returnValue. Thus the stack-pointer (sp) needs to be updated to
			// apply this change correctly and push the returnValue to the right position on the stack.
			vm.sp = frame.basePointer - 1
			err := vm.push(returnValue)
			if err != nil {
				return err
			}

		// Execute OpReturn instruction. It should just push a Null value to the stack for the function.
		case code.OpReturn:
			frame := vm.popFrame()
			vm.sp = frame.basePointer - 1

			err := vm.push(Null)
			if err != nil {
				return err
			}

		// Execute the OpNull instructin. Simply push the Null constant on to the stack
		case code.OpNull:
			err := vm.push(Null)
			if err != nil {
				return err
			}

		// OpPop has no operands and simply pops an element from the stack
		case code.OpPop:
			// EXECUTE: pop the element before the stack pointer
			vm.pop()
		}
	}

	return nil
}

// isTruthy simply asserts the type of the provided object
// and returns whether whether its value is truthy or falsey
func isTruthy(obj object.Object) bool {
	switch obj := obj.(type) {
	case *object.Boolean:
		return obj.Value
	case *object.Null:
		return false
	default:
		return true
	}
}

// push validates the stack size and adds the provided object (o) to the
// next available slot in the stack, finally it preps the stackpointer (sp),
// incrementing it to designate the next slot to be allocated
func (vm *VM) push(o object.Object) error {
	if vm.sp >= StackSize {
		return fmt.Errorf("stack overflow")
	}

	vm.stack[vm.sp] = o
	vm.sp++

	return nil
}

// LastPoppedStackElem helps identify the last element that was popped from the stack as the VM executes through it.
// If a stack had two elements [a, b], sp would be at index 2. If the vm pops an element,
// it would pop the element at [sp-1], so index 1, and then sp is moved to index 1.
// Leaving b to be the last popped stack element.
func (vm *VM) LastPoppedStackElem() object.Object {
	return vm.stack[vm.sp]
}

// pop simply grabs the constant sittng 1 position before the stackpointer,
// it then decrements the stack pointer to be aware of the updated position,
// leaving that slot to be eventually overwritten with a new constant
func (vm *VM) pop() object.Object {
	o := vm.stack[vm.sp-1]
	vm.sp--
	return o
}

// executeBinaryOperation pops the two constants before the stack-pointer
// and validates what type of binary operation to run with them. If the combination
// of types do not have a valid operation an error is returned.
func (vm *VM) executeBinaryOperation(op code.Opcode) error {
	right := vm.pop()
	left := vm.pop()

	leftType := left.Type()
	rightType := right.Type()

	switch {
	case leftType == object.INTEGER_OBJ && rightType == object.INTEGER_OBJ:
		return vm.executeBinaryIntegerOperation(op, left, right)
	case leftType == object.STRING_OBJ && rightType == object.STRING_OBJ:
		return vm.executeBinaryStringOperation(op, left, right)
	default:
		return fmt.Errorf("unsupported types for binary operation: %s, %s",
			leftType, rightType)
	}
}

// executeBinaryIntegerOperation will perform an arithmetic operation
// with the provided operator and objects. If the operation is successful,
// the new evaluated object is pushed on to the stack.
func (vm *VM) executeBinaryIntegerOperation(
	op code.Opcode,
	left, right object.Object,
) error {
	// assert the Objects to grab their integer value
	leftValue := left.(*object.Integer).Value
	rightValue := right.(*object.Integer).Value

	var result int64
	// handle arithmetic operation
	switch op {
	case code.OpAdd:
		result = leftValue + rightValue
	case code.OpSub:
		result = leftValue - rightValue
	case code.OpMul:
		result = leftValue * rightValue
	case code.OpDiv:
		result = leftValue / rightValue
	default:
		return fmt.Errorf("unknown integer operation: %d", op)
	}

	// push the Object to the stack
	return vm.push(&object.Integer{Value: result})
}

// executeBinaryStringOperation will assert that the provided Objects are
// string literals, it will concatenate them and push the new string to the stack.
// If the Opcode is invalid (not OpAdd) it will return an error.
func (vm *VM) executeBinaryStringOperation(
	op code.Opcode,
	left, right object.Object,
) error {
	if op != code.OpAdd {
		return fmt.Errorf("unknown string operation: %d", op)
	}

	// assert the Objects to grab their string values
	leftValue := left.(*object.String).Value
	rightValue := right.(*object.String).Value

	// push the Object to the stack
	return vm.push(&object.String{Value: fmt.Sprint(leftValue, rightValue)})
}

// executeComparison will compare the two constants directly before the stack-pointer
// and then push the result on to the stack. It validates the type of the two constants (object.Object)
// to determine what comparison helper to run this pattern.
func (vm *VM) executeComparison(op code.Opcode) error {
	right := vm.pop()
	left := vm.pop()

	leftType := left.Type()
	rightType := right.Type()

	if leftType == object.INTEGER_OBJ && rightType == object.INTEGER_OBJ {
		return vm.executeIntegerComparison(op, left, right)
	}

	// compare of pointer-addresses. For boolean objects,
	// right and left are holding the constants TRUE and FALSE listed, and we
	// are reusing those constants so we can compare their pointer-addresses.
	switch op {
	case code.OpEqual:
		return vm.push(nativeBoolToBooleanObject(right == left))
	case code.OpNotEqual:
		return vm.push(nativeBoolToBooleanObject(right != left))
	default:
		return fmt.Errorf("unknown operator: %d, (%s %s)",
			op, leftType, rightType)
	}
}

// executeIntegerComparison is the helper to compare two integer constants. It asserts
// the two constants as *object.Integers and compares their values. With the result
// of the comparison, it constructs a Boolean Object and pushes it to the stack.
func (vm *VM) executeIntegerComparison(
	op code.Opcode,
	left, right object.Object,
) error {
	leftValue := left.(*object.Integer).Value
	rightValue := right.(*object.Integer).Value

	var result *object.Boolean
	switch op {
	case code.OpGreaterThan:
		result = nativeBoolToBooleanObject(leftValue > rightValue)
	case code.OpEqual:
		result = nativeBoolToBooleanObject(leftValue == rightValue)
	case code.OpNotEqual:
		result = nativeBoolToBooleanObject(leftValue != rightValue)
	default:
		return fmt.Errorf("unknown operator: %d", op)
	}

	return vm.push(result)

}

// nativeBoolToBooleanObject simply converts a traditional boolean
// to an *object.Boolean
func nativeBoolToBooleanObject(b bool) *object.Boolean {
	if b {
		return True
	}
	return False
}

// executeBangOperator handles the execution of an instruction for a OpBang Opcode.
// It pops the constant before the stack pointer and negates it with the "!" prefix.
// If the constant is truthy we will push False to the stack. If the constant is falsey
// we will push True to the stack.
func (vm *VM) executeBangOperator() error {
	operand := vm.pop()

	switch operand {
	case True:
		return vm.push(False)
	case False:
		return vm.push(True)
	case Null:
		return vm.push(True)
	default:
		return vm.push(False)
	}
}

// executeMinusOperator handles the execution of an isntruction for an OpMinus Opcode.
// It pops the constant before the stack pointer and negates it with the "-" prefix.
// It will construct a new Integer Object, with its value inversed and push that to the stack.
func (vm *VM) executeMinusOperator() error {
	right := vm.pop()

	if right.Type() != object.INTEGER_OBJ {
		return fmt.Errorf("unsupported type for negation: %s", right.Type())
	}

	rightValue := right.(*object.Integer).Value

	return vm.push(&object.Integer{Value: -rightValue})
}

// buildArray constructs a new Object.Array using existing elements
// on the stack. With a given startIndex and endIndex, it will construct
// an array using all elements from the startIndex up until the endIndex (not inclusive).
func (vm *VM) buildArray(startIndex, endIndex int) object.Object {
	elements := make([]object.Object, endIndex-startIndex)
	for i := startIndex; i < endIndex; i++ {
		elements[i-startIndex] = vm.stack[i]
	}

	return &object.Array{Elements: elements}
}

// buildHash constructs a new Object.hash using existing elements
// on the stack. With a given startIndex and endIndex, it will construct a hash
// using all elements from the startIndex up until the endIndex (not inclusive).
func (vm *VM) buildHash(
	startIndex, endIndex int,
) (object.Object, error) {
	hashedPairs := make(map[object.HashKey]object.HashPair)

	for i := startIndex; i < endIndex; i += 2 {
		// build hashPair
		key := vm.stack[i]
		value := vm.stack[i+1]
		pair := object.HashPair{Key: key, Value: value}

		// build hashKey
		hashKey, ok := key.(object.Hashable)
		if !ok {
			return nil, fmt.Errorf("unusable as hash key: %s", key.Type())
		}

		// assign new key value pair to hash map
		hashedPairs[hashKey.HashKey()] = pair
	}

	return &object.Hash{Pairs: hashedPairs}, nil
}

// executeIndexExpression performs an index operation with the provided arguments.
// Depending on the type of the arguments, it will delegate execute to the
// matching helper method.
func (vm *VM) executeIndexExpression(left, index object.Object) error {
	switch {
	case left.Type() == object.ARRAY_OBJ && index.Type() == object.INTEGER_OBJ:
		return vm.executeArrayIndex(left, index)
	case left.Type() == object.HASH_OBJ:
		return vm.executeHashIndex(left, index)
	default:
		return fmt.Errorf("index operator not supported: %s", left.Type())
	}
}

// executeArrayIndex is the helper method that performs an index operation
// on an array object and pushes the result to the stack
func (vm *VM) executeArrayIndex(left, index object.Object) error {
	arrayObject := left.(*object.Array)
	i := index.(*object.Integer).Value
	max := int64(len(arrayObject.Elements) - 1)

	if i < 0 || i > max {
		return vm.push(Null)
	}

	return vm.push(arrayObject.Elements[i])
}

// executeHashIndex is the helper method that performs an index operation
// on a hash object and pushes the result to the stack
func (vm *VM) executeHashIndex(hash, index object.Object) error {
	hashObject := hash.(*object.Hash)

	key, ok := index.(object.Hashable)
	if !ok {
		return fmt.Errorf("unusable as hash key: %s", index.Type())
	}

	pair, ok := hashObject.Pairs[key.HashKey()]
	if !ok {
		return vm.push(Null)
	}

	return vm.push(pair.Value)
}

// executeCall is invoked when the VM executes the OpCall expression. When a function is called,
// we want to grab it from the stack and apply the helper method that it matches with.
func (vm *VM) executeCall(numArgs int) error {
	// grab the function object from the stack and determine how to call it
	callee := vm.stack[vm.sp-1-numArgs]
	switch callee := callee.(type) {
	case *object.CompiledFunction:
		return vm.callFunction(callee, numArgs)
	case *object.Builtin:
		return vm.callBuiltin(callee, numArgs)
	default:
		return fmt.Errorf("calling non-function and non-built-in")
	}
}

// callFunction creates a new frame for the calling function and updates the stack-pointer accordingly
// so the VM can execute the function.
func (vm *VM) callFunction(fn *object.CompiledFunction, numArgs int) error {
	if numArgs != fn.NumParameters {
		return fmt.Errorf("wrong number of arguments: want=%d, got=%d",
			fn.NumParameters, numArgs)
	}

	basePointer := vm.sp - numArgs
	// create a new frame for this function, we need to initialize the basePointer so
	// it starts directly after the index of the function - being the start of its local-bindings.
	frame := NewFrame(fn, basePointer)
	vm.pushFrame(frame)
	// the stack pointer is `increased` to allocate space ("the hole") for the local-bindings and any new values
	// generated in the function will start at the updated stack pointer (above the "hole").
	vm.sp = frame.basePointer + fn.NumLocals
	return nil
}

// callBuiltin executes the builtin function and pushes the return value onto the stack
func (vm *VM) callBuiltin(builtin *object.Builtin, numArgs int) error {
	// grab the arguments for this function on the stack
	args := vm.stack[vm.sp-numArgs : vm.sp]
	// execute the builtin function
	result := builtin.Fn(args...)
	// set sp to the position of the built-in function on the stack
	vm.sp = vm.sp - numArgs - 1
	// replace function with return value
	if result != nil {
		vm.push(result)
	} else {
		vm.push(Null)
	}

	return nil
}

// NewWithGlobalStore keeps global state in the REPL so the VM can execute
// with the byteode and global store from a previous compilation.
func NewWithGlobalStore(bytecode *compiler.Bytecode, s []object.Object) *VM {
	vm := New(bytecode)
	vm.globals = s
	return vm
}
