package morph

import (
	"strings"
	"testing"
)

func TestEvalDel(t *testing.T) {
	input := `
	set val = "abcd"
	del val
	set val_2.a = "1"
	del val_2.a
	set val_2.b = ["c", "d", "e"] 
	if val_2.a == null :: set val_2.c.d = val_2.b
	if len(val_2.c.d) == len(val_2.b) :: {
		del val_2.d // no-op
		del val_2.c // deletes .c and .c.d with it
	}
	// so now this should be true and delete .b
	if val_2.c.d == null :: del val_2.b
	
	//and we should get all are null, setting result to 'true'
	set result = val == null && val_2.b == val && val_2.b == val_2.c.d
	`
	env := newEnvironment(newBuiltinFuncStore())
	parser := setupEvalTestParser(input)
	program, err := parser.parseProgram()
	if err != nil {
		t.Fatal(err)
	}
	res := program.eval(env)
	if isObjectErr(res) {
		t.Fatal(objectToError(res))
	}
	got, ok := env.get("result")
	if !ok {
		t.Fatalf("expected an existing env entry for %q, but got no result", "result")
	}
	testConvertObject(t, got, true)
}

func TestEvalPipe(t *testing.T) {
	input := `
	set result = append([1, 2, "3"], 4) | append(5)
	`
	env := newEnvironment(newBuiltinFuncStore())
	parser := setupEvalTestParser(input)
	program, err := parser.parseProgram()
	if err != nil {
		t.Fatal(err)
	}
	res := program.eval(env)
	if isObjectErr(res) {
		t.Fatal(objectToError(res))
	}
	got, ok := env.get("result")
	if !ok {
		t.Fatalf("expected an existing env entry for %q, but got no result", "result")
	}
	testConvertObject(t, got, []interface{}{1, 2, "3", 4, 5})
}

func TestEvalNull(t *testing.T) {
	input := `
	set mynull = NULL
	`
	env := newEnvironment(nil)
	parser := setupEvalTestParser(input)
	program, err := parser.parseProgram()
	if err != nil {
		t.Fatal(err)
	}
	res := program.eval(env)
	if isObjectErr(res) {
		t.Fatal(objectToError(res))
	}
	got, ok := env.get("mynull")
	if !ok {
		t.Fatalf("expected an existing env entry for %q, but got no result", "mynull")
	}
	if got != obj_global_null {
		t.Errorf("expected result to be a globally-shared null object. instead got %+v", got)
	}
}

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
	res := program.eval(env)
	if isObjectErr(res) {
		t.Fatal(objectToError(res))
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
	res := program.eval(env)
	if isObjectErr(res) {
		t.Fatal(objectToError(res))
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
		res := stmt.eval(env)
		if isObjectErr(res) {
			t.Fatal(objectToError(res))
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
		res := stmt.eval(env)
		if isObjectErr(res) {
			t.Fatal(objectToError(res))
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
		res := stmt.eval(env)
		if isObjectErr(res) {
			t.Fatal(objectToError(res))
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
		res := stmt.eval(env)
		if isObjectErr(res) {
			t.Fatal(objectToError(res))
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
	res := program.eval(env)
	if isObjectErr(res) {
		t.Fatal(objectToError(res))
	}

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
	got := stmt1.eval(env)
	if isObjectErr(got) {
		t.Fatal(objectToError(got))
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
	got = stmt2.eval(env)
	if isObjectErr(got) {
		t.Fatal(objectToError(got))
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
	got := program.eval(env)
	if isObjectErr(got) {
		t.Fatal(objectToError(got))
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
	got = program.eval(env)
	if isObjectErr(got) {
		t.Fatal(objectToError(got))
	}
	if got.getType() != t_float {
		t.Fatalf("expected result type to be %s. got=%s", t_float, got.getType())
	}
	gotFloat := got.(*objectFloat)
	if gotFloat.value != -.123 {
		t.Errorf("wrong value for exclamation prefix. want=%f got=%f", -.123, gotFloat.value)
	}
}

func TestEvalIfStatement(t *testing.T) {
	env := newEnvironment(nil)
	parser := setupEvalTestParser("IF true :: SET my.path.var = 10")
	if len(parser.errors) > 0 {
		t.Fatalf("parser error: %s", parser.errors[0])
	}
	program := parser.parseStatement()
	res := program.eval(env)
	if isObjectErr(res) {
		t.Fatal(objectToError(res))
	}

	parser = setupEvalTestParser("my.path.var")
	program = parser.parseStatement()
	if len(parser.errors) > 0 {
		t.Fatalf("parser error: %s", parser.errors[0])
	}
	got := program.eval(env)
	if isObjectErr(got) {
		t.Fatal(objectToError(got))
	}
	if got.getType() != t_integer {
		t.Fatalf("expected result type to be %s. got=%s", t_integer, got.getType())
	}
	asInt := got.(*objectInteger)
	if asInt.value != 10 {
		t.Errorf("expected integer val to be 10. got=%d", asInt.value)
	}
}

// trying to write to a path where one of the path items is a non-map object should result in an error
func TestEvalSetStatementInvalidPaths(t *testing.T) {
	testInputs := []string{`
		set myvar.next = 5
		set myvar.next.sub = 10
		`,

		`set myvar = 5
		set myvar.next.sub = 10
		`,
	}
	for _, input := range testInputs {

		parser := setupEvalTestParser(input)
		program, err := parser.parseProgram()
		if err != nil {
			t.Fatalf("parser error: %s", parser.errors[0])
		}
		env := newEnvironment(nil)
		res := program.eval(env)
		if !isObjectErr(res) {
			t.Fatalf("expected resulting object to be an error describing an invalid SET statement on a non-map object. instead got %+v", res)
		}

		errObj := res.(*objectError)
		if !strings.Contains(errObj.message, "invalid path part for SET statement: cannot use a path expression on a non-map object.") {
			t.Errorf("expected an error output describing invalid SET statment due to path operations on a non-map object. instead got=%q", errObj.message)
		}
	}
}

func TestEvalSetStatement(t *testing.T) {
	parser := setupEvalTestParser("SET my.path.var = 5")
	program := parser.parseStatement()
	if len(parser.errors) > 0 {
		t.Fatalf("parser error: %s", parser.errors[0])
	}
	env := newEnvironment(nil)
	res := program.eval(env)
	if isObjectErr(res) {
		t.Fatal(objectToError(res))
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
	attrPathMap, ok := attrPathObj.(*objectMap)
	if !ok {
		t.Fatalf("expected my.path to be an *objectMap. got=%T", attrPathObj)
	}
	attrVarObj, ok := attrPathMap.kvPairs["var"]
	if !ok {
		t.Fatal("expected my.path.var to exist")
	}
	attrVarInt, ok := attrVarObj.(*objectInteger)
	if !ok {
		t.Fatalf("expected my.path.var to be an integer. got=%T", attrVarObj)
	}

	if attrVarInt.value != 5 {
		t.Errorf("expected my.path.var to be equal to 5. got=%d", attrVarInt.value)
	}
}

func TestEvalPathExpression(t *testing.T) {
	env := newEnvironment(nil)
	dataMap := convertBytesToObject([]byte(`{
		"nested": {
			"key": 5
		}
	}`))
	if isObjectErr(dataMap) {
		t.Fatal(objectToError(dataMap))
	}
	env.set("myobj", dataMap)
	parser := setupEvalTestParser("myobj.nested.key")
	program := parser.parseStatement()
	if len(parser.errors) > 0 {
		t.Fatalf("parser error: %s", parser.errors[0])
	}
	got := program.eval(env)
	if isObjectErr(got) {
		t.Fatal(objectToError(got))
	}
	if got.getType() != t_integer {
		t.Fatalf("expected result type to be %s. got=%s", t_integer, got.getType())
	}
	asInt := got.(*objectInteger)
	if asInt.value != 5 {
		t.Errorf("expected integer val to be 5. got=%d", asInt.value)
	}
}

func TestEvalIndexOutOfBoundsReturnsError(t *testing.T) {
	env := newEnvironment(nil)
	dataMap := convertBytesToObject([]byte(`{
		"nested": {
			"key": 5,
			"arr": [
				4,
				{
					"arrkey": 10
				}
			]
		}
	}`))
	if isObjectErr(dataMap) {
		t.Fatal(objectToError(dataMap))
	}
	env.set("myobj", dataMap)

	testInputs := []string{
		"myobj.nested.arr[-1]",
		"myobj.nested.arr[2]",
		"myobj.nested.arr[4].arrkey[5]",
	}
	for idx, input := range testInputs {
		parser := setupEvalTestParser(input)
		program := parser.parseStatement()
		if len(parser.errors) > 0 {
			t.Fatalf("parser error: %s", parser.errors[0])
		}
		got := program.eval(env)
		if !isObjectErr(got) {
			t.Fatalf("case %d: expected input %q to return an error regarding path expressions not being applicable on non-map objects.", idx+1, input)
		}
		errObj := got.(*objectError)
		if !strings.Contains(errObj.message, "index is out of range for target array") {
			t.Errorf("case %d: expected input %q to return an error regarding index expressions not being applicable on non-array objects. intead got %s", idx+1, input, errObj.message)
		}
	}
}

func TestEvalIndexOnNonArrayReturnsError(t *testing.T) {
	env := newEnvironment(nil)
	dataMap := convertBytesToObject([]byte(`{
		"nested": {
			"key": 5,
			"arr": [
				4,
				{
					"arrkey": 10
				}
			]
		}
	}`))
	if isObjectErr(dataMap) {
		t.Fatal(objectToError(dataMap))
	}
	env.set("myobj", dataMap)

	testInputs := []string{
		"myobj[4]",
		"myobj.nested[0]",
		"myobj.nested.key[999]",
		"myobj.nested.arr[1].arrkey[5]",
	}
	for idx, input := range testInputs {
		parser := setupEvalTestParser(input)
		program := parser.parseStatement()
		if len(parser.errors) > 0 {
			t.Fatalf("parser error: %s", parser.errors[0])
		}
		got := program.eval(env)
		if !isObjectErr(got) {
			t.Fatalf("case %d: expected input %q to return an error regarding path expressions not being applicable on non-map objects.", idx+1, input)
		}
		errObj := got.(*objectError)
		if !strings.Contains(errObj.message, "cannot call index expression on non-array object") {
			t.Errorf("case %d: expected input %q to return an error regarding index expressions not being applicable on non-array objects. intead got %s", idx+1, input, errObj.message)
		}
	}
}

func TestEvalPathOnNonMapReturnsError(t *testing.T) {
	env := newEnvironment(nil)
	dataMap := convertBytesToObject([]byte(`{
		"nested": {
			"key": 5,
			"arr": [
				4,
				{
					"arrkey": 10
				}
			]
		}
	}`))
	if isObjectErr(dataMap) {
		t.Fatal(objectToError(dataMap))
	}
	env.set("myobj", dataMap)

	testInputs := []string{
		"myobj.nested.key.nonexistent",
		"myobj.nested.key.arr.nope",
		"myobj.nested.key.arr[0].notthiseither",
	}
	for idx, input := range testInputs {
		parser := setupEvalTestParser(input)
		program := parser.parseStatement()
		if len(parser.errors) > 0 {
			t.Fatalf("parser error: %s", parser.errors[0])
		}
		got := program.eval(env)
		if !isObjectErr(got) {
			t.Fatalf("case %d: expected input %q to return an error regarding path expressions not being applicable on non-map objects.", idx+1, input)
		}
		errObj := got.(*objectError)
		if !strings.Contains(errObj.message, "cannot access a path") && !strings.Contains(errObj.message, "on a non-map object") {
			t.Errorf("case %d: expected input %q to return an error regarding path expressions not being applicable on non-map objects. intead got %s", idx+1, input, errObj.message)
		}
	}
}

// any path or index expression as part of a path MUST return null if the lefthand side is a nonexistent/null expression.
// that is, if any path prefix is null, and is followed by a [index] or additional .path notation, the result should be null
// however, any existing/non-null items that are followed by either an [index] or .path notation, which are not of the appropriate type (array or map, respectively), should throw an error
func TestEvalNonexistentPathReturnsNull(t *testing.T) {
	env := newEnvironment(nil)
	dataMap := convertBytesToObject([]byte(`{
		"nested": {
			"key": 5,
			"arr": [
				4,
				{
					"arrkey": 10
				}
			]
		}
	}`))
	if isObjectErr(dataMap) {
		t.Fatal(objectToError(dataMap))
	}
	env.set("myobj", dataMap)

	testInputs := []string{
		"mynonexistentarr[0]",
		"mynonexistentobj",
		"myobj.nonexistent_nested",
		"myobj.nested.nonexistentkey",
		"myobj.nested.nonexistentarr[999]",
		"myobj.nested.arr[1].fdsa",
	}
	for idx, input := range testInputs {
		parser := setupEvalTestParser(input)
		program := parser.parseStatement()
		if len(parser.errors) > 0 {
			t.Fatalf("parser error: %s", parser.errors[0])
		}
		got := program.eval(env)
		if isObjectErr(got) {
			t.Fatal(objectToError(got))
		}
		if got != obj_global_null {
			t.Fatalf("case %d: expected result type to be a shared global object of type %s. got=%s", idx+1, t_null, got.getType())
		}
	}
}

func TestEvalArrayInfixPl(t *testing.T) {
	env := newEnvironment(nil)
	dataMap := convertBytesToObject([]byte(`{
		"nested": {
			"arr": ["e", "f"]
		}
	}`))
	if isObjectErr(dataMap) {
		t.Fatal(objectToError(dataMap))
	}
	env.set("myobj", dataMap)
	testInputs := []struct {
		input string
		want  []interface{}
	}{
		{
			`set result = ["a", "b"] + ["c", "d"] + myobj.nested.arr`,
			[]interface{}{"a", "b", "c", "d", "e", "f"},
		},
	}
	for idx, tt := range testInputs {
		parser := setupEvalTestParser(tt.input)
		program, err := parser.parseProgram()
		if err != nil {
			t.Fatal(err)
		}
		runRes := program.eval(env)
		if isObjectErr(runRes) {
			t.Fatalf("case %d: %s", idx+1, runRes.inspect())
		}
		got, ok := env.get("result")
		if !ok {
			t.Fatal("expected environment to contain 'result' value")
		}
		testConvertObject(t, got, tt.want)
	}
}

func setupEvalTestParser(input string) *parser {
	l := newLexer([]rune(input))
	return newParser(l)
}
