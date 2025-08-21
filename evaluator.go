package parser

var (
	OBJ_GLOBAL_NULL  = &objectNull{}
	OBJ_GLOBAL_TRUE  = &objectBoolean{value: true}
	OBJ_GLOBAL_FALSE = &objectBoolean{value: false}
)

// TODO
func runProgram(program program, env *environment) {

}

// TODO
func runExpression(statement expressionStatement, env *environment) object {
	return nil
}

func eval(astNode node, env *environment) object {
	switch astNode := astNode.(type) {
	case *expressionStatement:
		return eval(astNode.expression, env)
	case *integerLiteral:
		return &objectInteger{value: astNode.value}
	case *floatLiteral:
		return &objectFloat{value: astNode.value}
	case *booleanLiteral:
		return &objectBoolean{value: astNode.value}
	}
	return nil
}
