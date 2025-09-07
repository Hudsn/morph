package morph

import (
	"fmt"
	"testing"
)

func TestFunctionConvert(t *testing.T) {
	fn, err := functionFromNative(myTestFunction)
	if err != nil {
		t.Fatal(err)
	}
	m := &objectMap{
		kvPairs: make(map[string]objectMapPair),
	}
	m.kvPairs["asdf"] = objectMapPair{
		key:   "asdf",
		value: &objectInteger{value: 5},
	}

	res, err := fn(m)
	if err != nil {
		t.Fatal(err)
	}

	fmt.Println(res)
}

func myTestFunction(m map[string]interface{}) (int, error) {
	return 0, fmt.Errorf("my error")
}
