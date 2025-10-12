package morph

import "fmt"

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
	Name        string
	Description string
	Fn          Function
	Args        []FunctionArg
	Return      FunctionReturn
	Attributes  []FunctionAttribute
	Tags        []FunctionTag
	Examples    []ProgramExample
}

func NewFunctionEntry(name string, description string, fn Function, opts ...functionEntryOpt) *FunctionEntry {
	return &FunctionEntry{
		Name:        name,
		Description: description,
		Fn:          fn,
		Args:        []FunctionArg{},
		Return:      FunctionReturn{},
		Attributes:  []FunctionAttribute{},
		Tags:        []FunctionTag{"General"},
		Examples:    []ProgramExample{},
	}
}

type functionEntryOpt func(*FunctionEntry)

func WithArgs(args ...FunctionArg) functionEntryOpt {
	return func(fe *FunctionEntry) {
		fe.Args = args
	}
}
func WithReturn(ret FunctionReturn) functionEntryOpt {
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

type FunctionReturn struct {
	Description string
	Types       []PublicType
}

func NewFunctionReturn(description string, types ...PublicType) FunctionReturn {
	return FunctionReturn{
		Description: description,
		Types:       types,
	}
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
