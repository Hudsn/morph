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
	obj_global_term  = &objectTerminate{}
)

type evaluator struct {
	parser *parser
}

func newEvaluator(p *parser) *evaluator {
	return &evaluator{parser: p}
}

func (e *evaluator) eval(astNode node, env *environment) (object, error) {
	switch astNode := astNode.(type) {
	case *program:
		return e.evalProgramStatements(astNode, env)
	case *setStatement:
		return e.evalSetStatement(astNode, env)
	case *whenStatement:
		return e.evalWhenStatement(astNode, env)
	case *expressionStatement:
		return e.eval(astNode.expression, env)
	case *callExpression:
		return e.evalCallExpression(astNode, env)
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
		return &objectString{value: astNode.value}, nil
	case *integerLiteral:
		return &objectInteger{value: astNode.value}, nil
	case *floatLiteral:
		return &objectFloat{value: astNode.value}, nil
	case *booleanLiteral:
		if astNode.value {
			return obj_global_true, nil
		} else {
			return obj_global_false, nil
		}
	case *mapLiteral:
		return e.evalMapLiteral(astNode, env)
	case *arrayLiteral:
		return e.evalArrayLiteral(astNode, env)
	case *indexExpression:
		return e.evalIndexExpression(astNode, env)
	default:
		return obj_global_null, fmt.Errorf("%s: unsupported statement", e.lineColForNode(astNode))
	}
}

func (e *evaluator) evalProgramStatements(programNode *program, env *environment) (object, error) {
	for _, stmt := range programNode.statements {
		_, err := e.eval(stmt, env)
		if err != nil {
			return obj_global_null, err
		}
	}
	return obj_global_null, nil
}

// infix

func (e *evaluator) evalInfixExpression(infix *infixExpression, env *environment) (object, error) {
	leftObj, err := e.eval(infix.left, env)
	if err != nil {
		return obj_global_null, err
	}
	rightObj, err := e.eval(infix.right, env)
	if err != nil {
		return obj_global_null, err
	}
	switch {
	case slices.Contains([]objectType{t_integer, t_float}, leftObj.getType()) && slices.Contains([]objectType{t_integer, t_float}, rightObj.getType()):
		ret, err := e.evalNumberInfixExpression(leftObj, infix.operator, rightObj)
		if err != nil {
			return obj_global_null, fmt.Errorf("%s: %w", e.lineColForNode(infix), err)
		}
		return ret, nil
	case leftObj.getType() != rightObj.getType():
		return obj_global_null, fmt.Errorf("type mismatch: %s %s %s", leftObj.getType(), infix.operator, rightObj.getType())
	case leftObj.getType() == t_string && rightObj.getType() == t_string:
		ret, err := e.evalStringInfixExpression(leftObj, infix.operator, rightObj)
		if err != nil {
			return obj_global_null, fmt.Errorf("%s: %w", e.lineColForNode(infix), err)
		}
		return ret, nil
	case infix.operator == "==":
		return objectFromBoolean(leftObj == rightObj), nil
	case infix.operator == "!=":
		return objectFromBoolean(leftObj != rightObj), nil
	default:
		return obj_global_null, fmt.Errorf("invalid operator for types: %s %s %s", leftObj.getType(), infix.operator, rightObj.getType())
	}
}
func (e *evaluator) evalStringInfixExpression(leftObj object, operator string, rightObj object) (object, error) {
	l := leftObj.(*objectString).value
	r := rightObj.(*objectString).value
	if operator != "+" {
		return obj_global_null, fmt.Errorf("invalid operator for types: %s %s %s", leftObj.getType(), operator, rightObj.getType())
	}

	return &objectString{value: l + r}, nil
}

func (e *evaluator) evalNumberInfixExpression(leftObj object, operator string, rightObj object) (object, error) {
	leftNum, err := objectNumberToFloat64(leftObj)
	if err != nil {
		return obj_global_null, err
	}
	rightNum, err := objectNumberToFloat64(rightObj)
	if err != nil {
		return obj_global_null, err
	}

	areBothInteger := leftObj.getType() == t_integer && rightObj.getType() == t_integer
	if slices.Contains([]string{"+", "-", "*", "/"}, operator) {
		return objHandleMathOperation(leftNum, operator, rightNum, areBothInteger), nil
	}

	switch operator {
	case "%":
		if !areBothInteger {
			return obj_global_null, fmt.Errorf("invalid operator for input types: %s %s %s", leftObj.getType(), operator, rightObj.getType())
		}
		res := int64(leftNum) % int64(rightNum)
		return &objectInteger{value: res}, nil
	case "<":
		return objectFromBoolean(leftNum < rightNum), nil
	case "<=":
		return objectFromBoolean(isFloatEqual(leftNum, rightNum) || leftNum < rightNum), nil
	case ">":
		return objectFromBoolean(leftNum > rightNum), nil
	case ">=":
		return objectFromBoolean(isFloatEqual(leftNum, rightNum) || leftNum > rightNum), nil
	case "==":
		return objectFromBoolean(isFloatEqual(leftNum, rightNum)), nil
	case "!=":
		return objectFromBoolean(!isFloatEqual(leftNum, rightNum)), nil
	default:
		return obj_global_null, fmt.Errorf("unsupported operator: %s", operator)
	}
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

//prefix

func (e *evaluator) evalPrefixExpression(prefix *prefixExpression, env *environment) (object, error) {
	rightObj, err := e.eval(prefix.right, env)
	if err != nil {
		return obj_global_null, err
	}
	switch prefix.operator {
	case "!":
		return e.handlePrefixExclamation(prefix, rightObj)
	case "-":
		return e.handlePrefixMinus(prefix, rightObj)
	default:
		return obj_global_null, fmt.Errorf("%s: unknown operator: %s", e.lineColForNode(prefix), prefix.operator)
	}
}
func (e *evaluator) handlePrefixExclamation(rightExpr *prefixExpression, rightObj object) (object, error) {
	switch rightObj {
	case obj_global_false:
		return obj_global_true, nil
	case obj_global_true:
		return obj_global_false, nil
	default:
		return obj_global_null, fmt.Errorf("%s: incompatible non-boolean right-side exprssion for operator: !%s", e.lineColForNode(rightExpr), rightExpr.string())
	}
}
func (e *evaluator) handlePrefixMinus(rightExpr *prefixExpression, rightObj object) (object, error) {
	switch v := rightObj.(type) {
	case *objectInteger:
		return &objectInteger{value: -v.value}, nil
	case *objectFloat:
		return &objectFloat{value: -v.value}, nil
	default:
		return obj_global_null, fmt.Errorf("%s: incompatible non-numeric right-side expression for operator: -%s", e.lineColForNode(rightExpr), rightExpr.string())
	}
}

//set stmt

func (e *evaluator) evalSetStatement(setStmt *setStatement, env *environment) (object, error) {
	valToSet, err := e.eval(setStmt.value, env)
	if err != nil {
		return obj_global_null, err
	}
	valToSet = valToSet.clone()

	var objHandle object // reference to object at current path. may be unused in instances where we're just assigning a regular variable without dot-path syntax
	currentPath := setStmt.target.toAssignPath()
	for currentPath != nil {
		switch currentPath.stepType {
		case assign_step_env:
			objHandle = e.setStatementHandleENV(currentPath, valToSet, env)
		case assign_step_map_key:
			objHandle, err = e.setStatementHandleMAP(objHandle, currentPath, valToSet, setStmt.target)
			if err != nil {
				return obj_global_null, err
			}
		default:
			return obj_global_null, fmt.Errorf("%s: invalid path part for SET statement", e.lineColForNode(setStmt.target))
		}
		currentPath = currentPath.next
	}
	return obj_global_null, nil
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

func (e *evaluator) setStatementHandleMAP(objHandle object, current *assignPath, valToSet object, setTarget assignable) (object, error) {
	mapObj, ok := objHandle.(*objectMap)
	if !ok {
		return obj_global_null, fmt.Errorf("%s: invalid path part for SET statement: cannot use a path expression on a non-map object", e.lineColForNode(setTarget))
	}
	if current.next == nil {
		pair := objectMapPair{
			key:   current.partName,
			value: valToSet,
		}
		mapObj.kvPairs[current.partName] = pair
		return obj_global_null, nil
	}
	existing, ok := mapObj.kvPairs[current.partName]
	if !ok {
		newMap := &objectMap{kvPairs: make(map[string]objectMapPair)}
		pair := objectMapPair{
			key:   current.partName,
			value: newMap,
		}
		mapObj.kvPairs[current.partName] = pair
		return newMap, nil
	}
	return existing.value, nil
}

//when

func (e *evaluator) evalWhenStatement(whenStmt *whenStatement, env *environment) (object, error) {
	conditionObj, err := e.eval(whenStmt.condition, env)
	if err != nil {
		return obj_global_null, err
	}
	if conditionObj.isTruthy() {
		if _, err := e.eval(whenStmt.consequence, env); err != nil {
			return obj_global_null, err
		}
	}
	return obj_global_null, nil
}

//ident

func (e *evaluator) evalIdentifierExpression(identExpr *identifierExpression, env *environment) (object, error) {
	if res, ok := env.get(identExpr.value); ok {
		return res, nil
	}
	return obj_global_null, fmt.Errorf("%s: identifier not found: %s", e.lineColForNode(identExpr), identExpr.value)
}

//template

func (e *evaluator) evalTemplateExpression(templateExpr *templateExpression, env *environment) (object, error) {
	stringParts := []string{}
	for _, entry := range templateExpr.parts {
		res, err := e.eval(entry, env)
		if err != nil {
			return obj_global_null, err
		}
		stringParts = append(stringParts, res.inspect())
	}
	return &objectString{value: strings.Join(stringParts, "")}, nil
}

//path

func (e *evaluator) evalPathExpression(pathExpr *pathExpression, env *environment) (object, error) {
	// apply attribute value to
	switch v := pathExpr.attribute.(type) {
	case *stringLiteral:
		return e.resolvePathEntryForKey(pathExpr, v.value, env)
	case *identifierExpression:
		return e.resolvePathEntryForKey(pathExpr, v.value, env)
	default:
		return obj_global_null, fmt.Errorf("%s: invalid path part: %s", e.lineColForNode(v), v.string())
	}
}

func (e *evaluator) resolvePathEntryForKey(pathExpr *pathExpression, key string, env *environment) (object, error) {
	// get left side value via eval
	leftObj, err := e.eval(pathExpr.left, env)
	if err != nil {
		return obj_global_null, err
	}
	leftMap, ok := leftObj.(*objectMap)
	if !ok {
		return obj_global_null, fmt.Errorf("%s: cannot access a path on a non-map object", e.lineColForNode(pathExpr.left))
	}
	res, ok := leftMap.kvPairs[key]
	if !ok {
		return obj_global_null, fmt.Errorf("%s: key not found: %s", e.lineColForNode(pathExpr.left), key)
	}
	return res.value, nil
}

// map
func (e *evaluator) evalMapLiteral(mapLit *mapLiteral, env *environment) (object, error) {
	objPairs := make(map[string]objectMapPair)
	for key, expr := range mapLit.pairs {

		toAdd, err := e.eval(expr, env)
		if err != nil {
			return obj_global_null, err
		}
		pair := objectMapPair{
			key:   key,
			value: toAdd,
		}
		objPairs[key] = pair
	}
	return &objectMap{kvPairs: objPairs}, nil
}

// arr
func (e *evaluator) evalArrayLiteral(arrayLit *arrayLiteral, env *environment) (object, error) {
	objEntries := []object{}
	for _, entryExpr := range arrayLit.entries {
		toAdd, err := e.eval(entryExpr, env)
		if err != nil {
			return obj_global_null, err
		}
		objEntries = append(objEntries, toAdd)
	}
	return &objectArray{entries: objEntries}, nil
}

//index

func (e *evaluator) evalIndexExpression(indexExpr *indexExpression, env *environment) (object, error) {
	identResult, err := e.eval(indexExpr.left, env)
	if err != nil {
		return obj_global_null, err
	}
	arrObj, ok := identResult.(*objectArray)
	if !ok {
		return obj_global_null, fmt.Errorf("%s: cannot call index expression on non-array object", e.lineColForNode(indexExpr.left))
	}

	indexObj, err := e.eval(indexExpr.index, env)
	if err != nil {
		return obj_global_null, err
	}
	if indexObj.getType() != t_integer {
		return obj_global_null, fmt.Errorf("%s: index is not of type %s. got=%s", e.lineColForNode(indexExpr.index), t_integer, indexObj.getType())
	}
	idxInt, ok := indexObj.(*objectInteger)
	if !ok {
		return obj_global_null, fmt.Errorf("%s: index is not of type %s. got=%s", e.lineColForNode(indexExpr.index), t_integer, indexObj.getType())
	}

	targetIdx := int(idxInt.value)
	if int(targetIdx) >= len(arrObj.entries) || targetIdx < 0 {
		return obj_global_null, fmt.Errorf("%s: index is out of range for target array", e.lineColForNode(indexExpr.index))
	}
	return arrObj.entries[targetIdx], nil
}

//

func (e *evaluator) evalCallExpression(callExpr *callExpression, env *environment) (object, error) {
	return nil, nil
}

func (e *evaluator) evalFunctionPath(pathExpr *pathExpression, env *environment) (object, error) {
	return nil, nil
}

// helpers

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
