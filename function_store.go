package morph

import (
	"context"
	"fmt"
	"slices"
	"strings"
)

type FunctionStore struct {
	namespaces map[string]*functionNamespace
}

type functionNamespace struct {
	name      string
	functions map[string]*FunctionEntry
}

func newFunctionNamespace(name string) *functionNamespace {
	return &functionNamespace{
		name:      name,
		functions: make(map[string]*FunctionEntry),
	}
}

func NewFunctionStore() *FunctionStore {
	s := &FunctionStore{
		namespaces: make(map[string]*functionNamespace),
	}
	s.namespaces["std"] = newFunctionNamespace("std")
	return s
}

func (fs *FunctionStore) get(namespace string, name string) (*FunctionEntry, error) {
	if len(namespace) == 0 {
		namespace = "std"
	}
	ns, ok := fs.namespaces[namespace]
	if !ok {
		return nil, fmt.Errorf("function namespace %q does not exist", namespace)
	}
	fn, ok := ns.functions[name]
	if !ok {
		return nil, fmt.Errorf("function %q in namespace %q does not exist", name, namespace)
	}
	return fn, nil
}
func (fs *FunctionStore) Register(fe *FunctionEntry) {
	fs.register("std", fe)
}
func (fs *FunctionStore) RegisterToNamespace(namespace string, fe *FunctionEntry) {
	fs.register(namespace, fe)
}
func (fs *FunctionStore) register(namespace string, fe *FunctionEntry) {
	if len(namespace) == 0 {
		namespace = "std"
	}
	fe.namespace = namespace
	if ns, ok := fs.namespaces[namespace]; ok {
		ns.functions[fe.Name] = fe
		return
	}
	ns := newFunctionNamespace(namespace)
	ns.functions[fe.Name] = fe
	fs.namespaces[namespace] = ns
}

// function entries contain documentation information AND runnable instances of functions
type FunctionEntry struct {
	namespace   string // populated when the function is registered.
	Name        string
	Description string
	Fn          Function
	Args        []FunctionArg
	Return      *FunctionReturn
	Attributes  []FunctionAttribute
	Tags        []FunctionTag
	Examples    []ProgramExample
}

func NewFunctionEntry(name string, description string, fn Function, opts ...functionEntryOpt) *FunctionEntry {
	ret := &FunctionEntry{
		Name:        name,
		namespace:   "",
		Description: description,
		Fn:          fn,
		Args:        []FunctionArg{},
		Return:      nil,
		Attributes:  []FunctionAttribute{},
		Tags:        []FunctionTag{"General"},
		Examples:    []ProgramExample{},
	}
	for _, fn := range opts {
		fn(ret)
	}
	return ret
}

// retruns functions namespace.name as a string
func (fe *FunctionEntry) fullName() string {
	if len(fe.namespace) == 0 {
		return fe.Name
	}
	return fmt.Sprintf("%s.%s", fe.namespace, fe.Name)
}

// returns the function signature string
func (fe *FunctionEntry) Signature() string {
	argStrList := []string{}
	for _, a := range fe.Args {
		argStrList = append(argStrList, a.typesString())
	}
	args := strings.Join(argStrList, ", ")
	ret := ""
	if fe.Return != nil {
		ret = fe.Return.typesString()
	}
	return fmt.Sprintf("%s(%s) %s", fe.fullName(), args, ret)
}

func (fe *FunctionEntry) run(ctx context.Context, args ...object) object {
	if len(args) < len(fe.Args) {
		return newObjectErrWithoutLC("function %q too few arguments supplied. want=%d got=%d\n\tfunction signature: %s", fe.fullName(), len(fe.Args), len(args), fe.Signature())
	}

	for argIdx, wantArg := range fe.Args {
		if len(wantArg.Types) == 0 {
			continue
		}
		arg := args[argIdx]
		if !slices.Contains(wantArg.Types, PublicType(arg.getType())) {
			return newObjectErrWithoutLC("function %q invalid argument type for %q. want=%s. got=%s\n\tfunction signature: %s", fe.fullName(), wantArg.Name, wantArg.typesString(), arg.getType(), fe.Signature())
		}
	}
	if err := fe.checkVariadic(args...); err != nil {
		return newObjectErrWithoutLC(err.Error())
	}
	ret := evalFunction(ctx, fe.Fn, args...)
	if isObjectErr(ret) {
		return ret
	}
	if fe.Return != nil {
		if !slices.Contains(fe.Return.Types, PublicType(ret.getType())) {
			return newObjectErrWithoutLC("function %q invalid return type. want=%s got=%s\n\tfunction signature: %s", fe.Name, fe.Return.typesString(), ret.getType(), fe.Signature())
		}
	}
	return ret
}

func evalFunction(ctx context.Context, fn Function, args ...object) object {
	objList := []*Object{}
	for _, arg := range args {
		objList = append(objList, &Object{inner: arg})
	}
	obj := fn(ctx, objList...)
	return obj.inner
}

func (fe *FunctionEntry) checkVariadic(args ...object) error {
	if len(args) == 0 {
		return nil
	}
	isVariadic := slices.Contains(fe.Attributes, FUNCTION_ATTRIBUTE_VARIADIC)
	if len(args) > len(fe.Args) && !isVariadic {
		return fmt.Errorf("invalid number of args for function %q: too many arguments supplied. want=%d got=%d", fe.fullName(), len(fe.Args), len(args))
	}
	if !isVariadic {
		return nil
	}
	firstVariadicArg := fe.Args[len(fe.Args)-1]
	curIdx := len(fe.Args) - 1
	lastArgs := args[curIdx:]
	for _, arg := range lastArgs {
		if !slices.Contains(firstVariadicArg.Types, PublicType(arg.getType())) {
			return fmt.Errorf("type error for function %q: argument at zero-indexed position %d does not match any type of variadic parameter %q (%s)", fe.fullName(), curIdx, firstVariadicArg.Name, firstVariadicArg.typesString())
		}
		curIdx++
	}
	return nil
}

type functionEntryOpt func(*FunctionEntry)

func WithArgs(args ...FunctionArg) functionEntryOpt {
	return func(fe *FunctionEntry) {
		fe.Args = args
	}
}
func WithReturn(ret *FunctionReturn) functionEntryOpt {
	return func(fe *FunctionEntry) {
		fe.Return = ret
	}
}
func WithAtributes(attrs ...FunctionAttribute) functionEntryOpt {
	return func(fe *FunctionEntry) {
		fe.Attributes = attrs
	}
}
func WithTags(tags ...FunctionTag) functionEntryOpt {
	return func(fe *FunctionEntry) {
		if len(tags) == 1 {
			fe.Tags = tags
			return
		}
		// if multiple tags apply, also add the General tag
		fe.Tags = append([]FunctionTag{"General"}, tags...)
	}
}
func WithExamples(examples ...ProgramExample) functionEntryOpt {
	return func(fe *FunctionEntry) {

	}
}

type FunctionArg struct {
	Name        string
	Description string
	Types       []PublicType
}

func NewFunctionArg(name string, description string, types ...PublicType) FunctionArg {
	return FunctionArg{
		Name:        name,
		Description: description,
		Types:       types,
	}
}

func (fa FunctionArg) typesString() string {
	isAny := true
	for _, t := range ANY {
		if !slices.Contains(fa.Types, t) {
			isAny = false
			break
		}
	}
	if isAny {
		return "ANY"
	}
	isBasic := true
	for _, t := range BASIC {
		if !slices.Contains(fa.Types, t) {
			isBasic = false
			break
		}
	}
	if isBasic {
		return "BASIC"
	}
	strs := []string{}
	for _, t := range fa.Types {
		strs = append(strs, string(t))
	}
	return fmt.Sprintf("%s:%s", fa.Name, strings.Join(strs, "|"))
}

type FunctionReturn struct {
	Description string
	Types       []PublicType
}

func NewFunctionReturn(description string, types ...PublicType) *FunctionReturn {
	return &FunctionReturn{
		Description: description,
		Types:       types,
	}
}

func (fr *FunctionReturn) typesString() string {
	isAny := true
	for _, t := range ANY {
		if !slices.Contains(fr.Types, t) {
			isAny = false
			break
		}
	}
	if isAny {
		return "ANY"
	}
	isBasic := true
	for _, t := range BASIC {
		if !slices.Contains(fr.Types, t) {
			isBasic = false
			break
		}
	}
	if isBasic {
		return "BASIC"
	}
	strs := []string{}
	for _, t := range fr.Types {
		strs = append(strs, string(t))
	}
	return strings.Join(strs, "|")
}

type FunctionAttribute string

const (
	FUNCTION_ATTRIBUTE_VARIADIC FunctionAttribute = "VARIADIC"
)

type FunctionTag string

const (
	FUNCTION_TAG_GENERAL         FunctionTag = "General"
	FUNCTION_TAG_TYPE_COERCION   FunctionTag = "Type Coercion"
	FUNCTION_TAG_ERR_NULL_CHECKS FunctionTag = "Error and Null Checks"
	FUNCTION_TAG_FLOW_CONTROL    FunctionTag = "Flow Control"
)

type ProgramExample struct {
	In      string
	Program string
	Out     string
}

func NewProgramExample(in string, program string, out string) ProgramExample {
	return ProgramExample{
		In:      in,
		Program: program,
		Out:     out,
	}
}
