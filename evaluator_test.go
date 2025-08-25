package morph

import (
	"fmt"
	"testing"
)

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
		fmt.Printf("%+v", got)
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
