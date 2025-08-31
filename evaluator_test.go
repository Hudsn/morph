package morph

import (
	"testing"
)

func TestEvalTemplateExpression(t *testing.T) {
	env := newEnvironment()
	evaluator := setupEvalTest(`
	SET hellovar = "hello"
	SET worldvar = "world"
	SET helloworld = '${hellovar}
	${worldvar}!'
	`)
	if len(evaluator.parser.errors) > 0 {
		t.Fatalf("parser error: %s", evaluator.parser.errors[0])
	}
	program, err := evaluator.parser.parseProgram()
	if err != nil {
		t.Fatal(err)
	}
	evaluator.eval(program, env)

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
	env := newEnvironment()
	evaluator := setupEvalTest(`"this is my string" "this is my other string"`)
	if len(evaluator.parser.errors) > 0 {
		t.Fatalf("parser error: %s", evaluator.parser.errors[0])
	}
	stmt1 := evaluator.parser.parseStatement()
	got := evaluator.eval(stmt1, env)
	if got.getType() != t_string {
		t.Fatalf("expected result type to be %s. got=%s", t_string, got.getType())
	}
	gotStr := got.(*objectString)
	if gotStr.value != "this is my string" {
		t.Errorf("wrong value for string. want=%q got=%q", "this is my string", gotStr.value)
	}
	//
	//
	evaluator.parser.next()
	stmt2 := evaluator.parser.parseStatement()
	got = evaluator.eval(stmt2, env)
	if got.getType() != t_string {
		t.Fatalf("expected result type to be %s. got=%s", t_string, got.getType())
	}
	gotStr = got.(*objectString)
	if gotStr.value != "this is my other string" {
		t.Errorf("wrong value for string. want=%q got=%q", "this is my other string", gotStr.value)
	}
}

func TestEvalPrefixExpression(t *testing.T) {
	env := newEnvironment()
	evaluator := setupEvalTest("!false")
	if len(evaluator.parser.errors) > 0 {
		t.Fatalf("parser error: %s", evaluator.parser.errors[0])
	}
	program := evaluator.parser.parseStatement()
	got := evaluator.eval(program, env)
	if got.getType() != t_boolean {
		t.Fatalf("expected result type to be %s. got=%s", t_boolean, got.getType())
	}
	gotBool := got.(*objectBoolean)
	if gotBool.value != true {
		t.Errorf("wrong value for exclamation prefix. want=%t got=%t", true, gotBool.value)
	}

	env = newEnvironment()
	evaluator = setupEvalTest("-.123")
	if len(evaluator.parser.errors) > 0 {
		t.Fatalf("parser error: %s", evaluator.parser.errors[0])
	}
	program = evaluator.parser.parseStatement()
	got = evaluator.eval(program, env)
	if got.getType() != t_float {
		t.Fatalf("expected result type to be %s. got=%s", t_float, got.getType())
	}
	gotFloat := got.(*objectFloat)
	if gotFloat.value != -.123 {
		t.Errorf("wrong value for exclamation prefix. want=%f got=%f", -.123, gotFloat.value)
	}
}

func TestEvalWhenStatement(t *testing.T) {
	env := newEnvironment()
	evaluator := setupEvalTest("WHEN true :: SET my.path.var = 10")
	if len(evaluator.parser.errors) > 0 {
		t.Fatalf("parser error: %s", evaluator.parser.errors[0])
	}
	program := evaluator.parser.parseStatement()
	evaluator.eval(program, env)

	evaluator = setupEvalTest("my.path.var")
	program = evaluator.parser.parseStatement()
	if len(evaluator.parser.errors) > 0 {
		t.Fatalf("parser error: %s", evaluator.parser.errors[0])
	}
	got := evaluator.eval(program, env)
	if got.getType() != t_integer {
		t.Fatalf("expected result type to be %s. got=%s", t_integer, got.getType())
	}
	asInt := got.(*objectInteger)
	if asInt.value != 10 {
		t.Errorf("expected integer val to be 10. got=%d", asInt.value)
	}
}

func TestEvalSetExpression(t *testing.T) {
	evaluator := setupEvalTest("SET my.path.var = 5")
	program := evaluator.parser.parseStatement()
	if len(evaluator.parser.errors) > 0 {
		t.Fatalf("parser error: %s", evaluator.parser.errors[0])
	}
	env := newEnvironment()
	got := evaluator.eval(program, env)
	if objectIsError(got) {
		errObj := got.(*objectError)
		t.Fatal(errObj.message)
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
	env := newEnvironment()
	dataMap, err := newObjectFromBytes([]byte(`{
		"nested": {
			"key": 5
		}
	}`))
	if err != nil {
		t.Fatal(err)
	}
	env.set("myobj", dataMap)
	evaluator := setupEvalTest("myobj.nested.key")
	program := evaluator.parser.parseStatement()
	if len(evaluator.parser.errors) > 0 {
		t.Fatalf("parser error: %s", evaluator.parser.errors[0])
	}
	got := evaluator.eval(program, env)
	if got.getType() != t_integer {
		t.Fatalf("expected result type to be %s. got=%s", t_integer, got.getType())
	}
	asInt := got.(*objectInteger)
	if asInt.value != 5 {
		t.Errorf("expected integer val to be 5. got=%d", asInt.value)
	}
}

func setupEvalTest(input string) *evaluator {
	l := newLexer([]rune(input))
	p := newParser(l)
	return newEvaluator(p)
}
