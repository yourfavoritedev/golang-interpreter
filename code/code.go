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
	case 1:
		return fmt.Sprintf("%s %d", def.Name, operands[0])
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
// it will use the iota-generated-value (Opcode) as an index to retrieve
// the constant (the evaluted expression, object.Object) and push it to the stack.
const (
	OpConstant Opcode = iota
)

// Definition helps us understand Opcode defintions. A Definition
// gives more insight on an Opcode's human-readable name (Name) and its operands.
// An operand is simply an identifier, it is the position of a constant (evaluated object) in the constant pool.
// If an operand is 10000, it refers to a constant at the 10000 position of the constant pool.
// An Opcode can have a variable number of operands and each operand can have different byte-sizes
// which is what OperandWidths is trying to represent. For instance, the OpConstant is an Opcode
// that has 1 operand, and that operand will be two-byte wide (2). A two-byte wide operand can have a maximum value of
// 65535, which means for OpConstants, the operand can be a max value of 65535, the identifier for the 65535 positioned constant.
// With that connection, the operandWidth essentially sets the upper-boundary value for the operand, it puts a ceiling on
// the maximum identifier-position an operand can hold for a constant in the constant pool.
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
		}
		// prepare offset for the next byte to be read, if any
		offset += width
	}

	return operands, offset
}

// ReadUint16 simply converts the Instructions bytes into a readable uint16
func ReadUint16(ins Instructions) uint16 {
	return binary.BigEndian.Uint16(ins)
}
