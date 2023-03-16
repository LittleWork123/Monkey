package evaluator

import (
	"fmt"
	"interpreter/ast"
	"interpreter/object2"
)

//singleton only has the only TRUE and the only FALSE
var (
	TRUE  = &object2.Boolean{Value: true}
	FALSE = &object2.Boolean{Value: false}
	NULL  = &object2.Null{}
)

var builtins = map[string]object2.Object{
	// builtin function len
	"len": &object2.Builtin{
		Fn: func(args ...object2.Object) object2.Object {
			if len(args) != 1 {
				return newError("wrong number of arguments. got=%d, want=1", len(args))
			}
			switch arg := args[0].(type) {
			case *object2.String:
				return &object2.Integer{Value: int64(len(arg.Value))}
			case *object2.Array:
				return &object2.Integer{Value: int64(len(arg.Elements))}
			default:
				return newError("argument to `len` not supported, got=%s", args[0].Type())
			}
		},
	},
	"first": &object2.Builtin{
		Fn: func(args ...object2.Object) object2.Object {
			if len(args) != 1 {
				return newError("wrong number of arguments. got=%d, want=1", len(args))
			}
			if args[0].Type() != object2.ARRAY_OBJ {
				return newError("argument to `first` must be ARRAY,got=%s", args[0].Type())
			}
			arr := args[0].(*object2.Array)
			if len(arr.Elements) > 0 {
				return arr.Elements[0]
			} else {
				return NULL
			}
		},
	},
	"last": &object2.Builtin{
		Fn: func(args ...object2.Object) object2.Object {
			if len(args) != 1 {
				return newError("wrong number of arguments. got=%d, want=1", len(args))
			}
			if args[0].Type() != object2.ARRAY_OBJ {
				return newError("argument to `last` must be ARRAY, got=%s", args[0].Type())
			}
			arr := args[0].(*object2.Array)
			if len(arr.Elements) > 0 {
				return arr.Elements[len(arr.Elements)-1]
			} else {
				return NULL
			}
		},
	},
	"rest": &object2.Builtin{
		Fn: func(args ...object2.Object) object2.Object {
			if len(args) != 1 {
				return newError("wrong number of arguments. got=%d, want=1", len(args))
			}
			if args[0].Type() != object2.ARRAY_OBJ {
				return newError("argument to `last` must be ARRAY, got=%s", args[0].Type())
			}
			arr := args[0].(*object2.Array)
			length := len(arr.Elements)
			if length > 0 {
				newElements := make([]object2.Object, length-1, length-1)
				copy(newElements, arr.Elements[1:length])
				return &object2.Array{Elements: newElements}
			}
			return NULL
		},
	},
	"push": &object2.Builtin{
		Fn: func(args ...object2.Object) object2.Object {
			if len(args) != 2 {
				return newError("wrong number of arguments. got=%d, want=2", len(args))
			}
			if args[0].Type() != object2.ARRAY_OBJ {
				return newError("argument to `last` must be ARRAY, got=%s", args[0].Type())
			}
			arr := args[0].(*object2.Array)
			length := len(arr.Elements)
			newElements := make([]object2.Object, length+1, length+1)
			copy(newElements, arr.Elements)
			newElements[length] = args[1]
			return &object2.Array{Elements: newElements}
		},
	},
}

func Eval(node ast.Node, env *object2.Environment) object2.Object {
	switch node := node.(type) {
	case *ast.Program:
		return evalProgram(node, env)
	case *ast.ExpressionStatement:
		return Eval(node.Expression, env)
	case *ast.StringLiteral:
		return &object2.String{Value: node.Value}
	case *ast.IntegerLiteral:
		return &object2.Integer{Value: node.Value}
	case *ast.Boolean:
		return nativeBoolToBooleanObject(node.Value)
	case *ast.PrefixExpression:
		right := Eval(node.Right, env)
		if isError(right) {
			return right
		}
		return evalPrefixExpression(node.Operator, right)
	case *ast.InfixExpression:
		left := Eval(node.Left, env)
		if isError(left) {
			return left
		}
		right := Eval(node.Right, env)
		if isError(right) {
			return right
		}
		return evalInfixExpression(node.Operator, left, right)
	case *ast.BlockStatement:
		return evalBlockStatement(node, env)
	case *ast.IfExpression:
		return evalIfExpression(node, env)
	case *ast.ReturnStatement:
		val := Eval(node.ReturnValue, env)
		if isError(val) {
			return val
		}
		return &object2.ReturnValue{Value: val}
	case *ast.LetStatement:
		val := Eval(node.Value, env)
		if isError(val) {
			return val
		}
		env.Set(node.Name.Value, val)
	case *ast.Identifier:
		return evalIdentifier(node, env)
	case *ast.FunctionLiteral:
		params := node.Parameters
		body := node.Body
		// store function
		return &object2.Function{Parameters: params, Env: env, Body: body}
	case *ast.CallExpression:
		function := Eval(node.Function, env)
		if isError(function) {
			return function
		}
		args := evalExpressions(node.Arguments, env)
		if len(args) == 1 && isError(args[0]) {
			return args[0]
		}
		return applyFunction(function, args)
	case *ast.ArrayLiteral:
		elements := evalExpressions(node.Elements, env)
		if len(elements) == 1 && isError(elements[0]) {
			return elements[0]
		}
		return &object2.Array{Elements: elements}
	case *ast.IndexExpression:
		left := Eval(node.Left, env)
		if isError(left) {
			return left
		}
		index := Eval(node.Index, env)
		if isError(index) {
			return index
		}
		return evalIndexExpression(left, index)
	}
	return nil
}

func evalProgram(program *ast.Program, env *object2.Environment) object2.Object {
	var result object2.Object
	for _, statement := range program.Statements {
		result = Eval(statement, env)
		switch result := result.(type) {
		case *object2.ReturnValue:
			return result.Value
		case *object2.Error:
			return result
		}
	}
	return result
}

func nativeBoolToBooleanObject(input bool) *object2.Boolean {
	if input {
		return TRUE
	}
	return FALSE
}

func evalPrefixExpression(operator string, right object2.Object) object2.Object {
	switch operator {
	case "!":
		return evalBangOperatorExpression(right)
	case "-":
		return evalMinusOperatorExpression(right)
	default:
		return newError("unknown operator: %s%s", operator, right.Type())
	}
}

func evalBangOperatorExpression(right object2.Object) object2.Object {
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

func evalMinusOperatorExpression(right object2.Object) object2.Object {
	if right.Type() != object2.INTEGER_OBJ {
		return newError("unknown operator: -%s", right.Type())
	}
	value := right.(*object2.Integer).Value
	return &object2.Integer{Value: -value}
}

func evalInfixExpression(operator string, left object2.Object, right object2.Object) object2.Object {
	switch {
	case left.Type() == object2.STRING_OBJ && right.Type() == object2.STRING_OBJ:
		return evalStringInfixExpression(operator, left, right)
	case left.Type() == object2.INTEGER_OBJ && right.Type() == object2.INTEGER_OBJ:
		return evalIntegerInfixExpression(operator, left, right)
	case operator == "==":
		return nativeBoolToBooleanObject(left == right)
	case operator == "!=":
		return nativeBoolToBooleanObject(left != right)
	case left.Type() != right.Type():
		return newError("type mismatch: %s %s %s",
			left.Type(), operator, right.Type())
	default:
		return newError("unknown operator: %s %s %s",
			left.Type(), operator, right.Type())
	}
}

func evalIntegerInfixExpression(operator string, left object2.Object, right object2.Object) object2.Object {
	leftVal := left.(*object2.Integer).Value
	rightVal := right.(*object2.Integer).Value

	switch operator {
	case "+":
		return &object2.Integer{Value: leftVal + rightVal}
	case "*":
		return &object2.Integer{Value: leftVal * rightVal}
	case "/":
		return &object2.Integer{Value: leftVal / rightVal}
	case "-":
		return &object2.Integer{Value: leftVal - rightVal}
	case "<":
		return nativeBoolToBooleanObject(leftVal < rightVal)
	case ">":
		return nativeBoolToBooleanObject(leftVal > rightVal)
	case "==":
		return nativeBoolToBooleanObject(leftVal == rightVal)
	case "!=":
		return nativeBoolToBooleanObject(leftVal != rightVal)
	default:
		return newError("unknown operator: %s %s %s",
			left.Type(), operator, right.Type())
	}
}

func evalIfExpression(ie *ast.IfExpression, env *object2.Environment) object2.Object {
	condition := Eval(ie.Condition, env)
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

func isTruthy(obj object2.Object) bool {
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

func evalBlockStatement(block *ast.BlockStatement, env *object2.Environment) object2.Object {
	var result object2.Object
	for _, statement := range block.Statements {
		result = Eval(statement, env)
		if result != nil {
			rt := result.Type()
			// if error happened return error currently
			// And if detect return expression return value currently
			//if rt == object2.RETURN_VALUE_OBJ || rt == object2.ERROR_OBJ {
			//	return result
			//}
			if rt == object2.RETURN_VALUE_OBJ {
				return result.(*object2.ReturnValue).Value
			}
			if rt == object2.ERROR_OBJ {
				return result
			}
		}
	}
	return result
}

func newError(format string, a ...interface{}) *object2.Error {
	return &object2.Error{Message: fmt.Sprintf(format, a...)}
}

func isError(obj object2.Object) bool {
	if obj != nil {
		return obj.Type() == object2.ERROR_OBJ
	}
	return false
}

func evalIdentifier(node *ast.Identifier, env *object2.Environment) object2.Object {
	if val, ok := env.Get(node.Value); ok {
		return val
	}

	if builtin, ok := builtins[node.Value]; ok {
		return builtin
	}
	return newError("identifier not found: " + node.Value)
}

func evalExpressions(exps []ast.Expression, env *object2.Environment) []object2.Object {
	var result []object2.Object
	for _, e := range exps {
		evaluated := Eval(e, env)
		if isError(evaluated) {
			return []object2.Object{evaluated}
		}
		result = append(result, evaluated)
	}
	return result
}

func applyFunction(fn object2.Object, args []object2.Object) object2.Object {
	switch fn := fn.(type) {
	case *object2.Function:
		extendedEnv := extendFunctionEnv(fn, args)
		evaluated := Eval(fn.Body, extendedEnv)
		return evaluated
	case *object2.Builtin:
		return fn.Fn(args...)
	default:
		return newError("not a function: %s", fn.Type())
	}
	// Because we unwrapReturnValue in evalBlockStatement function
	// So we do not need to call unwrapReturnValue function
	//return unwrapReturnValue(evaluated)
}

// map identifier to param value
func extendFunctionEnv(fn *object2.Function, args []object2.Object) *object2.Environment {
	env := object2.NewEnclosedEnvironment(fn.Env)
	for paramIds, param := range fn.Parameters {
		env.Set(param.Value, args[paramIds])
	}
	return env
}

func unwrapReturnValue(obj object2.Object) object2.Object {
	if returnValue, ok := obj.(*object2.ReturnValue); ok {
		return returnValue.Value
	}
	return obj
}

func evalStringInfixExpression(operator string, left object2.Object, right object2.Object) object2.Object {
	if operator != "+" {
		return newError("unknown operator: %s %s %s", left.Type(), operator, right.Type())
	}
	leftVal := left.(*object2.String).Value
	rightVal := right.(*object2.String).Value
	return &object2.String{
		Value: leftVal + rightVal,
	}
}

func evalIndexExpression(left object2.Object, index object2.Object) object2.Object {
	switch {
	case left.Type() == object2.ARRAY_OBJ && index.Type() == object2.INTEGER_OBJ:
		return evalArrayIndexExpression(left, index)
	default:
		return newError("index operator not supported: %s", left.Type())
	}
}

// Notion: error handle
func evalArrayIndexExpression(array object2.Object, index object2.Object) object2.Object {
	arrayObject := array.(*object2.Array)
	idx := index.(*object2.Integer).Value
	max := int64(len(arrayObject.Elements) - 1)
	if idx < 0 || idx > max {
		return NULL
	}
	return arrayObject.Elements[idx]
}
