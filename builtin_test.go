package morph

import "testing"

func TestBuiltinLen(t *testing.T) {
	fnStore := newBuiltinFuncStore()
	env := newEnvironment(fnStore)
	arr, err := convertArrayToObject([]interface{}{"a", "b", "c"}, false)
	if err != nil {
		t.Fatal(err)
	}
	env.set("myarr", arr)
	m, err := convertMapToObject(map[string]interface{}{"mykey": 1}, false)
	if err != nil {
		t.Fatal(err)
	}
	env.set("mymap", m)
	str, err := convertAnyToObject("mystringvalue", false)
	if err != nil {
		t.Fatal(err)
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
	_, err = eval(stmt, env)
	if err != nil {
		t.Fatal(err)
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
	_, err := eval(stmt, env)
	if err != nil {
		t.Fatal(err)
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
	set num = "mydatastring"
	drop()
	set otherThing = 5
	`
	l := newLexer([]rune(input))
	p := newParser(l)
	program, err := p.parseProgram()
	if err != nil {
		t.Fatal(err)
	}

	_, err = evalProgram(program, env)
	if err != nil {
		t.Fatal(err)
	}
	if len(env.store) != 0 {
		t.Errorf("expected drop() to cause the env to be empty. got len=%d", len(env.store))
	}
}
