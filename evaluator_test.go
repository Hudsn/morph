package parser

import "testing"

func testEvaluateProgram(t *testing.T, input string, src interface{}, wantDest interface{})

func testEvaluateExpressionStatement(t *testing.T, input string, want interface{}) {
	l := newLexer([]rune(input))
	p := newParser(l)
	program, err := p.parseProgram()
	if err != nil {
		t.Fatalf("%s", err.Error())
	}
	env := newEnvironment()
}
