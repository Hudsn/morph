package morph

import "testing"

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
