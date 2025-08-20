package parser

import "testing"

func TestParseSetStatement(t *testing.T) {
	input := "SET myvar = 5"
	program := setupParserTest(t, input)
	checkParserProgramLength(t, program, 1)
	checkParserStatementType(t, program.statements[0], SET_STATEMENT)
	stmt := program.statements[0].(*setStatement)
	ident, ok := stmt.target.(*identifierExpression)
	if ok {
		testIdentifierExpression(t, ident, "myvar")
	}

	input = "SET my.path = true"
	program = setupParserTest(t, input)
	checkParserProgramLength(t, program, 1)
	checkParserStatementType(t, program.statements[0], SET_STATEMENT)
	stmt = program.statements[0].(*setStatement)
	path, ok := stmt.target.(*pathExpression)
	if ok {
		// testPathExpression(t, path)
	}

}

func TestParsePrefix(t *testing.T) {
	tests := []struct {
		input    string
		operator string
		right    interface{}
	}{
		{"-5", "-", 5},
		{"!5", "!", 5},
		{"-.123", "-", .123},
	}

	for _, tt := range tests {
		program := setupParserTest(t, tt.input)
		checkParserProgramLength(t, program, 1)
		checkParserStatementType(t, program.statements[0], EXPRESSION_STATEMENT)
		exprStatement := program.statements[0].(*expressionStatement)
		testPrefixExpression(t, exprStatement.expression, tt.operator, tt.right)
	}
}

func TestParseNumbers(t *testing.T) {
	tests := []struct {
		input string
		want  interface{}
	}{
		{"5", 5},
		{"1.234", 1.234},
		{".111", 0.111},
	}
	for _, tt := range tests {
		program := setupParserTest(t, tt.input)
		checkParserProgramLength(t, program, 1)
		checkParserStatementType(t, program.statements[0], EXPRESSION_STATEMENT)
		exprStatment := program.statements[0].(*expressionStatement)
		testLiteralExpression(t, exprStatment.expression, tt.want)
	}
}

func TestParseBooleans(t *testing.T) {
	tests := []struct {
		input string
		want  interface{}
	}{
		{"true", true},
		{"false", false},
	}
	for _, tt := range tests {
		program := setupParserTest(t, tt.input)
		checkParserProgramLength(t, program, 1)
		checkParserStatementType(t, program.statements[0], EXPRESSION_STATEMENT)
		exprStatment := program.statements[0].(*expressionStatement)
		testLiteralExpression(t, exprStatment.expression, tt.want)
	}
}

// sub-parser helpers

func testPrefixExpression(t *testing.T, expr expression, wantOperator string, wantRight interface{}) bool {
	prefixExpr, ok := expr.(*prefixExpression)
	if !ok {
		t.Fatalf("exprStatement.expression is not type *prefixExpression. got=%T", expr)
	}
	if prefixExpr.operator != wantOperator {
		t.Errorf("expected prefix operator to be %s. got=%s", wantOperator, prefixExpr.operator)
		return false
	}
	return testLiteralExpression(t, prefixExpr.right, wantRight)
}

func testLiteralExpression(t *testing.T, expr expression, value interface{}) bool {
	switch v := value.(type) {
	case int:
		return testIntegerLiteral(t, expr, int64(v))
	case int32:
		return testIntegerLiteral(t, expr, int64(v))
	case int64:
		return testIntegerLiteral(t, expr, v)
	case float32:
		return testFloatLiteral(t, expr, float64(v))
	case float64:
		return testFloatLiteral(t, expr, v)
	case bool:
		return testBooleanLiteral(t, expr, v)
	default:
		return false
	}
}

func testFloatLiteral(t *testing.T, expr expression, want float64) bool {
	floatLit, ok := expr.(*floatLiteral)
	if !ok {
		t.Errorf("expression is not of type *floatLiteral. got=%T", expr)
		return false
	}
	if floatLit.value != want {
		t.Errorf("expected float expression to be equal to %f. got=%f", want, floatLit.value)
		return false
	}
	return true
}

func testIntegerLiteral(t *testing.T, expr expression, want int64) bool {
	intLit, ok := expr.(*integerLiteral)
	if !ok {
		t.Errorf("expression is not of type *integerLiteral. got=%T", expr)
		return false
	}

	if intLit.value != want {
		t.Errorf("expected integer expression to be equal to %d. got=%d", want, intLit.value)
		return false
	}
	return true
}

func testBooleanLiteral(t *testing.T, expr expression, want bool) bool {
	boolLit, ok := expr.(*booleanLiteral)
	if !ok {
		t.Errorf("expression is not of type *booleanLiteral. got=%T", expr)
		return false
	}

	if boolLit.value != want {
		t.Errorf("expected integer expression to be equal to %t. got=%t", want, boolLit.value)
		return false
	}
	return true
}

func testIdentifierExpression(t *testing.T, expr expression, want string) bool {
	ident, ok := expr.(*identifierExpression)
	if !ok {
		t.Errorf("expression is not of type *identifierExpression. got=%T")
		return false
	}
	if ident.value != want {
		t.Errorf("expected identifier string to be equal to %s. got=%s", want, ident.value)
		return false
	}
	return true
}

// setup helpers

func setupParserTest(t *testing.T, input string) *program {
	l := newLexer([]rune(input))
	p := newParser(l)
	program, err := p.parseProgram()
	if err != nil {
		t.Fatalf("parsing error: %s", err.Error())
		return nil
	}
	return program
}

func checkParserProgramLength(t *testing.T, program *program, wantLen int) {
	if len(program.statements) != wantLen {
		t.Fatalf("expected program statments to be length %d. got=%d", len(program.statements), wantLen)
	}
}

type statementType int

const (
	_ statementType = iota
	EXPRESSION_STATEMENT
	SET_STATEMENT
)

func checkParserStatementType(t *testing.T, statement statement, stype statementType) {
	switch stype {
	case EXPRESSION_STATEMENT:
		if _, ok := statement.(*expressionStatement); !ok {
			t.Fatalf("statment is not of type *expressionStatement. got=%T", statement)
		}
	case SET_STATEMENT:
		if _, ok := statement.(*setStatement); !ok {
			t.Fatalf("statement is not of type *setStatement. got=%T", statement)
		}
	default:
		t.Errorf("statment type not supported")
	}
}
