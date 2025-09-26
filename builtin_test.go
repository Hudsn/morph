package morph

import (
	"testing"
)

func TestBuiltinMap(t *testing.T) {
	tests := []struct {
		nameDesc string
		data     string
		input    string
		want     interface{}
	}{
		{
			nameDesc: "map() with key and val reassignment",
			data: `{
				"a": 1,
				"b": 2,
				"c": 3
			}`,
			input: `
			SET res = map(mydata, entry ~> {
				set return.key = "prefix_" + entry.key 
				SET return.value = entry.value * 2
			})
			`,
			want: map[string]interface{}{
				"prefix_a": 2,
				"prefix_b": 4,
				"prefix_c": 6,
			},
		},
		{
			nameDesc: "map() with key reassignment; without val assignment",
			data: `{
				"a": 1,
				"b": 2,
				"c": 3
			}`,
			input: `
			SET res = map(mydata, entry ~> {
				SET return.key = "prefix_" + entry.key
			})
			`,
			want: map[string]interface{}{
				"prefix_a": 1,
				"prefix_b": 2,
				"prefix_c": 3,
			},
		},
		{
			nameDesc: "map() with val reassignment; without key assignment",
			data: `{
				"a": 1,
				"b": 2,
				"c": 3
			}`,
			input: `
			SET res = map(mydata, entry ~> {
				SET return.value = entry.value * 2
			})
			`,
			want: map[string]interface{}{
				"a": 2,
				"b": 4,
				"c": 6,
			},
		},
		{
			nameDesc: "map() with no valid assignments",
			data: `{
				"a": 1,
				"b": 2,
				"c": 3
			}`,
			input: `
			SET res = map(mydata, entry ~> {
				set fdsa.value = entry.key
				SET asdf.value = entry.value * 2
			})
			`,
			want: map[string]interface{}{
				"a": 1,
				"b": 2,
				"c": 3,
			},
		},
		{
			nameDesc: "map() with array",
			data: `
				[1, 2.5, 3]
			`,
			input: `
			SET res = map(mydata, entry ~> {
				set return = entry - 1
			})
			`,
			want: []interface{}{0, 1.5, 2},
		},
		{
			nameDesc: "map() with array; without assignment",
			data: `
				[1, 2.5, 3]
			`,
			input: `
			SET res = map(mydata, entry ~> {
				set not_return = entry - 1
			})
			`,
			want: []interface{}{1, 2.5, 3},
		},
	}
	for _, tt := range tests {
		inputObj := convertBytesToObject([]byte(tt.data))
		env := newEnvironment(newBuiltinFuncStore())
		env.set("mydata", inputObj)
		lexer := newLexer([]rune(tt.input))
		parser := newParser(lexer)
		program, err := parser.parseProgram()
		if err != nil {
			t.Fatalf("%s: %s", tt.nameDesc, err.Error())
		}
		res := program.eval(env)
		if isObjectErr(res) {
			t.Fatalf("%s: %s", tt.nameDesc, objectToError(res).Error())
		}
		gotObj, ok := env.get("res")
		if !ok {
			t.Fatalf("%s: expected env var res to exist. got null", tt.nameDesc)
		}
		if !testConvertObject(t, gotObj, tt.want) {
			t.Errorf("%s: incorrect result value", tt.nameDesc)
		}
	}
}
func TestBuiltinReduce(t *testing.T) {
	tests := []struct {
		nameDesc string
		data     string
		input    string
		want     interface{}
	}{
		{
			nameDesc: "reduce() on map with proper acc int assignment",
			data: `{
				"a": 1,
				"b": 2,
				"c": 3
			}`,
			input: `
			SET res = reduce(mydata, 0, entry ~> {
				SET return = entry.current + entry.value
			})
			`,
			want: 6,
		},
		{
			nameDesc: "reduce() on map with proper acc arr assignment",
			data: `{
				"a": 1,
				"b": 2,
				"c": 2.5
			}`,
			input: `
			SET res = reduce(mydata, null, entry ~> {
				WHEN entry.current == null :: SET entry.current = []
				SET return = append(entry.current, int(entry.value * 2))
			})
			`,
			want: []interface{}{2, 4, 5},
		},
		{
			nameDesc: "reduce() on map with null acc without assignment",
			data: `{
				"a": 1,
				"b": 2,
				"c": 2.5
			}`,
			input: `
			SET res = reduce(mydata, null, entry ~> {
				SET whocares = entry.current + entry.value
			})
			`,
			want: nil,
		},
		{
			nameDesc: "reduce() on map with existing acc without assignment",
			data: `{
				"a": 1,
				"b": 2,
				"c": 2.5
			}`,
			input: `
			SET res = reduce(mydata, string(999), entry ~> {
				SET whocares = entry.current + entry.value
			})
			`,
			want: "999",
		},
		{
			nameDesc: "reduce() on array with existing acc with assignment",
			data:     `[1, 2, "3"]`,
			input: `
			SET res = reduce(mydata, 6, entry ~> {
				SET return = entry.current + int(entry.value)
			})
			`,
			want: 12,
		},
		{
			nameDesc: "reduce() on array with out acc without assignment",
			data:     `[1, 2, "3"]`,
			input: `
			SET res = reduce(mydata, null, entry ~> {
				SET whocares = entry.current + int(entry.value)
			})
			`,
			want: nil,
		},
		{
			nameDesc: "reduce() on array with out acc with assignment",
			data:     `[1, 2, "3"]`,
			input: `
			SET res = reduce(mydata, null, entry ~> {
				WHEN entry.current == NULL :: SET entry.current = 0
				SET return = entry.current + int(entry.value)
			})
			`,
			want: 6,
		},
		{
			nameDesc: "reduce() on array with out acc with null assignment",
			data:     `[1, 2, "3"]`,
			input: `
			SET res = reduce(mydata, 0, entry ~> {
				SET whocares = entry.current + int(entry.value)
				SET return = entry.current + int(entry.value) 
				WHEN entry.current >= 3 :: SET return = NULL
			})
			`,
			want: nil,
		},
	}
	for _, tt := range tests {
		inputObj := convertBytesToObject([]byte(tt.data))
		env := newEnvironment(newBuiltinFuncStore())
		env.set("mydata", inputObj)
		lexer := newLexer([]rune(tt.input))
		parser := newParser(lexer)
		program, err := parser.parseProgram()
		if err != nil {
			t.Fatalf("%s: %s", tt.nameDesc, err.Error())
		}
		res := program.eval(env)
		if isObjectErr(res) {
			t.Fatalf("%s: %s", tt.nameDesc, objectToError(res).Error())
		}
		gotObj, ok := env.get("res")
		if !ok {
			t.Fatalf("%s: expected env var res to exist. got null", tt.nameDesc)
		}
		if !testConvertObject(t, gotObj, tt.want) {
			t.Errorf("%s: incorrect result value", tt.nameDesc)
		}
	}
}

func TestBuiltinAppend(t *testing.T) {
	fnStore := newBuiltinFuncStore()
	env := newEnvironment(fnStore)
	arr := convertArrayToObject([]interface{}{"a", "b", "c"}, false)
	if isObjectErr(arr) {
		t.Fatal(objectToError(arr))
	}
	env.set("myarr", arr)
	input := `
	set res = append(myarr, "d")
	`
	want := []interface{}{"a", "b", "c", "d"}
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
	res, ok := env.get("res")
	if !ok {
		t.Fatalf("expected sum field to be populated in env")
	}

	testConvertObject(t, res, want)
}

func TestBuiltinLen(t *testing.T) {
	fnStore := newBuiltinFuncStore()
	env := newEnvironment(fnStore)
	arr := convertArrayToObject([]interface{}{"a", "b", "c"}, false)
	if isObjectErr(arr) {
		t.Fatal(objectToError(arr))
	}
	env.set("myarr", arr)
	m := convertMapToObject(map[string]interface{}{"mykey": 1}, false)
	if isObjectErr(m) {
		t.Fatal(objectToError(m))
	}
	env.set("mymap", m)
	str := convertAnyToObject("mystringvalue", false)
	if isObjectErr(str) {
		t.Fatal(objectToError(str))
	}
	env.set("mystring", str)

	want := 17
	input := `
	set sum = len(myarr) + len(mymap) + len(mystring)
	`
	l := newLexer([]rune(input))
	p := newParser(l)
	stmt := p.parseStatement()
	if len(p.errors) > 0 {
		t.Fatal(p.errors[0])
	}
	evalRes := stmt.eval(env)
	if isObjectErr(evalRes) {
		t.Fatal(objectToError(evalRes))
	}
	res, ok := env.get("sum")
	if !ok {
		t.Fatalf("expected sum field to be populated in env")
	}

	testConvertObjectInt(t, res, int64(want))
}

func TestBuiltinMin(t *testing.T) {
	fnStore := newBuiltinFuncStore()
	env := newEnvironment(fnStore)
	want := 5
	input := `
	set my.num = min(5, len("mylongstring"))
	`
	l := newLexer([]rune(input))
	p := newParser(l)
	stmt := p.parseStatement()
	if len(p.errors) > 0 {
		t.Fatal(p.errors[0])
	}
	evalRes := stmt.eval(env)
	if isObjectErr(evalRes) {
		t.Fatal(objectToError(evalRes))
	}
	res, ok := env.get("my")
	if !ok {
		t.Fatalf("expected my.num field to be populated in env")
	}
	testConvertObject(t, res, map[string]interface{}{
		"num": want,
	})
}
func TestBuiltinMax(t *testing.T) {
	fnStore := newBuiltinFuncStore()
	env := newEnvironment(fnStore)
	want := 12
	input := `
	set my.num = max(5, len("mylongstring"))
	`
	l := newLexer([]rune(input))
	p := newParser(l)
	stmt := p.parseStatement()
	if len(p.errors) > 0 {
		t.Fatal(p.errors[0])
	}
	evalRes := stmt.eval(env)
	if isObjectErr(evalRes) {
		t.Fatal(objectToError(evalRes))
	}
	res, ok := env.get("my")
	if !ok {
		t.Fatalf("expected my.num field to be populated in env")
	}
	testConvertObject(t, res, map[string]interface{}{
		"num": want,
	})
}

func TestBuiltinDrop(t *testing.T) {
	fnStore := newBuiltinFuncStore()
	env := newEnvironment(fnStore)
	input := `
	set str = "mydatastring"
	drop()
	set otherThing = 5
	`
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
	if len(env.store) != 0 {
		t.Errorf("expected drop() to cause the env to be empty. got len=%d", len(env.store))
	}
}
func TestBuiltinEmit(t *testing.T) {
	fnStore := newBuiltinFuncStore()
	env := newEnvironment(fnStore)
	input := `
	set str = "mydatastring"
	emit()
	set otherThing = 5
	`
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

	if len(env.store) != 1 {
		t.Errorf("expected emit() to cause the env to have a len of 1. got len=%d", len(env.store))
	}
	res, ok := env.get("str")
	if !ok {
		t.Fatalf("expected str field to be populated in env")
	}
	testConvertObject(t, res, "mydatastring")
}

func TestBuiltinInt(t *testing.T) {
	fnStore := newBuiltinFuncStore()
	env := newEnvironment(fnStore)
	input := `
	set myvar = "5"
	set res = int(myvar)`
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
	want := 5
	res, ok := env.get("res")
	if !ok {
		t.Fatalf("expected res field to exist in env")
	}
	testConvertObjectInt(t, res, int64(want))
}
func TestBuiltinFloat(t *testing.T) {
	fnStore := newBuiltinFuncStore()
	env := newEnvironment(fnStore)
	input := `
	set myvar = "5"
	set res = float(myvar)`
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
	want := 5
	res, ok := env.get("res")
	if !ok {
		t.Fatalf("expected res field to exist in env")
	}
	testConvertObjectFloat(t, res, float64(want))
}
func TestBuiltinString(t *testing.T) {
	fnStore := newBuiltinFuncStore()
	env := newEnvironment(fnStore)
	input := `
	set myvar = 5.5
	set res = string(myvar)`
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
	want := "5.5"
	res, ok := env.get("res")
	if !ok {
		t.Fatalf("expected res field to exist in env")
	}
	testConvertObjectString(t, res, want)
}

func TestBuiltinCatch(t *testing.T) {
	fnStore := newBuiltinFuncStore()
	env := newEnvironment(fnStore)
	input := `
	set myvar = 5
	set res = catch(myvar.invalidpath, myvar)`
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
	want := 5
	res, ok := env.get("res")
	if !ok {
		t.Fatalf("expected res field to exist in env")
	}
	testConvertObjectInt(t, res, int64(want))
}

func TestBuiltinCoalesce(t *testing.T) {
	fnStore := newBuiltinFuncStore()
	env := newEnvironment(fnStore)
	input := `
	set myvar = 5
	set res = coalesce(empty.nullpathresult, myvar)`
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
	want := 5
	res, ok := env.get("res")
	if !ok {
		t.Fatalf("expected res field to exist in env")
	}
	testConvertObjectInt(t, res, int64(want))
}
func TestBuiltinFallback(t *testing.T) {
	fnStore := newBuiltinFuncStore()
	env := newEnvironment(fnStore)
	input := `
	set myvar = 5
	set res = fallback(empty.nullpathresult, myvar)`
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
	want := 5
	res, ok := env.get("res")
	if !ok {
		t.Fatalf("expected res field to exist in env")
	}
	testConvertObjectInt(t, res, int64(want))
}

func TestBuiltinContains(t *testing.T) {
	fnStore := newBuiltinFuncStore()
	env := newEnvironment(fnStore)
	input := `
	set myarr = [1, 2, "three"]
	set res1 = contains(myarr, 2)
	set res2 = contains(myarr, "three")
	set res3 = contains(myarr, 0)
	set mystring = "abcd"
	set res4 = contains(mystring, "bc")
	set res5 = contains(mystring, "def")
	`
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
	res, ok := env.get("res1")
	if !ok {
		t.Fatalf("expected res1 field to exist in env")
	}
	testConvertObject(t, res, true)
	res, ok = env.get("res2")
	if !ok {
		t.Fatalf("expected res2 field to exist in env")
	}
	testConvertObject(t, res, true)
	res, ok = env.get("res3")
	if !ok {
		t.Fatalf("expected res3 field to exist in env")
	}
	testConvertObject(t, res, false)
	res, ok = env.get("res4")
	if !ok {
		t.Fatalf("expected res4 field to exist in env")
	}
	testConvertObject(t, res, true)
	res, ok = env.get("res5")
	if !ok {
		t.Fatalf("expected res5 field to exist in env")
	}
	testConvertObject(t, res, false)
}
