package evaluator

import (
	"fmt"

	"github.com/yourfavoritedev/golang-interpreter/ast"
	"github.com/yourfavoritedev/golang-interpreter/object"
)

var (
	// null can be referenced instead of allocating a new object each time we evaluate a node.
	NULL = &object.Null{}
	// there will only ever be two variations of object.Booleans,
	// it is more beneficial to reference them instead of allocating new ones.
	TRUE  = &object.Boolean{Value: true}
	FALSE = &object.Boolean{Value: false}
)

// Eval accepts an AST Node and determines the best way to evaluate it.
// We store the evaluated value in an Object, which can be later referenced.
// Eval is expected to run recursively, following the "tree-walking pattern".
// It should traverse the tree (AST), starting with the top-level *ast.Program,
// going into all its statements and evaluating each one. It traverses each Statement,
// evaluating its own nodes. This will lead to evaluating the actual Expression Nodes,
// where the Value of the node can be consumed and stored in an Object.
func Eval(node ast.Node) object.Object {
	switch node := node.(type) {
	// Statements
	case *ast.Program:
		// evaluates all statements in the program
		return evalProgram(node)
	case *ast.ExpressionStatement:
		// recursively calls itself to evaluate the entire expression statement
		return Eval(node.Expression)
	case *ast.BlockStatement:
		// evaluate all statements in the BlockStatement
		return evalBlockStatement(node)
	case *ast.ReturnStatement:
		// evaluate the expression associated with the return statement and then wrap the value
		val := Eval(node.ReturnValue)
		// if there is an error, prevent it from being passed around
		// and bubbling up far away from their origin
		if isError(val) {
			return val
		}
		return &object.ReturnValue{Value: val}

	// Expressions
	case *ast.PrefixExpression:
		// Evaluate its operand and then use the result with the operator
		right := Eval(node.Right)
		// if there is an error, prevent it from being passed around
		// and bubbling up far away from their origin
		if isError(right) {
			return right
		}
		return evalPrefixExpression(node.Operator, right)
	case *ast.InfixExpression:
		// evaluate the left and right operands and then use the results with the operator
		left := Eval(node.Left)
		// return error if encountered when evaluating left node
		if isError(left) {
			return left
		}

		right := Eval(node.Right)
		// return error if encountered when evaluating right node
		if isError(right) {
			return right
		}
		return evalInfixExpression(node.Operator, left, right)
	case *ast.IfExpression:
		// evaluate if expression
		return evalIfExpression(node)
	case *ast.IntegerLiteral:
		// Simply evaluates an integer literal
		return &object.Integer{Value: node.Value}
	case *ast.Boolean:
		// Simply evaluates a Boolean
		return nativeBoolToBooleanObject(node.Value)
	}

	return nil
}

// evalProgram accepts an ast.Program and evaluates its
// statements, constructing an object.Object for every
// evaluated ast.Node it encounters
func evalProgram(program *ast.Program) object.Object {
	var result object.Object

	for _, statement := range program.Statements {
		result = Eval(statement)

		switch result := result.(type) {
		// if we encounter a ReturnValue after successfully evaluating a statement,
		// then we should return it immediately and early-exit
		case *object.ReturnValue:
			return result.Value
		// return the Error immediately
		case *object.Error:
			return result
		}
	}

	return result
}

// evalBlockStatement evaluates a block statement and identifies
// if we should immediately return the evaluated value if
// it is of type object.RETURN_VALUE_OBJ
func evalBlockStatement(block *ast.BlockStatement) object.Object {
	var result object.Object

	for _, statement := range block.Statements {
		result = Eval(statement)

		if result != nil {
			rt := result.Type()
			// should return the Object and early-exit if the statement has evalated to an object of type
			// RETURN_VALUE_OBJ or ERROR_OBJ, these are objects that should stop the evaluation.
			// This happens after we evaluate a return statement or encounter an error
			if rt == object.RETURN_VALUE_OBJ || rt == object.ERROR_OBJ {
				return result
			}
		}
	}

	return result
}

// evalPrefixExpression will construct a new Object for an evaluated prefix expression.
// It validates the given operator to determine the best evaluating
// function to use for the scenario.
func evalPrefixExpression(operator string, right object.Object) object.Object {
	switch operator {
	case "!":
		return evalBangOperatorExpression(right)
	case "-":
		return evalMinusPrefixOperatorExpression(right)
	default:
		return newError("unknown operator: %s%s", operator, right.Type())
	}
}

// evalInfixExpression will construct a new Object for an evaluated infix expression.
// It validates the given arguments to determine the best evaluating
// function to use for the scenario.
func evalInfixExpression(
	operator string,
	left, right object.Object,
) object.Object {
	switch {
	// evaluate the infix expression where both left and right nodes are operating on integers
	case left.Type() == object.INTEGER_OBJ && right.Type() == object.INTEGER_OBJ:
		return evalIntegerInfixExpression(operator, left, right)
	// When the nodes are not integers then they are object.Booleans.
	// We can do a pointer comparison here to check for equality between booleans.
	// This is possible because the nodes here have already been evaluated
	// to pointers of the object.Booleans we defined as TRUE and FALSE. So if left
	// is &Object.Boolean{Value: true}, its the same TRUE we defined, the same memory address.
	case operator == "==":
		return nativeBoolToBooleanObject(left == right)
	case operator == "!=":
		return nativeBoolToBooleanObject(left != right)
	// infix expression is trying to perform an operation of mismatched types,
	// this should return an error
	case left.Type() != right.Type():
		return newError("type mismatch: %s %s %s",
			left.Type(), operator, right.Type())
	default:
		return newError("unknown operator: %s %s %s",
			left.Type(), operator, right.Type())
	}
}

// evalIntegerInfixExpression will construct a new Object for an
// infix expression where both nodes are of type object.Integer.
// The operator will help determine what type of Object to construct.
// Upon evaluation, the Object Value should be the result of
// the performed operation between the left and right nodes.
func evalIntegerInfixExpression(
	operator string,
	left, right object.Object,
) object.Object {
	leftValue := left.(*object.Integer).Value
	rightValue := right.(*object.Integer).Value

	switch operator {
	case "+":
		return &object.Integer{Value: leftValue + rightValue}
	case "-":
		return &object.Integer{Value: leftValue - rightValue}
	case "*":
		return &object.Integer{Value: leftValue * rightValue}
	case "/":
		return &object.Integer{Value: leftValue / rightValue}
	case "<":
		return nativeBoolToBooleanObject(leftValue < rightValue)
	case ">":
		return nativeBoolToBooleanObject(leftValue > rightValue)
	case "==":
		return nativeBoolToBooleanObject(leftValue == rightValue)
	case "!=":
		return nativeBoolToBooleanObject(leftValue != rightValue)
	default:
		return newError("unknown operator: %s %s %s",
			left.Type(), operator, right.Type())
	}
}

// evalBangOperatorExpression will return the inverse object.Boolean
// for the provided Object. object.Integers are treated as truthy values,
// so their inverse should be falsey.
func evalBangOperatorExpression(right object.Object) object.Object {
	switch right {
	case TRUE:
		return FALSE
	case FALSE:
		return TRUE
	case NULL:
		return TRUE
	default:
		return FALSE
	}
}

// evalMinusPrefixOperatorExpression construct a new object.Integer with
// a Value that is oppositely charged to the provided object.Integer, right.
// 5 -> -5 and -5 -> 5
func evalMinusPrefixOperatorExpression(right object.Object) object.Object {
	// validate that an integer is provided
	if right.Type() != object.INTEGER_OBJ {
		return newError("unknown operator: -%s", right.Type())
	}

	value := right.(*object.Integer).Value
	return &object.Integer{Value: -value}
}

// nativeBoolToBooleanObject determines which object.Boolean struct
// to return depending on the provided input
func nativeBoolToBooleanObject(input bool) *object.Boolean {
	if input {
		return TRUE
	}
	return FALSE
}

// evalIfExpression constructs a new Object by evaluating either
// the if expression's Consequence or Alternative.
func evalIfExpression(ie *ast.IfExpression) object.Object {
	// evaluate the condition and determine whether it is truthy or falsey
	condition := Eval(ie.Condition)
	// if there is an error, prevent it from being passed around
	// and bubbling up far away from its origin
	if isError(condition) {
		return condition
	}

	if isTruthy(condition) {
		return Eval(ie.Consequence)
	} else if ie.Alternative != nil {
		return Eval(ie.Alternative)
	} else {
		return NULL
	}
}

func isTruthy(obj object.Object) bool {
	switch obj {
	case NULL:
		return false
	case TRUE:
		return true
	case FALSE:
		return false
	default:
		return true
	}
}

// newError constructs a object.Error with the given format and
// a, which is a variadic slice of error message(s) which can be for any type
func newError(format string, a ...interface{}) *object.Error {
	return &object.Error{Message: fmt.Sprintf(format, a...)}
}

// isError simply validates whether the given object is
// of type object.ERROR_OBJ
func isError(obj object.Object) bool {
	if obj != nil {
		return obj.Type() == object.ERROR_OBJ
	}
	return false
}
