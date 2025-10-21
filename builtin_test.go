package morph

import (
	"fmt"
	"testing"
	"time"
)

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
func TestBuiltinCoalesce(t *testing.T) {
	tests := []struct {
		input   string
		wantKey string
		want    interface{}
	}{
		{
			input: `
			set this_exists = "hello world"
			set res = coalesce(this_exists, "goodbye world")`,
			wantKey: "res",
			want:    "hello world",
		},
		{
			input:   `set res = coalesce(this.doesnt.exist, "saved the world")`,
			wantKey: "res",
			want:    "saved the world",
		},
		{
			input:   `set res = coalesce(int("goodbye world"), "saved the world")`,
			wantKey: "res",
			want:    "saved the world",
		},
		{
			input:   `set res = this.doesnt.exist | coalesce("saved the world")`,
			wantKey: "res",
			want:    "saved the world",
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
func TestBuiltinInt(t *testing.T) {
	tests := []struct {
		input   string
		wantKey string
		want    interface{}
	}{
		{
			input:   `SET res = int("5")`,
			wantKey: "res",
			want:    5,
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

func TestBuiltinFloat(t *testing.T) {
	tests := []struct {
		input   string
		wantKey string
		want    interface{}
	}{
		{
			input:   `SET res = float("5.5")`,
			wantKey: "res",
			want:    5.5,
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
func TestBuiltinString(t *testing.T) {
	tests := []struct {
		input   string
		wantKey string
		want    interface{}
	}{
		{
			input:   `SET res = string(5.5)`,
			wantKey: "res",
			want:    "5.5",
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
func TestBuiltinTime(t *testing.T) {
	tInt := 1759782264
	start := time.Unix(int64(tInt), 0).UTC()
	addStr := start.Format(time.RFC3339)
	input := `
	SET time_int = 1759782264
	SET time_float = 1759782264.0
	SET time_string_unix = "1759782264"
	SET result = time(time_int) == time(time_float)
	SET result = result && time(time_float) == time(time_string_unix) && time(time_string_unix) == time(time_string)  
	`
	input = fmt.Sprintf("SET time_string = %q\n%s", addStr, input)
	gotEnv := newBuiltinTestEnv(t, input)
	got, ok := gotEnv.get("result")
	if !ok {
		t.Fatal("expected env key result to exist. got none")
	}
	testConvertObject(t, got, true)
	tests := []struct {
		input   string
		wantKey string
		want    interface{}
	}{
		{
			input:   `SET res = time(1759782264) | string()`,
			wantKey: "res",
			want:    "2025-10-06T20:24:24Z",
		},
		{
			input:   `SET res = time(1759782264.0) | string()`,
			wantKey: "res",
			want:    "2025-10-06T20:24:24Z",
		},
		{
			input:   `SET res = time("1759782264") | string()`,
			wantKey: "res",
			want:    "2025-10-06T20:24:24Z",
		},
		{
			input:   `SET res = time("2025-10-06T20:24:24Z") | string()`,
			wantKey: "res",
			want:    "2025-10-06T20:24:24Z",
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
func TestBuiltinDrop(t *testing.T) {
	tests := []struct {
		input   string
		wantKey string
		want    interface{}
	}{
		{
			input: `SET @out.result = 100
			drop()`,
			wantKey: "res",
			want:    nil,
		},
	}
	for _, tt := range tests {
		gotEnv := newBuiltinTestEnv(t, tt.input)
		if len(gotEnv.store) != 0 {
			t.Error("expected drop() to cause the resulting env to be empty")
		}
	}
}
func TestBuiltinEmit(t *testing.T) {
	tests := []struct {
		input   string
		wantKey string
		want    interface{}
	}{
		{
			input: `SET res = 100
			emit()
			SET res = 0`,
			wantKey: "res",
			want:    100,
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
