package morph

import (
	"fmt"
	"reflect"
)

type functionStore map[string]function

type function func(args ...object) (object, error)

// TODO/ WIP
func functionFromNative(fnToRegister func(args ...interface{}) (interface{}, error)) (function, error) {
	fnType := reflect.TypeOf(fnToRegister)
	if fnType.Kind() != reflect.Func {
		return nil, fmt.Errorf("input is not a valid function")
	}

	for idx := range fnType.NumIn() {
		paramType := fnType.In(idx)
		fmt.Println("pos:", idx, "type:", paramType.Kind().String())
	}
	for idx := range fnType.NumOut() {
		retType := fnType.Out(idx)
		fmt.Println("ret_pos:", idx, "type:", retType.Kind().String())
		fmt.Println(retType.Elem())
	}

	return nil, nil
}
