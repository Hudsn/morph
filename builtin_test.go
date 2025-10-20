package morph

import "testing"

func TestBuiltinCatch(t *testing.T) {
	tests := []struct {
		input   string
		wantKey string
		want    interface{}
	}{
		{
			input:   `set res = catch("hello world", "goodbye world")`,
			wantKey: "res",
			want:    "hello world",
		},
		{
			input:   `set res = catch(int("goodbye world"), "saved the world")`,
			wantKey: "res",
			want:    "saved the world",
		},
		{
			input:   `set res = int("goodbye world") | catch("saved the world")`,
			wantKey: "res",
			want:    "saved the world",
		},
		{
			input: `set res = catch(int("goodbye world"), err ~> {
	SET return = {
		"err_msg": err
	} 
})`,
			wantKey: "res",
			want: map[string]interface{}{
				"err_msg": "unable to cast string as INTEGER. invalid string",
			},
		},
	}
	for _, tt := range tests {
		gotEnv := newBuiltinTestEnv(t, tt.input)
		got, ok := gotEnv.get(tt.wantKey)
		if !ok {
			t.Fatalf("expected env key %q to exist. got none", tt.wantKey)
		}
		testConvertObject(t, got, tt.want)
	}

}

func newBuiltinTestEnv(t *testing.T, input string) *environment {
	env := newEnvironment(newBuiltinFuncStore(), EnvWithNewFNStore(newBuiltinFunctionStore()))
	l := newLexer([]rune(input))
	p := newParser(l)
	program, err := p.parseProgram()
	if err != nil {
		t.Fatal(err)
	}
	evalRes := program.eval(env)
	if isObjectErr(evalRes) {
		t.Fatal(objectToError(evalRes))
	}
	return env
}
