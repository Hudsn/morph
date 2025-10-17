package morph

import (
	"testing"
)

func TestParsePipeCall(t *testing.T) {
	input := "asdf | myfunc(2)"
	program := setupParserTest(t, input)
	checkParserProgramLength(t, program, 1)
	checkParserStatementType(t, program.statements[0], EXPRESSION_STATEMENT)
	stmt := program.statements[0].(*expressionStatement)
	ce, ok := stmt.expression.(*callExpression)
	if !ok {
		t.Fatalf("stmt.expression is not of type *callExpression. got=%T", stmt.expression)
	}
	if len(ce.arguments) != 2 {
		t.Fatalf("expected ce.arguments to be of length 2. got=%d", len(ce.arguments))
	}
	arg1, ok := ce.arguments[0].(*identifierExpression)
	if !ok {
		t.Fatalf("ce.arguments[0] is not of type *identifierExpression. got=%T", ce.arguments[0])
	}
	if arg1.value != "asdf" {
		t.Errorf("wrong value for left argument. want=%s got=%s", "asdf", arg1.value)
	}
	arg2, ok := ce.arguments[1].(*integerLiteral)
	if !ok {
		t.Fatalf("ce.arguments[1] is not of type *integerLiteral. got=%T", ce.arguments[1])
	}
	if arg2.value != int64(2) {
		t.Errorf("expected arg of function to be 2. got=%d", arg2.value)
	}
}

func TestParseNullLiteral(t *testing.T) {
	input := "null"
	program := setupParserTest(t, input)
	checkParserProgramLength(t, program, 1)
	checkParserStatementType(t, program.statements[0], EXPRESSION_STATEMENT)
	stmt := program.statements[0].(*expressionStatement)
	_, ok := stmt.expression.(*nullLiteral)
	if !ok {
		t.Fatalf("stmt.expression is not of type *nullLiteral. got=%T", stmt.expression)
	}
}

func TestParseArrowFunction(t *testing.T) {
	input := `myvar ~> {
		set innervar = 10
		randomFunc()
		set inner2 = otherFunc(innervar)
	}`
	program := setupParserTest(t, input)
	checkParserProgramLength(t, program, 1)
	checkParserStatementType(t, program.statements[0], EXPRESSION_STATEMENT)
	stmt := program.statements[0].(*expressionStatement)
	arrowFnExp, ok := stmt.expression.(*arrowFunctionExpression)
	if !ok {
		t.Fatalf("stmt.expression is not of type *arrowFunctionExpression. got=%T", stmt.expression)
	}
	if len(arrowFnExp.block) != 3 {
		t.Errorf("expected length of arrowFunc statements to be 2. got=%d", len(arrowFnExp.block))
	}
	if arrowFnExp.paramName.value != "myvar" {
		t.Errorf("expected arrowFnExp value to be %s. got=%s", "myvar", arrowFnExp.paramName.value)
	}
}

func TestParseFunctionCall(t *testing.T) {
	input := `myfunc(ident1, "three", 6)`
	program := setupParserTest(t, input)
	checkParserProgramLength(t, program, 1)
	checkParserStatementType(t, program.statements[0], EXPRESSION_STATEMENT)
	stmt := program.statements[0].(*expressionStatement)
	callExpr, ok := stmt.expression.(*callExpression)
	if !ok {
		t.Fatalf("stmt.expression is not of type *callExpression. got=%T", stmt.expression)
	}
	tests := []interface{}{
		"ident1",
		"three",
		6,
	}
	if len(callExpr.arguments) != len(tests) {
		t.Fatalf("expected arguments to be len %d. got=%d", len(tests), len(callExpr.arguments))
	}
	for idx, want := range tests {
		gotArg := callExpr.arguments[idx]
		testLiteralExpression(t, gotArg, want)
	}
}

func TestParseIndexExpression(t *testing.T) {
	input := "myArray[3+2]"
	program := setupParserTest(t, input)
	checkParserProgramLength(t, program, 1)
	checkParserStatementType(t, program.statements[0], EXPRESSION_STATEMENT)
	stmt := program.statements[0].(*expressionStatement)
	indexExpr, ok := stmt.expression.(*indexExpression)
	if !ok {
		t.Fatalf("stmt.expression is not of type *indexExpression. got=%T", stmt.expression)
	}
	testIdentifierExpression(t, indexExpr.left, "myArray")
	testInfixExpression(t, indexExpr.index, 3, "+", 2)

	input = "myArray[0]"
	program = setupParserTest(t, input)
	checkParserProgramLength(t, program, 1)
	checkParserStatementType(t, program.statements[0], EXPRESSION_STATEMENT)
	stmt = program.statements[0].(*expressionStatement)
	indexExpr, ok = stmt.expression.(*indexExpression)
	if !ok {
		t.Fatalf("stmt.expression is not of type *indexExpression. got=%T", stmt.expression)
	}
	testIdentifierExpression(t, indexExpr.left, "myArray")
	testLiteralExpression(t, indexExpr.index, 0)
}

func TestParseMapLiteral(t *testing.T) {
	input := `{"key1": 1+1, "key2": {"nested1": 1 * 1, "nested2": "nested value"}, "key3": [1]}`
	program := setupParserTest(t, input)
	checkParserProgramLength(t, program, 1)
	checkParserStatementType(t, program.statements[0], EXPRESSION_STATEMENT)
	stmt := program.statements[0].(*expressionStatement)
	mapLit, ok := stmt.expression.(*mapLiteral)
	if !ok {
		t.Fatalf("stmt.expression is not of type *mapLiteral. got=%T", stmt.expression)
	}
	if len(mapLit.pairs) != 3 {
		t.Errorf("expected len of map literal to be 3. got=%d", len(mapLit.pairs))
	}
	wantStr := `{"key1": (1 + 1), "key2": {"nested1": (1 * 1), "nested2": "nested value"}, "key3": [1]}`
	if mapLit.string() != wantStr {
		t.Errorf("wrong map value.\n\twant=%s\n\tgot=%s", wantStr, mapLit.string())
	}
}

func TestParseArrayLiteral(t *testing.T) {
	input := `[1, "two", 2 * 2, 3 + 3]`
	program := setupParserTest(t, input)
	checkParserProgramLength(t, program, 1)
	checkParserStatementType(t, program.statements[0], EXPRESSION_STATEMENT)
	stmt := program.statements[0].(*expressionStatement)
	arr, ok := stmt.expression.(*arrayLiteral)
	if !ok {
		t.Fatalf("stmt.expression is not of type *arrayLiteral. got=%T", stmt.expression)
	}
	if len(arr.entries) != 4 {
		t.Errorf("expected len of array literal to be 4. got=%d", len(arr.entries))
	}
	wantStr := `[1, "two", (2 * 2), (3 + 3)]`
	if arr.string() != wantStr {
		t.Errorf("wrong array value. want=%s got=%s", wantStr, arr.string())
	}
}

func TestParseTemplateLiteral(t *testing.T) {
	input := `'${my_var} world ${myother_var} ${myLastVar}'`
	program := setupParserTest(t, input)
	checkParserProgramLength(t, program, 1)
	checkParserStatementType(t, program.statements[0], EXPRESSION_STATEMENT)
	stmt := program.statements[0].(*expressionStatement)
	template, ok := stmt.expression.(*templateExpression)
	if !ok {
		t.Fatalf("stmt.expression is not of type *templateExpression. got=%T", stmt.expression)
	}
	wantList := []string{
		"",
		"my_var",
		" world ",
		"myother_var",
		" ",
		"myLastVar",
		"",
	}
	if len(wantList) != len(template.parts) {
		t.Fatalf("expected template.parts to be of len %d. got=%d", len(wantList), len(template.parts))
	}
	for idx, want := range wantList {
		expr := template.parts[idx]
		testLiteralExpression(t, expr, want)
	}
}

func TestParseStringLiteral(t *testing.T) {
	input := `"hello world!"`
	program := setupParserTest(t, input)
	checkParserProgramLength(t, program, 1)
	checkParserStatementType(t, program.statements[0], EXPRESSION_STATEMENT)
	stmt := program.statements[0].(*expressionStatement)
	testLiteralExpression(t, stmt.expression, "hello world!")
}

func TestParseIfStatementMulti(t *testing.T) {
	input := `IF true :: {
		SET myvar = 5
		SET myothervar = 10
	}`
	program := setupParserTest(t, input)
	checkParserProgramLength(t, program, 1)
	checkParserStatementType(t, program.statements[0], IF_STATEMENT)
	stmt := program.statements[0].(*ifStatement)
	testLiteralExpression(t, stmt.condition, true)
	if len(stmt.consequence) != 2 {
		t.Fatalf("stmt.consequence is not of length 2. got=%d", len(stmt.consequence))
	}
	setStmt, ok := stmt.consequence[0].(*setStatement)
	if !ok {
		t.Fatalf("stmt.consequence[0] is not of type *setStatement. got=%T", stmt.consequence)
	}
	targetIdent, ok := setStmt.target.(*identifierExpression)
	if !ok {
		t.Fatalf("setStmt.target is not of type *identifierExpression. got=%T", setStmt.target)
	}
	testIdentifierExpression(t, targetIdent, "myvar")
	testLiteralExpression(t, setStmt.value, 5)
	setStmt, ok = stmt.consequence[1].(*setStatement)
	if !ok {
		t.Fatalf("stmt.consequence[1] is not of type *setStatement. got=%T", stmt.consequence)
	}
	targetIdent, ok = setStmt.target.(*identifierExpression)
	if !ok {
		t.Fatalf("setStmt.target is not of type *identifierExpression. got=%T", setStmt.target)
	}
	testIdentifierExpression(t, targetIdent, "myothervar")
	testLiteralExpression(t, setStmt.value, 10)
}
func TestParseIfStatement(t *testing.T) {
	input := "IF true :: SET myvar = 5"
	program := setupParserTest(t, input)
	checkParserProgramLength(t, program, 1)
	checkParserStatementType(t, program.statements[0], IF_STATEMENT)
	stmt := program.statements[0].(*ifStatement)
	testLiteralExpression(t, stmt.condition, true)
	if len(stmt.consequence) != 1 {
		t.Fatalf("stmt.consequence is not of length 1. got=%d", len(stmt.consequence))
	}
	setStmt, ok := stmt.consequence[0].(*setStatement)
	if !ok {
		t.Fatalf("stmt.consequence[0] is not of type *setStatement. got=%T", stmt.consequence)
	}
	targetIdent, ok := setStmt.target.(*identifierExpression)
	if !ok {
		t.Fatalf("setStmt.target is not of type *identifierExpression. got=%T", setStmt.target)
	}
	testIdentifierExpression(t, targetIdent, "myvar")
	testLiteralExpression(t, setStmt.value, 5)
}

func TestParseDelStatement(t *testing.T) {
	input := `DEL myvar."sub"`
	program := setupParserTest(t, input)
	checkParserProgramLength(t, program, 1)
	checkParserStatementType(t, program.statements[0], DEL_STATEMENT)
	stmt := program.statements[0].(*delStatement)
	path, ok := stmt.target.(*pathExpression)
	if !ok {
		t.Fatalf("stmt.target is not *pathExpression. got=%T", stmt.target)
	}
	subStr, ok := path.attribute.(*stringLiteral)
	if !ok {
		t.Fatalf("path.attribute is not *stringLiteral. got=%T", path.attribute)
	}
	testLiteralExpression(t, subStr, "sub")
	myvarIndent, ok := path.left.(*identifierExpression)
	if !ok {
		t.Fatalf("path.left is not *identifierExpression. got=%T", path.left)
	}
	testLiteralExpression(t, myvarIndent, "myvar")
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
		{assign_step_env, "my"},
		{assign_step_map_key, "path"},
		{assign_step_map_key, "var"},
	}
	curPath := assignPath
	for idx := 0; curPath != nil; idx++ {
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

func TestParseInfixExpression(t *testing.T) {
	infixTests := []struct {
		input      string
		leftValue  interface{}
		operator   string
		rightValue interface{}
	}{
		{"5 + 5", 5, "+", 5},
		{"5 - 5", 5, "-", 5},
		{"5 * 5", 5, "*", 5},
		{"5 / 5", 5, "/", 5},
		{"5 % 5", 5, "%", 5},
		{"5 > 5", 5, ">", 5},
		{"5 < 5", 5, "<", 5},
		{"5 == 5", 5, "==", 5},
		{"5 != 5", 5, "!=", 5},
		{"5 <= 5", 5, "<=", 5},
		{"5 >= 5", 5, ">=", 5},
		{"5 != 5", 5, "!=", 5},
		{"true == true", true, "==", true},
		{"true != false", true, "!=", false},
		{"false == false", false, "==", false},
	}
	for _, tt := range infixTests {
		program := setupParserTest(t, tt.input)
		checkParserProgramLength(t, program, 1)
		checkParserStatementType(t, program.statements[0], EXPRESSION_STATEMENT)
		stmt := program.statements[0].(*expressionStatement)
		testInfixExpression(t, stmt.expression, tt.leftValue, tt.operator, tt.rightValue)
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

func TestParseOperatorPrecedence(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{
			"5 + 5",
			"(5 + 5)",
		},
		{
			"5 + 5 * 5",
			"(5 + (5 * 5))",
		},
		{
			"5 + 5 * 5 - 5 / 5 % 5",
			"((5 + (5 * 5)) - ((5 / 5) % 5))",
		},
		{
			"5 * (-5 + 5)",
			"(5 * ((-5) + 5))",
		},
		{
			"-5 * 10 <= 5.12 % 3 - 1",
			"(((-5) * 10) <= ((5.12 % 3) - 1))",
		},
		{
			"!blue != false",
			"((!blue) != false)",
		},
		{
			"true || !false == false && true",
			"(true || (((!false) == false) && true))",
		},
		{
			"func(a + b) + x * y",
			"(func((a + b)) + (x * y))",
		},
	}
	for _, tt := range tests {
		program := setupParserTest(t, tt.input)
		got := program.string()
		if tt.want != got {
			t.Errorf("expected=%q got=%q", tt.want, got)
		}
	}
}

// sub-parser helpers

func testInfixExpression(t *testing.T, exp expression, left interface{}, operator string, right interface{}) bool {
	infix, ok := exp.(*infixExpression)
	if !ok {
		t.Errorf("exp is not *ast.InfixExpression. got=%T", exp)
		return false
	}
	if !testLiteralExpression(t, infix.left, left) {
		return false
	}
	if infix.operator != operator {
		t.Errorf("exp.Operator is not %s. got=%q", operator, infix.operator)
		return false
	}
	if !testLiteralExpression(t, infix.right, right) {
		return false
	}
	return true
}

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
		switch expr.(type) {
		case *identifierExpression:
			return testIdentifierExpression(t, expr, v)
		case *stringLiteral:
			return testStringLiteral(t, expr, v)
		default:
			t.Errorf("testLiteral expression type not supported for string. got=%T", value)
			return false
		}
	default:
		t.Errorf("testLiteralExpression type not supported. got=%T", value)
		return false
	}
}

func testStringLiteral(t *testing.T, expr expression, want string) bool {
	strLit, ok := expr.(*stringLiteral)
	if !ok {
		t.Errorf("expression is not of type *floatLiteral. got=%T", expr)
		return false
	}
	if strLit.value != want {
		t.Errorf("expected float expression to be equal to %s. got=%s", want, strLit.value)
		return false
	}
	return true
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
		// if we get here, our left node is NOT a path expr so we are probably at the first path part like an ident, string, or indexExprssion
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
		t.Fatalf("expected program statments to be length %d. got=%d", wantLen, len(program.statements))
	}
}

type statementType int

const (
	_ statementType = iota
	EXPRESSION_STATEMENT
	SET_STATEMENT
	IF_STATEMENT
	DEL_STATEMENT
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
	case DEL_STATEMENT:
		if _, ok := statement.(*delStatement); !ok {
			t.Fatalf("statement is not of type *delStatement. got=%T", statement)
		}
	case IF_STATEMENT:
		if _, ok := statement.(*ifStatement); !ok {
			t.Fatalf("statement is not of type *ifStatement. got=%T", statement)
		}
	default:
		t.Errorf("statment type not supported")
	}
}
