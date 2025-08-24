package morph

import "fmt"

var (
	OBJ_GLOBAL_NULL  = &objectNull{}
	OBJ_GLOBAL_TRUE  = &objectBoolean{value: true}
	OBJ_GLOBAL_FALSE = &objectBoolean{value: false}
)

type evaluator interface {
	eval(env *environment) object
}

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
	// TODO
	return nil
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
