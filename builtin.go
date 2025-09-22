package morph

import (
	"slices"
)

func newBuiltinFuncStore() *functionStore {
	store := newFunctionStore()

	store.Register(builtinLenEntry())
	store.Register(builtinMinEntry())
	store.Register(builtinMaxEntry())
	store.Register(builtinDropEntry())
	store.Register(builtinEmitEntry())
	store.Register(builtinIntEntry())

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

//
// map

//
// filter

//
// reduce

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
	if len(args) != 1 {
		return ObjectError("function int() should have 1 argument. got=%d", len(args))
	}
	a := args[0]
	val, err := a.AsAny()
	if err != nil {
		return ObjectError("unable to convert item to INTEGER. invalid input type: %s", a.Type())
	}
	return CastInt(val)
}

//
// float

//
// string

//
// catch
