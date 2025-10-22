package morph

import (
	"context"
)

type environment struct {
	ctx           context.Context
	store         map[string]object
	functions     *functionStore
	functionStore *FunctionStore
}

func newEnvironment(fstore *functionStore, opts ...newEnvArg) *environment {
	e := &environment{functions: fstore, store: make(map[string]object)}
	for _, fn := range opts {
		fn(e)
	}
	return e
}

type newEnvArg func(*environment)

func EnvWithContext(ctx context.Context) newEnvArg {
	return func(e *environment) {
		e.ctx = ctx
	}
}

func EnvWithNewFNStore(fnStore *FunctionStore) newEnvArg {
	return func(e *environment) {
		e.functionStore = fnStore
	}
}

func (e *environment) get(name string) (object, bool) {
	ret, ok := e.store[name]
	return ret, ok
}

func (e *environment) set(name string, val object) object {
	e.store[name] = val
	return val
}

type runnableFN interface {
	run(ctx context.Context, args ...object) object
}

func (e *environment) getFunction(name string) (runnableFN, error) {
	if e.functionStore != nil {
		return e.functionStore.get("", name)
	}
	return e.functions.get(name)
}
func (e *environment) getFunctionByNamespace(namespace string, name string) (runnableFN, error) {
	if e.functionStore != nil {
		return e.functionStore.get(namespace, name)
	}
	return e.functions.getNamespace(namespace, name)
}
