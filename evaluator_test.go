package morph

import (
	"testing"
)

func TestEvalPrefixExpression(t *testing.T) {
	env := newEnvironment()
	evaluator := setupEvalTest("!false")
	if len(evaluator.parser.errors) > 0 {
		t.Fatalf("parser error: %s", evaluator.parser.errors[0])
	}
	program := evaluator.parser.parseStatement()
	got := evaluator.eval(program, env)
	if got.getType() != T_BOOLEAN {
		t.Fatalf("expected result type to be %s. got=%s", T_BOOLEAN, got.getType())
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
	if got.getType() != T_FLOAT {
		t.Fatalf("expected result type to be %s. got=%s", T_FLOAT, got.getType())
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
	if got.getType() != T_INTEGER {
		t.Fatalf("expected result type to be %s. got=%s", T_INTEGER, got.getType())
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
	if got.getType() != T_INTEGER {
		t.Fatalf("expected result type to be %s. got=%s", T_INTEGER, got.getType())
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
