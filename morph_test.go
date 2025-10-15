package morph

import (
	"encoding/json"
	"fmt"
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
		program: `
			IF src.mood == "happy" :: SET dest = "🙂"
		`,
		wantJSON: `
		"🙂"
		`,
	}
	checkTestMorphCase(t, test, NewEmptyFunctionStore())
}

func TestMorphCustomFunction(t *testing.T) {

	fs := NewDefaultFunctionStore()
	funcEntry := NewFunctionEntryOld("mycoolfunc", testMorphCustomFn999)
	fs.RegisterToNamespace("myfuncs", funcEntry)
	fs.RegisterToNamespace("std", funcEntry)
	test := testMorphCase{
		description: "namespace test",
		srcJSON: `
		{
			"mood": "happy"
		}
		`,
		program: `
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

func TestMorphComments(t *testing.T) {
	test := testMorphCase{
		description: "check that comments don't impact the program",
		srcJSON: `
		{}
		`,
		program: `
		SET x = 5 // some comment
		SET dest = x*2 // a second comment
		// SET dest = 0
		// another comment`,
		wantJSON: `10`,
	}
	checkTestMorphCase(t, test, NewDefaultFunctionStore())
}

func TestMorphInvalidPathErr(t *testing.T) {
	tests := []testMorphError{
		{
			description:     "check that a string used as a path throws an error",
			srcJSON:         `{}`,
			program:         `SET "asdf"."bdsa" = true`,
			wantErrContains: []string{"parsing error at 1:5:", "unexpected token type"},
		},
	}
	for _, tt := range tests {
		checkTestMorphParseError(t, tt, NewEmptyFunctionStore())
	}
}

func TestMorphSetSrcErr(t *testing.T) {
	tests := []testMorphError{
		{
			description:     "check that src cannot be set",
			program:         `SET src = true`,
			srcJSON:         `{}`,
			wantErrContains: []string{"parsing error at 1:5:", "SET statement cannot modify src data"},
		},
		{
			description:     "check that src subfields cannot be set",
			program:         `SET src.subfield = 5`,
			srcJSON:         `{}`,
			wantErrContains: []string{"parsing error at 1:5:", "SET statement cannot modify src data"},
		},
	}
	for _, tt := range tests {
		checkTestMorphParseError(t, tt, NewEmptyFunctionStore())
	}
}

func TestMorphPathWithStrings(t *testing.T) {
	tests := []testMorphCase{
		{
			description: "check that multi-line if statements work as intended",
			srcJSON: `
			{"my spaced out path": 10}
			`,
			program: `
				SET dest = src."my spaced out path"
				`,
			wantJSON: `10`,
		},
		{
			description: "check that multi-line if statements work as intended",
			srcJSON: `
			{"my spaced out path": 10}
			`,
			program: `
				SET part = "ce"
				SET missing_piece = '${"s" + '${"pa" + part}'}d'
				SET dest = src.'my ${missing_piece} out path'
				`,
			wantJSON: `10`,
		},
	}
	for _, tt := range tests {
		checkTestMorphCase(t, tt, NewDefaultFunctionStore())
	}
}

func TestMorphIfMulti(t *testing.T) {
	test := testMorphCase{
		description: "check that multi-line if statements work as intended",
		srcJSON: `
		{}
		`,
		program: `
		SET x = 5
		IF x >= 5 :: {
			SET y = x
			SET x = 10
			if y < x :: {
				if y > 0 && y < 6 :: SET dest = y
			}
		}
		`,
		wantJSON: `5`,
	}
	checkTestMorphCase(t, test, NewDefaultFunctionStore())
}

func TestMorphIfErr(t *testing.T) {
	tests := []testMorphError{
		{
			description:     "check that a single-line if statment can only point to a SET statement",
			program:         `IF true :: IF false :: set dest = 0`,
			srcJSON:         `{}`,
			wantErrContains: []string{"parsing error at 1:12:", "expected one of", "{", "SET"},
		},
	}
	for _, tt := range tests {
		checkTestMorphParseError(t, tt, NewEmptyFunctionStore())
	}
}

func TestMorphReturnNull(t *testing.T) {
	tests := []testMorphCase{
		{
			description: "check unset dest returns null",
			srcJSON: `
			{}
			`,
			program:  ``,
			wantJSON: `null`,
		},
		{
			description: "check set dest with nonexistent value returns null",
			srcJSON: `
			{}
			`,
			program:  `SET dest = src.i_dont_exist`,
			wantJSON: `null`,
		},
	}
	for _, tt := range tests {
		checkTestMorphCase(t, tt, NewDefaultFunctionStore())
	}
}

func TestMorphExclamationOnIndirectBool(t *testing.T) {
	tests := []testMorphCase{
		{
			description: "check that ! prefix works properly",
			srcJSON: `
			{}
			`,
			program: `
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
		program:  `SET dest = 'my ${1300 + 37} ${"str" + "ing"}'`,
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
			program: `
			SET dest = filter(src.my_arr, entry ~> {
				IF entry.index >= 2 && ((entry.value % 2 == 0) | catch(false) || (int(entry.value) >= 4) | catch(false)) :: SET return = true
			})
			`,
			wantJSON: `["4", 4]`,
		},
		{
			description: "check pipes bind lower than math ops",
			srcJSON: `
				{}
				`,
			program: `
			SET dest = 3 * 2 / 3 + 4 - 2 | min(100)
			`,
			wantJSON: `4`,
		},
		{
			description: "check pipes bind higher than equality",
			srcJSON: `
				{}
				`,
			program: `
			SET dest = 4 == 4 | max(100)
			`,
			wantJSON: `false`,
		},
		{
			description: "check pipes bind higher than bool ops",
			srcJSON: `
				{}
				`,
			program: `
			SET dest = true && "pizza" | contains("iz")
			`,
			wantJSON: `true`,
		},
		{
			description: "check pipes can func chain",
			srcJSON: `
				{}
				`,
			program: `
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
	test := testMorphCase{
		description: "check edge case for map() key re-assignment to pre-existing key",
		srcJSON: `
			{
				"name": "Daniel",
				"cool_factor": 999
			}
			`,
		program: `
			SET is_cool = src.cool_factor >= 500
			SET dest.name = src.name
			IF src.name == "Daniel" || is_cool :: SET dest.name = 'The cooler ${src.name}'
			`,
		wantJSON: `{
			"name": "The cooler Daniel"
		}`,
	}
	checkTestMorphCase(t, test, NewDefaultFunctionStore())

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
		program: `
			SET dest = map(src, entry ~> {
				IF entry.value == 3 :: SET return.key = "a"
				IF entry.value == 3 :: SET return.value = 3
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
			program: `
			SET dest = filter(src.my_arr, entry ~> {
				IF entry.index >= 2 && (catch(entry.value % 2 == 0, false) || catch(int(entry.value) >= 4, false)) :: SET return = true
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
			program: `
			SET dest = filter(src, entry ~> {
				IF entry.key == "a" :: SET return = true
				if entry.value == 3 :: SET return = true
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
			program: `
			SET dest.result = reduce(src.my_arr, null, entry ~> {
				IF entry.current == NULL :: SET entry.current = 0
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
			program: `
			SET dest.result = reduce(src, null, entry ~> {
				IF entry.current == NULL :: SET entry.current = 0
				IF entry.key != "a" :: SET return = entry.current + int(entry.value)
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

func TestMorphComparisonChecks(t *testing.T) {
	tests := []testMorphCase{
		{
			description: "check that comparisons are working as intended",
			srcJSON: `
			{}
			`,
			program: `
			set dest = "abc" == "abc"
			`,
			wantJSON: `true`,
		},
		{
			description: "check that comparisons are working as intended",
			srcJSON: `
			{}
			`,
			program: `
			set a = "abc"
			set dest = a == "abc"
			`,
			wantJSON: `true`,
		},
		{
			description: "check that comparisons are working as intended",
			srcJSON: `
			{}
			`,
			program: `
			set a = 1
			set dest = a <= 5
			`,
			wantJSON: `true`,
		},
		{
			description: "check that comparisons are working as intended",
			srcJSON: `
			{}
			`,
			program: `
			set a = 1
			set dest = a != NULL
			`,
			wantJSON: `true`,
		},
		{
			description: "check that comparisons are working as intended",
			srcJSON: `
			{}
			`,
			program: `
			set dest = a == NULL
			`,
			wantJSON: `true`,
		},
	}
	for _, tt := range tests {
		checkTestMorphCase(t, tt, NewDefaultFunctionStore())
	}
}

func TestMorphMultiLineDQuoteStringError(t *testing.T) {
	test := testMorphError{
		description: "check multiline string is error",
		srcJSON: `
		{}
		`,
		program: `
		SET dest = "holy
		smokes"
		`,
		wantErrContains: []string{"string literal not terminated"},
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
			program: `
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
			program: `
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
			program: `
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
		program: `
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
		program: `
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
		program: `
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
	_, err := New(tt.program, WithFunctionStore(fnStore))
	if err == nil {
		t.Fatalf("expected error to contain %q. instead got no error", tt.wantErrContains)
	}

	if !testMorphCheckContainsAll(err.Error(), tt.wantErrContains...) {
		strList := []string{}
		for _, s := range tt.wantErrContains {
			strList = append(strList, fmt.Sprintf("%q", s))
		}
		wantString := strings.Join(strList, ", ")
		wantString = fmt.Sprintf("[%s]", wantString)
		t.Errorf("expected error to contain one of %s. got=%s", wantString, err.Error())
		return false
	}
	return true
}

func checkTestMorphCase(t *testing.T, tt testMorphCase, fnStore *functionStore) bool {
	m, err := New(tt.program, WithFunctionStore(fnStore))
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

func testMorphCheckContainsAll(mainString string, checkStrings ...string) bool {
	count := 0
	for _, want := range checkStrings {
		if strings.Contains(mainString, want) {
			count++
		}
	}
	return count == len(checkStrings)
}

type testMorphError struct {
	description     string
	program         string
	srcJSON         string
	wantErrContains []string
}
type testMorphCase struct {
	description string
	program     string
	srcJSON     string
	wantJSON    string
}
