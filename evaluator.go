package morph

import "fmt"

var (
	OBJ_GLOBAL_NULL  = &objectNull{}
	OBJ_GLOBAL_TRUE  = &objectBoolean{value: true}
	OBJ_GLOBAL_FALSE = &objectBoolean{value: false}
)

func eval(astNode node, env *environment) object {
	switch astNode := astNode.(type) {
	case *setStatement:
		return evalSetStatement(astNode, env)
	case *expressionStatement:
		return eval(astNode.expression, env)
	case *pathExpression:
		return evalPathExpression(astNode, env)
	case *integerLiteral:
		return &objectInteger{value: astNode.value}
	case *floatLiteral:
		return &objectFloat{value: astNode.value}
	case *booleanLiteral:
		return &objectBoolean{value: astNode.value}
	}
	return nil
}

func evalSetStatement(whenStmt *setStatement, env *environment) object {
	//TODO
	return nil
}

func evalPathExpression(pathExpr *pathExpression, env *environment) object {
	if len(pathExpr.parts) < 1 {
		return objectNewErr("%d:%d: invalid path expression: %s", pathExpr.position().start, pathExpr.position().end, pathExpr.string())
	}
	obj := eval(pathExpr.parts[0], env)
	for _, part := range pathExpr.parts[1:] {
		obj = evalObjectPart(obj, part)
	}
	return obj
}

func evalObjectPart(obj object, part pathPart) object {
	switch v := part.(type) {
	case *identifierExpression:
		return evalObjectPartIdentifier(obj, v)
	default:
		return objectNewErr("%d:%d: invalid path part: %s", part.position().start, part.position().end, part.string())
	}
}

func evalObjectPartMapString(obj object, tryKey string) object {
	objMap, ok := obj.(*objectMap)
	if !ok {
		return objectNewErr("attempted attribute access on a non-map: %s", tryKey)
	}
	ret, ok := objMap.kvPairs[tryKey]
	if !ok {
		return objectNewErr("map key does not exist: %s", tryKey)
	}
	return ret.value
}

func evalObjectPartIdentifier(obj object, ident *identifierExpression) object {
	ret := evalObjectPartMapString(obj, ident.value)
	if objectIsError(ret) {
		return objectNewErr("%s:%s: %s", ident.position().start, ident.position().end, ret.inspect())
	}
	return ret
}

// err helpers

func objectNewErr(format string, a ...interface{}) *objectError {
	return &objectError{message: fmt.Sprintf(format, a...)}
}

func objectIsError(obj object) bool {
	if obj != nil {
		return obj.getType() == T_ERROR
	}
	return false
}
