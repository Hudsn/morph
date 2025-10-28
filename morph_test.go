package morph

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/hudsn/morph/lang"
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
			IF @in.mood == "happy" :: SET @out = "🙂"
		`,
		wantJSON: `
		"🙂"
		`,
	}
	checkTestMorphCase(t, test, lang.NewFunctionStore())
}

func TestMorphCustomFunction(t *testing.T) {

	fs := lang.DefaultFunctionStore()
	funcEntry := lang.NewFunctionEntry("mycoolfunc", "does come cool custom stuff", testMorphCustomFn999)
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
			set @out.mood = std.string(@in.mood)
			set @out.num = myfuncs.mycoolfunc()
			set @out.num2 = mycoolfunc()
			set @out.num3 = std.mycoolfunc()
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
		SET @out = x*2 // a second comment
		// SET @out = 0
		// another comment`,
		wantJSON: `10`,
	}
	checkTestMorphCase(t, test, lang.DefaultFunctionStore())
}

func TestMorphInvalidPathErr(t *testing.T) {
	tests := []testMorphError{
		{
			description:     "check that a string used to begin a path throws an error",
			srcJSON:         `{}`,
			program:         `SET "asdf"."bdsa" = true`,
			wantErrContains: []string{"parsing error at 1:5:", "unexpected token type"},
		},
	}
	for _, tt := range tests {
		if err := checkTestMorphParseError(t, tt, lang.NewFunctionStore()); err != nil {
			t.Error(err.Error())
		}
	}
}

func TestMorphSetInErr(t *testing.T) {
	tests := []testMorphError{
		{
			description:     "check that @in cannot be set",
			program:         `SET @in = true`,
			srcJSON:         `{}`,
			wantErrContains: []string{"parsing error at 1:5:", "SET statement cannot modify @in data"},
		},
		{
			description:     "check that @in subfields cannot be set",
			program:         `SET @in.subfield = 5`,
			srcJSON:         `{}`,
			wantErrContains: []string{"parsing error at 1:5:", "SET statement cannot modify @in data"},
		},
	}
	for _, tt := range tests {
		if err := checkTestMorphParseError(t, tt, lang.NewFunctionStore()); err != nil {
			t.Error(err.Error())
		}
	}
}

func TestMorphPathWithStrings(t *testing.T) {
	tests := []testMorphCase{
		{
			description: "check that strings can be used for path parts as long as they are not the first",
			srcJSON: `
			{"my spaced out path": 10}
			`,
			program: `
				SET @out = @in."my spaced out path"
				`,
			wantJSON: `10`,
		},
		{
			description: "check that template strings can be used for path parts as long as they are not the first",
			srcJSON: `
			{"my spaced out path": 10}
			`,
			program: `
				SET part = "ce"
				SET missing_piece = '${"s" + '${"pa" + part}'}d'
				SET @out = @in.'my ${missing_piece} out path'
				`,
			wantJSON: `10`,
		},
	}
	for _, tt := range tests {
		checkTestMorphCase(t, tt, lang.DefaultFunctionStore())
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
				if y > 0 && y < 6 :: SET @out = y
			}
		}
		`,
		wantJSON: `5`,
	}
	checkTestMorphCase(t, test, lang.DefaultFunctionStore())
}

func TestMorphIfErr(t *testing.T) {
	tests := []testMorphError{
		{
			description:     "check that a single-line if statment can only point to a SET statement",
			program:         `IF true :: IF false :: set @out = 0`,
			srcJSON:         `{}`,
			wantErrContains: []string{"parsing error at 1:12:", "expected one of", "{", "SET"},
		},
	}
	for _, tt := range tests {
		if err := checkTestMorphParseError(t, tt, lang.NewFunctionStore()); err != nil {
			t.Error(err.Error())
		}
	}
}

func TestMorphReturnNull(t *testing.T) {
	tests := []testMorphCase{
		{
			description: "check unset @out returns null",
			srcJSON: `
			{}
			`,
			program:  ``,
			wantJSON: `null`,
		},
		{
			description: "check set @out with nonexistent value returns null",
			srcJSON: `
			{}
			`,
			program:  `SET @out = @in.i_dont_exist`,
			wantJSON: `null`,
		},
	}
	for _, tt := range tests {
		checkTestMorphCase(t, tt, lang.DefaultFunctionStore())
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
			SET @out.a = !("a" == 2)
			SET @out.b = !false
			SET @out.c = !true
			SET @out.d = !!true
			SET my_var = true
			SET @out.e = !my_var
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
		checkTestMorphCase(t, tt, lang.DefaultFunctionStore())
	}
}

func TestMorphTemplateStrings(t *testing.T) {
	test := testMorphCase{
		description: "check that template strings work as intended",
		srcJSON: `
		{}
		`,
		program:  `SET @out = 'my ${1300 + 37} ${"str" + "ing"}'`,
		wantJSON: `"my 1337 string"`,
	}
	checkTestMorphCase(t, test, lang.DefaultFunctionStore())

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
			SET @out = filter(@in.my_arr, entry ~> {
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
			SET @out = 3 * 2 / 3 + 4 - 2 | min(1)
			`,
			wantJSON: `1`,
		},
		{
			description: "check pipes bind higher than equality",
			srcJSON: `
				{}
				`,
			program: `
			SET @out = 4 == 4 | max(100)
			`,
			wantJSON: `false`,
		},
		{
			description: "check pipes bind higher than bool ops",
			srcJSON: `
				{}
				`,
			program: `
			SET @out = true && "pizza" | contains("iz")
			`,
			wantJSON: `true`,
		},
		{
			description: "check pipes can func chain",
			srcJSON: `
				{}
				`,
			program: `
			SET @out = 2 + 2 | max(50) | min(100) | string() | contains("5")
			`,
			wantJSON: `true`,
		},
	}
	for _, tt := range tests {
		checkTestMorphCase(t, tt, lang.DefaultFunctionStore())
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
			SET is_cool = @in.cool_factor >= 500
			SET @out.name = @in.name
			IF @in.name == "Daniel" || is_cool :: SET @out.name = 'The cooler ${@in.name}'
			`,
		wantJSON: `{
			"name": "The cooler Daniel"
		}`,
	}
	checkTestMorphCase(t, test, lang.DefaultFunctionStore())

}

func TestMorphMapEdgeCase(t *testing.T) {
	test := testMorphCase{
		description: "check edge case for map() key re-assignment to pre-existing key does not work",
		srcJSON: `
			{
				"a": 1,
				"b": 2,
				"c": 3
			}
			`,
		program: `
			SET @out = map(@in, entry ~> {
				IF entry.value == 3 :: SET entry.key = "a"
				IF entry.value == 3 :: SET entry.value = 3
			})
			`,
		wantJSON: `{
			"a": 1,
			"b": 2,
			"c": 3
		}`,
	}
	checkTestMorphCase(t, test, lang.DefaultFunctionStore())
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
			SET @out = filter(@in.my_arr, entry ~> {
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
			SET @out = filter(@in, entry ~> {
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
		checkTestMorphCase(t, tt, lang.DefaultFunctionStore())
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
			SET @out.result = reduce(@in.my_arr, null, entry ~> {
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
			SET @out.result = reduce(@in, null, entry ~> {
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
		checkTestMorphCase(t, tt, lang.DefaultFunctionStore())
	}
}

func TestMorphComparisonChecks(t *testing.T) {
	tests := []testMorphCase{
		{
			description: "check that comparisons are working as intended (same literal type ==)",
			srcJSON: `
			{}
			`,
			program: `
			set @out = "abc" == "abc"
			`,
			wantJSON: `true`,
		},
		{
			description: "check that comparisons are working as intended (variable and literal ==)",
			srcJSON: `
			{}
			`,
			program: `
			set a = "abc"
			set @out = a == "abc"
			`,
			wantJSON: `true`,
		},
		{
			description: "check that comparisons are working as intended (<=)",
			srcJSON: `
			{}
			`,
			program: `
			set a = 1
			set @out = a <= 5
			`,
			wantJSON: `true`,
		},
		{
			description: "check that comparisons are working as intended (!=)",
			srcJSON: `
			{}
			`,
			program: `
			set a = 1
			set @out = a != NULL
			`,
			wantJSON: `true`,
		},
		{
			description: "check that comparisons are working as intended",
			srcJSON: `
			{}
			`,
			program: `
			set @out = a == NULL
			`,
			wantJSON: `true`,
		},
	}
	for _, tt := range tests {
		checkTestMorphCase(t, tt, lang.DefaultFunctionStore())
	}
}

func TestMorphMultiLineDQuoteStringError(t *testing.T) {
	test := testMorphError{
		description: "check multiline double-quoted string is error",
		srcJSON: `
		{}
		`,
		program: `
		SET @out = "holy
		smokes"
		`,
		wantErrContains: []string{"string literal not terminated"},
	}
	if err := checkTestMorphParseError(t, test, lang.DefaultFunctionStore()); err != nil {
		t.Error(err.Error())
	}
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
			SET @out = map(@in.arr, entry ~> {
				set return = 0
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
			SET @out = filter(@in.arr, entry ~> {
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
			SET @out = reduce(@in.arr, 0, entry ~> {
				SET return = entry.current + 5
				drop()
			})
			`,
			wantJSON: `0`,
		},
	}
	for _, tt := range tests {
		checkTestMorphCase(t, tt, lang.DefaultFunctionStore())
	}
}

func TestMorphDropBase(t *testing.T) {
	test := testMorphCase{
		description: "drop function returns null",
		srcJSON: `
		{}
		`,
		program: `
		SET @out = "holy smokes"
		drop()
		SET @out = "no way dude"
		`,
		wantJSON: `null`,
	}
	checkTestMorphCase(t, test, lang.DefaultFunctionStore())
}

func TestMorphEmit(t *testing.T) {
	test := testMorphCase{
		description: "emit function returns @out",
		srcJSON: `
		{}
		`,
		program: `
		SET @out = "holy smokes"
		emit()
		SET @out = "no way dude"
		`,
		wantJSON: `"holy smokes"`,
	}
	checkTestMorphCase(t, test, lang.DefaultFunctionStore())
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
			set @out = y
		`,
		wantJSON: `
		[5, 1]
		`,
	}
	checkTestMorphCase(t, test, lang.NewFunctionStore())
}

// helpers
func testMorphCustomFn999(ctx context.Context, args ...*lang.Object) *lang.Object {
	if ret, ok := lang.IsArgCountEqual(0, args); !ok {
		return ret
	}
	return lang.CastInt(999)
}

func checkTestMorphParseError(t *testing.T, tt testMorphError, fnStore *lang.FunctionStore) error {
	_, err := New(tt.program)
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
		return fmt.Errorf("expected error to contain all of %s. got=%s", wantString, err.Error())
	}
	return nil
}

func checkTestMorphCase(t *testing.T, tt testMorphCase, fnStore *lang.FunctionStore) bool {
	m, err := New(tt.program, WithFunctionStore(fnStore))
	if err != nil {
		t.Fatal(err)
	}
	got, err := m.Exec([]byte(tt.srcJSON))
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

// Builtin tests

func TestBuiltinFunctions(t *testing.T) {
	err := lang.RunFunctionStoreExamples(lang.DefaultFunctionStore())
	if err != nil {
		t.Fatal(err)
	}
}
