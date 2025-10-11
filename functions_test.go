package morph

import (
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
