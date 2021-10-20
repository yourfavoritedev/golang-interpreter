package code

import (
	"encoding/binary"
	"fmt"
)

// Instructions is used to encapsulate many Instruction(s). A single Instruction
// consists of an opcode and an optional number of operands, which
// is effectively a []byte. We defined Instructions, plural for simplicity to
// work with a series of Instruction
type Instructions []byte

// Opcode is used as the first byte in an instruction.
// An Opcode specifies a unique instruction for the VM to execute.
// ie: pushing something onto the stack
type Opcode byte

// Opcodes, when defined, will have ever increasing byte values. (+1 from the previous definition)
// The value is not relevant to us, they only need to be distinct from
// each other and fit in one byte. When the VM executes a specific Op like OpConstant,
// it will use the iota-generated-value (Opcode) as an index to retrieve
// the constant (the evaluted expression, object.Object) and push it to the stack.
const (
	OpConstant Opcode = iota
)

// Definition helps us understand Opcode defintions. A Definition
// gives more insight on an Opcode's human-readable name (Name) and its operands.
// An Opcode can have a variable number of operands and each operand can have different byte-sizes
// which is what OperandWidths is trying to represent. For instance, the OpConstant is an Opcode
// that has 1 operand, and that operand will be two-byte.
type Definition struct {
	Name          string
	OperandWidths []int
}

var definitions = map[Opcode]*Definition{
	OpConstant: {"OpConstant", []int{2}}, //OpConstant has one two-byte operand
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
	// and then for each operand, we want to increment it by the total operand widths
	for _, w := range def.OperandWidths {
		instructionLen += w
	}

	// initialize the instruction
	instruction := make([]byte, instructionLen)
	instruction[0] = byte(op)

	offset := 1
	// Iterate over the provided operands. Un theory we should only have
	// len(operands) with matching len(def.OperandWidths)
	for i, o := range operands {
		// find the operandWidth from how to encode the argument provided operand (o)
		width := def.OperandWidths[i]
		switch width {
		// for two-byte sized operands, encode o with BigEndian
		case 2:
			binary.BigEndian.PutUint16(instruction[offset:], uint16(o))
		}
		offset += width
	}

	return instruction
}
