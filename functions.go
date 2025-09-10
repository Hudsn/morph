package morph

import (
	"encoding/json"
	"fmt"
	"reflect"
)

// public wrapper of object to be used for implementing custom functions
type Object struct {
	inner object
}

// wrappers for public types
const (
	INTEGER = t_integer
	FLOAT   = t_float
	BOOLEAN = t_boolean
	MAP     = t_map
	ARRAY   = t_array
)

func (o *Object) Type() string {
	return string(o.inner.getType())
}

type Function func(args ...*Object) (*Object, error)

func evalFunction(fn Function, args ...object) (object, error) {
	objList := []*Object{}
	for _, arg := range args {
		objList = append(objList, &Object{inner: arg})
	}
	obj, err := fn(objList...)
	if err != nil {
		return obj_global_null, err
	}
	return obj.inner, err
}

func EnforceFunctionArgCount(wantArgNum int, args []*Object) error {
	if wantArgNum != len(args) {
		return fmt.Errorf("incorrect number of arguments. expected=%d got=%d", wantArgNum, len(args))
	}
	return nil
}

func (o *Object) AsInt() (int64, error) {
	i, ok := o.inner.(*objectInteger)
	if !ok {
		return 0, fmt.Errorf("unable to convert object to Integer: underlying structure is not an integer type. got=%s", o.inner.getType())
	}
	return i.value, nil
}
func (o *Object) AsFloat() (float64, error) {
	f, ok := o.inner.(*objectFloat)
	if !ok {
		return 0, fmt.Errorf("unable to convert object to Float: underlying structure is not a float type. got=%s", o.inner.getType())
	}
	return f.value, nil
}
func (o *Object) AsBool() (bool, error) {
	b, ok := o.inner.(*objectBoolean)
	if !ok {
		return false, fmt.Errorf("unable to convert object to Boolean: underlying structure is not a boolean type. got=%s", o.inner.getType())
	}
	return b.value, nil
}
func (o *Object) AsString() (string, error) {
	s, ok := o.inner.(*objectString)
	if !ok {
		return "", fmt.Errorf("unable to convert object to String: underlying structure is not a string type. got=%s", o.inner.getType())
	}
	return s.value, nil
}
func (o *Object) AsMap() (map[string]interface{}, error) {
	m, ok := o.inner.(*objectMap)
	if !ok {
		return nil, fmt.Errorf("unable to convert object to Map: underlying structure is not a map type. got=%s", o.inner.getType())
	}
	res, err := convertMapToNative(m)
	if err != nil {
		return nil, err
	}
	ret, ok := res.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("unable to convert object to Map: underlying structure does not convert to a map[string]interface{}. got=%s", o.inner.getType())
	}
	return ret, nil
}

// takes an object with an underlying map type, and attempts to marshal it into the target *struct.
// you can check the underlying type as a string with Object.Type()
// NOTE: requires that the fields you want to access in your struct are exported. Cannot marshal data into private fields
func (o *Object) MapStruct(target interface{}) error {
	m, err := o.AsMap()
	if err != nil {
		return err
	}
	targetVal := reflect.ValueOf(target)
	if targetVal.Kind() != reflect.Pointer {
		return fmt.Errorf("target must be a pointer")
	}
	if targetVal.IsNil() {
		return fmt.Errorf("target must not be nil")
	}
	if targetVal.Elem().Kind() != reflect.Struct {
		return fmt.Errorf("target must be pointer to a struct")
	}

	b, err := json.Marshal(m)
	if err != nil {
		return fmt.Errorf("MapStruct: cannot convert target to intermediate json format: %w", err)
	}
	err = json.Unmarshal(b, target)
	if err != nil {
		return err
	}
	return nil
}

func (o *Object) AsArray() ([]interface{}, error) {
	a, ok := o.inner.(*objectArray)
	if !ok {
		return nil, fmt.Errorf("unable to convert object to Array: underlying structure is not an array type. got=%s", o.inner.getType())
	}
	res, err := convertArrayToNative(a)
	if err != nil {
		return nil, err
	}
	ret, ok := res.([]interface{})
	if !ok {
		return nil, fmt.Errorf("unable to convert object to Array: underlying structure does not convert to a []interface{}. got=%T", res)
	}
	return ret, nil
}

// typecast helpers

// casts a Go number to a morph Integer Object so it can be used when defining custom functions
// input must be one of: int, int8, int16, int32, int64, float32, float64
func CastInt(value interface{}) (*Object, error) {
	ret := &Object{
		inner: obj_global_null,
	}
	switch v := value.(type) {
	case int:
		ret.inner = &objectInteger{value: int64(v)}
	case int8:
		ret.inner = &objectInteger{value: int64(v)}
	case int16:
		ret.inner = &objectInteger{value: int64(v)}
	case int32:
		ret.inner = &objectInteger{value: int64(v)}
	case int64:
		ret.inner = &objectInteger{value: int64(v)}
	case float32:
		ret.inner = &objectInteger{value: int64(v)}
	case float64:
		ret.inner = &objectInteger{value: int64(v)}
	default:
		return ret, fmt.Errorf("unable to cast type as Int. unsupported type: %T", v)
	}
	return ret, nil
}

// casts a Go number to a morph Float Object so it can be used when defining custom functions
// input must be one of: int, int8, int16, int32, int64, float32, float64
func CastFloat(value interface{}) (*Object, error) {
	ret := &Object{
		inner: obj_global_null,
	}
	switch v := value.(type) {
	case float32:
		ret.inner = &objectFloat{value: float64(v)}
	case float64:
		ret.inner = &objectFloat{value: float64(v)}
	case int:
		ret.inner = &objectFloat{value: float64(v)}
	case int8:
		ret.inner = &objectFloat{value: float64(v)}
	case int16:
		ret.inner = &objectFloat{value: float64(v)}
	case int32:
		ret.inner = &objectFloat{value: float64(v)}
	case int64:
		ret.inner = &objectFloat{value: float64(v)}
	default:
		return ret, fmt.Errorf("unable to cast type as Float. unsupported type: %T", v)
	}
	return ret, nil
}

// casts a Go type to a morph String Object so it can be used when defining custom functions
// input must be one of: int, int8, int16, int32, int64, float32, float64, string, bool
func CastString(value interface{}) (*Object, error) {
	ret := &Object{
		inner: obj_global_null,
	}
	switch v := value.(type) {
	case string:
		ret.inner = &objectString{value: v}
	case bool:
		ret.inner = &objectString{value: fmt.Sprintf("%t", v)}
	case float32:
		ret.inner = &objectString{value: fmt.Sprintf("%f", v)}
	case float64:
		ret.inner = &objectString{value: fmt.Sprintf("%f", v)}
	case int:
		ret.inner = &objectString{value: fmt.Sprintf("%d", v)}
	case int8:
		ret.inner = &objectString{value: fmt.Sprintf("%d", v)}
	case int16:
		ret.inner = &objectString{value: fmt.Sprintf("%d", v)}
	case int32:
		ret.inner = &objectString{value: fmt.Sprintf("%d", v)}
	case int64:
		ret.inner = &objectString{value: fmt.Sprintf("%d", v)}
	default:
		return ret, fmt.Errorf("unable to cast type as Float. unsupported type: %T", v)
	}
	return ret, nil
}

// casts a Go type to a morph Boolean Object so it can be used when defining custom functions
// input must be a bool
func CastBool(value interface{}) (*Object, error) {
	ret := &Object{
		inner: obj_global_null,
	}
	switch v := value.(type) {
	case bool:
		ret.inner = &objectBoolean{value: v}
	default:
		return ret, fmt.Errorf("unable to cast type as Boolean. unsupported type: %T", v)
	}
	return ret, nil
}

// casts a Go type to a morph Map Object so it can be used when defining custom functions
// input must be a map[string]interface{}, which is the default format of raw data maps being passed via morph statements and expressions
func CastMap(value interface{}) (*Object, error) {
	ret := &Object{
		inner: obj_global_null,
	}
	switch v := value.(type) {
	case map[string]interface{}:
		m, err := convertMapToObject(v, false)
		if err != nil {
			return ret, err
		}
		ret.inner = m
	default:
		return ret, fmt.Errorf("unable to cast type as Boolean. unsupported type: %T", v)
	}
	return ret, nil
}

// casts a Go type to a morph Map Object so it can be used when defining custom functions
// input must be a []interface{}, which is the default format of raw data arrays being passed via morph statements and expressions
func CastArray(value interface{}) (*Object, error) {
	ret := &Object{
		inner: obj_global_null,
	}
	switch v := value.(type) {
	case []interface{}:
		a, err := convertArrayToObject(v, false)
		if err != nil {
			return ret, err
		}
		ret.inner = a
	default:
		return ret, fmt.Errorf("unable to cast type as Array. unsupported type: %T", v)
	}
	return ret, nil
}
