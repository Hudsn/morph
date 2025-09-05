package morph

import (
	"testing"
)

func TestDataObjectFromBytes(t *testing.T) {
	b := []byte(`{
	"key": 5,
	"arr": [1, 2, "three"]
	}`)
	obj, err := newObjectFromBytes(b)
	if err != nil {
		t.Fatal(err.Error())
	}
	want := map[string]interface{}{
		"key": 5,
		"arr": []interface{}{
			1, 2, "three",
		},
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
	case string:
		return testDataObjectString(t, data, v)
	case map[string]interface{}:
		return testDataObjectMap(t, data, v)
	case []interface{}:
		return testDataObjectArray(t, data, v)
	default:
		t.Errorf("testDataObject: unsupported data object type. got=%T", want)
		return false
	}
}

func testDataObjectString(t *testing.T, data object, want string) bool {
	strObj, ok := data.(*objectString)
	if !ok {
		t.Errorf("data is not of type *objectString. got=%T", data)
		return false
	}
	if strObj.value != want {
		t.Errorf("integer value is incorrect. want=%q got=%q", want, strObj.value)
		return false
	}
	return true
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

func testDataObjectArray(t *testing.T, data object, want []interface{}) bool {
	arrObj, ok := data.(*objectArray)
	if !ok {
		t.Errorf("data is not of type *objectArray. got=%T", data)
		return false
	}
	if len(arrObj.entries) != len(want) {
		t.Fatalf("array test length not equal to actual data length. wantLen=%d got=%d", len(want), len(arrObj.entries))
	}
	for idx, entry := range want {
		got := arrObj.entries[idx]
		if !testDataObject(t, got, entry) {
			return false
		}
	}
	return true
}

func testDataObjectMap(t *testing.T, data object, want map[string]interface{}) bool {
	mapObject, ok := data.(*objectMap)
	if !ok {
		t.Errorf("data is not of type *objectMap. got=%T", data)
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
