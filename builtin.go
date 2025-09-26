package morph

import (
	"slices"
	"strings"
)

func newBuiltinFuncStore() *functionStore {
	store := newFunctionStore()

	store.Register(builtinLenEntry())
	store.Register(builtinMinEntry())
	store.Register(builtinMaxEntry())
	store.Register(builtinDropEntry())
	store.Register(builtinEmitEntry())
	store.Register(builtinIntEntry())
	store.Register(builtinFloatEntry())
	store.Register(builtinStringEntry())
	store.Register(builtinCatchEntry())
	store.Register(builtinCoalesceEntry())
	store.Register(builtinFallbackEntry())
	store.Register(builtinContainsEntry())
	store.Register(builtinAppendEntry())

	store.Register(builtinMapEntry())
	store.Register(builtinReduceEntry())

	return store
}

// len
func builtinLenEntry() *functionEntry {
	l := NewFunctionEntry("len", builtinLen)
	l.SetDescription("return the length of the target object. object must be a string, array, or map")
	l.SetArgument("target", "target object for length operation", STRING, ARRAY, MAP)
	l.SetReturn("length", "length of the object", INTEGER)
	l.SetCategory(FUNC_CAT_GENERAL)
	l.SetExampleInput("mystring")
	l.SetExampleOut("8")
	return l
}

func builtinLen(args ...*Object) *Object {
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
	l := NewFunctionEntry("contains", builtinContains)
	l.SetDescription("return the length of the target object. object must be a string, array, or map")
	l.SetArgument("main", "larger object to check for the presence of sub", STRING, ARRAY)
	l.SetArgument("sub", "sub object to check if it is inside the larger object", ANY...)
	l.SetReturn("result", "whether the main object contains the sub object", BOOLEAN)
	l.SetCategory(FUNC_CAT_GENERAL)
	l.SetExampleInput("mystring", "string")
	l.SetExampleOut("true")
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
	fe := NewFunctionEntry("min", builtinMin)
	fe.SetDescription("returns the minimum value of two numbers")
	fe.SetArgument("num1", "first number to compare", INTEGER, FLOAT)
	fe.SetArgument("num2", "second number to compare", INTEGER, FLOAT)
	fe.SetReturn("minimum", "smallest number of num1 and num2", INTEGER, FLOAT)
	fe.SetCategory(FUNC_CAT_GENERAL)
	fe.SetExampleInput("1", "1.234")
	fe.SetExampleOut("1")
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
	fe := NewFunctionEntry("max", builtinMax)
	fe.SetDescription("returns the maximum value of two numbers")
	fe.SetArgument("num1", "first number to compare", INTEGER, FLOAT)
	fe.SetArgument("num2", "second number to compare", INTEGER, FLOAT)
	fe.SetReturn("minimum", "largest number of num1 and num2", INTEGER, FLOAT)
	fe.SetCategory(FUNC_CAT_GENERAL)
	fe.SetExampleInput("1", "1.234")
	fe.SetExampleOut("1.234")
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
func builtinDropEntry() *functionEntry {
	fe := NewFunctionEntry("drop", builtinDrop)
	fe.SetDescription("return early, effectively clearing all current data being processed and returning nothing")
	fe.SetCategory(FUNC_CAT_CONTROL)
	return fe
}

func builtinDrop(args ...*Object) *Object {
	if len(args) != 0 {
		return ObjectError("function drop() should have 0 arguments. got=%d", len(args))
	}
	return ObjectTerminateDrop
}

// emit
func builtinEmitEntry() *functionEntry {
	fe := NewFunctionEntry("emit", builtinEmit)
	fe.SetDescription("return early, returning data as-is")
	fe.SetCategory(FUNC_CAT_CONTROL)
	return fe
}

func builtinEmit(args ...*Object) *Object {
	if len(args) != 0 {
		ObjectError("function drop() should have 0 arguments. got=%d", len(args))
	}
	return ObjectTerminate
}

// int
func builtinIntEntry() *functionEntry {
	fe := NewFunctionEntry("int", builtinInt)
	fe.SetDescription("attempts to cast the input as an integer")
	fe.SetArgument("target", "the target object to convert to an integer", INTEGER, FLOAT, STRING)
	fe.SetReturn("result", "the resulting integer", INTEGER)
	fe.SetCategory(FUNC_CAT_GENERAL)
	fe.SetExampleInput("1.234")
	fe.SetExampleOut("1")
	return fe
}

func builtinInt(args ...*Object) *Object {
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
func builtinFloatEntry() *functionEntry {
	fe := NewFunctionEntry("float", builtinFloat)
	fe.SetDescription("attempts to cast the input as a float")
	fe.SetArgument("target", "the target object to convert to a float", INTEGER, FLOAT, STRING)
	fe.SetReturn("result", "the resulting float", FLOAT)
	fe.SetCategory(FUNC_CAT_GENERAL)
	fe.SetExampleInput("1")
	fe.SetExampleOut("1.0")
	return fe
}

func builtinFloat(args ...*Object) *Object {
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
func builtinStringEntry() *functionEntry {
	fe := NewFunctionEntry("string", builtinString)
	fe.SetDescription("attempts to cast the input as a string")
	fe.SetArgument("target", "the target object to convert to a string", INTEGER, FLOAT, STRING, BOOLEAN)
	fe.SetReturn("result", "the resulting string", STRING)
	fe.SetCategory(FUNC_CAT_GENERAL)
	fe.SetExampleInput("1.0")
	fe.SetExampleOut(`"1.0"`)
	return fe
}

func builtinString(args ...*Object) *Object {
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
func builtinCatchEntry() *functionEntry {
	fe := NewFunctionEntry("catch", builtinCatch)
	fe.SetDescription("checks if the value is an error. if it is, the second value is returned")
	fe.SetArgument("target", "the target object to check for errors", ANY...)
	fe.SetArgument("fallback", "the value to return in case of an error", INTEGER, FLOAT, STRING, ARRAY, MAP)
	fe.SetReturn("result", "either the original value or the fallback, depending on if an error occurred", ANY...)
	fe.SetCategory(FUNC_CAT_GENERAL)
	fe.SetExampleInput("nonexistent_variable + 1", "1000")
	fe.SetExampleOut("5")
	return fe
}

func builtinCatch(args ...*Object) *Object {
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
func builtinCoalesceEntry() *functionEntry {
	fe := NewFunctionEntry("coalesce", builtinCoalesce)
	fe.SetDescription("checks if the value is null; if it is, the second value is returned")
	fe.SetArgument("target", "the target object to check for null", ANY...)
	fe.SetArgument("fallback", "the value to return in case of a null", INTEGER, FLOAT, STRING, ARRAY, MAP)
	fe.SetReturn("result", "either the original value or the fallback, depending on if a null occurred", ANY...)
	fe.SetCategory(FUNC_CAT_GENERAL)
	fe.SetExampleInput("nonexistent_variable + 1", "1000")
	fe.SetExampleOut("5")
	return fe
}

func builtinCoalesce(args ...*Object) *Object {
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
func builtinFallbackEntry() *functionEntry {
	fe := NewFunctionEntry("fallback", builtinFallback)
	fe.SetDescription("checks if the value is null or an error; if it is, the second value is returned")
	fe.SetArgument("target", "the target object to check for null or error", ANY...)
	fe.SetArgument("fallback", "the value to return in case of a null or error", INTEGER, FLOAT, STRING, ARRAY, MAP)
	fe.SetReturn("result", "either the original value or the fallback, depending on if a null occurred", ANY...)
	fe.SetCategory(FUNC_CAT_GENERAL)
	fe.SetExampleInput("nonexistent_variable + 1", "1000")
	fe.SetExampleOut("5")
	return fe
}

func builtinFallback(args ...*Object) *Object {
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
	fe := NewFunctionEntry("map", builtinMap)
	fe.SetDescription("apply a series of morph statements to each element of an array or map, and return an array/map of entries resulting from the statements applied")
	fe.SetArgument("input_data", "the array or map on which to apply the statements", ARRAY, MAP)
	fe.SetArgument("function", `An arrow function containing the statements to transform each entry of the input data. 
The item before the arrow will be the variable name you use to access the input_data within the arrow function body.
Assign the desired entry to replace the current array/map value to the "return" variable.
When a MAP is the first argument, the "key" and "value" can both be available as subfields of the passed input variable. For example: myfield ~> {set return.key = myfield.key}
When an ARRAY is the first argument, the entry will be directly available by accessing the passed input variable.
	`, ARROWFUNC)
	fe.SetReturn("result", "the final array or map from the series of mapping operations", ARRAY, MAP)
	fe.SetCategory(FUNC_CAT_AGGREGATE)
	fe.SetExampleInput(`{"a": 1, "b": 2}`, `in ~> {
	SET return.key = "prefix_" + in.key
	SET return.value = in.value * 2
}`)
	fe.SetExampleOut(`{"prefix_a": 2, "prefix_b": 4}`)
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
		for key, value := range in {
			input := make(map[string]interface{})
			input["key"] = key
			input["value"] = value
			subEnv, err := arrowFn.Run(input)
			if err != nil {
				return ObjectError("error calling map() arrow function: %s", err.Error())
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
				if newKeyStr, ok := newKey.(string); ok {
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
		for _, entry := range in {
			subEnv, err := arrowFn.Run(entry)
			if err != nil {
				return ObjectError("error calling map() arrow function: %s", err.Error())
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

//
// filter

// reduce
func builtinReduceEntry() *functionEntry {
	fe := NewFunctionEntry("reduce", builtinReduce)
	fe.SetDescription("apply a series of morph statements to each element of an array or map, and return an array/map of entries resulting from the statements applied")
	fe.SetArgument("input_data", "the array or map on which to apply the reducer statements", ARRAY, MAP)
	fe.SetArgument("accumulator", "the initial value of the accumulator to use", STRING, INTEGER, FLOAT, BOOLEAN, ARRAY, MAP, NULL)
	fe.SetArgument("function", `An arrow function containing the statements to modify the accumulator, given the input_data. 
The item before the arrow will be the variable name you use to access the input_data within the arrow function body.
Assign the desired accumulator output to the "return" variable.
The current accumulator state can be accessed by referencing the "current" subfield of the passed input variable.
When a MAP is the first argument, the "key" and "value" can both be available as subfields of the passed input variable.
When an ARRAY is the first argument, the entry will be davailable by accessing the "value" subfield of the input variable; no "key" subfield will be present.
	`, ARROWFUNC)
	fe.SetReturn("result", "the final accumulator value from the series of mapping operations", STRING, INTEGER, FLOAT, BOOLEAN, ARRAY, MAP, NULL)
	fe.SetCategory(FUNC_CAT_AGGREGATE)
	fe.SetExampleInput(`{"a": 1, "b": 2}`, "0", `in ~> {
	SET return = in.current + in.value
}`)
	fe.SetExampleOut(`3`)
	return fe
}

func builtinReduce(args ...*Object) *Object {
	if res, ok := IsArgCountEqual(3, args); !ok {
		return res
	}
	arrowFn, err := args[2].AsArrowFunction()
	if err != nil {
		return ObjectError("reduce() second argument must be a valid ARROWFUNC. got type of %s", args[1].Type())
	}

	acc, err := args[1].AsAny()
	if err != nil {
		return ObjectError("reduce() data issue with accumulator argument: %s", err.Error())
	}

	switch args[0].Type() {
	case string(MAP):
		in, err := args[0].AsMap()
		if err != nil {
			return ObjectError("reduce() input data argument issue. type is not compatible with map operation: %s", args[0].Type())
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
			subEnv, err := arrowFn.Run(input)
			if err != nil {
				return ObjectError("reduce() arrow function error: %s", err.Error())
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
		for _, entry := range in {
			input := make(map[string]interface{})
			input["value"] = entry
			input["current"] = ret
			subEnv, err := arrowFn.Run(input)
			if err != nil {
				return ObjectError("error calling reduce() arrow function: %s", err.Error())
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
		return ObjectError("invalid argument for map(): first argument must be an ARRAY or MAP. got type of %s", args[0].Type())
	}
}

// append
func builtinAppendEntry() *functionEntry {
	fe := NewFunctionEntry("append", builtinAppend)
	fe.SetDescription("appends a value to an array")
	fe.SetArgument("arr", "the target object to check for null or error", ARRAY)
	fe.SetArgument("to_add", "the value to add to the array", INTEGER, FLOAT, STRING, ARRAY, MAP, NULL)
	fe.SetReturn("result", "the new array with the item added", ARRAY)
	fe.SetCategory(FUNC_CAT_GENERAL)
	fe.SetExampleInput("[1, 2]", "3")
	fe.SetExampleOut("[1, 2, 3]")
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
