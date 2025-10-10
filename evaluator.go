package morph

import (
	"fmt"
	"math"
	"slices"
	"strings"
)

var (
	obj_global_null      = &objectNull{}
	obj_global_true      = &objectBoolean{value: true}
	obj_global_false     = &objectBoolean{value: false}
	obj_global_term      = &objectTerminate{shouldReturnNull: false}
	obj_global_term_drop = &objectTerminate{shouldReturnNull: true}
)

//
//program

func (p *program) eval(env *environment) object {
	for _, stmt := range p.statements {
		obj := stmt.eval(env)
		if isObjectErr(obj) {
			return unWrapErr(stmt.token().lineCol, obj)
		}
		if obj.getType() == t_terminate {
			term := obj.(*objectTerminate)
			if term.shouldReturnNull {
				env.store = map[string]object{}
			}
			break
		}
	}
	return obj_global_null
}

//
//SET

func (s *setStatement) eval(env *environment) object {
	valToSet := s.value.eval(env)
	if isObjectErr(valToSet) {
		return unWrapErr(s.value.token().lineCol, valToSet)
	}
	valToSet = valToSet.clone()

	var objHandle object // reference to object at current path. may be unused in instances where we're just assigning a regular variable without dot-path syntax
	currentPath := s.target.toAssignPath()
	for currentPath != nil {
		switch currentPath.stepType {
		case assign_step_env:
			objHandle = evalSetStatementHandleENV(currentPath, valToSet, env)
			if isObjectErr(objHandle) {
				return unWrapErr(s.target.token().lineCol, objHandle)
			}
		case assign_step_map_key:
			objHandle = evalSetStatementHandleMAP(objHandle, currentPath, valToSet)
			if isObjectErr(objHandle) {
				return unWrapErr(s.target.token().lineCol, objHandle)
			}
		default:
			return newObjectErr(s.target.token().lineCol, "invalid path part for SET statement")
		}
		currentPath = currentPath.next
	}
	return obj_global_null
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
	}
	if existing.getType() != t_map {
		return newObjectErrWithoutLC("invalid path part for SET statement: cannot use a path expression on a non-map object. Object is of type %s", existing.getType())
	}
	return existing
}

func evalSetStatementHandleMAP(objHandle object, current *assignPath, valToSet object) object {
	mapObj, ok := objHandle.(*objectMap)
	if !ok {
		return newObjectErrWithoutLC("invalid path part for SET statement: cannot use a path expression on a non-map object. Object is of type %s", objHandle.getType())
	}
	if current.next == nil {
		mapObj.kvPairs[current.partName] = valToSet
		return obj_global_null
	}
	existing, ok := mapObj.kvPairs[current.partName]
	if !ok {
		newMap := &objectMap{kvPairs: make(map[string]object)}
		mapObj.kvPairs[current.partName] = newMap
		return newMap
	}
	if existing.getType() != t_map {
		return newObjectErrWithoutLC("invalid path part for SET statement: cannot use a path expression on a non-map object. Object is of type %s", existing.getType())
	}
	return existing
}

//
// when statement

func (i *ifStatement) eval(env *environment) object {
	conditionObj := i.condition.eval(env)
	if isObjectErr(conditionObj) {
		return unWrapErr(i.condition.token().lineCol, conditionObj)
	}

	if conditionObj.isTruthy() {
		for _, c := range i.consequence {
			res := c.eval(env)
			if isObjectErr(res) {
				return unWrapErr(c.token().lineCol, res)
			}
		}
	}
	return obj_global_null
}

//
//expression statement

func (e *expressionStatement) eval(env *environment) object {
	res := e.expression.eval(env)
	if isObjectErr(res) {
		return unWrapErr(e.expression.token().lineCol, res)
	}
	return res
}

//
//call expression

func (c *callExpression) eval(env *environment) object {
	var fnEntry *functionEntry
	var err error
	switch v := c.name.(type) {
	case *identifierExpression:
		fnEntry, err = env.functions.get(v.value)
		if err != nil {
			return newObjectErr(v.token().lineCol, err.Error())
		}
	case *pathExpression:
		fnEntry, err = evalFunctionNamePath(v, env)
		if err != nil {
			return newObjectErr(v.token().lineCol, err.Error())
		}
	}
	args := []object{}
	for _, argExpr := range c.arguments {
		toAdd := argExpr.eval(env)
		args = append(args, toAdd)
	}
	ret := fnEntry.eval(args...)
	if isObjectErr(ret) {
		return unWrapErr(c.name.token().lineCol, ret)
	}
	return ret
}

func evalFunctionNamePath(pathExpr *pathExpression, env *environment) (*functionEntry, error) {
	switch v := pathExpr.attribute.(type) {
	case *identifierExpression:
		return evalResolvePathForFunction(pathExpr, v.value, env)
	}
	return nil, fmt.Errorf("function path must be composed of valid identifiers")
}

func evalResolvePathForFunction(pathExpr *pathExpression, key string, env *environment) (*functionEntry, error) {
	switch v := pathExpr.left.(type) {
	case *identifierExpression:
		namespace := v.value
		return env.functions.getNamespace(namespace, key)
	}
	return nil, fmt.Errorf("function path must be composed of valid identifiers")
}

//
// identifier expression

func (i *identifierExpression) eval(env *environment) object {
	if res, ok := env.get(i.value); ok {
		return res
	}
	return obj_global_null
}

// path expression
func (p *pathExpression) eval(env *environment) object {
	switch v := p.attribute.(type) {
	case *stringLiteral:
		ret := evalResolvePathEntryForKey(p, v.value, env)
		if isObjectErr(ret) {
			return unWrapErr(v.tok.lineCol, ret)
		}
		return ret
	case *identifierExpression:
		ret := evalResolvePathEntryForKey(p, v.value, env)
		if isObjectErr(ret) {
			return unWrapErr(v.tok.lineCol, ret)
		}
		return ret
	case *templateExpression:
		str := v.eval(env)
		if isObjectErr(str) {
			return unWrapErr(v.tok.lineCol, str)
		}
		if strVal, ok := str.(*objectString); ok {
			ret := evalResolvePathEntryForKey(p, strVal.value, env)
			if isObjectErr(ret) {
				return unWrapErr(v.tok.lineCol, ret)
			}
			return ret
		}
		return newObjectErr(v.tok.lineCol, "invalid path part: %s", v.string())

	default:
		return newObjectErr(v.token().lineCol, "invalid path part: %s", v.string())
	}
}

func evalResolvePathEntryForKey(pathExpr *pathExpression, key string, env *environment) object {
	leftObj := pathExpr.left.eval(env)
	if isObjectErr(leftObj) {
		return unWrapErr(pathExpr.left.token().lineCol, leftObj)
	}
	if leftObj == obj_global_null {
		return leftObj
	}
	leftMap, ok := leftObj.(*objectMap)
	if !ok {
		return newObjectErr(pathExpr.left.token().lineCol, "cannot access a path %q on a non-map object. %q is of type %s", pathExpr.string(), pathExpr.left.string(), leftObj.getType())
	}
	res, ok := leftMap.kvPairs[key] // if it is a map, but the item doesn't exist, we return null
	if !ok {
		return obj_global_null
	}
	return res
}

// template expr
func (t *templateExpression) eval(env *environment) object {
	stringParts := []string{}
	for _, entry := range t.parts {
		res := entry.eval(env)
		if isObjectErr(res) {
			return unWrapErr(entry.token().lineCol, res)
		}
		stringParts = append(stringParts, res.inspect())
	}
	return &objectString{value: strings.Join(stringParts, "")}
}

//
// prefix expr

func (p *prefixExpression) eval(env *environment) object {
	rightObj := p.right.eval(env)
	if isObjectErr(rightObj) {
		return unWrapErr(p.right.token().lineCol, rightObj)
	}
	switch p.operator {
	case "!":
		ret := evalHandlePrefixExclamation(p, rightObj)
		if isObjectErr(ret) {
			return unWrapErr(p.tok.lineCol, ret)
		}
		return ret
	case "-":
		ret := evalHandlePrefixMinus(p, rightObj)
		if isObjectErr(ret) {
			return unWrapErr(p.tok.lineCol, ret)
		}
		return ret
	default:
		return newObjectErr(p.tok.lineCol, "unknown operator: %s", p.operator)
	}
}
func evalHandlePrefixExclamation(rightExpr *prefixExpression, rightObj object) object {
	switch rightObj {
	case obj_global_false:
		return obj_global_true
	case obj_global_true:
		return obj_global_false
	default:
		return newObjectErr(rightExpr.tok.lineCol, "incompatible non-boolean right-side exprssion for ! operator: %s", rightExpr.string())
	}
}
func evalHandlePrefixMinus(rightExpr *prefixExpression, rightObj object) object {
	switch v := rightObj.(type) {
	case *objectInteger:
		return &objectInteger{value: -v.value}
	case *objectFloat:
		return &objectFloat{value: -v.value}
	default:
		return newObjectErr(rightExpr.tok.lineCol, "incompatible non-numeric right-side expression for operator: %s", rightExpr.string())
	}
}

//
// infix expr

func (i *infixExpression) eval(env *environment) object {
	leftObj := i.left.eval(env)
	if isObjectErr(leftObj) {
		return unWrapErr(i.left.token().lineCol, leftObj)
	}
	rightObj := i.right.eval(env)
	if isObjectErr(rightObj) {
		return unWrapErr(i.right.token().lineCol, rightObj)
	}
	switch {
	case slices.Contains([]objectType{t_integer, t_float}, leftObj.getType()) && slices.Contains([]objectType{t_integer, t_float}, rightObj.getType()):
		ret := evalNumberInfixExpression(leftObj, i.operator, rightObj)
		if isObjectErr(ret) {
			return unWrapErr(i.token().lineCol, ret)
		}
		return ret
	case leftObj.getType() == t_string && rightObj.getType() == t_string:
		ret := evalStringInfixExpression(leftObj, i.operator, rightObj)
		if isObjectErr(ret) {
			return unWrapErr(i.token().lineCol, ret)
		}
		return ret
	case leftObj.getType() == t_array && rightObj.getType() == t_array:
		ret := evalArrayInfixExpression(leftObj, i.operator, rightObj)
		if isObjectErr(ret) {
			return unWrapErr(i.token().lineCol, ret)
		}
		return ret
	case i.operator == "==":
		if leftObj.getType() != rightObj.getType() {
			return obj_global_false
		}
		if leftObj.getType() == t_time {
			lTime := leftObj.(*objectTime) // don't really need to bool check type assertions; we know that type are equal, and that they're of type t_time.
			rTime := rightObj.(*objectTime)
			return objectFromBoolean(lTime.value.Equal(rTime.value))
		}
		return objectFromBoolean(leftObj == rightObj)
	case i.operator == "!=":
		if leftObj.getType() != rightObj.getType() {
			return obj_global_true
		}
		return objectFromBoolean(leftObj != rightObj)
	case i.operator == "&&":
		return objectFromBoolean(leftObj.isTruthy() && rightObj.isTruthy())
	case i.operator == "||":
		return objectFromBoolean(leftObj.isTruthy() || rightObj.isTruthy())
	// case leftObj.getType() != rightObj.getType():
	// 	return newObjectErr()
	default:
		return newObjectErr(i.tok.lineCol, "invalid operator for types: %s %s %s", leftObj.getType(), i.operator, rightObj.getType())
	}
}

func evalStringInfixExpression(leftObj object, operator string, rightObj object) object {
	l := leftObj.(*objectString).value
	r := rightObj.(*objectString).value
	switch operator {
	case "+":
		return &objectString{value: l + r}
	case "==":
		return objectFromBoolean(l == r)
	case "!=":
		return objectFromBoolean(l != r)
	default:
		return newObjectErrWithoutLC("invalid operator for types: %s %s %s", leftObj.getType(), operator, rightObj.getType())
	}
}

func evalNumberInfixExpression(leftObj object, operator string, rightObj object) object {
	leftNum, err := objectNumberToFloat64(leftObj)
	if err != nil {
		return newObjectErrWithoutLC("invalid number on left side of expression")
	}
	rightNum, err := objectNumberToFloat64(rightObj)
	if err != nil {
		return newObjectErrWithoutLC("invalid number on right side of expression")
	}

	areBothInteger := leftObj.getType() == t_integer && rightObj.getType() == t_integer
	if slices.Contains([]string{"+", "-", "*", "/"}, operator) {
		return objHandleMathOperation(leftNum, operator, rightNum, areBothInteger)
	}

	switch operator {
	case "%":
		if !areBothInteger {
			return newObjectErrWithoutLC("invalid operator for input types: %s %s %s", leftObj.getType(), operator, rightObj.getType())
		}
		res := int64(leftNum) % int64(rightNum)
		return &objectInteger{value: res}
	case "<":
		return objectFromBoolean(leftNum < rightNum)
	case "<=":
		return objectFromBoolean(isFloatEqual(leftNum, rightNum) || leftNum < rightNum)
	case ">":
		return objectFromBoolean(leftNum > rightNum)
	case ">=":
		return objectFromBoolean(isFloatEqual(leftNum, rightNum) || leftNum > rightNum)
	case "==":
		return objectFromBoolean(isFloatEqual(leftNum, rightNum))
	case "!=":
		return objectFromBoolean(!isFloatEqual(leftNum, rightNum))
	default:
		return newObjectErrWithoutLC("unsupported operator: %s", operator)
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

func evalArrayInfixExpression(leftObj object, operator string, rightObj object) object {
	l := leftObj.(*objectArray).entries
	r := rightObj.(*objectArray).entries
	switch operator {
	case "+":
		combined := []object{}
		for _, add := range l {
			combined = append(combined, add.clone())
		}
		for _, add := range r {
			combined = append(combined, add.clone())
		}
		return &objectArray{entries: combined}
	default:
		return newObjectErrWithoutLC("unsupported operator for arrays: %s", operator)
	}
}

//
// index expression

func (i *indexExpression) eval(env *environment) object {
	identResult := i.left.eval(env)
	if isObjectErr(identResult) {
		return unWrapErr(i.left.token().lineCol, identResult)
	}

	if identResult == obj_global_null {
		return identResult
	}
	arrObj, ok := identResult.(*objectArray)
	if !ok {
		return newObjectErr(i.left.token().lineCol, "cannot call index expression on non-array object %q. object type is %s", i.left.string(), identResult.getType())
	}

	indexObj := i.index.eval(env)
	if isObjectErr(indexObj) {
		return unWrapErr(i.index.token().lineCol, indexObj)
	}
	if indexObj.getType() != t_integer {
		return newObjectErr(i.index.token().lineCol, "index is not of type %s. got=%s", t_integer, indexObj.getType())
	}
	idxInt, ok := indexObj.(*objectInteger)
	if !ok {
		return newObjectErr(i.index.token().lineCol, "index is not of type %s. got=%s", t_integer, indexObj.getType())
	}

	targetIdx := int(idxInt.value)
	if int(targetIdx) >= len(arrObj.entries) || targetIdx < 0 {
		return newObjectErr(i.index.token().lineCol, "index is out of range for target array")
	}
	return arrObj.entries[targetIdx]
}

//
//arrow expr

func (a *arrowFunctionExpression) eval(env *environment) object {
	return &objectArrowFunction{
		paramName:  a.paramName.value,
		statements: a.block,
		functions:  env.functions,
	}
}

//
// string lit

func (s *stringLiteral) eval(env *environment) object {
	return &objectString{value: s.value}
}

//
// int lit

func (i *integerLiteral) eval(env *environment) object {
	return &objectInteger{value: i.value}
}

//
// float lit

func (f *floatLiteral) eval(env *environment) object {
	return &objectFloat{value: f.value}
}

//
// boolean lit

func (b *booleanLiteral) eval(env *environment) object {
	if b.value {
		return obj_global_true
	} else {
		return obj_global_false
	}
}

// null lit
func (n *nullLiteral) eval(env *environment) object {
	return obj_global_null
}

//
// map lit

func (m *mapLiteral) eval(env *environment) object {
	objPairs := make(map[string]object)
	for key, expr := range m.pairs {
		objectToAdd := expr.eval(env)
		if isObjectErr(objectToAdd) {
			return unWrapErr(expr.token().lineCol, objectToAdd)
		}
		objPairs[key] = objectToAdd
	}
	return &objectMap{kvPairs: objPairs}
}

func (a *arrayLiteral) eval(env *environment) object {
	objEntries := []object{}
	for _, entryExpr := range a.entries {
		toAdd := entryExpr.eval(env)
		if isObjectErr(toAdd) {
			return unWrapErr(entryExpr.token().lineCol, toAdd)
		}
		objEntries = append(objEntries, toAdd)
	}
	return &objectArray{entries: objEntries}
}

// helpers

func objectFromBoolean(b bool) *objectBoolean {
	if b {
		return obj_global_true
	} else {
		return obj_global_false
	}
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

func newObjectErrWithoutLC(s string, fmtArgs ...interface{}) *objectError {
	return &objectError{
		lineCol: "",
		message: fmt.Sprintf(s, fmtArgs...),
	}
}
func newObjectErr(lc string, s string, fmtArgs ...interface{}) *objectError {
	return &objectError{
		lineCol: lc,
		message: fmt.Sprintf(s, fmtArgs...),
	}
}

func isObjectErr(o object) bool {
	return o.getType() == t_error
}

// takes a nested error and attaches a linecol if one does not already exist
// used to "bubble up" error messages from things like internal functions or helpers to the calling AST node
func unWrapErr(outerLC string, innerObj object) object {
	objErr, ok := innerObj.(*objectError)
	if !ok {
		return innerObj
	}
	if len(objErr.lineCol) == 0 {
		return newObjectErr(outerLC, objErr.message)
	}
	return objErr
}

func objectToError(o object) (err error) {
	if e, ok := o.(*objectError); ok {
		return fmt.Errorf(e.message)
	}
	return nil
}
