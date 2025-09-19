package morph

import (
	"fmt"
	"slices"
)

func newBuiltinFuncStore() *functionStore {
	store := newFunctionStore()

	store.Register(builtinLenEntry())

	return store
}

// len
func builtinLenEntry() *functionEntry {
	l := NewFunctionEntry("len", builtinLen)
	l.SetDescription("return the length of the target object. object must be a string, array, or map")
	l.SetArgument("target", "target object for length operation", STRING, ARRAY, MAP)
	l.SetReturn("length", "length of the object", INTEGER)
	return l
}

func builtinLen(args ...*Object) (*Object, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("function len() requires a single argument. got=%d", len(args))
	}
	arg := args[0]
	acceptableTypes := []string{string(ARRAY), string(STRING), string(MAP)}
	if !slices.Contains(acceptableTypes, arg.Type()) {
		return nil, fmt.Errorf("invalid argument type. Expect one of STRING, ARRAY, or MAP. got=%s", arg.Type())
	}
	ret := 0
	switch arg.Type() {
	case string(STRING):
		s, err := arg.AsString()
		if err != nil {
			return nil, err
		}
		ret = len(s)
	case string(ARRAY):
		a, err := arg.AsArray()
		if err != nil {
			return nil, err
		}
		ret = len(a)
	case string(MAP):
		m, err := arg.AsMap()
		if err != nil {
			return nil, err
		}
		ret = len(m)
	}
	return CastInt(ret)
}

//
// map

//
// filter

//
// reduce
