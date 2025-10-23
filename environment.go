package morph

import (
	"context"
	"fmt"
)

type environment struct {
	ctx           context.Context
	store         map[string]object
	functionStore *FunctionStore
}

func newEnvironment(fstore *FunctionStore, opts ...newEnvArg) *environment {
	e := &environment{functionStore: fstore, store: make(map[string]object)}
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

func (e *environment) get(name string) (object, bool) {
	ret, ok := e.store[name]
	return ret, ok
}

func (e *environment) set(name string, val object) object {
	e.store[name] = val
	return val
}

func (e *environment) getFunction(name string) (*FunctionEntry, error) {
	return e.functionStore.get("std", name)
}
func (e *environment) getFunctionByNamespace(namespace string, name string) (*FunctionEntry, error) {
	r, err := e.functionStore.get(namespace, name)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	return r, nil
}
