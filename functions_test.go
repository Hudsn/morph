package morph

import (
	"reflect"
	"slices"
	"testing"
)

func TestFunctionRegistry(t *testing.T) {
	registry := newFunctionStore()

	fnName1 := "my_cool_func"
	fnEntry := NewFunctionEntry(fnName1, testFunctionCustomSum)
	fnEntry.SetArgument("a", INTEGER)
	fnEntry.SetArgument("b", INTEGER)
	fnEntry.SetReturn("result", INTEGER)

	fnEntry.SetAttributes(FUNCTION_ATTRIBUTE_VARIADIC)
	registry.Register(fnEntry)

	fnName2 := "my_other_func"
	fnEntry2 := NewFunctionEntry(fnName2, testFunctionCustomSum)
	fnEntry2.SetArgument("a", INTEGER)
	fnEntry2.SetArgument("b", INTEGER)
	fnEntry2.SetReturn("result", INTEGER, STRING)
	registry.RegisterToNamespace("custom", fnEntry2)

	fnName3 := "my_any_func"
	fnEntry3 := NewFunctionEntry(fnName3, testFunctionCustomAny)
	fnEntry3.SetArgument("a", ANY...)
	fnEntry3.SetReturn("result", ANY...)
	registry.Register(fnEntry3)

	tests := []struct {
		fnEntry  *functionEntry
		wantName string
		wantStr  string
		args     []interface{}
		wantRes  int
	}{
		{
			fnEntry:  fnEntry,
			wantName: "my_cool_func",
			wantStr:  "my_cool_func(a:INTEGER, b:INTEGER) result:INTEGER",
			args:     []interface{}{2, 2, 3, 4},
			wantRes:  4,
		},
		{
			fnEntry:  fnEntry2,
			wantName: "my_other_func",
			wantStr:  "my_other_func(a:INTEGER, b:INTEGER) result:INTEGER|STRING",
			args:     []interface{}{7, 3},
			wantRes:  10,
		},
		{
			fnEntry:  fnEntry3,
			wantName: "my_any_func",
			wantStr:  "my_any_func(a:ANY) result:ANY",
			args:     []interface{}{7},
			wantRes:  7,
		},
	}
	for _, tt := range tests {
		inputObjs := []object{}
		for _, arg := range tt.args {
			toAdd := convertAnyToObject(arg, false)
			if isObjectErr(toAdd) {
				t.Fatal(objectToError(toAdd))
			}
			inputObjs = append(inputObjs, toAdd)
		}
		entry := tt.fnEntry
		resObj := entry.eval(inputObjs...)
		if isObjectErr(resObj) {
			t.Fatal(objectToError(resObj))
		}
		res, err := convertObjectToNative(resObj)
		if err != nil {
			t.Fatal(err)
		}
		resInt, ok := res.(int64)
		if !ok {
			t.Fatalf("res is not of type int64. got=%T", res)
		}
		if resInt != int64(tt.wantRes) {
			t.Errorf("expected result of custom function to be %d. got=%d", tt.wantRes, resInt)
		}

		if entry.name != tt.wantName {
			t.Errorf("expected name to be %q. got=%q", tt.wantName, entry.name)
		}
		if entry.string() != tt.wantStr {
			t.Errorf("expected string output to be %q. got=%q", tt.wantStr, entry.string())
		}

		// for idx, want := range tt.wantExIn {
		// 	got := entry.docInfo.exampleIn[idx]
		// 	if want != got {
		// 		t.Errorf("incorrect example arg. want=%s got=%s", want, got)
		// 	}
		// }
		// if entry.docInfo.exampleOut != tt.wantExOut {
		// 	t.Errorf("incorrect example return. want=%s got=%s", tt.wantExOut, entry.docInfo.exampleOut)
		// }
	}

	_, err := registry.getNamespace("custom", "my_other_func")
	if err != nil {
		t.Fatal(err)
	}
	_, err = registry.get("my_cool_func")
	if err != nil {
		t.Fatal(err)
	}
}

func testFunctionCustomAny(args ...*Object) *Object {
	return args[0]
}

func testFunctionCustomSum(args ...*Object) *Object {
	if res, ok := IsArgCountAtLeast(2, args); !ok {
		return res
	}
	argInt, err := args[0].AsInt()
	if err != nil {
		return ObjectError(err.Error())
	}
	argInt2, err := args[1].AsInt()
	if err != nil {
		return ObjectError(err.Error())
	}
	return CastInt(argInt + argInt2)
}

func TestFunctionInt(t *testing.T) {
	tests := []struct {
		start interface{}
		want  interface{}
	}{
		{10, 10},
		{float64(5), 5},
		{float32(15), 15},
		{int64(8), 8},
	}

	for _, tt := range tests {
		obj := CastInt(tt.start)
		if isObjectErr(obj.inner) {
			t.Fatal(objectToError(obj.inner))
		}
		testFunctionObjectMethods(t, obj, tt.want)
	}
}
func TestFunctionFloat(t *testing.T) {
	tests := []struct {
		start interface{}
		want  interface{}
	}{
		{10, float64(10)},
		{10.9, float64(10.9)},
		{int64(2), float64(2)},
		{float32(8), float64(8)},
	}

	for _, tt := range tests {
		obj := CastFloat(tt.start)
		if isObjectErr(obj.inner) {
			t.Fatal(objectToError(obj.inner))
		}
		testFunctionObjectMethods(t, obj, tt.want)
	}
}
func TestFunctionString(t *testing.T) {
	tests := []string{
		"teststring", "myotherstring",
	}

	for _, tt := range tests {
		obj := CastString(tt)
		if isObjectErr(obj.inner) {
			t.Fatal(objectToError(obj.inner))
		}
		testFunctionObjectMethods(t, obj, tt)
	}
}
func TestFunctionBool(t *testing.T) {
	tests := []bool{
		true, false,
	}

	for _, tt := range tests {
		obj := CastBool(tt)
		if isObjectErr(obj.inner) {
			t.Fatal(objectToError(obj.inner))
		}
		testFunctionObjectMethods(t, obj, tt)
	}
}

func TestFunctionArray(t *testing.T) {
	testArr := []interface{}{
		int64(1), float64(2), "three", false,
	}
	obj := CastArray(testArr)
	if isObjectErr(obj.inner) {
		t.Fatal(objectToError(obj.inner))
	}
	testFunctionObjectMethods(t, obj, testArr)
}

func TestFunctionMap(t *testing.T) {
	testMap := map[string]interface{}{
		"mykey": "mystringval",
		"myotherkey": map[string]interface{}{
			"nested": int64(5),
			"nested_arr": []interface{}{
				int64(10), int64(9), int64(8), int64(7),
			},
		},
	}
	testStruct := struct {
		MyKey      string `json:"mykey"`
		MyOtherKey struct {
			Nested    int   `json:"nested"`
			NestedArr []int `json:"nested_arr"`
		} `json:"myotherkey"`
	}{}

	obj := CastMap(testMap)
	if isObjectErr(obj.inner) {
		t.Fatal(objectToError(obj.inner))
	}

	testFunctionObjectMethods(t, obj, testMap)

	err := obj.MapStruct(&testStruct)
	if err != nil {
		t.Fatal(err)
	}
	if testStruct.MyKey != "mystringval" {
		t.Errorf("wrong value for testStruct.MyKey. want=%s got=%s", "mystringval", testStruct.MyKey)
	}
	if testStruct.MyOtherKey.Nested != 5 {
		t.Errorf("wrong value for testStruct.MyOtherKey.Nested. want=%d got=%d", 5, testStruct.MyOtherKey.Nested)
	}
	if !slices.Equal(testStruct.MyOtherKey.NestedArr, []int{10, 9, 8, 7}) {
		t.Errorf("wrong value for testStruct.MyOtherKey.NestedArry. want=%+v got=%+v", []int{10, 9, 8, 7}, testStruct.MyOtherKey.NestedArr)
	}
}

func testFunctionObjectMethods(t *testing.T, obj *Object, want interface{}) bool {

	switch v := want.(type) {
	case map[string]interface{}:
		return testFunctionObjectAsMap(t, obj, v)
	case []interface{}:
		return testFunctionObjectAsArray(t, obj, v)
	case bool:
		return testFunctionObjectAsBool(t, obj, v)
	case string:
		return testFunctionObjectAsString(t, obj, v)
	case int:
		return testFunctionObjectAsInt(t, obj, v)
	case float64:
		return testFunctionObjectAsFloat(t, obj, v)
	default:
		t.Errorf("unsupported target type for Object conversion method: %T", want)
		return false
	}
}

func testFunctionObjectAsMap(t *testing.T, obj *Object, want map[string]interface{}) bool {
	m, err := obj.AsMap()
	if err != nil {
		t.Error(err)
		return false
	}
	if !reflect.DeepEqual(m, want) {
		t.Errorf("maps are not equal in value.\n\twant=\n\t\t%+v\n\tgot=\n\t\t%+v", want, m)
	}
	return true
}
func testFunctionObjectAsArray(t *testing.T, obj *Object, want []interface{}) bool {
	a, err := obj.AsArray()
	if err != nil {
		t.Error(err)
		return false
	}
	if !reflect.DeepEqual(want, a) {
		t.Errorf("arrays are not equal in value.\n\twant=\n\t\t%+v\n\tgot=\n\t\t%+v", want, a)
	}
	return true
}
func testFunctionObjectAsBool(t *testing.T, obj *Object, want bool) bool {
	b, err := obj.AsBool()
	if err != nil {
		t.Error(err)
		return false
	}
	if b != want {
		t.Errorf("wrong value for obj.AsBool. want=%t got=%t", want, b)
		return false
	}
	return true
}
func testFunctionObjectAsString(t *testing.T, obj *Object, want string) bool {
	s, err := obj.AsString()
	if err != nil {
		t.Error(err)
		return false
	}
	if s != want {
		t.Errorf("wrong value for obj.AsString. want=%s got=%s", want, s)
		return false
	}
	return true
}
func testFunctionObjectAsInt(t *testing.T, obj *Object, want int) bool {
	i, err := obj.AsInt()
	if err != nil {
		t.Error(err)
		return false
	}
	if int(i) != want {
		t.Errorf("wrong value for obj.AsInt. want=%d got=%d", want, i)
		return false
	}
	return true
}
func testFunctionObjectAsFloat(t *testing.T, obj *Object, want float64) bool {
	f, err := obj.AsFloat()
	if err != nil {
		t.Error(err)
		return false
	}
	if !isFloatEqual(f, want) {
		t.Errorf("wrong value for obj.AsFloat. want=%f got=%f", want, f)
		return false
	}
	return true

}
