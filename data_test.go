package morph

import (
	"testing"
)

func TestDataObjectFromBytes(t *testing.T) {
	b := []byte(`{"key": 5}`)
	obj, err := newObjectFromBytes(b)
	if err != nil {
		t.Fatal(err.Error())
	}
	want := map[string]interface{}{
		"key": 5,
	}
	testDataObject(t, obj, want)
}

func testDataObject(t *testing.T, data object, want interface{}) bool {
	switch v := want.(type) {
	case int:
		return testDataObjectInt(t, data, int64(v))
	case float64:
		return testDataObjectFloat(t, data, v)
	case bool:
		return testDataObjectBool(t, data, v)
	case map[string]interface{}:
		return testDataObjectMap(t, data, v)
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

func testDataObjectFloat(t *testing.T, data object, want float64) bool {
	floatObj, ok := data.(*objectFloat)
	if !ok {
		t.Errorf("data is not of type *objectFloat. got=%T", data)
		return false
	}
	if floatObj.value != want {
		t.Errorf("float value is incorrect. want=%f got=%f", want, floatObj.value)
		return false
	}
	return true
}
func testDataObjectBool(t *testing.T, data object, want bool) bool {
	boolObject, ok := data.(*objectBoolean)
	if !ok {
		t.Errorf("data is not of type *objectBoolean. got=%T", data)
		return false
	}
	if boolObject.value != want {
		t.Errorf("bool value is incorrect. want=%t got=%t", want, boolObject.value)
		return false
	}
	return true
}

func testDataObjectMap(t *testing.T, data object, want map[string]interface{}) bool {
	mapObject, ok := data.(*objectMap)
	if !ok {
		t.Errorf("data is not of type *objectBoolean. got=%T", data)
		return false
	}
	for wantKey, wantVal := range want {
		gotVal, ok := mapObject.kvPairs[wantKey]
		if !ok {
			t.Errorf("objectMap does not contain desired key: %s", wantKey)
			return false
		}
		if !testDataObject(t, gotVal.value, wantVal) {
			return false
		}
	}

	return true
}
