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
	iany, err := convertObjectToNative(res)
	if err != nil {
		t.Error(err)
	}
	got, ok := iany.(int64)
	if !ok {
		t.Errorf("resulting sum field of env is not of type int. got=%T", iany)
	}

	if got != int64(want) {
		t.Errorf("wrong value for env.sum want=%d got=%d", want, got)
	}
}
