package evaluator

import (
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
		return evalStatements(node.Statements)
	case *ast.ExpressionStatement:
		// recursively calls itself to evaluate the entire expression statement
		return Eval(node.Expression)
	case *ast.BlockStatement:
		// evaluate all statements in the BlockStatement
		return evalStatements(node.Statements)

	// Expressions
	case *ast.PrefixExpression:
		// Evaluate its operand and then use the result with the operator
		right := Eval(node.Right)
		return evalPrefixExpression(node.Operator, right)
	case *ast.InfixExpression:
		// evaluate the left and right operands and then use the results with the operator
		left := Eval(node.Left)
		right := Eval(node.Right)
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

// evalStatements accepts a slice of ast.Statements
// and evaluates them, constructing an object.Object for every
// evaluated ast.Node it encounters
func evalStatements(statements []ast.Statement) object.Object {
	var result object.Object
	for _, stmt := range statements {
		result = Eval(stmt)
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
		return NULL
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
	default:
		return NULL
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
		return NULL
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
		return NULL
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
