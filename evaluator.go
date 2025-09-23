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
			return obj
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
		return valToSet
	}
	valToSet = valToSet.clone()

	var objHandle object // reference to object at current path. may be unused in instances where we're just assigning a regular variable without dot-path syntax
	currentPath := s.target.toAssignPath()
	for currentPath != nil {
		switch currentPath.stepType {
		case assign_step_env:
			objHandle = evalSetStatementHandleENV(currentPath, valToSet, env)
			if isObjectErr(objHandle) {
				return objHandle
			}
		case assign_step_map_key:
			objHandle = evalSetStatementHandleMAP(objHandle, currentPath, valToSet, s.target)
			if isObjectErr(objHandle) {
				return objHandle
			}
		default:
			return newObjectErr("%s: invalid path part for SET statement", s.target.token().lineCol)
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
	} else {
		return existing
	}
}

func evalSetStatementHandleMAP(objHandle object, current *assignPath, valToSet object, setTarget assignable) object {
	mapObj, ok := objHandle.(*objectMap)
	if !ok {
		return newObjectErr("%s: invalid path part for SET statement: cannot use a path expression on a non-map object", setTarget.token().lineCol)
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
	return existing
}

//
// when statement

func (w *whenStatement) eval(env *environment) object {
	conditionObj := w.condition.eval(env)
	if isObjectErr(conditionObj) {
		return conditionObj
	}
	if conditionObj.isTruthy() {
		res := w.consequence.eval(env)
		if isObjectErr(res) {
			return res
		}
	}
	return obj_global_null
}

//
//expression statement

func (e *expressionStatement) eval(env *environment) object {
	return e.expression.eval(env)
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
			return newObjectErr(err.Error())
		}
	case *pathExpression:
		fnEntry, err = evalFunctionNamePath(v, env)
		if err != nil {
			return newObjectErr(err.Error())
		}
	}
	args := []object{}
	for _, argExpr := range c.arguments {
		toAdd := argExpr.eval(env)
		args = append(args, toAdd)
	}
	ret := fnEntry.eval(args...)
	if isObjectErr(ret) {
		return newObjectErr("%s: %s", c.name.token().lineCol, objectToError(ret).Error())
	}
	return ret
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

//
// identifier expression

func (i *identifierExpression) eval(env *environment) object {
	if res, ok := env.get(i.value); ok {
		return res
	}
	return newObjectErr("%s: identifier not found: %s", i.token().lineCol, i.value)
}

// path expression
func (p *pathExpression) eval(env *environment) object {
	switch v := p.attribute.(type) {
	case *stringLiteral:
		return evalResolvePathEntryForKey(p, v.value, env)
	case *identifierExpression:
		return evalResolvePathEntryForKey(p, v.value, env)
	default:
		return newObjectErr("%s: invalid path part: %s", v.token().lineCol, v.string())
	}
}

func evalResolvePathEntryForKey(pathExpr *pathExpression, key string, env *environment) object {
	leftObj := pathExpr.left.eval(env)
	if isObjectErr(leftObj) {
		return leftObj
	}
	leftMap, ok := leftObj.(*objectMap)
	if !ok {
		return newObjectErr("%s: cannot access a path on a non-map object", pathExpr.left.token().lineCol)
	}
	res, ok := leftMap.kvPairs[key]
	if !ok {
		return newObjectErr("%s: key not found: %s", pathExpr.left.token().lineCol, key)
	}
	return res
}

// template expr
func (t *templateExpression) eval(env *environment) object {
	stringParts := []string{}
	for _, entry := range t.parts {
		res := entry.eval(env)
		if isObjectErr(res) {
			return res
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
		return rightObj
	}

	switch p.operator {
	case "!":
		return evalHandlePrefixExclamation(p, rightObj)
	case "-":
		return evalHandlePrefixMinus(p, rightObj)
	default:
		return newObjectErr("%s: unknown operator: %s", p.tok.lineCol, p.operator)
	}
}
func evalHandlePrefixExclamation(rightExpr *prefixExpression, rightObj object) object {
	switch rightObj {
	case obj_global_false:
		return obj_global_true
	case obj_global_true:
		return obj_global_false
	default:
		return newObjectErr("%s: incompatible non-boolean right-side exprssion for operator: !%s", rightExpr.tok.lineCol, rightExpr.string())
	}
}
func evalHandlePrefixMinus(rightExpr *prefixExpression, rightObj object) object {
	switch v := rightObj.(type) {
	case *objectInteger:
		return &objectInteger{value: -v.value}
	case *objectFloat:
		return &objectFloat{value: -v.value}
	default:
		return newObjectErr("%s: incompatible non-numeric right-side expression for operator: -%s", rightExpr.tok.lineCol, rightExpr.string())
	}
}

//
// infix expr

func (i *infixExpression) eval(env *environment) object {
	leftObj := i.left.eval(env)
	if isObjectErr(leftObj) {
		return leftObj
	}
	rightObj := i.right.eval(env)
	if isObjectErr(rightObj) {
		return rightObj
	}
	switch {
	case slices.Contains([]objectType{t_integer, t_float}, leftObj.getType()) && slices.Contains([]objectType{t_integer, t_float}, rightObj.getType()):
		ret := evalNumberInfixExpression(leftObj, i.operator, rightObj)
		if isObjectErr(ret) {
			errObj := ret.(*objectError)
			return newObjectErr("%s: %s", i.tok.lineCol, errObj.message)
		}
		return ret
	case leftObj.getType() != rightObj.getType():
		return newObjectErr("%s: type mismatch: %s %s %s", leftObj.getType(), i.tok.lineCol, i.operator, rightObj.getType())
	case leftObj.getType() == t_string && rightObj.getType() == t_string:
		ret := evalStringInfixExpression(leftObj, i.operator, rightObj)
		if isObjectErr(ret) {
			errObj := ret.(*objectError)
			return newObjectErr("%s: %s", i.tok.lineCol, errObj.message)
		}
		return ret
	case i.operator == "==":
		return objectFromBoolean(leftObj == rightObj)
	case i.operator == "!=":
		return objectFromBoolean(leftObj != rightObj)
	default:
		return newObjectErr("%s invalid operator for types: %s %s %s", leftObj.getType(), i.tok.lineCol, i.operator, rightObj.getType())
	}
}

func evalStringInfixExpression(leftObj object, operator string, rightObj object) object {
	l := leftObj.(*objectString).value
	r := rightObj.(*objectString).value
	if operator != "+" {
		return newObjectErr("invalid operator for types: %s %s %s", leftObj.getType(), operator, rightObj.getType())
	}

	return &objectString{value: l + r}
}

func evalNumberInfixExpression(leftObj object, operator string, rightObj object) object {
	leftNum, err := objectNumberToFloat64(leftObj)
	if err != nil {
		return newObjectErr("invalid number on left side of expression")
	}
	rightNum, err := objectNumberToFloat64(rightObj)
	if err != nil {
		return newObjectErr("invalid number on right side of expression")
	}

	areBothInteger := leftObj.getType() == t_integer && rightObj.getType() == t_integer
	if slices.Contains([]string{"+", "-", "*", "/"}, operator) {
		return objHandleMathOperation(leftNum, operator, rightNum, areBothInteger)
	}

	switch operator {
	case "%":
		if !areBothInteger {
			return newObjectErr("invalid operator for input types: %s %s %s", leftObj.getType(), operator, rightObj.getType())
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
		return newObjectErr("unsupported operator: %s", operator)
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

//
// index expression

func (i *indexExpression) eval(env *environment) object {
	identResult := i.left.eval(env)
	if isObjectErr(identResult) {
		return identResult
	}
	arrObj, ok := identResult.(*objectArray)
	if !ok {
		return newObjectErr("%s: cannot call index expression on non-array object", i.left.token().lineCol)
	}

	indexObj := i.index.eval(env)
	if isObjectErr(indexObj) {
		return indexObj
	}
	if indexObj.getType() != t_integer {
		return newObjectErr("%s: index is not of type %s. got=%s", i.index.token().lineCol, t_integer, indexObj.getType())
	}
	idxInt, ok := indexObj.(*objectInteger)
	if !ok {
		return newObjectErr("%s: index is not of type %s. got=%s", i.index.token().lineCol, t_integer, indexObj.getType())
	}

	targetIdx := int(idxInt.value)
	if int(targetIdx) >= len(arrObj.entries) || targetIdx < 0 {
		return newObjectErr("%s: index is out of range for target array", i.index.token().lineCol)
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

//
// map lit

func (m *mapLiteral) eval(env *environment) object {
	objPairs := make(map[string]object)
	for key, expr := range m.pairs {
		objectToAdd := expr.eval(env)
		if isObjectErr(objectToAdd) {
			return objectToAdd
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
			return toAdd
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

func newObjectErr(s string, fmtArgs ...interface{}) *objectError {
	return &objectError{
		message: fmt.Sprintf(s, fmtArgs...),
	}
}

func isObjectErr(o object) bool {
	return o.getType() == t_error
}

func objectToError(o object) error {
	if e, ok := o.(*objectError); ok {
		return fmt.Errorf(e.message)
	}
	return nil
}
