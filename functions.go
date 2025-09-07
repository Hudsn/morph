package morph

import (
	"fmt"
	"reflect"
)

type functionStore map[string]function

type function func(args ...object) (object, error)

var refErrPtr = (*error)(nil)
var refErrType = reflect.TypeOf(refErrPtr).Elem()

// TODO/ WIP
func functionFromNative(fnToRegister interface{}) (function, error) {
	fnRef := reflect.ValueOf(fnToRegister)
	fnType := fnRef.Type()
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
		if !fnType.Out(1).Implements(refErrType) {
			return nil, fmt.Errorf("second return value must be of type error")
		}
	}

	numIn := fnType.NumIn()
	return func(args ...object) (object, error) {
		if len(args) != numIn {
			return obj_global_null, fmt.Errorf("expected %d arguments, got %d", numIn, len(args))
		}

		argList := []reflect.Value{}
		for idx, arg := range args {
			expectType := fnType.In(idx)
			if expectType.Kind() == reflect.Ptr {
				expectType = expectType.Elem()
			}
			argV, err := objectToReflectValueByType(arg, expectType)
			if err != nil {
				return obj_global_false, err
			}
			argList = append(argList, argV)
		}

		retVals := fnRef.Call(argList)
		var retObj object = obj_global_null
		for _, entry := range retVals {
			maybeObj, err := reflectValueToObject(entry)
			if err != nil {
				return retObj, err
			}
			retObj = maybeObj
		}

		return retObj, nil

	}, nil
}

func reflectValueToObject(val reflect.Value) (object, error) {
	if val.Type().Implements(refErrType) {
		if val.IsNil() {
			return obj_global_null, nil
		}
		return obj_global_null, val.Interface().(error)
	}
	switch val.Kind() {
	case reflect.Int:
		return &objectInteger{value: val.Int()}, nil
		// case reflect.Map:
		// 	return &objectMap{}, nil
	}
	return obj_global_null, fmt.Errorf("invalid return type for custom function")
}

func objectToReflectValueByType(obj object, targetType reflect.Type) (reflect.Value, error) {
	switch targetType.Kind() {
	case reflect.Int, reflect.Int16, reflect.Int32, reflect.Int64:
		if o, ok := obj.(*objectInteger); ok {
			return reflect.ValueOf(o.value).Convert(targetType), nil
		}
	case reflect.Map:
		if o, ok := obj.(*objectMap); ok {
			return objectMapToReflectValue(o)
		}
	}

	return reflect.Value{}, fmt.Errorf("cannot convert object of type %T to %s", obj, targetType)
}

func objectMapToReflectValue(obj *objectMap) (reflect.Value, error) {
	retMap := make(map[string]interface{})
	for objK, objPairs := range obj.kvPairs {
		objV, err := objectToNativeType(objPairs.value)
		if err != nil {
			return reflect.Value{}, err
		}
		retMap[objK] = objV
	}
	return reflect.ValueOf(retMap), nil
}

func objectToNativeType(obj object) (interface{}, error) {
	switch v := obj.(type) {
	case *objectInteger:
		return v.value, nil
	case *objectMap:
		ret := make(map[string]interface{})
		for k, pair := range v.kvPairs {
			newVal, err := objectToNativeType(pair.value)
			if err != nil {
				return nil, err
			}
			ret[k] = newVal
		}
	}
	return nil, fmt.Errorf("unsupported object type for conversion")
}
