package morph

import "fmt"

var (
	OBJ_GLOBAL_NULL  = &objectNull{}
	OBJ_GLOBAL_TRUE  = &objectBoolean{value: true}
	OBJ_GLOBAL_FALSE = &objectBoolean{value: false}
)

type evaluator struct {
	parser *parser
}

func newEvaluator(p *parser) *evaluator {
	return &evaluator{parser: p}
}

func (e *evaluator) eval(astNode node, env *environment) object {
	switch astNode := astNode.(type) {
	case *setStatement:
		return e.evalSetStatement(astNode, env)
	case *expressionStatement:
		return e.eval(astNode.expression, env)
	case *pathExpression:
		return e.evalPathExpression(astNode, env)
	case *identifierExpression:
		return e.evalIdentifierExpression(astNode, env)
	case *integerLiteral:
		return &objectInteger{value: astNode.value}
	case *floatLiteral:
		return &objectFloat{value: astNode.value}
	case *booleanLiteral:
		return &objectBoolean{value: astNode.value}
	}
	return nil
}

func (e *evaluator) evalSetStatement(whenStmt *setStatement, env *environment) object {
	//TODO
	return nil
}

func (e *evaluator) evalIdentifierExpression(identExpr *identifierExpression, env *environment) object {
	if res, ok := env.get(identExpr.value); ok {
		return res
	}

	return objectNewErr("%s: identifier not found: %s", e.lineColForNode(identExpr), identExpr.value)
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
		return obj.getType() == T_ERROR
	}
	return false
}

//

func (e *evaluator) lineColForNode(n node) string {
	return lineColString(lineAndCol(e.parser.lexer.input, n.position().start))

}
