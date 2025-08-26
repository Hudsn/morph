package morph

import (
	"fmt"
	"testing"
)

func TestParseWhenStatement(t *testing.T) {
	input := "WHEN true :: SET myvar = 5"
	program := setupParserTest(t, input)
	checkParserProgramLength(t, program, 1)
	checkParserStatementType(t, program.statements[0], WHEN_STATEMENT)
	stmt := program.statements[0].(*whenStatement)
	testLiteralExpression(t, stmt.condition, true)
	setStmt, ok := stmt.consequence.(*setStatement)
	if !ok {
		t.Fatalf("stmt.consequence is not of type *setStatement. got=%T", stmt.consequence)
	}
	targetIdent, ok := setStmt.target.(*identifierExpression)
	if !ok {
		t.Fatalf("setStmt.target is not of type *identifierExpression. got=%T", setStmt.target)
	}
	testIdentifierExpression(t, targetIdent, "myvar")
	testLiteralExpression(t, setStmt.value, 5)
}

func TestParseSetStatement(t *testing.T) {
	input := "SET myvar = 5"
	program := setupParserTest(t, input)
	checkParserProgramLength(t, program, 1)
	checkParserStatementType(t, program.statements[0], SET_STATEMENT)
	stmt := program.statements[0].(*setStatement)
	ident, ok := stmt.target.(*identifierExpression)
	if !ok {
		t.Errorf("stmt.target is not *identifierExpression. got=%T", stmt.target)
	}
	testIdentifierExpression(t, ident, "myvar")

	input = "SET my.path.var = true"
	program = setupParserTest(t, input)
	checkParserProgramLength(t, program, 1)
	checkParserStatementType(t, program.statements[0], SET_STATEMENT)
	stmt = program.statements[0].(*setStatement)
	assignPath := stmt.target.toAssignPath()
	// steps := stmt.target.assignPathSteps()
	wantSteps := []struct {
		stepType assignStepType
		partName string
	}{
		{ASSIGN_STEP_ENV, "my"},
		{ASSIGN_STEP_MAP_KEY, "path"},
		{ASSIGN_STEP_MAP_KEY, "var"},
	}
	curPath := assignPath
	for idx := 0; curPath != nil; idx++ {
		fmt.Printf("%+v\n\n", curPath)
		if idx >= len(wantSteps) {
			t.Fatalf("too many path parts. expected=%d got=%d", len(wantSteps), idx+1)
		}
		want := wantSteps[idx]
		if want.stepType != curPath.stepType {
			t.Errorf("wrong path stepType at test index %d: want=%s got=%s", idx, want.stepType, curPath.stepType)
		}
		if want.partName != curPath.partName {
			t.Errorf("wrong path part name at test index %d: want=%s got=%s", idx, want.partName, curPath.partName)
		}
		curPath = curPath.next
	}
	path, ok := stmt.target.(*pathExpression)
	if !ok {
		t.Errorf("stmt.target is not *pathExpression. got=%T", stmt.target)
	}
	want := []interface{}{
		"my",
		"path",
		"var",
	}
	testPathExpression(t, path, want)

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
	case string:
		return testIdentifierExpression(t, expr, v)
	default:
		t.Errorf("testLiteralExpression type not supported. got=%T", value)
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
		t.Errorf("expression is not of type *identifierExpression. got=%T", expr)
		return false
	}
	if ident.value != want {
		t.Errorf("expected identifier string to be equal to %q. got=%q", want, ident.value)
		return false
	}
	return true
}

func testPathExpression(t *testing.T, expr expression, wantList []interface{}) bool {
	path, ok := expr.(*pathExpression)
	if !ok {
		t.Errorf("expression is not of type *pathExpression. got=%T", expr)
		return false
	}
	idx := 0
	for i := len(wantList) - 1; i >= 0; i-- {
		wantVal := wantList[i]
		gotAttr, ok := path.attribute.(expression)
		if !ok {
			t.Fatalf("testPathExpression case idx=%d path.attribute is not of type expression. got=%T", idx, path.attribute)
			return false
		}
		testLiteralExpression(t, gotAttr, wantVal)
		idx++
		if _, ok := path.left.(*pathExpression); ok {
			path = path.left.(*pathExpression)
			continue
		}
		// if we get here, out left node is NOT a path expr so we are probably at the first path part like an ident, string, or indexExprssion
		i--
		wantVal = wantList[i]
		final, ok := path.left.(expression)
		if !ok {
			t.Fatalf("testPathExpression case idx=%d path.left is not of type expression. got=%T", idx, path.attribute)
			return false
		}
		testLiteralExpression(t, final, wantVal)
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
	WHEN_STATEMENT
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
	case WHEN_STATEMENT:
		if _, ok := statement.(*whenStatement); !ok {
			t.Fatalf("statement is not of type *whenStatement. got=%T", statement)
		}
	default:
		t.Errorf("statment type not supported")
	}
}
