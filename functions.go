package morph

import (
	"fmt"
	"slices"
	"strings"
)

// creates a new function store using default builtin functions
// this store can still be extended by registering custom functions
func NewDefaultFunctionStore() *functionStore {
	return newBuiltinFuncStore()
}

// creates a new function store without any builtin functions
// you must implment any builtin functions you wish to use
func NewEmptyFunctionStore() *functionStore {
	return newFunctionStore()
}

type functionStore struct {
	std        *functionNamespaceOld
	namespaces map[string]*functionNamespaceOld
}

func (s *functionStore) Register(fn *functionEntry) {
	s.std.register(fn)
}
func (s *functionStore) RegisterToNamespace(namespace string, fn *functionEntry) {
	namespace = strings.ToLower(namespace)
	namespace = strings.TrimSpace(namespace)
	if namespace == "std" {
		s.Register(fn)
		return
	}
	if _, ok := s.namespaces[namespace]; !ok {
		s.namespaces[namespace] = newFunctionNamespaceOld(namespace)
	}
	ns := s.namespaces[namespace]
	ns.register(fn)
}
func (s *functionStore) get(name string) (*functionEntry, error) {
	return s.std.get(name)
}
func (s *functionStore) getNamespace(namespace string, name string) (*functionEntry, error) {
	namespace = strings.ToLower(namespace)
	namespace = strings.TrimSpace(namespace)
	if namespace == "std" {
		return s.get(name)
	}
	if ns, ok := s.namespaces[namespace]; ok {
		return ns.get(name)
	}
	return nil, fmt.Errorf("namespace %q does not found", namespace)
}

type functionNamespaceOld struct {
	name  string
	store map[string]*functionEntry
}

func newFunctionNamespaceOld(name string) *functionNamespaceOld {
	return &functionNamespaceOld{
		name:  name,
		store: make(map[string]*functionEntry),
	}
}

func (n *functionNamespaceOld) get(name string) (*functionEntry, error) {
	if ret, ok := n.store[name]; ok {
		return ret, nil
	}
	msg := fmt.Sprintf("function %q does not exist", name)
	if n.name == "std" {
		msg = fmt.Sprintf("%s in namespace %q", msg, n.name)
	}
	return nil, fmt.Errorf("%s", msg)
}
func (n *functionNamespaceOld) register(fe *functionEntry) {
	n.store[fe.name] = fe
}

func newFunctionStore() *functionStore {
	return &functionStore{
		std: &functionNamespaceOld{
			name:  "std",
			store: make(map[string]*functionEntry),
		},
		namespaces: make(map[string]*functionNamespaceOld),
	}
}

type functionEntry struct {
	name       string
	ret        *functionIO
	args       []functionIO
	attributes []functionAttribute
	function   FunctionOld
}

func NewFunctionEntryOld(name string, function FunctionOld) *functionEntry {
	return &functionEntry{
		name:       name,
		function:   function,
		args:       []functionIO{},
		attributes: []functionAttribute{},
	}
}

func (fe *functionEntry) SetArgument(name string, types ...PublicType) *functionEntry {
	toAdd := functionIO{
		name:  name,
		types: types,
	}
	fe.args = append(fe.args, toAdd)
	return fe
}
func (fe *functionEntry) SetReturn(name string, types ...PublicType) *functionEntry {
	fe.ret = &functionIO{
		name:  name,
		types: types,
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
	isAny := true
	for _, t := range ANY {
		if !slices.Contains(fio.types, t) {
			isAny = false
			break
		}
	}
	if isAny {
		return "ANY"
	}
	strs := []string{}
	for _, t := range fio.types {
		strs = append(strs, string(t))
	}
	return strings.Join(strs, "|")
}

func evalFunctionOld(fn FunctionOld, args ...object) object {
	objList := []*Object{}
	for _, arg := range args {
		objList = append(objList, &Object{inner: arg})
	}
	obj := fn(objList...)
	return obj.inner
}

func (fe *functionEntry) eval(args ...object) object {
	if len(args) < len(fe.args) {
		return newObjectErrWithoutLC("function %q too few arguments supplied. want=%d got=%d\n\tfunction signature: %s", fe.name, len(fe.args), len(args), fe.string())
	}

	for argIdx, wantArg := range fe.args {
		if len(wantArg.types) == 0 {
			continue
		}
		arg := args[argIdx]
		if !slices.Contains(wantArg.types, PublicType(arg.getType())) {
			return newObjectErrWithoutLC("function %q invalid argument type for %q. want=%s. got=%s\n\tfunction signature: %s", fe.name, wantArg.name, wantArg.typesString(), arg.getType(), fe.string())
		}
	}
	if err := fe.checkVariadic(args...); err != nil {
		return newObjectErrWithoutLC(err.Error())
	}
	ret := evalFunctionOld(fe.function, args...)
	if isObjectErr(ret) {
		return ret
	}
	if fe.ret != nil {
		if !slices.Contains(fe.ret.types, PublicType(ret.getType())) {
			return newObjectErr("function %q invalid return type. want=%s got=%s\n\tfunction signature: %s", fe.name, fe.ret.typesString(), ret.getType(), fe.string())
		}
	}
	return ret
}

func (fe *functionEntry) checkVariadic(args ...object) error {
	if len(args) == 0 {
		return nil
	}
	isVariadic := slices.Contains(fe.attributes, FUNCTION_ATTRIBUTE_VARIADIC_OLD)
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
		if !slices.Contains(lastWantArg.types, PublicType(arg.getType())) {
			return fmt.Errorf("type error for function %q: argument at zero-indexed position %d does not match any type of variadic parameter %q (%s)", fe.name, curIdx, lastWantArg.name, lastWantArg.typesString())
		}
		curIdx++
	}
	return nil
}

type functionAttribute string

const (
	FUNCTION_ATTRIBUTE_VARIADIC_OLD = "VARIADIC"
)

type functionIO struct {
	name  string
	types []PublicType
}
