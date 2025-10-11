package morph

import (
	"reflect"
	"slices"
	"testing"
)

func TestPublicInt(t *testing.T) {
	tests := []struct {
		start interface{}
		want  interface{}
	}{
		{10, 10},
		{float64(5), 5},
		{float32(15), 15},
		{int64(8), 8},
	}

	for _, tt := range tests {
		obj := CastInt(tt.start)
		if isObjectErr(obj.inner) {
			t.Fatal(objectToError(obj.inner))
		}
		testPublicObjectMethods(t, obj, tt.want)
	}
}
func TestPublicFloat(t *testing.T) {
	tests := []struct {
		start interface{}
		want  interface{}
	}{
		{10, float64(10)},
		{10.9, float64(10.9)},
		{int64(2), float64(2)},
		{float32(8), float64(8)},
	}

	for _, tt := range tests {
		obj := CastFloat(tt.start)
		if isObjectErr(obj.inner) {
			t.Fatal(objectToError(obj.inner))
		}
		testPublicObjectMethods(t, obj, tt.want)
	}
}
func TestPublicString(t *testing.T) {
	tests := []string{
		"teststring", "myotherstring",
	}

	for _, tt := range tests {
		obj := CastString(tt)
		if isObjectErr(obj.inner) {
			t.Fatal(objectToError(obj.inner))
		}
		testPublicObjectMethods(t, obj, tt)
	}
}
func TestPublicBool(t *testing.T) {
	tests := []bool{
		true, false,
	}

	for _, tt := range tests {
		obj := CastBool(tt)
		if isObjectErr(obj.inner) {
			t.Fatal(objectToError(obj.inner))
		}
		testPublicObjectMethods(t, obj, tt)
	}
}

func TestPublicArray(t *testing.T) {
	testArr := []interface{}{
		int64(1), float64(2), "three", false,
	}
	obj := CastArray(testArr)
	if isObjectErr(obj.inner) {
		t.Fatal(objectToError(obj.inner))
	}
	testPublicObjectMethods(t, obj, testArr)
}

func TestPublicMap(t *testing.T) {
	testMap := map[string]interface{}{
		"mykey": "mystringval",
		"myotherkey": map[string]interface{}{
			"nested": int64(5),
			"nested_arr": []interface{}{
				int64(10), int64(9), int64(8), int64(7),
			},
		},
	}
	testStruct := struct {
		MyKey      string `json:"mykey"`
		MyOtherKey struct {
			Nested    int   `json:"nested"`
			NestedArr []int `json:"nested_arr"`
		} `json:"myotherkey"`
	}{}

	obj := CastMap(testMap)
	if isObjectErr(obj.inner) {
		t.Fatal(objectToError(obj.inner))
	}

	testPublicObjectMethods(t, obj, testMap)

	err := obj.MapStruct(&testStruct)
	if err != nil {
		t.Fatal(err)
	}
	if testStruct.MyKey != "mystringval" {
		t.Errorf("wrong value for testStruct.MyKey. want=%s got=%s", "mystringval", testStruct.MyKey)
	}
	if testStruct.MyOtherKey.Nested != 5 {
		t.Errorf("wrong value for testStruct.MyOtherKey.Nested. want=%d got=%d", 5, testStruct.MyOtherKey.Nested)
	}
	if !slices.Equal(testStruct.MyOtherKey.NestedArr, []int{10, 9, 8, 7}) {
		t.Errorf("wrong value for testStruct.MyOtherKey.NestedArry. want=%+v got=%+v", []int{10, 9, 8, 7}, testStruct.MyOtherKey.NestedArr)
	}
}

func testPublicObjectMethods(t *testing.T, obj *Object, want interface{}) bool {

	switch v := want.(type) {
	case map[string]interface{}:
		return testPublicObjectAsMap(t, obj, v)
	case []interface{}:
		return testPublicObjectAsArray(t, obj, v)
	case bool:
		return testPublicObjectAsBool(t, obj, v)
	case string:
		return testPublicObjectAsString(t, obj, v)
	case int:
		return testPublicObjectAsInt(t, obj, v)
	case float64:
		return testPublicObjectAsFloat(t, obj, v)
	default:
		t.Errorf("unsupported target type for Object conversion method: %T", want)
		return false
	}
}

func testPublicObjectAsMap(t *testing.T, obj *Object, want map[string]interface{}) bool {
	m, err := obj.AsMap()
	if err != nil {
		t.Error(err)
		return false
	}
	if !reflect.DeepEqual(m, want) {
		t.Errorf("maps are not equal in value.\n\twant=\n\t\t%+v\n\tgot=\n\t\t%+v", want, m)
	}
	return true
}
func testPublicObjectAsArray(t *testing.T, obj *Object, want []interface{}) bool {
	a, err := obj.AsArray()
	if err != nil {
		t.Error(err)
		return false
	}
	if !reflect.DeepEqual(want, a) {
		t.Errorf("arrays are not equal in value.\n\twant=\n\t\t%+v\n\tgot=\n\t\t%+v", want, a)
	}
	return true
}
func testPublicObjectAsBool(t *testing.T, obj *Object, want bool) bool {
	b, err := obj.AsBool()
	if err != nil {
		t.Error(err)
		return false
	}
	if b != want {
		t.Errorf("wrong value for obj.AsBool. want=%t got=%t", want, b)
		return false
	}
	return true
}
func testPublicObjectAsString(t *testing.T, obj *Object, want string) bool {
	s, err := obj.AsString()
	if err != nil {
		t.Error(err)
		return false
	}
	if s != want {
		t.Errorf("wrong value for obj.AsString. want=%s got=%s", want, s)
		return false
	}
	return true
}
func testPublicObjectAsInt(t *testing.T, obj *Object, want int) bool {
	i, err := obj.AsInt()
	if err != nil {
		t.Error(err)
		return false
	}
	if int(i) != want {
		t.Errorf("wrong value for obj.AsInt. want=%d got=%d", want, i)
		return false
	}
	return true
}
func testPublicObjectAsFloat(t *testing.T, obj *Object, want float64) bool {
	f, err := obj.AsFloat()
	if err != nil {
		t.Error(err)
		return false
	}
	if !isFloatEqual(f, want) {
		t.Errorf("wrong value for obj.AsFloat. want=%f got=%f", want, f)
		return false
	}
	return true

}
