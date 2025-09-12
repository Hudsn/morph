package morph

import (
	"encoding/json"
	"fmt"
	"reflect"
	"slices"
	"strings"
)

type functionStore struct {
	std        *functionNamespace
	namespaces map[string]*functionNamespace
}

func (s *functionStore) Register(fn *functionEntry) {
	s.std.register(fn)
}
func (s *functionStore) RegisterToNamespace(namespace string, fn *functionEntry) {
	if _, ok := s.namespaces[namespace]; !ok {
		s.namespaces[namespace] = newFunctionNamespace(namespace)
	}
	ns := s.namespaces[namespace]
	ns.register(fn)
}
func (s *functionStore) get(name string) (*functionEntry, error) {
	return s.std.get(name)
}
func (s *functionStore) getNamespace(namespace string, name string) (*functionEntry, error) {
	if ns, ok := s.namespaces[namespace]; ok {
		return ns.get(name)
	}
	return nil, fmt.Errorf("namespace %q does not found", namespace)
}

type functionNamespace struct {
	name  string
	store map[string]*functionEntry
}

func newFunctionNamespace(name string) *functionNamespace {
	return &functionNamespace{
		name:  name,
		store: make(map[string]*functionEntry),
	}
}

func (n *functionNamespace) get(name string) (*functionEntry, error) {
	if ret, ok := n.store[name]; ok {
		return ret, nil
	}
	msg := fmt.Sprintf("function %q does not exist", name)
	if n.name == "std" {
		msg = fmt.Sprintf("%s in namespace %q", msg, n.name)
	}
	return nil, fmt.Errorf("%s", msg)
}
func (n *functionNamespace) register(fe *functionEntry) {
	n.store[fe.name] = fe
}

func newFunctionStore() *functionStore {
	return &functionStore{
		std: &functionNamespace{
			name:  "std",
			store: make(map[string]*functionEntry),
		},
		namespaces: make(map[string]*functionNamespace),
	}
}

type functionEntry struct {
	name        string
	description string
	ret         *functionIO
	args        []functionIO
	function    Function
	attributes  []functionAttribute
}

func NewFunctionEntry(name string, function Function) *functionEntry {
	return &functionEntry{
		name:       name,
		function:   function,
		args:       []functionIO{},
		attributes: []functionAttribute{},
	}
}
func (fe *functionEntry) SetDescription(desc string) *functionEntry {
	fe.description = desc
	return fe
}
func (fe *functionEntry) SetArgument(name string, description string, types ...publicObject) *functionEntry {
	toAdd := functionIO{
		name:        name,
		description: description,
		types:       types,
	}
	fe.args = append(fe.args, toAdd)
	return fe
}
func (fe *functionEntry) SetReturn(name string, description string, types ...publicObject) *functionEntry {
	fe.ret = &functionIO{
		name:        name,
		description: description,
		types:       types,
	}
	return fe
}
func (fe *functionEntry) SetAttributes(attrs ...functionAttribute) *functionEntry {
	fe.attributes = attrs
	return fe
}
func (fe *functionEntry) string() string {
	argStrList := []string{}
	for _, a := range fe.args {
		argStrList = append(argStrList, a.formatString())
	}
	args := strings.Join(argStrList, ", ")
	ret := ""
	if fe.ret != nil {
		ret = fe.ret.formatString()
	}
	return fmt.Sprintf("%s(%s) %s", fe.name, args, ret)
}

func (fio *functionIO) formatString() string {
	typeString := fio.typesString()
	return fmt.Sprintf("%s:%s", fio.name, typeString)
}

func (fio *functionIO) typesString() string {
	strs := []string{}
	for _, t := range fio.types {
		strs = append(strs, string(t))
	}
	return strings.Join(strs, "|")
}

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

func (fe *functionEntry) eval(args ...object) (object, error) {
	if len(args) < len(fe.args) {
		return obj_global_null, fmt.Errorf("invalid number of args for function %q: too few arguments supplied. want=%d got=%d", fe.name, len(fe.args), len(args))
	}
	for argIdx, wantArg := range fe.args {
		arg := args[argIdx]
		if !slices.Contains(wantArg.types, publicObject(arg.getType())) {
			return obj_global_null, fmt.Errorf("function %q invalid argument type for %q. want=%s. got=%s", fe.name, wantArg.name, wantArg.typesString(), arg.getType())
		}
	}
	if err := fe.checkVariadic(args...); err != nil {
		return obj_global_null, err
	}
	return evalFunction(fe.function, args...)
}

func (fe *functionEntry) checkVariadic(args ...object) error {
	if len(args) == 0 {
		return nil
	}
	isVariadic := slices.Contains(fe.attributes, FUNCTION_ATTRIBUTE_VARIADIC)
	if len(args) > len(fe.args) && !isVariadic {
		return fmt.Errorf("invalid number of args for function %q: too many arguments supplied. want=%d got=%d", fe.name, len(fe.args), len(args))
	}
	if !isVariadic {
		return nil
	}
	lastWantArg := fe.args[len(fe.args)-1]
	curIdx := len(fe.args) - 1
	lastArgs := args[curIdx:]
	for _, arg := range lastArgs {
		if !slices.Contains(lastWantArg.types, publicObject(arg.getType())) {
			return fmt.Errorf("type error for function %q: argument at zero-indexed position %d does not match any type of variadic parameter %q (%s)", fe.name, curIdx, lastWantArg.name, lastWantArg.typesString())
		}
		curIdx++
	}
	return nil
}

type functionAttribute string

const (
	FUNCTION_ATTRIBUTE_VARIADIC = "VARIADIC"
)

type functionIO struct {
	name        string
	description string
	types       []publicObject
}

type Function func(args ...*Object) (*Object, error)

// public wrapper of object to be used for implementing custom functions
type Object struct {
	inner object
}

func (o *Object) Type() string {
	return string(o.inner.getType())
}

type publicObject string

// wrappers for public types
const (
	INTEGER publicObject = publicObject(t_integer)
	FLOAT   publicObject = publicObject(t_float)
	BOOLEAN publicObject = publicObject(t_boolean)
	STRING  publicObject = publicObject(t_string)
	MAP     publicObject = publicObject(t_map)
	ARRAY   publicObject = publicObject(t_array)
)

// func EnforceFunctionArgCount(wantArgNum int, args []*Object) error {
// 	if wantArgNum != len(args) {
// 		return fmt.Errorf("incorrect number of arguments. expected=%d got=%d", wantArgNum, len(args))
// 	}
// 	return nil
// }

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
