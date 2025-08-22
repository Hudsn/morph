package morph

import "testing"

func TestDataFromBytes(t *testing.T) {
	b := []byte(`{5}`)
	data, err := newDataFromBytes(b)
	if err != nil {
		t.Fatal(err.Error())
	}
	testDataObject(t, data.contents, 5)
}

func testDataObject(t *testing.T, data object, want interface{}) bool {
	switch v := want.(type) {
	case int:
		return testDataObjectInt(t, data, int64(v))
	default:
		t.Errorf("testDataObject: unsupported data object type. got=%T", want)
		return false
	}
}

func testDataObjectInt(t *testing.T, data object, want int64) bool {
	intobj, ok := data.(*objectInteger)
	if !ok {
		t.Errorf("data is not of type *objectInteger. got=%T", data)
		return false
	}
	if intobj.value != want {
		t.Errorf("integer value is incorrect. want=%d got=%d", want, intobj.value)
		return false
	}
	return true
}
