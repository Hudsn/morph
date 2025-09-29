package morph

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"
)

func TestMorphBasicExample(t *testing.T) {
	test := testMorphCase{
		description: "basic io test",
		srcJSON: `
		{
			"mood": "happy"
		}
		`,
		in: `
			WHEN src.mood == "happy" :: SET dest = "🙂"
		`,
		wantJSON: `
		"🙂"
		`,
	}
	checkTestMorphCase(t, test, NewEmptyFunctionStore())
}

func TestMorphCustomFunction(t *testing.T) {

	fs := NewDefaultFunctionStore()
	funcEntry := NewFunctionEntry("mycoolfunc", testMorphCustomFn999)
	fs.RegisterToNamespace("myfuncs", funcEntry)
	fs.RegisterToNamespace("std", funcEntry)
	test := testMorphCase{
		description: "namespace test",
		srcJSON: `
		{
			"mood": "happy"
		}
		`,
		in: `
			set dest.mood = std.string(src.mood)
			set dest.num = myfuncs.mycoolfunc()
			set dest.num2 = mycoolfunc()
			set dest.num3 = std.mycoolfunc()
		`,
		wantJSON: `
		{
			"mood": "happy",
			"num": 999,
			"num2": 999,
			"num3": 999
		}
		`,
	}
	checkTestMorphCase(t, test, fs)
}

func TestMorphReturnNull(t *testing.T) {
	test := testMorphCase{
		description: "check unset dest returns null",
		srcJSON: `
		{}
		`,
		in:       ``,
		wantJSON: `null`,
	}
	checkTestMorphCase(t, test, NewDefaultFunctionStore())
}

func TestMorphExclamationOnIndirectBool(t *testing.T) {
	tests := []testMorphCase{
		{
			description: "check that ! prefix works properly",
			srcJSON: `
			{}
			`,
			in: `
			SET dest.a = !("a" == 2)
			SET dest.b = !false
			SET dest.c = !true
			SET dest.d = !!true
			SET my_var = true
			SET dest.e = !my_var
			`,
			wantJSON: `
			{
				"a": true,
				"b": true,
				"c": false,
				"d": true,
				"e": false
			}`,
		},
	}
	for _, tt := range tests {
		checkTestMorphCase(t, tt, NewDefaultFunctionStore())
	}
}

func TestMorphTemplateStrings(t *testing.T) {
	test := testMorphCase{
		description: "null dest check",
		srcJSON: `
		{}
		`,
		in:       `SET dest = 'my ${1300 + 37} ${"str" + "ing"}'`,
		wantJSON: `"my 1337 string"`,
	}
	checkTestMorphCase(t, test, NewDefaultFunctionStore())

}

func TestMorphPipes(t *testing.T) {
	tests := []testMorphCase{
		{
			description: "check pipes work as expected + in recursive arrow funcs",
			srcJSON: `
				{
					"my_arr": [1, 2, "three", "4", 4]
				}
				`,
			in: `
			SET dest = filter(src.my_arr, entry ~> {
				WHEN entry.index >= 2 && ((entry.value % 2 == 0) | catch(false) || (int(entry.value) >= 4) | catch(false)) :: SET return = true
			})
			`,
			wantJSON: `["4", 4]`,
		},
		{
			description: "check pipes bind lower than math ops",
			srcJSON: `
				{}
				`,
			in: `
			SET dest = 3 * 2 / 3 + 4 - 2 | min(100)
			`,
			wantJSON: `4`,
		},
		{
			description: "check pipes bind higher than equality",
			srcJSON: `
				{}
				`,
			in: `
			SET dest = 4 == 4 | max(100)
			`,
			wantJSON: `false`,
		},
		{
			description: "check pipes bind higher than bool ops",
			srcJSON: `
				{}
				`,
			in: `
			SET dest = true && "pizza" | contains("iz")
			`,
			wantJSON: `true`,
		},
		{
			description: "check pipes can func chain",
			srcJSON: `
				{}
				`,
			in: `
			SET dest = 2 + 2 | max(50) | min(100) | string() | contains("5")
			`,
			wantJSON: `true`,
		},
	}
	for _, tt := range tests {
		checkTestMorphCase(t, tt, NewDefaultFunctionStore())
	}

}

func TestMorphTheCoolerDaniel(t *testing.T) {
	// 	SET is_cool = cool_factor >= 500
	// SET dest.name = src.name
	// WHEN src.name == "Daniel" || is_cool :: SET dest.name = 'The cooler ${src.name}'

}

func TestMorphMapEdgeCase(t *testing.T) {
	test := testMorphCase{
		description: "check edge case for map() key re-assignment to pre-existing key",
		srcJSON: `
			{
				"a": 1,
				"b": 2,
				"c": 3
			}
			`,
		in: `
			SET dest = map(src, entry ~> {
				WHEN entry.value == 3 :: SET return.key = "a"
				WHEN entry.value == 3 :: SET return.value = 3
			})
			`,
		wantJSON: `{
			"a": 1,
			"b": 2,
			"c": 3
		}`,
	}
	checkTestMorphCase(t, test, NewDefaultFunctionStore())
}

func TestMorphFilter(t *testing.T) {
	tests := []testMorphCase{
		{
			description: "check array filter works as expected",
			srcJSON: `
			{
				"my_arr": [1, 2, "three", "4", 4]
			}
			`,
			in: `
			SET dest = filter(src.my_arr, entry ~> {
				WHEN entry.index >= 2 && (catch(entry.value % 2 == 0, false) || catch(int(entry.value) >= 4, false)) :: SET return = true
			})
			`,
			wantJSON: `["4", 4]`,
		},
		{
			description: "check maps filter works as expected",
			srcJSON: `
			{
				"a": 1,
				"b": 2,
				"c": 3
			}
			`,
			in: `
			SET dest = filter(src, entry ~> {
				WHEN entry.key == "a" :: SET return = true
				WHEN entry.value == 3 :: SET return = true
			})
			`,
			wantJSON: `{
				"a": 1,
				"c": 3
			}`,
		},
	}
	for _, tt := range tests {
		checkTestMorphCase(t, tt, NewDefaultFunctionStore())
	}
}

func TestMorphReduce(t *testing.T) {
	tests := []testMorphCase{
		{
			description: "check arrays reduce works as expected",
			srcJSON: `
			{
				"my_arr": [1, 2, "3"]
			}
			`,
			in: `
			SET dest.result = reduce(src.my_arr, null, entry ~> {
				WHEN entry.current == NULL :: SET entry.current = 0
				SET return = entry.current + int(entry.value)
			})
			`,
			wantJSON: `{
				"result": 6
			}`,
		},
		{
			description: "check maps reduce works as expected",
			srcJSON: `
			{
				"a": 1,
				"b": 2,
				"c": 3
			}
			`,
			in: `
			SET dest.result = reduce(src, null, entry ~> {
				WHEN entry.current == NULL :: SET entry.current = 0
				WHEN entry.key != "a" :: SET return = entry.current + int(entry.value)
			})
			`,
			wantJSON: `{
				"result": 5
			}`,
		},
	}

	for _, tt := range tests {
		checkTestMorphCase(t, tt, NewDefaultFunctionStore())
	}
}

func TestMultiLineDQuoteStringError(t *testing.T) {
	test := testMorphError{
		description: "check multiline string is error",
		srcJSON: `
		{}
		`,
		in: `
		SET dest = "holy
		smokes"
		`,
		wantErrContains: "string literal not terminated",
	}
	checkTestMorphParseError(t, test, NewDefaultFunctionStore())
}

func TestMorphDropArrow(t *testing.T) {
	tests := []testMorphCase{
		{
			description: "drop function in map arrowfunc returns original val",
			srcJSON: `
			{
				"arr": [1, 2, 3]
			}
			`,
			in: `
			SET dest = map(src.arr, entry ~> {
				drop()
			})
			`,
			wantJSON: `[1, 2, 3]`,
		},
		{
			description: "drop function in filter arrowfunc removes entry",
			srcJSON: `
			{
				"arr": [1, 2, 3]
			}
			`,
			in: `
			SET dest = filter(src.arr, entry ~> {
				SET return = true
				drop()
			})
			`,
			wantJSON: `[]`,
		},
		{
			description: "drop function in reduce arrowfunc doesn't affect accumulator",
			srcJSON: `
			{
				"arr": [1, 2, 3]
			}
			`,
			in: `
			SET dest = reduce(src.arr, 0, entry ~> {
				SET return = entry.current + 5
				drop()
			})
			`,
			wantJSON: `0`,
		},
	}
	for _, tt := range tests {
		checkTestMorphCase(t, tt, NewDefaultFunctionStore())
	}
}

func TestMorphDrop(t *testing.T) {
	test := testMorphCase{
		description: "drop function returns null",
		srcJSON: `
		{}
		`,
		in: `
		SET dest = "holy smokes"
		drop()
		SET dest = "no way dude"
		`,
		wantJSON: `null`,
	}
	checkTestMorphCase(t, test, NewDefaultFunctionStore())
}

func TestMorphEmit(t *testing.T) {
	test := testMorphCase{
		description: "emit function returns dest",
		srcJSON: `
		{}
		`,
		in: `
		SET dest = "holy smokes"
		emit()
		SET dest = "no way dude"
		`,
		wantJSON: `"holy smokes"`,
	}
	checkTestMorphCase(t, test, NewDefaultFunctionStore())
}

func TestMorphSetByValue(t *testing.T) {
	test := testMorphCase{
		description: "ensure objs are set by value, not reference",
		srcJSON: `
		{}
		`,
		in: `
			set x = 1
			set y = x
			set x = 5
			set z = [x, y]
			set y = z
			set z = x
			set dest = y
		`,
		wantJSON: `
		[5, 1]
		`,
	}
	checkTestMorphCase(t, test, NewEmptyFunctionStore())
}

// helpers
func testMorphCustomFn999(args ...*Object) *Object {
	if ret, ok := IsArgCountEqual(0, args); !ok {
		return ret
	}
	return CastInt(999)
}

func checkTestMorphParseError(t *testing.T, tt testMorphError, fnStore *functionStore) bool {
	_, err := New(tt.in, WithFunctionStore(fnStore))
	if err == nil {
		t.Fatalf("expected error to contain %q. instead got no error", tt.wantErrContains)
	}
	// _, err = m.ToJSON([]byte(tt.srcJSON))
	// if err != nil {
	// 	t.Fatal(err)
	// }

	if !strings.Contains(err.Error(), tt.wantErrContains) {
		t.Errorf("expected error to contain %q. got=%s", tt.wantErrContains, err.Error())
		return false
	}
	return true
}

func checkTestMorphCase(t *testing.T, tt testMorphCase, fnStore *functionStore) bool {
	m, err := New(tt.in, WithFunctionStore(fnStore))
	if err != nil {
		t.Fatal(err)
	}
	got, err := m.ToJSON([]byte(tt.srcJSON))
	if err != nil {
		t.Fatal(err)
	}

	var gotInterface interface{}
	err = json.Unmarshal(got, &gotInterface)
	if err != nil {
		t.Fatal(err)
	}
	var wantInteface interface{}
	err = json.Unmarshal([]byte(tt.wantJSON), &wantInteface)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(wantInteface, gotInterface) {
		t.Errorf("%s: WRONG VALUE\n\twant:\n\t\t%s\n\tgot\n\t\t%s\n", tt.description, tt.wantJSON, string(got))
		return false
	}
	return true
}

type testMorphError struct {
	description     string
	in              string
	srcJSON         string
	wantErrContains string
}
type testMorphCase struct {
	description string
	in          string
	srcJSON     string
	wantJSON    string
}
