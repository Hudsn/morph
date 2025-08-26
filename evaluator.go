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

func (e *evaluator) evalSetStatement(setStmt *setStatement, env *environment) object {
	valToSet := e.eval(setStmt.value, env)
	if objectIsError(valToSet) {
		return valToSet
	}
	var objHandle object // reference to object at current path. may be unused in instances where we're just assigning a regular variable without dot-path syntax
	nextPath := setStmt.target.toAssignPath()
	for nextPath != nil {
		currentPath := nextPath
		nextPath = currentPath.next

		switch currentPath.stepType {
		case ASSIGN_STEP_ENV:
			objHandle = e.setStatementHandleENV(objHandle, currentPath, valToSet, env)
		case ASSIGN_STEP_MAP_KEY:
			objHandle = e.setStatementHandleMAP(objHandle, currentPath, valToSet, setStmt.target)
			if objectIsError(objHandle) {
				return objHandle
			}
		default:
			return objectNewErr("%s: invalid path part for SET statement", e.lineColForNode(setStmt.target))
		}
	}
	return OBJ_GLOBAL_NULL
}

func (e *evaluator) setStatementHandleENV(objHandle object, current *assignPath, valToSet object, env *environment) object {
	if current.next == nil {
		env.set(current.partName, valToSet)
		return OBJ_GLOBAL_NULL
	}
	existing, ok := env.get(current.partName)
	if !ok {
		newMap := &objectMap{kvPairs: make(map[string]objectMapPair)}
		env.set(current.partName, newMap)
		return newMap
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
		return OBJ_GLOBAL_NULL
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
