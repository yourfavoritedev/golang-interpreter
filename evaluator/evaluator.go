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
func Eval(node ast.Node, env *object.Environment) object.Object {
	switch node := node.(type) {
	// Statements
	case *ast.Program:
		// evaluates all statements in the program
		return evalProgram(node, env)
	case *ast.ExpressionStatement:
		// recursively calls itself to evaluate the entire expression statement
		return Eval(node.Expression, env)
	case *ast.BlockStatement:
		// evaluate all statements in the BlockStatement
		return evalBlockStatement(node, env)
	case *ast.ReturnStatement:
		// evaluate the expression associated with the return statement and then wrap the value
		val := Eval(node.ReturnValue, env)
		// if there is an error, prevent it from being passed around
		// and bubbling up far away from their origin
		if isError(val) {
			return val
		}
		return &object.ReturnValue{Value: val}
	case *ast.LetStatement:
		// first we need to evaluate the expression of the LetStatement
		val := Eval(node.Value, env)
		if isError(val) {
			return val
		}
		// set the identifier name and the evaluated value to the environment
		env.Set(node.Name.Value, val)

	// Expressions
	case *ast.PrefixExpression:
		// Evaluate its operand and then use the result with the operator
		right := Eval(node.Right, env)
		// if there is an error, prevent it from being passed around
		// and bubbling up far away from their origin
		if isError(right) {
			return right
		}
		return evalPrefixExpression(node.Operator, right)
	case *ast.InfixExpression:
		// evaluate the left and right operands and then use the results with the operator
		left := Eval(node.Left, env)
		// return error if encountered when evaluating left node
		if isError(left) {
			return left
		}

		right := Eval(node.Right, env)
		// return error if encountered when evaluating right node
		if isError(right) {
			return right
		}
		return evalInfixExpression(node.Operator, left, right)
	case *ast.IfExpression:
		// evaluate if expression
		return evalIfExpression(node, env)
	case *ast.IntegerLiteral:
		// Simply evaluates an integer literal
		return &object.Integer{Value: node.Value}
	case *ast.Boolean:
		// Simply evaluates a Boolean
		return nativeBoolToBooleanObject(node.Value)
		// Simply evaluates a string literal
	case *ast.StringLiteral:
		return &object.String{Value: node.Value}
	case *ast.ArrayLiteral:
		// Evaluate the array literal with its elements
		elements := evalExpressions(node.Elements, env)
		// Should stop evaluating as soon as we encounter an error while evaluating the elements
		if len(elements) == 1 && isError(elements[0]) {
			return elements[0]
		}
		return &object.Array{Elements: elements}
	case *ast.IndexExpression:
		// Evaluate the index operator expression. First evaluate the actual array which
		// can take the form of any expression. Then evaluate the index which is also an expression.
		left := Eval(node.Left, env)
		if isError(left) {
			return left
		}
		index := Eval(node.Index, env)
		if isError(index) {
			return index
		}
		return evalIndexExpression(left, index)

	// Identifiers
	case *ast.Identifier:
		// Evaluate the identifier and get back the object with its evaluated value
		return evalIdentifier(node, env)

	// Functions
	case *ast.FunctionLiteral:
		// Evaluate the function literal, store its Parameters
		// and Body nodes in the object.Function for future reference
		// they will be evaluated during function calls
		params := node.Parameters
		body := node.Body
		return &object.Function{Parameters: params, Body: body, Env: env}
	case *ast.CallExpression:
		// Evaluate the call expression, simply getting back the function we want to call,
		// it can be the form of an ast.Identifier or an ast.FunctionLiteral, it still
		// returns an object.Function
		function := Eval(node.Function, env)
		if isError(function) {
			return function
		}

		// Evaluate the arguments of the function and keep track of the produced Object values.
		// Should stop evaluating as soon as we encounter an error
		args := evalExpressions(node.Arguments, env)
		if len(args) == 1 && isError(args[0]) {
			return args[0]
		}

		// call the function!
		return applyFunction(function, args)
	}

	return nil
}

// applyFunction accepts an already evaluated function and evaluated arguments.
// If fn is of type object.Function, it will bind the function and arguments to a new inner environment then evaluate it.
// If fn is type object.Builtin, it will call the built-in function with the given arguments.
func applyFunction(fn object.Object, args []object.Object) object.Object {
	switch fn := fn.(type) {
	case *object.Function:
		// bind function and arguments to a new inner environment
		extendedEnv := extendFunctionEnv(fn, args)
		// evaluate the function body within this extended environemnt
		evaluated := Eval(fn.Body, extendedEnv)
		// unwrap object if its a return value object
		return unwrapReturnValue(evaluated)
	case *object.Builtin:
		// call the built-in function with the evaluated arguments
		return fn.Fn(args...)
	default:
		return newError("not a function: %s", fn.Type())
	}
}

// extendFunctionEnv creates a new inner environment for an object.Function
// It binds the function's parameters and already evaluated arguments to
// the new inner environment. The environment is enclosed by the initial environment (outer)
// of which the function was defined in (Function.Env)
func extendFunctionEnv(
	fn *object.Function,
	args []object.Object,
) *object.Environment {
	// Create inner environment, enclosed by the outer environment that defined the function
	env := object.NewEnclosedEnvironment(fn.Env)

	// set inner environment store with the function's parameters and evaluated arguments
	for paramIdx, param := range fn.Parameters {
		env.Set(param.Value, args[paramIdx])
	}

	return env
}

// unwrapReturnValue asserts if the evaluated object is an object.ReturnValue.
// If it is a ReturnValue object, we need to return the unwrapped value,
// otherwise simply return the already evaluated object.
func unwrapReturnValue(obj object.Object) object.Object {
	// assert that the given object is a ReturnValue, if it is return it
	if returnValue, ok := obj.(*object.ReturnValue); ok {
		return returnValue.Value
	}
	return obj
}

// evalProgram accepts an ast.Program and evaluates its
// statements, constructing an object.Object for every
// evaluated ast.Node it encounters
func evalProgram(program *ast.Program, env *object.Environment) object.Object {
	var result object.Object

	for _, statement := range program.Statements {
		result = Eval(statement, env)

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
func evalBlockStatement(block *ast.BlockStatement, env *object.Environment) object.Object {
	var result object.Object

	for _, statement := range block.Statements {
		result = Eval(statement, env)

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
	// evaluate the infix expression where both left and right nodes are operating on strings
	case left.Type() == object.STRING_OBJ && right.Type() == object.STRING_OBJ:
		return evalStringInfixExpression(operator, left, right)
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

// evalStringInfixExpression validates that a concatentation (+) is
// attempted on two Object.Strings (left) and (right).
// It concatenates the left and right Values to form a new Object.String
func evalStringInfixExpression(
	operator string,
	left, right object.Object,
) object.Object {
	if operator != "+" {
		return newError("unknown operator: %s %s %s",
			left.Type(), operator, right.Type())
	}

	leftVal := left.(*object.String).Value
	rightVal := right.(*object.String).Value
	return &object.String{Value: leftVal + rightVal}
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
func evalIfExpression(ie *ast.IfExpression, env *object.Environment) object.Object {
	// evaluate the condition and determine whether it is truthy or falsey
	condition := Eval(ie.Condition, env)
	// if there is an error, prevent it from being passed around
	// and bubbling up far away from its origin
	if isError(condition) {
		return condition
	}

	if isTruthy(condition) {
		return Eval(ie.Consequence, env)
	} else if ie.Alternative != nil {
		return Eval(ie.Alternative, env)
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

// evalIdentifier verifies if an identifier has been previously associated
// in the environment. Ff an identifier was found, return its mapped object.
// If not found, then check if there is a built-in function with that identifier.
// If there is a built-in function then return it. Otherwise, we've encountered an error,
// we're attempting to evaluate an identifier that has not been introduced to the environment.
func evalIdentifier(node *ast.Identifier, env *object.Environment) object.Object {
	if val, ok := env.Get(node.Value); ok {
		return val
	}

	if builtin, ok := builtins[node.Value]; ok {
		return builtin
	}

	return newError("identifier not found: %s", node.Value)
}

// evalExpressions evaluates the given list of expressions and if no error is encountered
// then it will return the evaluated Objects in their respective argument order.
// However, if an error was encountered, it will only return the Object.Error.
func evalExpressions(
	exps []ast.Expression,
	env *object.Environment,
) []object.Object {
	var result []object.Object

	for _, e := range exps {
		evaluated := Eval(e, env)
		if isError(evaluated) {
			return []object.Object{evaluated}
		}
		result = append(result, evaluated)
	}

	return result
}

// evalIndexExpression evaluates an index operation. If left and index
// meet the necessary conditions, it will return the evaluated value in that array
// at that index. Otherwise, it will return an error for the unsupported index operation.
func evalIndexExpression(left, index object.Object) object.Object {
	switch {
	case left.Type() == object.ARRAY_OBJ && index.Type() == object.INTEGER_OBJ:
		return evalArrayIndexExpression(left, index)
	default:
		return newError("index operator not supported: %s", left.Type())
	}
}

// evalArrayIndexExpression will return the evaluated element in the array (left)
// at the given index.Value. If the index is outside the bounds of the array,
// it will return NULL.
func evalArrayIndexExpression(left, index object.Object) object.Object {
	// assert that left is an object.Array so that we can access its Elements
	array := left.(*object.Array)
	// assert that index is an object.Integer so that we can access its Value
	idx := index.(*object.Integer).Value
	maxIdx := int64(len(array.Elements) - 1)
	if idx > maxIdx || idx < 0 {
		return NULL
	}
	return array.Elements[idx]
}
