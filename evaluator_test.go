package morph

import (
	"testing"
)

func TestEvalMaps(t *testing.T) {
	input := `
	set msgparts = {"h": "hello", "w": "world", "num": 2 + 2}
	set string = msgparts.h + " " + msgparts.w
	set num = msgparts.num
	`
	tests := []struct {
		key  string
		want interface{}
	}{
		{"string", "hello world"},
		{"num", 4},
	}
	env := newEnvironment(nil)
	parser := setupEvalTestParser(input)
	if len(parser.errors) > 0 {
		t.Fatalf("parser error: %s", parser.errors[0])
	}
	program, err := parser.parseProgram()
	if err != nil {
		t.Fatal(err)
	}
	_, err = eval(program, env)
	if err != nil {
		t.Fatal(err)
	}
	for _, tt := range tests {
		got, found := env.get(tt.key)
		if !found {
			t.Fatalf("wanted result for env key %q but got no result", tt.key)
		}
		switch want := tt.want.(type) {
		case int:
			gotInt, ok := got.(*objectInteger)
			if !ok {
				t.Fatalf("result is not of type *objectInteger. got=%T", got)
			}
			if want != int(gotInt.value) {
				t.Errorf("expected value for key %q to be %d. got=%d", tt.key, want, gotInt.value)
			}
		case string:
			gotStr, ok := got.(*objectString)
			if !ok {
				t.Fatalf("result is not of type *objectString. got=%T", got)
			}
			if want != gotStr.value {
				t.Errorf("expected value for key %q to be %s. got=%s", tt.key, want, gotStr.value)
			}
		default:
			t.Fatalf("unsupported assertion type: %T", want)
		}
	}
}

func TestEvalArrays(t *testing.T) {

	input := `
	set worldVar = "world"
	SET myarr = ["zero", 1, 1 + 1, 'hello ${worldVar}!']
	SET myresult0 = myarr[0]
	SET myresult1 = myarr[1]
	SET myresult2 = myarr[1 + 1]
	set myresult3 = myarr[9 / 3]
	`
	env := newEnvironment(nil)
	parser := setupEvalTestParser(input)
	if len(parser.errors) > 0 {
		t.Fatalf("parser error: %s", parser.errors[0])
	}
	program, err := parser.parseProgram()
	if err != nil {
		t.Fatal(err)
	}
	_, err = eval(program, env)
	if err != nil {
		t.Fatal(err)
	}
	tests := []struct {
		key  string
		want interface{}
	}{
		{"myresult0", "zero"},
		{"myresult1", 1},
		{"myresult2", 2},
		{"myresult3", "hello world!"},
	}
	for _, tt := range tests {
		got, found := env.get(tt.key)
		if !found {
			t.Fatalf("wanted result for env key %q but got no result", tt.key)
		}
		switch want := tt.want.(type) {
		case int:
			gotInt, ok := got.(*objectInteger)
			if !ok {
				t.Fatalf("result is not of type *objectInteger. got=%T", got)
			}
			if want != int(gotInt.value) {
				t.Errorf("expected value for key %q to be %d. got=%d", tt.key, want, gotInt.value)
			}
		case string:
			gotStr, ok := got.(*objectString)
			if !ok {
				t.Fatalf("result is not of type *objectString. got=%T", got)
			}
			if want != gotStr.value {
				t.Errorf("expected value for key %q to be %s. got=%s", tt.key, want, gotStr.value)
			}
		default:
			t.Fatalf("unsupported assertion type: %T", want)
		}
	}
}

func TestEvalStringAdd(t *testing.T) {
	cases := []struct {
		input string
		want  string
	}{
		{`"asdf" + "ghjkl"`, "asdfghjkl"},
		{`"hello" + " " + "world" + "!"`, "hello world!"},
		{`"raw" + " " + 'templ${"ate"}'`, "raw template"},
	}
	for _, tt := range cases {
		env := newEnvironment(nil)
		parser := setupEvalTestParser(tt.input)
		if len(parser.errors) > 0 {
			t.Fatalf("parser error: %s", parser.errors[0])
		}
		stmt := parser.parseExpressionStatement()
		res, err := eval(stmt, env)
		if err != nil {
			t.Fatal(err)
		}
		strRes, ok := res.(*objectString)
		if !ok {
			t.Fatalf("expected result to be type *objectString. got=%T", res)
		}
		if strRes.value != tt.want {
			t.Errorf("expected string value to be %q. got=%q", tt.want, strRes.value)
		}
	}
}

func TestEvalNumberEquality(t *testing.T) {
	cases := []struct {
		input string
		want  bool
	}{
		{"1 == 1", true},
		{"1 != 1", false},
		{"1.1 != 1.2", true},
		{"1.6 == 1.5", false},
		{"0.1 > 0.001", true},
		{"0.5 >= 0.5", true},
		{"5 <= 2", false},
		{"5 < 5.5", true},
	}
	for _, tt := range cases {
		env := newEnvironment(nil)
		parser := setupEvalTestParser(tt.input)
		if len(parser.errors) > 0 {
			t.Fatalf("parser error: %s", parser.errors[0])
		}
		stmt := parser.parseExpressionStatement()
		res, err := eval(stmt, env)
		if err != nil {
			t.Fatal(err)
		}
		boolRes, ok := res.(*objectBoolean)
		if !ok {
			t.Fatalf("expected result to be type *objectFloat. got=%T", res)
		}
		if boolRes.value != tt.want {
			t.Errorf("expected float value to be %t. got=%t", tt.want, boolRes.value)
		}
	}
}

func TestEvalMathFloats(t *testing.T) {
	cases := []struct {
		input string
		want  float64
	}{
		{"1 + 1.5", 2.5},
		{"1.5 + 1.6", 3.1},
		{"5 - .5", 4.5},
		{"1.6 - 1.5", 0.1},
		{"0.1 * 5", .5},
		{"0.5 * 0.5", .25},
		{"5 / 2", 2.5},
		{"5.5 / 5.5", 1.0},
	}
	for _, tt := range cases {
		env := newEnvironment(nil)
		parser := setupEvalTestParser(tt.input)
		if len(parser.errors) > 0 {
			t.Fatalf("parser error: %s", parser.errors[0])
		}
		stmt := parser.parseExpressionStatement()
		res, err := eval(stmt, env)
		if err != nil {
			t.Fatal(err)
		}
		flRes, ok := res.(*objectFloat)
		if !ok {
			t.Fatalf("expected result to be type *objectFloat. got=%T", res)
		}
		if !isFloatEqual(flRes.value, tt.want) {
			t.Errorf("expected float value to be %f. got=%f", tt.want, flRes.value)
		}
	}
}

func TestEvalMathInts(t *testing.T) {
	cases := []struct {
		input string
		want  int64
	}{
		{"1 + 1", 2},
		{"5 - 1", 4},
		{"5 * 5", 25},
		{"36 / 6", 6},
	}
	for _, tt := range cases {
		env := newEnvironment(nil)
		parser := setupEvalTestParser(tt.input)
		if len(parser.errors) > 0 {
			t.Fatalf("parser error: %s", parser.errors[0])
		}
		stmt := parser.parseExpressionStatement()
		res, err := eval(stmt, env)
		if err != nil {
			t.Fatal(err)
		}
		intRes, ok := res.(*objectInteger)
		if !ok {
			t.Fatalf("expected result to be type *objectInteger. got=%T", res)
		}
		if intRes.value != tt.want {
			t.Errorf("expected integer value to be %d. got=%d", tt.want, intRes.value)
		}
	}
}

func TestEvalTemplateExpression(t *testing.T) {
	env := newEnvironment(nil)
	parser := setupEvalTestParser(`
	SET hellovar = "hello"
	SET worldvar = "world"
	SET helloworld = '${hellovar}
	${worldvar}!'
	`)
	if len(parser.errors) > 0 {
		t.Fatalf("parser error: %s", parser.errors[0])
	}
	program, err := parser.parseProgram()
	if err != nil {
		t.Fatal(err)
	}
	eval(program, env)

	obj, found := env.get("helloworld")
	if !found {
		t.Errorf("env variable not found for helloworld. expected value of %q", "hello\n\tworld!")
	}
	if obj.getType() != t_string {
		t.Errorf("env variable at helloworld is not of type %s. got=%s", t_string, obj.getType())
	}
	objString := obj.(*objectString)
	if objString.value != "hello\n\tworld!" {
		t.Errorf("wrong value for env var string helloworld. want=%q got=%q", "hello\n\tworld!", objString.value)
	}
}

func TestEvalStringLitera(t *testing.T) {
	env := newEnvironment(nil)
	parser := setupEvalTestParser(`"this is my string" "this is my other string"`)
	if len(parser.errors) > 0 {
		t.Fatalf("parser error: %s", parser.errors[0])
	}
	stmt1 := parser.parseStatement()
	got, err := eval(stmt1, env)
	if err != nil {
		t.Fatal(err)
	}
	if got.getType() != t_string {
		t.Fatalf("expected result type to be %s. got=%s", t_string, got.getType())
	}
	gotStr := got.(*objectString)
	if gotStr.value != "this is my string" {
		t.Errorf("wrong value for string. want=%q got=%q", "this is my string", gotStr.value)
	}
	//
	//
	parser.next()
	stmt2 := parser.parseStatement()
	got, err = eval(stmt2, env)
	if err != nil {
		t.Fatal(err)
	}
	if got.getType() != t_string {
		t.Fatalf("expected result type to be %s. got=%s", t_string, got.getType())
	}
	gotStr = got.(*objectString)
	if gotStr.value != "this is my other string" {
		t.Errorf("wrong value for string. want=%q got=%q", "this is my other string", gotStr.value)
	}
}

func TestEvalPrefixExpression(t *testing.T) {
	env := newEnvironment(nil)
	parser := setupEvalTestParser("!false")
	if len(parser.errors) > 0 {
		t.Fatalf("parser error: %s", parser.errors[0])
	}
	program := parser.parseStatement()
	got, err := eval(program, env)
	if err != nil {
		t.Fatal(err)
	}
	if got.getType() != t_boolean {
		t.Fatalf("expected result type to be %s. got=%s", t_boolean, got.getType())
	}
	gotBool := got.(*objectBoolean)
	if gotBool.value != true {
		t.Errorf("wrong value for exclamation prefix. want=%t got=%t", true, gotBool.value)
	}

	env = newEnvironment(nil)
	parser = setupEvalTestParser("-.123")
	if len(parser.errors) > 0 {
		t.Fatalf("parser error: %s", parser.errors[0])
	}
	program = parser.parseStatement()
	got, err = eval(program, env)
	if err != nil {
		t.Fatal(err)
	}
	if got.getType() != t_float {
		t.Fatalf("expected result type to be %s. got=%s", t_float, got.getType())
	}
	gotFloat := got.(*objectFloat)
	if gotFloat.value != -.123 {
		t.Errorf("wrong value for exclamation prefix. want=%f got=%f", -.123, gotFloat.value)
	}
}

func TestEvalWhenStatement(t *testing.T) {
	env := newEnvironment(nil)
	parser := setupEvalTestParser("WHEN true :: SET my.path.var = 10")
	if len(parser.errors) > 0 {
		t.Fatalf("parser error: %s", parser.errors[0])
	}
	program := parser.parseStatement()
	eval(program, env)

	parser = setupEvalTestParser("my.path.var")
	program = parser.parseStatement()
	if len(parser.errors) > 0 {
		t.Fatalf("parser error: %s", parser.errors[0])
	}
	got, err := eval(program, env)
	if err != nil {
		t.Fatal(err)
	}
	if got.getType() != t_integer {
		t.Fatalf("expected result type to be %s. got=%s", t_integer, got.getType())
	}
	asInt := got.(*objectInteger)
	if asInt.value != 10 {
		t.Errorf("expected integer val to be 10. got=%d", asInt.value)
	}
}

func TestEvalSetExpression(t *testing.T) {
	parser := setupEvalTestParser("SET my.path.var = 5")
	program := parser.parseStatement()
	if len(parser.errors) > 0 {
		t.Fatalf("parser error: %s", parser.errors[0])
	}
	env := newEnvironment(nil)
	_, err := eval(program, env)
	if err != nil {
		t.Fatal(err)
	}
	objRoot, ok := env.get("my")
	if !ok {
		t.Fatalf("expected env to have an item at %q", "my")
	}
	objRootMap, ok := objRoot.(*objectMap)
	if !ok {
		t.Fatalf("expected env my.path to be an *objectMap. got=%T", objRoot)
	}
	attrPathObj, ok := objRootMap.kvPairs["path"]
	if !ok {
		t.Fatal("expected my.path to exist")
	}
	attrPathMap, ok := attrPathObj.value.(*objectMap)
	if !ok {
		t.Fatalf("expected my.path to be an *objectMap. got=%T", attrPathObj.value)
	}
	attrVarObj, ok := attrPathMap.kvPairs["var"]
	if !ok {
		t.Fatal("expected my.path.var to exist")
	}
	attrVarInt, ok := attrVarObj.value.(*objectInteger)
	if !ok {
		t.Fatalf("expected my.path.var to be an integer. got=%T", attrVarObj.value)
	}

	if attrVarInt.value != 5 {
		t.Errorf("expected my.path.var to be equal to 5. got=%d", attrVarInt.value)
	}
}

func TestEvalPathExpression(t *testing.T) {
	env := newEnvironment(nil)
	dataMap, err := convertBytesToObject([]byte(`{
		"nested": {
			"key": 5
		}
	}`))
	if err != nil {
		t.Fatal(err)
	}
	env.set("myobj", dataMap)
	parser := setupEvalTestParser("myobj.nested.key")
	program := parser.parseStatement()
	if len(parser.errors) > 0 {
		t.Fatalf("parser error: %s", parser.errors[0])
	}
	got, err := eval(program, env)
	if err != nil {
		t.Fatal(err)
	}
	if got.getType() != t_integer {
		t.Fatalf("expected result type to be %s. got=%s", t_integer, got.getType())
	}
	asInt := got.(*objectInteger)
	if asInt.value != 5 {
		t.Errorf("expected integer val to be 5. got=%d", asInt.value)
	}
}

func setupEvalTestParser(input string) *parser {
	l := newLexer([]rune(input))
	return newParser(l)
}
