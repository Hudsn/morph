package morph

import "testing"

func TestFunctionConvert(t *testing.T) {
	functionFromNative(myTestFunction)
}

func myTestFunction(a int, b string) []string {
	return nil
}
