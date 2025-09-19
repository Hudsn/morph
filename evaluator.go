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

func eval(astNode node, env *environment) (object, error) {
	switch astNode := astNode.(type) {
	case *program:
		return evalProgramStatements(astNode, env)
	case *setStatement:
		return evalSetStatement(astNode, env)
	case *whenStatement:
		return evalWhenStatement(astNode, env)
	case *expressionStatement:
		return eval(astNode.expression, env)
	case *callExpression:
		return evalCallExpression(astNode, env)
	case *pathExpression:
		return evalPathExpression(astNode, env)
	case *identifierExpression:
		return evalIdentifierExpression(astNode, env)
	case *templateExpression:
		return evalTemplateExpression(astNode, env)
	case *prefixExpression:
		return evalPrefixExpression(astNode, env)
	case *infixExpression:
		return evalInfixExpression(astNode, env)
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
		return evalMapLiteral(astNode, env)
	case *arrayLiteral:
		return evalArrayLiteral(astNode, env)
	case *indexExpression:
		return evalIndexExpression(astNode, env)
	default:
		return obj_global_null, fmt.Errorf("%s: unsupported statement", astNode.token().lineCol)
	}
}

func evalProgramStatements(programNode *program, env *environment) (object, error) {
	for _, stmt := range programNode.statements {
		_, err := eval(stmt, env)
		if err != nil {
			return obj_global_null, err
		}
	}
	return obj_global_null, nil
}

// infix

func evalInfixExpression(infix *infixExpression, env *environment) (object, error) {
	leftObj, err := eval(infix.left, env)
	if err != nil {
		return obj_global_null, err
	}
	rightObj, err := eval(infix.right, env)
	if err != nil {
		return obj_global_null, err
	}
	switch {
	case slices.Contains([]objectType{t_integer, t_float}, leftObj.getType()) && slices.Contains([]objectType{t_integer, t_float}, rightObj.getType()):
		ret, err := evalNumberInfixExpression(leftObj, infix.operator, rightObj)
		if err != nil {
			return obj_global_null, fmt.Errorf("%s: %w", infix.tok.lineCol, err)
		}
		return ret, nil
	case leftObj.getType() != rightObj.getType():
		return obj_global_null, fmt.Errorf("type mismatch: %s %s %s", leftObj.getType(), infix.operator, rightObj.getType())
	case leftObj.getType() == t_string && rightObj.getType() == t_string:
		ret, err := evalStringInfixExpression(leftObj, infix.operator, rightObj)
		if err != nil {
			return obj_global_null, fmt.Errorf("%s: %w", infix.tok.lineCol, err)
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
func evalStringInfixExpression(leftObj object, operator string, rightObj object) (object, error) {
	l := leftObj.(*objectString).value
	r := rightObj.(*objectString).value
	if operator != "+" {
		return obj_global_null, fmt.Errorf("invalid operator for types: %s %s %s", leftObj.getType(), operator, rightObj.getType())
	}

	return &objectString{value: l + r}, nil
}

func evalNumberInfixExpression(leftObj object, operator string, rightObj object) (object, error) {
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

func evalPrefixExpression(prefix *prefixExpression, env *environment) (object, error) {
	rightObj, err := eval(prefix.right, env)
	if err != nil {
		return obj_global_null, err
	}
	switch prefix.operator {
	case "!":
		return evalHandlePrefixExclamation(prefix, rightObj)
	case "-":
		return evalHandlePrefixMinus(prefix, rightObj)
	default:
		return obj_global_null, fmt.Errorf("%s: unknown operator: %s", prefix.tok.lineCol, prefix.operator)
	}
}
func evalHandlePrefixExclamation(rightExpr *prefixExpression, rightObj object) (object, error) {
	switch rightObj {
	case obj_global_false:
		return obj_global_true, nil
	case obj_global_true:
		return obj_global_false, nil
	default:
		return obj_global_null, fmt.Errorf("%s: incompatible non-boolean right-side exprssion for operator: !%s", rightExpr.tok.lineCol, rightExpr.string())
	}
}
func evalHandlePrefixMinus(rightExpr *prefixExpression, rightObj object) (object, error) {
	switch v := rightObj.(type) {
	case *objectInteger:
		return &objectInteger{value: -v.value}, nil
	case *objectFloat:
		return &objectFloat{value: -v.value}, nil
	default:
		return obj_global_null, fmt.Errorf("%s: incompatible non-numeric right-side expression for operator: -%s", rightExpr.tok.lineCol, rightExpr.string())
	}
}

//set stmt

func evalSetStatement(setStmt *setStatement, env *environment) (object, error) {
	valToSet, err := eval(setStmt.value, env)
	if err != nil {
		return obj_global_null, err
	}
	valToSet = valToSet.clone()

	var objHandle object // reference to object at current path. may be unused in instances where we're just assigning a regular variable without dot-path syntax
	currentPath := setStmt.target.toAssignPath()
	for currentPath != nil {
		switch currentPath.stepType {
		case assign_step_env:
			objHandle = evalSetStatementHandleENV(currentPath, valToSet, env)
		case assign_step_map_key:
			objHandle, err = evalSetStatementHandleMAP(objHandle, currentPath, valToSet, setStmt.target)
			if err != nil {
				return obj_global_null, err
			}
		default:
			return obj_global_null, fmt.Errorf("%s: invalid path part for SET statement", setStmt.target.token().lineCol)
		}
		currentPath = currentPath.next
	}
	return obj_global_null, nil
}

func evalSetStatementHandleENV(current *assignPath, valToSet object, env *environment) object {
	if current.next == nil {
		env.set(current.partName, valToSet)
		return obj_global_null
	}
	existing, ok := env.get(current.partName)
	if !ok {
		newMap := &objectMap{kvPairs: make(map[string]object)}
		return env.set(current.partName, newMap)
	} else {
		return existing
	}
}

func evalSetStatementHandleMAP(objHandle object, current *assignPath, valToSet object, setTarget assignable) (object, error) {
	mapObj, ok := objHandle.(*objectMap)
	if !ok {
		return obj_global_null, fmt.Errorf("%s: invalid path part for SET statement: cannot use a path expression on a non-map object", setTarget.token().lineCol)
	}
	if current.next == nil {
		mapObj.kvPairs[current.partName] = valToSet
		return obj_global_null, nil
	}
	existing, ok := mapObj.kvPairs[current.partName]
	if !ok {
		newMap := &objectMap{kvPairs: make(map[string]object)}
		mapObj.kvPairs[current.partName] = newMap
		return newMap, nil
	}
	return existing, nil
}

//when

func evalWhenStatement(whenStmt *whenStatement, env *environment) (object, error) {
	conditionObj, err := eval(whenStmt.condition, env)
	if err != nil {
		return obj_global_null, err
	}
	if conditionObj.isTruthy() {
		if _, err := eval(whenStmt.consequence, env); err != nil {
			return obj_global_null, err
		}
	}
	return obj_global_null, nil
}

//ident

func evalIdentifierExpression(identExpr *identifierExpression, env *environment) (object, error) {
	if res, ok := env.get(identExpr.value); ok {
		return res, nil
	}
	return obj_global_null, fmt.Errorf("%s: identifier not found: %s", identExpr.token().lineCol, identExpr.value)
}

//template

func evalTemplateExpression(templateExpr *templateExpression, env *environment) (object, error) {
	stringParts := []string{}
	for _, entry := range templateExpr.parts {
		res, err := eval(entry, env)
		if err != nil {
			return obj_global_null, err
		}
		stringParts = append(stringParts, res.inspect())
	}
	return &objectString{value: strings.Join(stringParts, "")}, nil
}

//path

func evalPathExpression(pathExpr *pathExpression, env *environment) (object, error) {
	// apply attribute value to
	switch v := pathExpr.attribute.(type) {
	case *stringLiteral:
		return evalResolvePathEntryForKey(pathExpr, v.value, env)
	case *identifierExpression:
		return evalResolvePathEntryForKey(pathExpr, v.value, env)
	default:
		return obj_global_null, fmt.Errorf("%s: invalid path part: %s", v.token().lineCol, v.string())
	}
}

func evalResolvePathEntryForKey(pathExpr *pathExpression, key string, env *environment) (object, error) {
	// get left side value via eval
	leftObj, err := eval(pathExpr.left, env)
	if err != nil {
		return obj_global_null, err
	}
	leftMap, ok := leftObj.(*objectMap)
	if !ok {
		return obj_global_null, fmt.Errorf("%s: cannot access a path on a non-map object", pathExpr.left.token().lineCol)
	}
	res, ok := leftMap.kvPairs[key]
	if !ok {
		return obj_global_null, fmt.Errorf("%s: key not found: %s", pathExpr.left.token().lineCol, key)
	}
	return res, nil
}

// map
func evalMapLiteral(mapLit *mapLiteral, env *environment) (object, error) {
	objPairs := make(map[string]object)
	for key, expr := range mapLit.pairs {

		objToAdd, err := eval(expr, env)
		if err != nil {
			return obj_global_null, err
		}
		objPairs[key] = objToAdd
	}
	return &objectMap{kvPairs: objPairs}, nil
}

// arr
func evalArrayLiteral(arrayLit *arrayLiteral, env *environment) (object, error) {
	objEntries := []object{}
	for _, entryExpr := range arrayLit.entries {
		toAdd, err := eval(entryExpr, env)
		if err != nil {
			return obj_global_null, err
		}
		objEntries = append(objEntries, toAdd)
	}
	return &objectArray{entries: objEntries}, nil
}

//index

func evalIndexExpression(indexExpr *indexExpression, env *environment) (object, error) {
	identResult, err := eval(indexExpr.left, env)
	if err != nil {
		return obj_global_null, err
	}
	arrObj, ok := identResult.(*objectArray)
	if !ok {
		return obj_global_null, fmt.Errorf("%s: cannot call index expression on non-array object", indexExpr.left.token().lineCol)
	}

	indexObj, err := eval(indexExpr.index, env)
	if err != nil {
		return obj_global_null, err
	}
	if indexObj.getType() != t_integer {
		return obj_global_null, fmt.Errorf("%s: index is not of type %s. got=%s", indexExpr.index.token().lineCol, t_integer, indexObj.getType())
	}
	idxInt, ok := indexObj.(*objectInteger)
	if !ok {
		return obj_global_null, fmt.Errorf("%s: index is not of type %s. got=%s", indexExpr.index.token().lineCol, t_integer, indexObj.getType())
	}

	targetIdx := int(idxInt.value)
	if int(targetIdx) >= len(arrObj.entries) || targetIdx < 0 {
		return obj_global_null, fmt.Errorf("%s: index is out of range for target array", indexExpr.index.token().lineCol)
	}
	return arrObj.entries[targetIdx], nil
}

//

func evalCallExpression(callExpr *callExpression, env *environment) (object, error) {
	var fnEntry *functionEntry
	var err error
	switch v := callExpr.name.(type) {
	case *identifierExpression:
		fnEntry, err = env.functions.get(v.value)
		if err != nil {
			return obj_global_null, err
		}
	case *pathExpression:
		fnEntry, err = evalFunctionNamePath(v, env)
		if err != nil {
			return obj_global_null, err
		}
	}
	args := []object{}
	for _, argExpr := range callExpr.arguments {
		toAdd, err := eval(argExpr, env)
		if err != nil {
			return obj_global_null, err
		}
		args = append(args, toAdd)
	}
	return fnEntry.eval(args...)
}

func evalFunctionNamePath(pathExpr *pathExpression, env *environment) (*functionEntry, error) {
	switch v := pathExpr.attribute.(type) {
	case *identifierExpression:
		return evalResolvePathForFunction(pathExpr, v.value, env)
	}
	return nil, fmt.Errorf("%s: function path must be composed of valid identifiers", pathExpr.token().lineCol)
}

func evalResolvePathForFunction(pathExpr *pathExpression, key string, env *environment) (*functionEntry, error) {
	switch v := pathExpr.left.(type) {
	case *identifierExpression:
		namespace := v.value
		return env.functions.getNamespace(namespace, key)
	}
	return nil, fmt.Errorf("%s: function path must be composed of valid identifiers", pathExpr.token().lineCol)
}

// helpers

func objectFromBoolean(b bool) *objectBoolean {
	if b {
		return obj_global_true
	} else {
		return obj_global_false
	}
}
