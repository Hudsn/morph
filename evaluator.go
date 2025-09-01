package morph

import (
	"fmt"
	"math"
	"slices"
	"strings"
)

var (
	obj_global_null  = &objectNull{}
	obj_global_true  = &objectBoolean{value: true}
	obj_global_false = &objectBoolean{value: false}
)

type evaluator struct {
	parser *parser
}

func newEvaluator(p *parser) *evaluator {
	return &evaluator{parser: p}
}

func (e *evaluator) eval(astNode node, env *environment) object {
	switch astNode := astNode.(type) {
	case *program:
		return e.evalProgramStatements(astNode, env)
	case *setStatement:
		return e.evalSetStatement(astNode, env)
	case *whenStatement:
		return e.evalWhenStatement(astNode, env)
	case *expressionStatement:
		return e.eval(astNode.expression, env)
	case *pathExpression:
		return e.evalPathExpression(astNode, env)
	case *identifierExpression:
		return e.evalIdentifierExpression(astNode, env)
	case *templateExpression:
		return e.evalTemplateExpression(astNode, env)
	case *prefixExpression:
		return e.evalPrefixExpression(astNode, env)
	case *infixExpression:
		return e.evalInfixExpression(astNode, env)
	case *stringLiteral:
		return &objectString{value: astNode.value}
	case *integerLiteral:
		return &objectInteger{value: astNode.value}
	case *floatLiteral:
		return &objectFloat{value: astNode.value}
	case *booleanLiteral:
		if astNode.value {
			return obj_global_true
		} else {
			return obj_global_false
		}
	default:
		return obj_global_null
	}
}

func (e *evaluator) evalProgramStatements(programNode *program, env *environment) object {
	for _, stmt := range programNode.statements {
		e.eval(stmt, env)
	}
	return obj_global_null
}

func (e *evaluator) evalInfixExpression(infix *infixExpression, env *environment) object {
	leftObj := e.eval(infix.left, env)
	if objectIsError(leftObj) {
		return leftObj
	}
	rightObj := e.eval(infix.right, env)
	if objectIsError(rightObj) {
		return rightObj
	}
	switch {
	case slices.Contains([]objectType{t_integer, t_float}, leftObj.getType()) && slices.Contains([]objectType{t_integer, t_float}, rightObj.getType()):
		return e.evalNumberInfixExpression(leftObj, infix.operator, rightObj)
	case leftObj.getType() != rightObj.getType():
		return objectNewErr("type mismatch: %s %s %s", leftObj.getType(), infix.operator, rightObj.getType())
	case leftObj.getType() == t_string && rightObj.getType() == t_string:
		return e.evalStringInfixExpression(leftObj, infix.operator, rightObj)
	case infix.operator == "==":
		return objectFromBoolean(leftObj == rightObj)
	case infix.operator == "!=":
		return objectFromBoolean(leftObj != rightObj)
	default:
		return objectNewErr("invalid operator for types: %s %s %s", leftObj.getType(), infix.operator, rightObj.getType())
	}
}
func (e *evaluator) evalStringInfixExpression(leftObj object, operator string, rightObj object) object {
	l := leftObj.(*objectString).value
	r := rightObj.(*objectString).value
	if operator != "+" {
		return objectNewErr("invalid operator for types: %s %s %s", leftObj.getType(), operator, rightObj.getType())
	}

	return &objectString{value: l + r}
}

func (e *evaluator) evalNumberInfixExpression(leftObj object, operator string, rightObj object) object {
	leftNum, err := objectNumberToFloat64(leftObj)
	if err != nil {
		return objectNewErr("invalid type for numeric operation: %s", leftObj.getType())
	}
	rightNum, err := objectNumberToFloat64(rightObj)
	if err != nil {
		return objectNewErr("invalid type for numeric operation: %s", rightObj.getType())
	}

	areBothInteger := leftObj.getType() == t_integer && rightObj.getType() == t_integer
	if slices.Contains([]string{"+", "-", "*", "/"}, operator) {
		return objHandleMathOperation(leftNum, operator, rightNum, areBothInteger)
	}

	switch operator {
	case "%":
		if !areBothInteger {
			return objectNewErr("invalid operator for input types: %s %s %s", leftObj.getType(), operator, rightObj.getType())
		}
		res := int64(leftNum) % int64(rightNum)
		return &objectInteger{value: res}
	case "<":
		return objectFromBoolean(leftNum < rightNum)
	case "<=":
		return objectFromBoolean(isFloatEqual(leftNum, rightNum) || leftNum <= rightNum)
	case ">":
		return objectFromBoolean(leftNum > rightNum)
	case ">=":
		return objectFromBoolean(isFloatEqual(leftNum, rightNum) || leftNum > rightNum)
	case "==":
		return objectFromBoolean(isFloatEqual(leftNum, rightNum))
	case "!=":
		return objectFromBoolean(!isFloatEqual(leftNum, rightNum))
	default:
		return objectNewErr("")
	}
}

const float_equality_tolerance = 1e-9

func isFloatEqual(a float64, b float64) bool {
	return math.Abs(a-b) <= float_equality_tolerance
}

func objHandleMathOperation(l float64, operator string, r float64, areBothInteger bool) object {

	var result float64
	avoidInteger := false
	switch operator {
	case "+":
		result = l + r
	case "-":
		result = l - r
	case "*":
		result = l * r
	case "/":
		result = l / r
		if result != math.Trunc(result) {
			avoidInteger = true
		}
	}
	if areBothInteger && !avoidInteger {
		return &objectInteger{value: int64(result)}
	}
	return &objectFloat{value: result}
}

func objectNumberToFloat64(obj object) (float64, error) {
	switch v := obj.(type) {
	case *objectInteger:
		return float64(v.value), nil
	case *objectFloat:
		return v.value, nil
	default:
		return 0, fmt.Errorf("not a valid number object")
	}
}

func (e *evaluator) evalPrefixExpression(prefix *prefixExpression, env *environment) object {
	rightObj := e.eval(prefix.right, env)
	if objectIsError(rightObj) {
		return rightObj
	}
	switch prefix.operator {
	case "!":
		return e.handlePrefixExclamation(prefix, rightObj)
	case "-":
		return e.handlePrefixMinus(prefix, rightObj)
	default:
		return objectNewErr("%s: unknown operator: %s", e.lineColForNode(prefix), prefix.operator)
	}
}
func (e *evaluator) handlePrefixExclamation(rightExpr *prefixExpression, rightObj object) object {
	switch rightObj {
	case obj_global_false:
		return obj_global_true
	case obj_global_true:
		return obj_global_false
	default:
		return objectNewErr("%s: incompatible non-boolean right-side exprssion for operator: !%s", e.lineColForNode(rightExpr), rightExpr.string())
	}
}
func (e *evaluator) handlePrefixMinus(rightExpr *prefixExpression, rightObj object) object {
	switch v := rightObj.(type) {
	case *objectInteger:
		return &objectInteger{value: -v.value}
	case *objectFloat:
		return &objectFloat{value: -v.value}
	default:
		return objectNewErr("%s: incompatible non-numeric right-side expression for operator: -%s", e.lineColForNode(rightExpr), rightExpr.string())
	}
}

func (e *evaluator) evalSetStatement(setStmt *setStatement, env *environment) object {
	valToSet := e.eval(setStmt.value, env)
	valToSet = valToSet.clone()
	if objectIsError(valToSet) {
		return valToSet
	}
	var objHandle object // reference to object at current path. may be unused in instances where we're just assigning a regular variable without dot-path syntax
	currentPath := setStmt.target.toAssignPath()
	for currentPath != nil {
		switch currentPath.stepType {
		case assign_step_env:
			objHandle = e.setStatementHandleENV(currentPath, valToSet, env)
			if objectIsError(objHandle) {
				return objHandle
			}
		case assign_step_map_key:
			objHandle = e.setStatementHandleMAP(objHandle, currentPath, valToSet, setStmt.target)
			if objectIsError(objHandle) {
				return objHandle
			}
		default:
			return objectNewErr("%s: invalid path part for SET statement", e.lineColForNode(setStmt.target))
		}
		currentPath = currentPath.next
	}
	return obj_global_null
}

func (e *evaluator) setStatementHandleENV(current *assignPath, valToSet object, env *environment) object {
	if current.next == nil {
		env.set(current.partName, valToSet)
		return obj_global_null
	}
	existing, ok := env.get(current.partName)
	if !ok {
		newMap := &objectMap{kvPairs: make(map[string]objectMapPair)}
		return env.set(current.partName, newMap)
	} else {
		return existing
	}
}
func (e *evaluator) setStatementHandleMAP(objHandle object, current *assignPath, valToSet object, setTarget assignable) object {
	mapObj, ok := objHandle.(*objectMap)
	if !ok {
		return objectNewErr("%s: invalid path part for SET statement: cannot use a path expression on a non-map object", e.lineColForNode(setTarget))
	}
	if current.next == nil {
		pair := objectMapPair{
			key:   current.partName,
			value: valToSet,
		}
		mapObj.kvPairs[current.partName] = pair
		return obj_global_null
	}
	existing, ok := mapObj.kvPairs[current.partName]
	if !ok {
		newMap := &objectMap{kvPairs: make(map[string]objectMapPair)}
		pair := objectMapPair{
			key:   current.partName,
			value: newMap,
		}
		mapObj.kvPairs[current.partName] = pair
		return newMap
	}
	return existing.value
}

func (e *evaluator) evalWhenStatement(whenStmt *whenStatement, env *environment) object {
	conditionObj := e.eval(whenStmt.condition, env)
	if objectIsError(conditionObj) {
		return conditionObj
	}
	if conditionObj.isTruthy() {
		e.eval(whenStmt.consequence, env)
	}
	return obj_global_null
}

func (e *evaluator) evalIdentifierExpression(identExpr *identifierExpression, env *environment) object {
	if res, ok := env.get(identExpr.value); ok {
		return res
	}
	return objectNewErr("%s: identifier not found: %s", e.lineColForNode(identExpr), identExpr.value)
}

func (e *evaluator) evalTemplateExpression(templateExpr *templateExpression, env *environment) object {
	stringParts := []string{}
	for _, entry := range templateExpr.parts {
		stringParts = append(stringParts, e.eval(entry, env).inspect())
	}
	return &objectString{value: strings.Join(stringParts, "")}
}

func (e *evaluator) evalPathExpression(pathExpr *pathExpression, env *environment) object {

	// get left side value via eval
	leftObj := e.eval(pathExpr.left, env)

	// apply attribute value to
	switch v := pathExpr.attribute.(type) {
	case *identifierExpression:
		leftMap, ok := leftObj.(*objectMap)
		if !ok {
			return objectNewErr("%s: cannot access a path on a non-map object", e.lineColForNode(pathExpr.left))
		}
		res, ok := leftMap.kvPairs[v.value]
		if !ok {
			return objectNewErr("%s: key not found: %s", e.lineColForNode(pathExpr.left), v.value)
		}
		return res.value
	default:
		msg := fmt.Sprintf("%s: invalid path part: %s", e.lineColForNode(v), v.string())
		return objectNewErr(msg)
	}
}

// err helpers

func objectNewErr(format string, a ...interface{}) *objectError {
	return &objectError{message: fmt.Sprintf(format, a...)}
}

func objectIsError(obj object) bool {
	if obj != nil {
		return obj.getType() == t_error
	}
	return false
}

//

func (e *evaluator) lineColForNode(n node) string {
	return lineColString(lineAndCol(e.parser.lexer.input, n.position().start))
}

func objectFromBoolean(b bool) *objectBoolean {
	if b {
		return obj_global_true
	} else {
		return obj_global_false
	}
}
