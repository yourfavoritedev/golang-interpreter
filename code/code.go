package code

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

// Instructions is used to encapsulate many Instruction(s). A single Instruction
// consists of an opcode and an optional number of operands, which
// is effectively a []byte. We defined Instructions, plural for simplicity to
// work with a series of Instruction
type Instructions []byte

// String builds all the instructions's bytes into human-readable text
// For a fully decoded instruction, we can expect String to build each one with
// the position of the first byte that starts the instruction, the Opcode name and its operands.
func (ins Instructions) String() string {
	var out bytes.Buffer

	i := 0
	// iterate through all instruction bytes
	for i < len(ins) {
		// grab opcode definition
		def, err := Lookup(ins[i])
		if err != nil {
			fmt.Fprintf(&out, "ERROR: %s\n", err)
			continue
		}

		// read operands of opcode
		operands, read := ReadOperands(def, ins[i+1:])

		// build string to output with the decoded instruction
		fmt.Fprintf(&out, "%04d %s\n", i, ins.fmtInstruction(def, operands))

		// prep reader for next instruction
		i += 1 + read
	}

	return out.String()
}

// fmtInstruction builds a string that comprises the Opcode's human readable name
// and the provided operands. First it asserts that provided operands and the
// Opcode's operandWidths are the same length. Then it evaluates the operand count
// to determine how to to build the string.
func (ins Instructions) fmtInstruction(def *Definition, operands []int) string {
	operandCount := len(def.OperandWidths)

	if len(operands) != operandCount {
		return fmt.Sprintf("ERROR: operand len %d does not match defined %d\n",
			len(operands), operandCount)
	}

	switch operandCount {
	case 0:
		return def.Name
	case 1:
		return fmt.Sprintf("%s %d", def.Name, operands[0])
	case 2:
		return fmt.Sprintf("%s %d %d", def.Name, operands[0], operands[1])
	}

	return fmt.Sprintf("ERROR: unhandled operandCount for %s\n", def.Name)
}

// Opcode is used as the first byte in an instruction.
// An Opcode specifies a unique instruction for the VM to execute.
// ie: pushing something onto the stack
type Opcode byte

// Opcodes, when defined, will have ever increasing byte values. (+1 from the previous definition)
// The value is not relevant to us, they only need to be distinct from
// each other and fit in one byte. When the VM executes a specific Op like OpConstant,
// it will decode the corresponding operand, getting back the index
// the constant (the evaluted expression, object.Object) in the constants pool and push it to the stack.
const (
	OpConstant Opcode = iota
	OpAdd
	OpPop
	OpSub
	OpMul
	OpDiv
	OpTrue
	OpFalse
	OpEqual
	OpNotEqual
	OpGreaterThan
	OpMinus
	OpBang
	OpJumpNotTruthy
	OpJump
	OpNull
	OpGetGlobal
	OpSetGlobal
	OpArray
	OpHash
	OpIndex
	OpCall
	OpReturnValue
	OpReturn
	OpSetLocal
	OpGetLocal
	OpGetBuiltin
	OpClosure
	OpGetFree
	OpCurrentClosure
)

// Definition helps us understand Opcode defintions. A Definition
// gives more insight on an Opcode's human-readable name (Name) and its operands.
// OperandWidths records the unique bytewidth that each operand may have
type Definition struct {
	Name          string
	OperandWidths []int
}

var definitions = map[Opcode]*Definition{
	OpConstant: {"OpConstant", []int{2}}, /**OpConstant has one two-byte operand. The operand refers to the index (position) of the constant in the constants pool.
	Its operand is simply an identifier, it is the position of a constant (evaluated object) in the constant pool.
	If an operand is 10000, it refers to a constant at the 10000 position of the constant pool.
	An Opcode can have a variable number of operands and each operand can have different byte-sizes
	which is what OperandWidths is trying to represent. For instance, the OpConstant is an Opcode
	that has 1 operand, and that operand will be two-byte wide (2). A two-byte wide operand can have a maximum value of
	65535, which means for OpConstants, the operand can be a max value of 65535, the identifier for the 65535 positioned constant.
	With that connection, the operandWidth essentially sets the upper-boundary value for the operand, it puts a ceiling on
	the maximum identifier-position an operand can hold for a constant in the constant pool. */
	OpAdd:           {"OpAdd", []int{}},            //OpAdd does not have any operands
	OpPop:           {"OpPop", []int{}},            //OpPop does not have any operands
	OpSub:           {"OpSub", []int{}},            //OpSub does not have any operands
	OpMul:           {"OpMul", []int{}},            //OpMul does not have any operands
	OpDiv:           {"OpDiv", []int{}},            //OpDiv does not have any operands
	OpTrue:          {"OpTrue", []int{}},           //OpTrue does not have any operands
	OpFalse:         {"OpFalse", []int{}},          //OpFalse does not have any operands
	OpEqual:         {"OpEqual", []int{}},          //OpEqual does not have any operands
	OpNotEqual:      {"OpNotEqual", []int{}},       //OpNotEqual does not have any operands
	OpGreaterThan:   {"OpGreaterThan", []int{}},    //OpGreaterThan does not have any operands
	OpMinus:         {"OpMinus", []int{}},          //OpMinus does not have any operands
	OpBang:          {"OpBang", []int{}},           //OpBang does not have any operands
	OpJumpNotTruthy: {"OpJumpNotTruthy", []int{2}}, //OpJumpNotTruthy has one two-byte operand. The operand refers to where in the instructions to jump to.
	OpJump:          {"OpJump", []int{2}},          //OpJump has one two-byte operand. The operand refers to where in the instructions to jump to.
	OpNull:          {"OpNull", []int{}},           //OpNull does not have any operands
	OpGetGlobal:     {"OpGetGlobal", []int{2}},     //OpGetGlobal has one two-byte operand. The operand refers to the unique index of a global binding.
	OpSetGlobal:     {"OpSetGlobal", []int{2}},     //OpSetGlobal has one two-byte operand. The operand refers to the unique index of a global binding.
	OpArray:         {"OpArray", []int{2}},         //OpArray has one two-byte operand. The operand is the number of elements in an array literal.
	OpHash:          {"OpHash", []int{2}},          //OpHash has one two-byte opereand. The operand is the combined number of keys and values in the hash literal.
	OpIndex:         {"OpIndex", []int{}},          //OpIndex does not have any operands
	OpCall:          {"OpCall", []int{1}},          //OpCall has one one-byte operand. The operand refers to the number of arguments of the calling function.
	OpReturnValue:   {"OpReturnValue", []int{}},    //OpReturnValue does not have any operands
	OpReturn:        {"OpReturn", []int{}},         //OpReturn does not have any operands
	OpSetLocal:      {"OpSetLocal", []int{1}},      //OpSetLocal has one one-byte operand. The operand refers to the unique index of a local binding
	OpGetLocal:      {"OpGetLocal", []int{1}},      //OpGetLocal has one one-byte operand. The operand refers to the unique index of a local binding
	OpGetBuiltin:    {"OpGetBuiltin", []int{1}},    //OpGetBuiltin has one one-byte operand. The operand refers to the unique index of the BuiltIn function in object.Builtins.
	OpClosure:       {"OpClosure", []int{2, 1}},    /**OpClosure has two operands. The first operand is two-bytes wide and refers to the
	index of the object.CompiledFunction in the constants pool. The second operand is one-byte wide and specifies how many free variables sit on the stack and need to
	be transferred to the about-to-be-created closure **/
	OpGetFree:        {"OpGetFree", []int{1}},       //OpGetFree has one one-byte operand. The operand refers to the unique index of a free variable.
	OpCurrentClosure: {"OpCurrentClosure", []int{}}, //OpCurrentClosure does not have any operands
}

// Lookup simply finds the definition of the provided op (Opcode)
func Lookup(op byte) (*Definition, error) {
	def, ok := definitions[Opcode(op)]
	if !ok {
		return nil, fmt.Errorf("opcode %d undefined", op)
	}
	return def, nil
}

// Make creates a single bytecode instruction. The instruction
// consists of an Opcode and an optional number of operands.
func Make(op Opcode, operands ...int) []byte {
	// verify that the Opcode definition exists
	def, ok := definitions[op]
	if !ok {
		return []byte{}
	}

	// set initial length at 1 for the first byte, the Opcode.
	instructionLen := 1
	// and then for each operand, we want to increment tnstructionLen by its operand width
	for _, w := range def.OperandWidths {
		instructionLen += w
	}

	// initialize the instruction
	instruction := make([]byte, instructionLen)
	instruction[0] = byte(op)

	offset := 1
	// Iterate over the provided operands. In theory we should only have
	// len(operands) with matching len(def.OperandWidths)
	for i, o := range operands {
		// find the operandWidth to determine how to encode
		// the argument provided operand
		width := def.OperandWidths[i]
		switch width {
		// for two-byte sized operands, encode o with BigEndian
		case 2:
			binary.BigEndian.PutUint16(instruction[offset:], uint16(o))
		// for one-byte size operands, simply set the instruction byte at the offset position to be the operand byte
		case 1:
			instruction[offset] = byte(o)
		}
		offset += width
	}

	return instruction
}

// ReadOperands decodes the operands for the given instruction
// It returns the decoded operands and tells us how many bytes it read to do that.
func ReadOperands(def *Definition, ins Instructions) ([]int, int) {
	// iniitalize slice with the expected number of operands
	operands := make([]int, len(def.OperandWidths))
	// offset has two purposes, 1 as a running number of total bytes we read
	// and 2 as the number of bytes to offset after successfully reading an operand
	offset := 0

	for i, width := range def.OperandWidths {
		switch width {
		// execute when the operandWidth is size two (two-byte width)
		case 2:
			// decode the two-byte width operand in the given instruction
			operands[i] = int(ReadUint16(ins[offset:]))
			// decode the one-byte width operand in the given instruction
		case 1:
			operands[i] = int(ins[offset])
		}
		// prepare offset for the next byte to be read, if any
		offset += width
	}

	return operands, offset
}

// ReadUint16 helps decode the operand correctly. Typically, when we
// call this function to decode an operand, we pass the entire
// instructions ([]byte) starting with the operand and then everything else.
// BigEndian.Uint16 will only return the first decodable int in the []byte
// which works perfectly to decode operand.
func ReadUint16(ins Instructions) uint16 {
	return binary.BigEndian.Uint16(ins)
}
