package morph

import (
	"fmt"
	"reflect"
)

type functionStore map[string]function

type function func(args ...object) (object, error)

// TODO/ WIP
func functionFromNative(fnToRegister interface{}) (function, error) {
	fnType := reflect.TypeOf(fnToRegister)
	if fnType.Kind() == reflect.Ptr { // if pointer, get the underlying func
		fnType = fnType.Elem()
	}
	if fnType.Kind() != reflect.Func {
		return nil, fmt.Errorf("cannot register function: item is not a function")
	}

	numOut := fnType.NumOut()
	if numOut > 2 {
		return nil, fmt.Errorf("function must return at most 2 values. if 2 values are returned, the second must be of type 'error'")
	}

	if numOut == 2 {
		errPtr := (*error)(nil)
		errorType := reflect.TypeOf(errPtr).Elem()
		if !fnType.Out(1).Implements(errorType) {
			return nil, fmt.Errorf("second return value must be of type error")
		}
	}

	numIn := fnType.NumIn()
	return func(args ...object) (object, error) {
		if len(args) != numIn {
			return obj_global_null, fmt.Errorf("expected %d arguments, got %d", numIn, len(args))
		}

		for idx, arg := range args {
			expectType := fnType.In(idx)
			if expectType.Kind() == reflect.Ptr {
				expectType = expectType.Elem()
			}
			argV, err := objectToReflectValueByType(arg, expectType)
		}

		// TODO
		return nil, nil

	}, nil
}

func objectToReflectValueByType(obj object, targetType reflect.Type) (reflect.Value, error) {
	switch targetType.Kind() {
	case reflect.Int, reflect.Int64:
		if o, ok := obj.(*objectInteger); ok {
			return reflect.ValueOf(o.value).Convert(targetType), nil
		}
	case reflect.Map:
		if o, ok := obj.(*objectMap); ok {
			//TODO
			_ = o
			return reflect.Value{}, nil
		}
	}

	return reflect.Value{}, fmt.Errorf("cannot convert object of type %T to %s", obj, targetType)
}
