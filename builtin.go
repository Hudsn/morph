package morph

import (
	"context"
	"slices"
	"strconv"
	"strings"
	"time"
)

// WIP, unused.
func newBuiltinFunctionStore() *FunctionStore {
	store := NewFunctionStore()

	//nulls and err handling
	store.Register(builtinCatchEntry())
	store.Register(builtinCoalesceEntry())
	store.Register(builtinFallbackEntry())

	//type coercion
	store.Register(builtinIntEntry())
	store.Register(builtinFloatEntry())
	store.Register(builtinStringEntry())
	store.Register(builtinTimeEntry())

	//flow control
	store.Register(builtinDropEntry())
	store.Register(builtinEmitEntry())

	//general
	store.Register(builtinLenEntry())

	return store
}

func builtinCatchEntry() *FunctionEntry {
	return NewFunctionEntry(
		"catch",
		"Checks a target item for errors. If the item is an error or evaluates to an error, the fallback is returned. If not, the item is returned",
		builtinCatch,
		WithArgs(
			NewFunctionArg(
				"item",
				"The expression to check for potential errors",
				BASIC...,
			),
			NewFunctionArg(
				"fallback",
				`The value to use as a fallback if the target item is an error.
The argument can also be an arrow function callback, where the target error string is accessible by the named parameter, and the resulting fallback must be assigned to the 'return' variable within the arrow function.`,
				ANY...,
			),
		),
		WithReturn(
			NewFunctionReturn(
				"the item or fallback, depending on whether the item is an error or not, respectively",
				BASIC...,
			),
		),
		WithTags(FUNCTION_TAG_ERR_NULL_CHECKS),
		WithExamples(
			NewProgramExample(
				`null`,
				`//using the target (no error)
SET @out.result = catch("hello world", "goodbye world")`,
				`{"result": "hello world"}`,
			),
			NewProgramExample(
				`null`,
				`//using the fallback (error)
SET @out.result = catch(int("goodbye world"), "saved the world")`,
				`{"result": "saved the world"}`,
			),
			NewProgramExample(
				`null`,
				`//using pipe syntax
SET @out.result = int("goodbye world") | catch("saved the world")`,
				`{"result": "saved the world"}`,
			),
			NewProgramExample(
				`null`,
				`//using arrow callback syntax
SET @out.result = catch(int("goodbye world"), err ~> {
	SET return = {
		"err_msg": err
	} 
})`,
				`{"result": {"err_msg": "unable to cast string as INTEGER. invalid string"}}`,
			),
		),
	)
}

func builtinCatch(ctx context.Context, args ...*Object) *Object {
	if res, ok := IsArgCountEqual(2, args); !ok {
		return res
	}
	target := args[0]
	fallback := args[1]
	target.Type()
	if target.Type() == string(ERROR) {
		if fallback.Type() == string(ARROWFUNC) {
			arrow, err := fallback.AsArrowFunction()
			if err != nil {
				return ObjectError(err.Error())
			}
			inputErr, err := target.AsError()
			if err != nil {
				return ObjectError(err.Error())
			}
			envOut := arrow.Run(inputErr.Error())
			mapOut, ok := envOut.(map[string]interface{})
			if !ok {
				return ObjectError("error calling fallback arrow function: unable to extract 'return' field from arrow function")
			}
			if ret, ok := mapOut["return"]; ok {
				return CastAuto(ret)
			}
		}
		return fallback
	}
	return target
}
func builtinFallbackEntry() *FunctionEntry {
	return NewFunctionEntry(
		"fallback",
		"Checks if target item is null or an error. If the item is null or evaluates to null, the fallback is returned. If not, the item is returned",
		builtinFallback,
		WithArgs(
			NewFunctionArg(
				"item",
				"The expression to check for being null or an error",
				BASIC...,
			),
			NewFunctionArg(
				"fallback",
				`The value to use as a fallback if the target item is null or an error.`,
				BASIC...,
			),
		),
		WithReturn(
			NewFunctionReturn(
				"the item or fallback, depending on whether the item is null or an error, or not, respectively",
				BASIC...,
			),
		),
		WithTags(FUNCTION_TAG_ERR_NULL_CHECKS),
		WithExamples(
			NewProgramExample(
				`null`,
				`//using the target (no null or error)
SET this_exists = "hello world"
SET @out.result = fallback(this_exists, "goodbye world")`,
				`{"result": "hello world"}`,
			),
			NewProgramExample(
				`null`,
				`//using the fallback (error)
SET @out.result = fallback(int("goodbye world"), "saved the world")`,
				`{"result": "saved the world"}`,
			),
			NewProgramExample(
				`null`,
				`//using the fallback (null)
SET @out.result = fallback(this.doesnt.exist, "saved the world")`,
				`{"result": "saved the world"}`,
			),
			NewProgramExample(
				`null`,
				`//using pipe syntax
SET @out.result = this.doesnt.exist | fallback("saved the world")`,
				`{"result": "saved the world"}`,
			),
		),
	)
}

func builtinFallback(ctx context.Context, args ...*Object) *Object {
	if res, ok := IsArgCountEqual(2, args); !ok {
		return res
	}
	target := args[0]
	fallback := args[1]
	target.Type()
	if target.Type() == string(NULL) || target.Type() == string(ERROR) {
		return fallback
	}
	return target
}
func builtinCoalesceEntry() *FunctionEntry {
	return NewFunctionEntry(
		"coalesce",
		"Checks if target item is null. If the item is null or evaluates to null, the fallback is returned. If not, the item is returned",
		builtinCoalesce,
		WithArgs(
			NewFunctionArg(
				"item",
				"The expression to check for being null",
				BASIC...,
			),
			NewFunctionArg(
				"fallback",
				`The value to use as a fallback if the target item is null.`,
				BASIC...,
			),
		),
		WithReturn(
			NewFunctionReturn(
				"the item or fallback, depending on whether the item is null or not, respectively",
				BASIC...,
			),
		),
		WithTags(FUNCTION_TAG_ERR_NULL_CHECKS),
		WithExamples(
			NewProgramExample(
				`null`,
				`//using the target (no null)
SET this_exists = "hello world"
SET @out.result = coalesce(this_exists, "goodbye world")`,
				`{"result": "hello world"}`,
			),
			NewProgramExample(
				`null`,
				`//using the fallback (null)
SET @out.result = coalesce(this.doesnt.exist, "saved the world")`,
				`{"result": "saved the world"}`,
			),
			NewProgramExample(
				`null`,
				`//using pipe syntax
SET @out.result = this.doesnt.exist | coalesce("saved the world")`,
				`{"result": "saved the world"}`,
			),
		),
	)
}

func builtinCoalesce(ctx context.Context, args ...*Object) *Object {
	if res, ok := IsArgCountEqual(2, args); !ok {
		return res
	}
	target := args[0]
	fallback := args[1]
	target.Type()
	if target.Type() == string(NULL) {
		return fallback
	}
	return target
}

func builtinIntEntry() *FunctionEntry {
	return NewFunctionEntry(
		"int",
		"attempts to convert the target item into an integer type",
		builtinInt,

		WithArgs(
			NewFunctionArg(
				"target",
				"The expression to convert into an integer",
				FLOAT, STRING, INTEGER,
			),
		),
		WithReturn(
			NewFunctionReturn(
				"The resultant integer after conversion. Throws an error if unable to convert.",
				INTEGER,
			),
		),
		WithTags(FUNCTION_TAG_TYPE_COERCION),
		WithExamples(
			NewProgramExample(
				`null`,
				`SET @out.result = int("5")`,
				`{"result": 5}`,
			),
		),
	)
}

func builtinInt(ctx context.Context, args ...*Object) *Object {
	if res, ok := IsArgCountEqual(1, args); !ok {
		return res
	}
	a := args[0]
	val, err := a.AsAny()
	if err != nil {
		return ObjectError("unable to cast item as INTEGER. invalid input type: %s", a.Type())
	}
	return CastInt(val)
}

func builtinFloat(ctx context.Context, args ...*Object) *Object {
	if res, ok := IsArgCountEqual(1, args); !ok {
		return res
	}
	a := args[0]
	val, err := a.AsAny()
	if err != nil {
		return ObjectError("unable to cast item as FLOAT. invalid input type: %s", a.Type())
	}
	return CastFloat(val)
}
func builtinFloatEntry() *FunctionEntry {
	return NewFunctionEntry(
		"float",
		"Attempts to convert the target item into a float type",
		builtinFloat,

		WithArgs(
			NewFunctionArg(
				"target",
				"The expression to convert into a float",
				INTEGER, STRING, FLOAT,
			),
		),
		WithReturn(
			NewFunctionReturn(
				"The resultant float after conversion. Throws an error if unable to convert.",
				FLOAT,
			),
		),
		WithTags(FUNCTION_TAG_TYPE_COERCION),
		WithExamples(
			NewProgramExample(
				`null`,
				`SET @out.result = float("5.5")`,
				`{"result": 5.5}`,
			),
		),
	)
}

func builtinStringEntry() *FunctionEntry {
	return NewFunctionEntry(
		"string",
		"Attempts to convert the target item into a string type",
		builtinString,

		WithArgs(
			NewFunctionArg(
				"target",
				"The expression to convert into a string",
				INTEGER, FLOAT, BOOLEAN, TIME, STRING,
			),
		),
		WithReturn(
			NewFunctionReturn(
				"The resultant string after conversion. Throws an error if unable to convert.",
				STRING,
			),
		),
		WithTags(FUNCTION_TAG_TYPE_COERCION),
		WithExamples(
			NewProgramExample(
				`null`,
				`SET @out.result = string(5.5)`,
				`{"result": "5.5"}`,
			),
		),
	)
}
func builtinString(ctx context.Context, args ...*Object) *Object {
	if res, ok := IsArgCountEqual(1, args); !ok {
		return res
	}
	a := args[0]
	val, err := a.AsAny()
	if err != nil {
		return ObjectError("unable to convert item to STRING. invalid input type: %s", a.Type())
	}
	return CastString(val)
}

func builtinTimeEntry() *FunctionEntry {
	return NewFunctionEntry(
		"time",
		`Attempts to convert the target item into a time type.`,
		builtinTime,

		WithArgs(
			NewFunctionArg(
				"target",
				`The expression to convert into a time.
Strings must be in RFC3339/RFC3339Nano format, or directly convertible to a number representing UNIX time as described below.
Integers must represent UNIX time in seconds.
Floats must represent UNIX time in seconds with nanosecond precision`,
				TIME, INTEGER, FLOAT, STRING,
			),
		),
		WithReturn(
			NewFunctionReturn(
				"The resultant time object after conversion. Throws an error if unable to convert.",
				TIME,
			),
		),
		WithTags(FUNCTION_TAG_TYPE_COERCION),
		WithExamples(
			NewProgramExample(
				`{"int_unix": 1759782264}`,
				`//parse an integer as unix time in seconds
SET @out.result = time(@in.int_unix) | string()`,
				`{"result": "2025-10-06T20:24:24Z"}`,
			),
		),
		WithExamples(
			NewProgramExample(
				`{"float_unix": 1759782264.0}`,
				`//parse a float as unix time in seconds w/ nanosecond precision
SET @out.result = time(@in.float_unix) | string()`,
				`{"result": "2025-10-06T20:24:24Z"}`,
			),
			NewProgramExample(
				`{"string": "2025-10-06T20:24:24Z"}`,
				`//parse an RFC3339 string
SET @out.result = time(@in.string) | catch(err ~> { SET return = err }) | string()  `,
				`{"result": "2025-10-06T20:24:24Z"}`,
			),
			NewProgramExample(
				`{"string_unix": "1759782264"}`,
				`//parse a string as unix time in seconds
SET @out.result = time(@in.string_unix) | string()`,
				`{"result": "2025-10-06T20:24:24Z"}`,
			),
		),
	)
}
func builtinTime(ctx context.Context, args ...*Object) *Object {
	if res, ok := IsArgCountEqual(1, args); !ok {
		return res
	}
	a := args[0]
	val, err := a.AsAny()
	if err != nil {
		return ObjectError("unable to convert item to TIME. invalid input type: %s", a.Type())
	}
	return CastTime(val)
}

func builtinDropEntry() *FunctionEntry {
	return NewFunctionEntry(
		"drop",
		"Stops the current run of Morph statements, and returns NULL",
		builtinDrop,

		WithTags(FUNCTION_TAG_FLOW_CONTROL),
		WithExamples(
			NewProgramExample(
				`null`,
				`SET @out.result = 100
drop()`,
				`null`,
			),
		),
	)
}
func builtinDrop(ctx context.Context, args ...*Object) *Object {
	if ret, ok := IsArgCountEqual(0, args); !ok {
		return ret
	}
	return ObjectTerminateDrop
}
func builtinEmitEntry() *FunctionEntry {
	return NewFunctionEntry(
		"emit",
		"Stops the current run of Morph statements, and returns data in its current state",
		builtinEmit,

		WithTags(FUNCTION_TAG_FLOW_CONTROL),
		WithExamples(
			NewProgramExample(
				`null`,
				`SET @out.result = 100
emit()
SET @out.result = 0`,
				`{"result": 100}`,
			),
		),
	)
}
func builtinEmit(ctx context.Context, args ...*Object) *Object {
	if ret, ok := IsArgCountEqual(0, args); !ok {
		return ret
	}
	return ObjectTerminate
}

func builtinLenEntry() *FunctionEntry {
	return NewFunctionEntry(
		"len",
		"Gets the length of the target string, array, or map",
		builtinLen,

		WithArgs(
			NewFunctionArg(
				"target",
				"The item to check for length",
				STRING, ARRAY, MAP,
			),
		),
		WithReturn(
			NewFunctionReturn(
				"The length of the target",
				INTEGER,
			),
		),
		WithTags(FUNCTION_TAG_GENERAL, FUNCTION_TAG_MAPS, FUNCTION_TAG_ARRAYS, FUNCTION_TAG_STRINGS),
		WithExamples(
			NewProgramExample(
				`null`,
				`//len of string
SET @out.result = len("car")`,
				`{"result": 3}`,
			),
			NewProgramExample(
				`null`,
				`//len of array
SET @out.result = len(["a", "b", "c"])`,
				`{"result": 3}`,
			),
			NewProgramExample(
				`{"k1": "v1", "k2": "v2", "k3": "v3"}`,
				`//len of map
SET @out.result = len(@in)`,
				`{"result": 3}`,
			),
		),
	)
}
func builtinLen(ctx context.Context, args ...*Object) *Object {
	if len(args) != 1 {
		return ObjectError("function len() requires a single argument. got=%d", len(args))
	}
	arg := args[0]
	acceptableTypes := []string{string(ARRAY), string(STRING), string(MAP)}
	if !slices.Contains(acceptableTypes, arg.Type()) {
		return ObjectError("invalid argument type. Expect one of STRING, ARRAY, or MAP. got=%s", arg.Type())
	}
	ret := 0
	switch arg.Type() {
	case string(STRING):
		s, err := arg.AsString()
		if err != nil {
			ObjectError(err.Error())
		}
		ret = len(s)
	case string(ARRAY):
		a, err := arg.AsArray()
		if err != nil {
			return ObjectError(err.Error())
		}
		ret = len(a)
	case string(MAP):
		m, err := arg.AsMap()
		if err != nil {
			return ObjectError(err.Error())
		}
		ret = len(m)
	}
	return CastInt(ret)
}

//
//
// OLD

func newBuiltinFuncStore() *functionStore {
	store := newFunctionStore()

	store.Register(builtinCatchEntryOld())
	store.Register(builtinCoalesceEntryOld())
	store.Register(builtinFallbackEntryOld())

	store.Register(builtinDropEntryOld())
	store.Register(builtinEmitEntryOld())

	store.Register(builtinIntEntryOld())
	store.Register(builtinFloatEntryOld())
	store.Register(builtinStringEntryOld())
	store.Register(builtinTimeEntryOld())

	store.Register(builtinLenEntryOld())
	store.Register(builtinMinEntry())
	store.Register(builtinMaxEntry())

	store.Register(builtinContainsEntry())
	store.Register(builtinAppendEntry())

	store.Register(builtinMapEntry())
	store.Register(builtinFilterEntry())
	store.Register(builtinReduceEntry())

	store.Register(builtinNowEntry())
	store.Register(builtinParseTimeEntry())

	return store
}

// len
func builtinLenEntryOld() *functionEntry {
	l := NewFunctionEntryOld("len", builtinLenOld)
	l.SetArgument("target", STRING, ARRAY, MAP)
	l.SetReturn("length", INTEGER)
	return l
}

func builtinLenOld(args ...*Object) *Object {
	if len(args) != 1 {
		return ObjectError("function len() requires a single argument. got=%d", len(args))
	}
	arg := args[0]
	acceptableTypes := []string{string(ARRAY), string(STRING), string(MAP)}
	if !slices.Contains(acceptableTypes, arg.Type()) {
		return ObjectError("invalid argument type. Expect one of STRING, ARRAY, or MAP. got=%s", arg.Type())
	}
	ret := 0
	switch arg.Type() {
	case string(STRING):
		s, err := arg.AsString()
		if err != nil {
			ObjectError(err.Error())
		}
		ret = len(s)
	case string(ARRAY):
		a, err := arg.AsArray()
		if err != nil {
			return ObjectError(err.Error())
		}
		ret = len(a)
	case string(MAP):
		m, err := arg.AsMap()
		if err != nil {
			return ObjectError(err.Error())
		}
		ret = len(m)
	}
	return CastInt(ret)
}

// contains
func builtinContainsEntry() *functionEntry {
	l := NewFunctionEntryOld("contains", builtinContains)
	l.SetArgument("item", STRING, ARRAY)
	l.SetArgument("target", ANY...)
	l.SetReturn("result", BOOLEAN)
	return l
}

func builtinContains(args ...*Object) *Object {
	if res, ok := IsArgCountEqual(2, args); !ok {
		return res
	}
	main := args[0]
	sub := args[1]

	ret := false
	switch main.Type() {
	case string(STRING):
		if sub.Type() != string(STRING) {
			return ObjectError("second argument of contains() cannot be a non-string, if the first argument is a string. got type=%s", sub.Type())
		}
		mainString, err := main.AsString()
		if err != nil {
			return ObjectError(err.Error())
		}
		subString, err := sub.AsString()
		if err != nil {
			return ObjectError(err.Error())
		}
		ret = strings.Contains(mainString, subString)
	case string(ARRAY):
		mainArr, err := main.AsArray()
		if err != nil {
			return ObjectError(err.Error())
		}
		subItem, err := sub.AsAny()
		if err != nil {
			return ObjectError(err.Error())
		}
		ret = slices.Contains(mainArr, subItem)
	}
	return CastBool(ret)
}

//min

func builtinMinEntry() *functionEntry {
	fe := NewFunctionEntryOld("min", builtinMin)
	fe.SetArgument("num1", INTEGER, FLOAT)
	fe.SetArgument("num2", INTEGER, FLOAT)
	fe.SetReturn("minimum", INTEGER, FLOAT)

	return fe
}

func builtinMin(args ...*Object) *Object {
	if len(args) != 2 {
		return ObjectError("function min() requires a single argument. got=%d", len(args))
	}
	bothInt := (args[0].Type() == args[1].Type()) && (args[0].Type() == string(INTEGER))

	cmpList := []float64{}
	for idx, arg := range args[:2] {
		switch arg.Type() {
		case string(INTEGER):
			i, err := arg.AsInt()
			if err != nil {
				return ObjectError("min(): argument at position %d is an invalid INTEGER", idx+1)
			}
			cmpList = append(cmpList, float64(i))
		case string(FLOAT):
			f, err := arg.AsFloat()
			if err != nil {
				return ObjectError("min(): argument at position %d is an invalid INTEGER", idx+1)
			}
			cmpList = append(cmpList, f)
		}
	}
	min := slices.Min(cmpList)
	if bothInt {
		return CastInt(min)
	}
	return CastFloat(min)
}

// max

func builtinMaxEntry() *functionEntry {
	fe := NewFunctionEntryOld("max", builtinMax)
	fe.SetArgument("num1", INTEGER, FLOAT)
	fe.SetArgument("num2", INTEGER, FLOAT)
	fe.SetReturn("minimum", INTEGER, FLOAT)
	return fe
}
func builtinMax(args ...*Object) *Object {
	if res, ok := IsArgCountEqual(2, args); !ok {
		return res
	}
	bothInt := (args[0].Type() == args[1].Type()) && (args[0].Type() == string(INTEGER))
	cmpList := []float64{}
	for idx, arg := range args[:2] {
		switch arg.Type() {
		case string(INTEGER):
			i, err := arg.AsInt()
			if err != nil {
				return ObjectError("min(): argument at position %d is an invalid INTEGER", idx+1)
			}
			cmpList = append(cmpList, float64(i))
		case string(FLOAT):
			f, err := arg.AsFloat()
			if err != nil {
				return ObjectError("min(): argument at position %d is an invalid INTEGER", idx+1)
			}
			cmpList = append(cmpList, f)
		}
	}
	max := slices.Max(cmpList)
	if bothInt {
		return CastInt(max)
	}
	return CastFloat(max)
}

// drop
func builtinDropEntryOld() *functionEntry {
	fe := NewFunctionEntryOld("drop", builtinDropOld)
	return fe
}

func builtinDropOld(args ...*Object) *Object {
	if ret, ok := IsArgCountEqual(0, args); !ok {
		return ret
	}
	return ObjectTerminateDrop
}

// emit
func builtinEmitEntryOld() *functionEntry {
	fe := NewFunctionEntryOld("emit", builtinEmitOld)
	return fe
}

func builtinEmitOld(args ...*Object) *Object {
	if ret, ok := IsArgCountEqual(0, args); !ok {
		return ret
	}
	return ObjectTerminate
}

// int
func builtinIntEntryOld() *functionEntry {
	fe := NewFunctionEntryOld("int", builtinIntOld)
	fe.SetArgument("target", INTEGER, FLOAT, STRING)
	fe.SetReturn("result", INTEGER)
	return fe
}

func builtinIntOld(args ...*Object) *Object {
	if res, ok := IsArgCountEqual(1, args); !ok {
		return res
	}
	a := args[0]
	val, err := a.AsAny()
	if err != nil {
		return ObjectError("unable to convert item to INTEGER. invalid input type: %s", a.Type())
	}
	return CastInt(val)
}

// float
func builtinFloatEntryOld() *functionEntry {
	fe := NewFunctionEntryOld("float", builtinFloatOld)
	fe.SetArgument("target", INTEGER, FLOAT, STRING)
	fe.SetReturn("result", FLOAT)
	return fe
}

func builtinFloatOld(args ...*Object) *Object {
	if res, ok := IsArgCountEqual(1, args); !ok {
		return res
	}
	a := args[0]
	val, err := a.AsAny()
	if err != nil {
		return ObjectError("unable to convert item to FLOAT. invalid input type: %s", a.Type())
	}
	return CastFloat(val)
}

// string
func builtinStringEntryOld() *functionEntry {
	fe := NewFunctionEntryOld("string", builtinStringOld)
	fe.SetArgument("target", INTEGER, FLOAT, STRING, BOOLEAN)
	fe.SetReturn("result", STRING)
	return fe
}

func builtinStringOld(args ...*Object) *Object {
	if res, ok := IsArgCountEqual(1, args); !ok {
		return res
	}
	a := args[0]
	val, err := a.AsAny()
	if err != nil {
		return ObjectError("unable to convert item to STRING. invalid input type: %s", a.Type())
	}
	return CastString(val)
}

// catch (handle errors)
func builtinCatchEntryOld() *functionEntry {
	fe := NewFunctionEntryOld("catch", builtinCatchOld)
	fe.SetArgument("target", ANY...)
	fe.SetArgument("fallback", BOOLEAN, INTEGER, FLOAT, STRING, ARRAY, MAP)
	fe.SetReturn("result", ANY...)
	return fe
}

func builtinCatchOld(args ...*Object) *Object {
	if res, ok := IsArgCountEqual(2, args); !ok {
		return res
	}
	target := args[0]
	fallback := args[1]
	target.Type()
	if target.Type() == string(ERROR) {
		return fallback
	}
	return target
}

// coalesce (catch nulls)
func builtinCoalesceEntryOld() *functionEntry {
	fe := NewFunctionEntryOld("coalesce", builtinCoalesceOld)
	fe.SetArgument("target", ANY...)
	fe.SetArgument("fallback", BOOLEAN, INTEGER, FLOAT, STRING, ARRAY, MAP)
	fe.SetReturn("result", ANY...)
	return fe
}

func builtinCoalesceOld(args ...*Object) *Object {
	if res, ok := IsArgCountEqual(2, args); !ok {
		return res
	}
	target := args[0]
	fallback := args[1]
	if target.Type() == string(NULL) {
		return fallback
	}
	return target
}

// fallback (catch errors or nulls)
func builtinFallbackEntryOld() *functionEntry {
	fe := NewFunctionEntryOld("fallback", builtinFallbackOld)
	fe.SetArgument("target", ANY...)
	fe.SetArgument("fallback", BOOLEAN, INTEGER, FLOAT, STRING, ARRAY, MAP)
	fe.SetReturn("result", ANY...)
	return fe
}

func builtinFallbackOld(args ...*Object) *Object {
	if res, ok := IsArgCountEqual(2, args); !ok {
		return res
	}
	target := args[0]
	fallback := args[1]
	if target.Type() == string(NULL) || target.Type() == string(ERROR) {
		return fallback
	}
	return target
}

// map
func builtinMapEntry() *functionEntry {
	fe := NewFunctionEntryOld("map", builtinMap)
	fe.SetArgument("input_data", "the array or map on which to apply the statements", ARRAY, MAP)
	fe.SetArgument("function", ARROWFUNC)
	fe.SetReturn("result", ARRAY, MAP)
	return fe
}

func builtinMap(args ...*Object) *Object {
	if res, ok := IsArgCountEqual(2, args); !ok {
		return res
	}
	arrowFn, err := args[1].AsArrowFunction()
	if err != nil {
		return ObjectError("invalid argument for map(): second argument must be a valid ARROWFUNC. got type of %s", args[1].Type())
	}
	switch args[0].Type() {
	case string(MAP):
		in, err := args[0].AsMap()
		if err != nil {
			return ObjectError("error calling map(): data issue with first argument of type %s", args[0].Type())
		}
		ret := make(map[string]interface{})
		keyList := []string{}
		for k := range in {
			keyList = append(keyList, k)
		}
		slices.Sort(keyList)
		for _, key := range keyList {
			value := in[key]
			input := make(map[string]interface{})
			input["key"] = key
			input["value"] = value
			subEnv := arrowFn.Run(input)
			if arrowFn.HasError() {
				return arrowFn.errObj
			}
			m, ok := subEnv.(map[string]interface{})
			if !ok {
				return ObjectError("error calling map() arrow function: unable to extract return value from arrow function")
			}
			out, ok := m["return"]
			if !ok {
				ret[key] = value // if return is nil, simply use the existing entry
				continue
			}
			retMap, ok := out.(map[string]interface{}) // if return is not a map (ie can't access the following: return.key, return.value), simply use the return as the map value, and use the existing key
			if !ok {
				ret[key] = out
				continue
			}
			if newKey, ok := retMap["key"]; ok { // if return.key exists and is a string, we use that as the new key
				if newKeyStr, ok := newKey.(string); ok && !slices.Contains(keyList, newKeyStr) {
					key = newKeyStr
				}
			}
			// either assign the new value if it exists, or keep the existing one if it doesn't.
			newVal, ok := retMap["value"]
			if !ok {
				ret[key] = value
				continue
			}
			ret[key] = newVal
		}
		return CastMap(ret)
	case string(ARRAY):
		in, err := args[0].AsArray()
		if err != nil {
			return ObjectError("error calling map(): data issue with first argument of type %s", args[0].Type())
		}
		ret := []interface{}{}
		for idx, entry := range in {
			input := make(map[string]interface{})
			input["index"] = int64(idx)
			input["value"] = entry
			subEnv := arrowFn.Run(input)
			if arrowFn.HasError() {
				return arrowFn.GetError()
			}
			m, ok := subEnv.(map[string]interface{})
			if !ok {
				return ObjectError("error calling map() arrow function: unable to extract return value from arrow function")
			}
			toAdd, ok := m["return"]
			if !ok {
				ret = append(ret, entry)
				continue
			}
			ret = append(ret, toAdd)
		}

		return CastArray(ret)
	default:
		return ObjectError("invalid argument for map(): first argument must be an ARRAY or MAP. got type of %s", args[0].Type())
	}
}

// filter
func builtinFilterEntry() *functionEntry {
	fe := NewFunctionEntryOld("filter", builtinFilter)
	fe.SetArgument("input_data", "the array or map on which to apply the filter statements", ARRAY, MAP)
	fe.SetArgument("function", ARROWFUNC)
	fe.SetReturn("result", ARRAY, MAP)
	return fe
}

func builtinFilter(args ...*Object) *Object {
	if res, ok := IsArgCountEqual(2, args); !ok {
		return res
	}
	arrowFn, err := args[1].AsArrowFunction()
	if err != nil {
		return ObjectError("filter() second argument must be a valid ARROWFUNC. got type of %s", args[1].Type())
	}

	switch args[0].Type() {
	case string(MAP):
		in, err := args[0].AsMap()
		if err != nil {
			return ObjectError("filter() input data argument issue. type is not compatible with map operation: %s", args[0].Type())
		}
		keyList := []string{}
		for k := range in {
			keyList = append(keyList, k)
		}
		slices.Sort(keyList)
		ret := make(map[string]interface{})
		for _, key := range keyList {
			value := in[key]
			input := make(map[string]interface{})
			input["key"] = key
			input["value"] = value
			subEnv := arrowFn.Run(input)
			if arrowFn.HasError() {
				return arrowFn.errObj
			}
			m, ok := subEnv.(map[string]interface{})
			if !ok {
				return ObjectError("filter() unable to extract return value from arrow function")
			}
			out, ok := m["return"]
			if !ok {
				continue
			}
			if resBool, ok := out.(bool); ok {
				if resBool {
					ret[key] = value
				}
			}
		}
		return CastMap(ret)
	case string(ARRAY):
		in, err := args[0].AsArray()
		if err != nil {
			return ObjectError("filter() input data argument issue. type is not compatible with array operation: %s", args[0].Type())
		}
		ret := []interface{}{}
		for idx, entry := range in {
			input := make(map[string]interface{})
			input["index"] = int64(idx)
			input["value"] = entry
			subEnv := arrowFn.Run(input)
			if arrowFn.HasError() {
				return arrowFn.GetError()
			}
			m, ok := subEnv.(map[string]interface{})
			if !ok {
				return ObjectError("filter() unable to extract return value from arrow function")
			}
			out, ok := m["return"]
			if !ok {
				continue
			}
			if resBool, ok := out.(bool); ok {
				if resBool {
					ret = append(ret, entry)
				}
			}
		}
		return CastAuto(ret)
	default:
		return ObjectError("invalid argument for filter(): first argument must be an ARRAY or MAP. got type of %s", args[0].Type())
	}
}

// reduce
func builtinReduceEntry() *functionEntry {
	fe := NewFunctionEntryOld("reduce", builtinReduce)
	fe.SetArgument("input_data", ARRAY, MAP)
	fe.SetArgument("accumulator", STRING, INTEGER, FLOAT, BOOLEAN, ARRAY, MAP, NULL)
	fe.SetArgument("function", ARROWFUNC)
	fe.SetReturn("result", STRING, INTEGER, FLOAT, BOOLEAN, ARRAY, MAP, NULL)
	return fe
}

func builtinReduce(args ...*Object) *Object {
	if res, ok := IsArgCountEqual(3, args); !ok {
		return res
	}
	arrowFn, err := args[2].AsArrowFunction()
	if err != nil {
		return ObjectError("reduce() third argument must be a valid ARROWFUNC. got type of %s", args[2].Type())
	}

	acc, err := args[1].AsAny()
	if err != nil {
		return ObjectError(err.Error())
	}

	switch args[0].Type() {
	case string(MAP):
		in, err := args[0].AsMap()
		if err != nil {
			return ObjectError("reduce() input data argument issue. type is not compatible with MAP operations: %s", args[0].Type())
		}
		ret := acc
		keyList := []string{}
		for k := range in {
			keyList = append(keyList, k)
		}
		slices.Sort(keyList)
		for _, key := range keyList {
			value := in[key]
			input := make(map[string]interface{})
			input["key"] = key
			input["value"] = value
			input["current"] = ret
			subEnv := arrowFn.Run(input)
			if arrowFn.HasError() {
				return arrowFn.GetError()
			}
			m, ok := subEnv.(map[string]interface{})
			if !ok {
				return ObjectError("reduce() unable to extract return value from arrow function")
			}
			if out, ok := m["return"]; ok {
				ret = out
			}
		}
		return CastAuto(ret)
	case string(ARRAY):
		in, err := args[0].AsArray()
		if err != nil {
			return ObjectError("reduce() input data argument issue. type is not compatible with array operation: %s", args[0].Type())
		}
		ret := acc
		for idx, entry := range in {
			input := make(map[string]interface{})
			input["value"] = entry
			input["index"] = int64(idx)
			input["current"] = ret
			subEnv := arrowFn.Run(input)
			if arrowFn.HasError() {
				return arrowFn.GetError()
			}
			m, ok := subEnv.(map[string]interface{})
			if !ok {
				return ObjectError("reduce() unable to extract return value from arrow function")
			}
			if out, ok := m["return"]; ok {
				ret = out
			}
		}
		return CastAuto(ret)
	default:
		return ObjectError("invalid argument for reduce(): first argument must be an ARRAY or MAP. got type of %s", args[0].Type())
	}
}

// append
func builtinAppendEntry() *functionEntry {
	fe := NewFunctionEntryOld("append", builtinAppend)
	fe.SetArgument("arr", ARRAY)
	fe.SetArgument("to_add", INTEGER, FLOAT, STRING, ARRAY, MAP, NULL)
	fe.SetReturn("result", ARRAY)
	return fe
}

func builtinAppend(args ...*Object) *Object {
	if res, ok := IsArgCountEqual(2, args); !ok {
		return res
	}
	arr, err := args[0].AsArray()
	if err != nil {
		return ObjectError("append() invalid array argument of type %s", args[0].Type())
	}
	toAdd, err := args[1].AsAny()
	if err != nil {
		return ObjectError("append() invalid second argument of type %s", args[1].Type())
	}
	arr = append(arr, toAdd)
	return CastArray(arr)
}

// time
func builtinTimeEntryOld() *functionEntry {
	fe := NewFunctionEntryOld("time", builtinTimeOld)
	fe.SetArgument("input", TIME, INTEGER, FLOAT, STRING)
	fe.SetReturn("result", TIME)
	return fe
}

func builtinTimeOld(args ...*Object) *Object {
	if res, ok := IsArgCountEqual(1, args); !ok {
		return res
	}
	a := args[0]
	val, err := a.AsAny()
	if err != nil {
		return ObjectError("unable to convert item to TIME. invalid input type: %s", a.Type())
	}
	return CastTime(val)
}

func builtinParseTimeEntry() *functionEntry {
	fe := NewFunctionEntryOld("parse_time", builtinParseTime)
	fe.SetArgument("input", INTEGER, FLOAT, STRING)
	fe.SetArgument("format string", STRING)
	fe.SetReturn("result", TIME)
	return fe
}

func builtinParseTime(args ...*Object) *Object {
	if res, ok := IsArgCountEqual(2, args); !ok {
		return res
	}
	a1 := args[0]
	a2 := args[1]
	fmtString, err := a2.AsString()
	if err != nil {
		return ObjectError("unable to convert item to TIME. invalid input type for second argument. got=%s", a2.Type())
	}
	switch strings.ToLower(fmtString) {
	case "rfc_3339":
		inputString, err := a1.AsString()
		if err != nil {
			return ObjectError("unable to convert item to TIME. parsing the time format from rfc_3339 requires the first argument to be a STRING type. got=%s", a1.Type())
		}
		t, err := time.Parse(time.RFC3339, inputString)
		if err != nil {
			return ObjectError("unable to convert item to TIME. invalid TIME string: %s", inputString)
		}
		return CastTime(t)
	case "rfc_3339_nano":
		inputString, err := a1.AsString()
		if err != nil {
			return ObjectError("unable to convert item to TIME. parsing the time format from rfc_3339 requires the first argument to be a STRING type. got=%s", a1.Type())
		}
		t, err := time.Parse(time.RFC3339Nano, inputString)
		if err != nil {
			return ObjectError("unable to convert item to TIME. invalid TIME string: %s", inputString)
		}
		return CastTime(t)
	case "unix":
		arg1Any, err := a1.AsAny()
		if err != nil {
			return ObjectError("unable to convert item to TIME. invalid first argument")
		}
		switch a1.Type() {
		case string(INTEGER), string(FLOAT), string(STRING):
			return CastTime(arg1Any)
		default:
			return ObjectError("unable to convert item to TIME. parsing the time from unix seconds requires the first argument to be an INT, FLOAT, or STRING type: %s", a1.Type())
		}
	case "unix_micro":
		if argInt, err := a1.AsInt(); err == nil {
			return CastTime(time.UnixMicro(argInt).UTC())
		}
		if argStr, err := a1.AsString(); err == nil {
			asInt, err := strconv.ParseInt(argStr, 10, 64)
			if err != nil {
				return ObjectError("unable to convert item to TIME. invalid TIME string: %s", argStr)
			}
			return CastTime(time.UnixMicro(asInt).UTC())
		}
		return ObjectError("unable to convert item to TIME. first argument must be an INTEGER or STRING when parsing from unix micro")
	case "unix_milli":
		if argInt, err := a1.AsInt(); err == nil {
			return CastTime(time.UnixMilli(argInt).UTC())
		}
		if argStr, err := a1.AsString(); err == nil {
			asInt, err := strconv.ParseInt(argStr, 10, 64)
			if err != nil {
				return ObjectError("unable to convert item to TIME. invalid TIME string: %s", argStr)
			}
			return CastTime(time.UnixMilli(asInt).UTC())
		}
		return ObjectError("unable to convert item to TIME. first argument must be an INTEGER or STRING when parsing from unix milli")
	case "unix_nano":
		if argInt, err := a1.AsInt(); err == nil {
			return CastTime(time.Unix(0, argInt).UTC())
		}
		if argStr, err := a1.AsString(); err == nil {
			asInt, err := strconv.ParseInt(argStr, 10, 64)
			if err != nil {
				return ObjectError("unable to convert item to TIME. invalid TIME string: %s", argStr)
			}
			return CastTime(time.Unix(0, asInt).UTC())
		}
		return ObjectError("unable to convert item to TIME. first argument must be an INTEGER or STRING when parsing from unix nano")
	default:
		inputString, err := a1.AsString()
		if err != nil {
			return ObjectError("unable to convert item to TIME. parsing an arbitrary time format requires the first argument to be a STRING tpye. got=%s", a1.Type())
		}
		t, err := time.Parse(fmtString, inputString)
		if err != nil {
			return ObjectError("unable to convert item to TIME. issue parsing time %s with format string %s: %s", inputString, fmtString, err.Error())
		}
		return CastTime(t)
	}

}

func builtinNowEntry() *functionEntry {
	fe := NewFunctionEntryOld("now", builtinNow)
	fe.SetReturn("current time", TIME)
	return fe
}

func builtinNow(args ...*Object) *Object {
	if res, ok := IsArgCountEqual(0, args); !ok {
		return res
	}
	return CastTime(time.Now().UTC())
}
