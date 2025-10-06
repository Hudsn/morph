package morph

import (
	"slices"
	"strings"
	"time"
)

func newBuiltinFuncStore() *functionStore {
	store := newFunctionStore()

	store.Register(builtinCatchEntry())
	store.Register(builtinCoalesceEntry())
	store.Register(builtinFallbackEntry())

	store.Register(builtinDropEntry())
	store.Register(builtinEmitEntry())

	store.Register(builtinIntEntry())
	store.Register(builtinFloatEntry())
	store.Register(builtinStringEntry())

	store.Register(builtinLenEntry())
	store.Register(builtinMinEntry())
	store.Register(builtinMaxEntry())

	store.Register(builtinContainsEntry())
	store.Register(builtinAppendEntry())

	store.Register(builtinMapEntry())
	store.Register(builtinReduceEntry())
	store.Register(builtinFilterEntry())

	store.Register(builtinTimeEntry())
	store.Register(builtinNowEntry())

	return store
}

// len
func builtinLenEntry() *functionEntry {
	l := NewFunctionEntry("len", builtinLen)
	l.SetArgument("target", STRING, ARRAY, MAP)
	l.SetReturn("length", INTEGER)
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
	fe := NewFunctionEntry("min", builtinMin)
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
	fe := NewFunctionEntry("max", builtinMax)
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
func builtinDropEntry() *functionEntry {
	fe := NewFunctionEntry("drop", builtinDrop)
	return fe
}

func builtinDrop(args ...*Object) *Object {
	if ret, ok := IsArgCountEqual(0, args); !ok {
		return ret
	}
	return ObjectTerminateDrop
}

// emit
func builtinEmitEntry() *functionEntry {
	fe := NewFunctionEntry("emit", builtinEmit)
	return fe
}

func builtinEmit(args ...*Object) *Object {
	if ret, ok := IsArgCountEqual(0, args); !ok {
		return ret
	}
	return ObjectTerminate
}

// int
func builtinIntEntry() *functionEntry {
	fe := NewFunctionEntry("int", builtinInt)
	fe.SetArgument("target", INTEGER, FLOAT, STRING)
	fe.SetReturn("result", INTEGER)
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
	fe.SetArgument("target", INTEGER, FLOAT, STRING)
	fe.SetReturn("result", FLOAT)
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
	fe.SetArgument("target", INTEGER, FLOAT, STRING, BOOLEAN)
	fe.SetReturn("result", STRING)
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
	fe.SetArgument("target", ANY...)
	fe.SetArgument("fallback", BOOLEAN, INTEGER, FLOAT, STRING, ARRAY, MAP)
	fe.SetReturn("result", ANY...)
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
	fe.SetArgument("target", ANY...)
	fe.SetArgument("fallback", BOOLEAN, INTEGER, FLOAT, STRING, ARRAY, MAP)
	fe.SetReturn("result", ANY...)
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
	fe.SetArgument("target", ANY...)
	fe.SetArgument("fallback", BOOLEAN, INTEGER, FLOAT, STRING, ARRAY, MAP)
	fe.SetReturn("result", ANY...)
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
	fe := NewFunctionEntry("filter", builtinFilter)
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
	fe := NewFunctionEntry("reduce", builtinReduce)
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
	fe := NewFunctionEntry("append", builtinAppend)
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
func builtinTimeEntry() *functionEntry {
	fe := NewFunctionEntry("time", builtinTime)
	fe.SetArgument("input", TIME, INTEGER, FLOAT, STRING)
	fe.SetReturn("result", TIME)
	return fe
}

func builtinTime(args ...*Object) *Object {
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

func builtinNowEntry() *functionEntry {
	fe := NewFunctionEntry("now", builtinNow)
	fe.SetReturn("current time", TIME)
	return fe
}

func builtinNow(args ...*Object) *Object {
	if res, ok := IsArgCountEqual(0, args); !ok {
		return res
	}
	return CastTime(time.Now().UTC())
}
