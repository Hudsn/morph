package morph

import (
	"fmt"
	"testing"
	"time"
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
			nameDesc: "map() with attempt to directly assign .value (shouldn't work)",
			data: `{
				"a": 1,
				"b": 2,
				"c": 3
			}`,
			input: `
			SET res = map(mydata, entry ~> {
				SET entry.value = entry.value * 2
				SET mydata = "ASDF"
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
				set return = entry.value - 1
			})
			`,
			want: []interface{}{0, 1.5, 2},
		},
		{
			nameDesc: "map() with array; with assignment, using indexes",
			data: `
				[1, 2.5, 3]
			`,
			input: `
			SET res = map(mydata, entry ~> {
				if entry.index == 2 :: set return = entry.value * 2
			})
			`,
			want: []interface{}{1, 2.5, 6},
		},
		{
			nameDesc: "map() with array; without assignment",
			data: `
				[1, 2.5, 3]
			`,
			input: `
			SET res = map(mydata, entry ~> {
				set not_return = entry.value - 1
			})
			`,
			want: []interface{}{1, 2.5, 3},
		},
		{
			nameDesc: "map() on array with attempt to directly assign .value and .index (shouldn't work)",
			data: `
				[1, 2.5, 3]
			`,
			input: `
			SET res = map(mydata, entry ~> {
				SET entry.value = 5
				SET entry.index = 0
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
			t.Fatalf("%s: %s", tt.nameDesc, res.inspect())
		}
		gotObj, ok := env.get("res")
		if !ok {
			t.Fatalf("%s: expected env var res to exist. got null", tt.nameDesc)
		}
		if !testConvertObject(t, gotObj, tt.want) {
			t.Errorf("%s: incorrect result value", tt.nameDesc)
		}
		outObj, ok := env.get("mydata")
		if !ok {
			t.Fatalf("%s: expected env var res to exist. got null", tt.nameDesc)
		}
		if outObj != inputObj {
			t.Fatal("original data object should be unchanged")
		}
	}
}
func TestBuiltinFilter(t *testing.T) {
	tests := []struct {
		nameDesc string
		data     string
		input    string
		want     interface{}
	}{
		{
			nameDesc: "filter() on map with key and val filtering",
			data: `{
				"a": 1,
				"b": 2,
				"c": 3
			}`,
			input: `
			SET res = filter(mydata, entry ~> {
				if entry.key == "a" :: SET return = true
				if entry.value == 3 :: SET return = true
			})
			`,
			want: map[string]interface{}{
				"a": 1,
				"c": 3,
			},
		},
		{
			nameDesc: "filter() with attempt to directly assign .value (shouldn't work)",
			data: `{
				"a": 1,
				"b": 2,
				"c": 3
			}`,
			input: `
			SET res = filter(mydata, entry ~> {
				SET entry.key = "0"
				SET entry.value = 0
			})
			`,
			want: map[string]interface{}{},
		},
		{
			nameDesc: "filter() on map with key and val reassignments to null",
			data: `{
				"a": 1,
				"b": 2,
				"c": 3
			}`,
			input: `
			SET res = filter(mydata, entry ~> {
				IF entry.key == "a" :: SET return = doesntexist
				IF entry.value == 3 :: SET return = thiseither
			})
			`,
			want: map[string]interface{}{},
		},
		{
			nameDesc: "filter() on array with correct filtering",
			data:     `[1, 2, "3", 4]`,
			input: `
			SET res = filter(mydata, entry ~> {
				IF entry.value == string(3) :: SET return = true
			})
			`,
			want: []interface{}{"3"},
		},
		{
			nameDesc: "filter() on array with index filtering",
			data:     `[1, 2, "3", 4]`,
			input: `
			SET res = filter(mydata, entry ~> {
				IF entry.index >= 2 :: SET return = true
			})
			`,
			want: []interface{}{"3", 4},
		},
		{
			nameDesc: "filter() on array with reassignments to null",
			data:     `[1, 2, "3", 4]`,
			input: `
			SET res = filter(mydata, entry ~> {
				IF entry.value == 3 :: SET return = thiseither
			})
			`,
			want: []interface{}{},
		},
		{
			nameDesc: "filter() on array with attempt to directly assign .value (shouldn't work)",
			data:     `[1, 2, "3", 4]`,
			input: `
			SET res = filter(mydata, entry ~> {
				SET entry.value = 0
				SET entry.index = 0
			})
			`,
			want: []interface{}{},
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
			t.Fatalf("%s: %s", tt.nameDesc, res.inspect())
		}
		gotObj, ok := env.get("res")
		if !ok {
			t.Fatalf("%s: expected env var res to exist. got null", tt.nameDesc)
		}
		if !testConvertObject(t, gotObj, tt.want) {
			t.Errorf("%s: incorrect result value", tt.nameDesc)
		}
		outObj, ok := env.get("mydata")
		if !ok {
			t.Fatalf("%s: expected env var res to exist. got null", tt.nameDesc)
		}
		if outObj != inputObj {
			t.Fatal("original data object should be unchanged")
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
				IF entry.current == null :: SET entry.current = []
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
				IF entry.current == NULL :: SET entry.current = 0
				SET whocares = entry.current + entry.value
			})
			`,
			want: nil,
		},
		{
			nameDesc: "reduce() on map with attempt to directly assign .value (shouldn't work)",
			data: `{
				"a": 1,
				"b": 2,
				"c": 2.5
			}`,
			input: `
			SET res = reduce(mydata, string(999), entry ~> {
				SET entry.value = 10
			})
			`,
			want: "999",
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
				SET whocares = int(entry.current) + entry.value
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
			nameDesc: "reduce() on array without acc without assignment",
			data:     `[1, 2, "3"]`,
			input: `
			SET res = reduce(mydata, null, entry ~> {
				IF entry.current == NULL :: SET entry.current = 0
				SET whocares = entry.current + int(entry.value)
			})
			`,
			want: nil,
		},
		{
			nameDesc: "reduce() on array without acc with assignment",
			data:     `[1, 2, "3"]`,
			input: `
			SET res = reduce(mydata, null, entry ~> {
				IF entry.current == NULL :: SET entry.current = 0
				SET return = entry.current + int(entry.value)
			})
			`,
			want: 6,
		},
		{
			nameDesc: "reduce() on array with acc with assignment using index",
			data:     `[1, 2, "3"]`,
			input: `
			SET res = reduce(mydata, 0, entry ~> {
				IF entry.index <=1 :: SET return = entry.current + int(entry.value)
			})
			`,
			want: 3,
		},
		{
			nameDesc: "reduce() on array with out acc with null assignment",
			data:     `[1, 2, "3"]`,
			input: `
			SET res = reduce(mydata, 0, entry ~> {
				SET whocares = entry.current + int(entry.value)
				SET return = entry.current + int(entry.value) 
				IF entry.current >= 3 :: SET return = NULL
			})
			`,
			want: nil,
		},
		{
			nameDesc: "reduce() on array with attempt to directly assign .value (shouldn't work)",
			data:     `[1, 2, "3"]`,
			input: `
			SET res = reduce(mydata, 0, entry ~> {
				SET entry.value = 5
				SET mydata = NULL
			})
			`,
			want: 0,
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
			t.Fatalf("%s: %s", tt.nameDesc, res.inspect())
		}
		gotObj, ok := env.get("res")
		if !ok {
			t.Fatalf("%s: expected env var res to exist. got null", tt.nameDesc)
		}
		if !testConvertObject(t, gotObj, tt.want) {
			t.Errorf("%s: incorrect result value", tt.nameDesc)
		}
		outObj, ok := env.get("mydata")
		if !ok {
			t.Fatalf("%s: expected env var res to exist. got null", tt.nameDesc)
		}
		if outObj != inputObj {
			t.Fatal("original data object should be unchanged")
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

func TestBuiltinCatchOLD(t *testing.T) {
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

// time

func TestBuiltinNow(t *testing.T) {
	realNow := time.Now()
	input := `
	SET now = now()
	`
	resEnv := newBuiltinTestEnvOLD(t, input)
	nowObj, ok := resEnv.get("now")
	if !ok {
		t.Fatal("expected now field to exist in env")
	}
	nowInterface, err := convertObjectToNative(nowObj)
	if err != nil {
		t.Fatal(err)
	}
	got, ok := nowInterface.(time.Time)
	if !ok {
		t.Fatalf("nowInt is not of type time.Time. got=%T", nowInterface)
	}
	if got.IsZero() {
		t.Errorf("now() generated a zero-based time. expected something recent")
	}
	if name, zInt := got.Zone(); zInt != 0 {
		t.Errorf("now generated a non-UTC based timestamp. got name=%s offset=%d", name, zInt)
	}
	if got.Sub(realNow).Abs() >= time.Duration(1*time.Second) {
		t.Fatalf("now() generated a value that is not reflective of the actual current time")
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
	env := newBuiltinTestEnvOLD(t, input)
	got, ok := env.get("result")
	if !ok {
		t.Fatal("expected key result to exist in environment")
	}
	testConvertObject(t, got, true)
}

func TestBuiltinParseTime(t *testing.T) {
	mdyTime, err := time.Parse("2006-01-02", "2025-10-07")
	if err != nil {
		t.Fatal(err)
	}
	wantMap := map[string]time.Time{
		"nano":     time.Unix(0, 1759875973453511000),
		"sec_nano": time.Unix(0, 1759875973453511000),
		"micro":    time.UnixMicro(1759875973453511),
		"milli":    time.UnixMilli(1759875973453),
		"sec":      time.Unix(1759875973, 0),
		"mdy":      mdyTime,
	}

	input := ` 
	SET sec_int = 1759875973
	SET sec_string = "1759875973"
	IF parse_time(sec_int, "unix") == parse_time(sec_string, "unix") :: SET sec = parse_time(sec_int, "unix")
	SET sec_nano_float = 1759875973.453511000
	SET sec_nano_string = "1759875973.453511000"
	IF parse_time(sec_nano_float, "unix") == parse_time(sec_nano_string, "unix") :: SET sec_nano = parse_time(sec_nano_string, "unix")
	SET nano_int = 1759875973453511000
	SET nano_string = "1759875973453511000"
	IF parse_time(nano_int, "unix_nano") == parse_time(nano_string, "unix_nano") :: SET nano = parse_time(nano_string, "unix_nano")
	SET micro_int = 1759875973453511
	SET micro_string = "1759875973453511"
	IF parse_time(micro_int, "unix_micro") == parse_time(micro_string, "unix_micro") :: SET micro = parse_time(micro_int, "unix_micro")
	SET milli_int = 1759875973453
	SET milli_string = "1759875973453"
	IF parse_time(milli_int, "unix_milli") == parse_time(milli_string, "unix_milli") :: SET milli = parse_time(milli_int, "unix_milli")
	SET mdy = parse_time("2025-10-07", "2006-01-02")
	`
	env := newBuiltinTestEnvOLD(t, input)

	for wantKey, wantTime := range wantMap {
		gotObj, ok := env.get(wantKey)
		if !ok {
			t.Fatalf("expected env key %q to exist", wantKey)
		}
		testConvertObject(t, gotObj, wantTime)
	}
}

func newBuiltinTestEnvOLD(t *testing.T, input string) *environment {
	env := newEnvironment(newBuiltinFuncStore())
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
