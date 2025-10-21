package morph

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"
)

type Function func(ctx context.Context, args ...*Object) *Object
type FunctionOld func(args ...*Object) *Object

// public wrapper of object to be used for implementing custom functions
type Object struct {
	inner object
}

func (o *Object) Type() string {
	return string(o.inner.getType())
}

type PublicType string

// wrappers for public types
const (
	INTEGER   PublicType = PublicType(t_integer)
	FLOAT     PublicType = PublicType(t_float)
	BOOLEAN   PublicType = PublicType(t_boolean)
	STRING    PublicType = PublicType(t_string)
	MAP       PublicType = PublicType(t_map)
	ARRAY     PublicType = PublicType(t_array)
	ARROWFUNC PublicType = PublicType(t_arrow)
	TIME      PublicType = PublicType(t_time)
	NULL      PublicType = PublicType(t_null)
	ERROR     PublicType = PublicType(t_error)
)

var BASIC = []PublicType{INTEGER, FLOAT, BOOLEAN, STRING, MAP, ARRAY, TIME, ERROR, NULL}
var ANY = []PublicType{INTEGER, FLOAT, BOOLEAN, STRING, MAP, ARRAY, TIME, ERROR, NULL, ARROWFUNC}

func (o *Object) AsAny() (interface{}, error) {
	switch o.Type() {
	case string(NULL):
		return nil, nil
	case string(INTEGER):
		return o.AsInt()
	case string(FLOAT):
		return o.AsFloat()
	case string(MAP):
		return o.AsMap()
	case string(ARRAY):
		return o.AsArray()
	case string(ARROWFUNC):
		return o.AsArrowFunction()
	case string(STRING):
		return o.AsString()
	case string(TIME):
		return o.AsTime()
	case string(ERROR):
		return o.AsError()
	default:
		return nil, fmt.Errorf("unable to convert Object: not a convertible type. got=%s", o.Type())
	}
}

func (o *Object) AsError() (error, error) {
	e, ok := o.inner.(*objectError)
	if !ok {
		return nil, fmt.Errorf("unable to convert object to error: underlying structure is not an integer type. got=%s", o.inner.getType())
	}
	return fmt.Errorf(e.message), nil
}

func (o *Object) AsTime() (time.Time, error) {
	t, ok := o.inner.(*objectTime)
	if !ok {
		return time.Time{}, fmt.Errorf("unable to convert object to Time: underlying structure is not an integer type. got=%s", o.inner.getType())
	}
	return t.value, nil
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

type ObjectArrowFN struct {
	inner  *objectArrowFunction
	errObj *Object
}

func (af *ObjectArrowFN) HasError() bool {
	return af.errObj != nil
}

func (af *ObjectArrowFN) GetError() *Object {
	return af.errObj
}

func (af *ObjectArrowFN) Run(input interface{}) interface{} {
	env := newEnvironment(af.inner.functions)
	startingObj := convertAnyToObject(input, false)
	if isObjectErr(startingObj) {
		af.errObj = &Object{inner: startingObj}
		return nil
	}
	env.set(af.inner.paramName, startingObj)
	for _, stmt := range af.inner.statements {
		obj := stmt.eval(env)
		if isObjectErr(obj) {
			af.errObj = &Object{inner: obj}
			return nil
		}
		if obj.getType() == t_terminate {
			term := obj.(*objectTerminate)
			if term.shouldReturnNull {
				env.store = map[string]object{}
			}
			break
		}
	}
	ret, err := convertMapStringObjectToNative(env.store)
	if err != nil {
		af.errObj = ObjectError(err.Error())
		return nil
	}
	return ret
}

func (o *Object) AsArrowFunction() (*ObjectArrowFN, error) {
	arrow, ok := o.inner.(*objectArrowFunction)
	if !ok {
		return nil, fmt.Errorf("unable to convert object to ArrowFunction: underlying structure is not an Arrow Function type. got=%s", o.inner.getType())
	}
	return &ObjectArrowFN{
		inner: arrow,
	}, nil
}

var ObjectNull = &Object{inner: obj_global_null}
var ObjectTerminate = &Object{inner: obj_global_term}
var ObjectTerminateDrop = &Object{inner: obj_global_term_drop}

func ObjectError(msg string, args ...interface{}) *Object {
	return &Object{
		inner: newObjectErrWithoutLC(msg, args...),
	}
}

//
// typecast helpers

// casts a Go time struct to a morph Time Object so it can be used when defining custom functions
// input must be one of: int, int8, int16, int32, int64, float32, float64
func CastTime(value interface{}) *Object {
	var err error
	var t time.Time
	switch v := value.(type) {
	case time.Time:
		t = v
	case *time.Time:
		t = *v
	case int, int32, int64, float32, float64:
		t, err = castTimeNumCase(v)
		if err != nil {
			return ObjectError(err.Error())
		}
	case string:
		t, err = castTimeStringCase(v)
		if err != nil {
			return ObjectError(err.Error())
		}
	default:
		return ObjectError("unable to cast underlying type as TIME. unsupported input type")
	}
	return &Object{inner: &objectTime{value: t}}
}

func castTimeStringCase(s string) (time.Time, error) {
	if t, err := time.Parse(time.RFC3339Nano, s); err == nil {
		return t, nil
	}
	if t, err := time.Parse(time.RFC3339, s); err == nil {
		return t, nil
	}
	if asInt, err := strconv.ParseInt(s, 10, 64); err == nil {
		return castTimeNumCase(asInt)
	}
	if asFloat, err := strconv.ParseFloat(s, 64); err == nil {
		return castTimeNumCase(asFloat)
	}
	return time.Time{}, fmt.Errorf("unable to cast STRING as TIME. Invalid STRING %s", s)
}

func castTimeNumCase(i interface{}) (time.Time, error) {
	var retTime time.Time
	var err error
	switch v := i.(type) {
	case float32:
		retTime, err = castTimeFloatToUnixSecN(float64(v))
		if err != nil {
			return time.Time{}, err
		}
	case float64:
		retTime, err = castTimeFloatToUnixSecN(v)
		if err != nil {
			return time.Time{}, err
		}
	case int:
		retTime = time.Unix(int64(v), 0)
	case int32:
		retTime = time.Unix(int64(v), 0)
	case int64:
		retTime = time.Unix(int64(v), 0)
	}
	return retTime.UTC(), nil
}

func castTimeFloatToUnixSecN(f float64) (time.Time, error) {
	fstring := strconv.FormatFloat(f, 'f', 9, 64)
	spl := strings.Split(fstring, ".")
	if len(spl) != 2 {
		return time.Time{}, fmt.Errorf("unable to cast FLOAT as TIME. invalid FLOAT %f", f)
	}
	sec, err := strconv.ParseInt(spl[0], 10, 64)
	if err != nil {
		return time.Time{}, fmt.Errorf("unable to cast FLOAT as TIME. invalid FLOAT %f", f)
	}
	nsec, err := strconv.ParseInt(spl[1], 10, 64)
	if err != nil {
		return time.Time{}, fmt.Errorf("unable to cast FLOAT as TIME. invalid FLOAT %f", f)
	}
	return time.Unix(sec, nsec), nil
}

// casts a Go number to a morph Integer Object so it can be used when defining custom functions
// input must be one of: int, int8, int16, int32, int64, float32, float64
func CastInt(value interface{}) *Object {
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
	case string:
		i, err := strconv.Atoi(v)
		if err != nil {
			return ObjectError("unable to cast string as INTEGER. invalid string")
		}
		ret.inner = &objectInteger{value: int64(i)}
	default:
		return ObjectError("unable to cast underlying type as INTEGER. unsupported input type")
	}
	return ret
}

// casts a Go number to a morph Float Object so it can be used when defining custom functions
// input must be one of: int, int8, int16, int32, int64, float32, float64
func CastFloat(value interface{}) *Object {
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
	case string:
		f, err := strconv.ParseFloat(v, 64)
		if err != nil {
			return ObjectError("unable to cast string as FLOAT. invalid string: %s", v)
		}
		ret.inner = &objectFloat{value: f}
	default:
		return ObjectError("unable to cast type as FLOAT. unsupported type")
	}
	return ret
}

// casts a Go type to a morph String Object so it can be used when defining custom functions
// input must be one of: int, int8, int16, int32, int64, float32, float64, string, bool, time
func CastString(value interface{}) *Object {
	ret := &Object{
		inner: obj_global_null,
	}
	switch v := value.(type) {
	case string:
		ret.inner = &objectString{value: v}
	case bool:
		ret.inner = &objectString{value: fmt.Sprintf("%t", v)}
	case float32:
		ret.inner = &objectString{value: strconv.FormatFloat(float64(v), 'g', -1, 32)}
	case float64:
		ret.inner = &objectString{value: strconv.FormatFloat(float64(v), 'g', -1, 64)}
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
	case time.Time:
		ret.inner = &objectString{value: v.Format(time.RFC3339Nano)}
	default:
		return ObjectError("unable to cast type as STRING. unsupported type")
	}
	return ret
}

// casts a Go type to a morph Boolean Object so it can be used when defining custom functions
// input must be a bool
func CastBool(value interface{}) *Object {
	ret := &Object{
		inner: obj_global_null,
	}
	switch v := value.(type) {
	case bool:
		ret.inner = objectFromBoolean(v)
	case *bool:
		ret.inner = objectFromBoolean(*v)
	default:
		return ObjectError("unable to cast type as BOOLEAN. unsupported type")
	}
	return ret
}

// casts a Go type to a morph Map Object so it can be used when defining custom functions
// input must be a map[string]interface{}, which is the default format of raw data maps being passed via morph statements and expressions
func CastMap(value interface{}) *Object {
	switch v := value.(type) {
	case map[string]interface{}:
		m := convertMapToObject(v, false)
		return &Object{inner: m}
	default:
		return ObjectError("unable to cast type as MAP. unsupported type")
	}
}

// casts a Go type to a morph Array Object so it can be used when defining custom functions
// input must be a []interface{}, which is the default format of raw data arrays being passed via morph statements and expressions
func CastArray(value interface{}) *Object {
	ret := &Object{
		inner: obj_global_null,
	}
	switch v := value.(type) {
	case []interface{}:
		a := convertArrayToObject(v, false)
		ret.inner = a
	default:
		return ObjectError("unable to cast type as Array. unsupported type")
	}
	return ret
}

func CastError(value interface{}) *Object {
	if e, ok := value.(error); ok {
		return &Object{
			inner: &objectError{message: e.Error()},
		}
	}
	return ObjectError("unable to cast type as error. unsupported type")
}

func CastAuto(value interface{}) *Object {
	if value == nil {
		return ObjectNull
	}
	switch v := value.(type) {

	case int, int8, int16, int32, int64:
		return CastInt(v)
	case float32, float64:
		return CastFloat(v)
	case bool:
		return CastBool(v)
	case string:
		return CastString(v)
	case map[string]interface{}:
		return CastMap(v)
	case []interface{}:
		return CastArray(v)
	case time.Time:
		return CastTime(v)
	case error:
		return CastError(v)
	default:
		return ObjectError("unable to automatically cast type for value")
	}
}

// helpers

// checks minimum function argument count. useful for variadic functions
func IsArgCountAtLeast(count int, args []*Object) (*Object, bool) {
	if len(args) < count {
		return ObjectError("function requires at least %d args. got=%d", count, len(args)), false
	}
	return ObjectNull, true
}

// checks function argument count
func IsArgCountEqual(count int, args []*Object) (*Object, bool) {
	if len(args) != count {
		return ObjectError("function requires %d args. got=%d", count, len(args)), false
	}
	return ObjectNull, true
}
