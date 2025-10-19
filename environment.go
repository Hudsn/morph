package morph

import "context"

type environment struct {
	ctx       context.Context
	store     map[string]object
	functions *functionStore
}

type newEnvArg func(*environment)

func EnvWithContext(ctx context.Context) newEnvArg {
	return func(e *environment) {
		e.ctx = ctx
	}
}

func newEnvironment(fstore *functionStore, args ...newEnvArg) *environment {
	ret := &environment{functions: fstore, store: make(map[string]object)}
	for _, fn := range args {
		fn(ret)
	}
	return ret
}

func (e *environment) get(name string) (object, bool) {
	ret, ok := e.store[name]
	return ret, ok
}

func (e environment) set(name string, val object) object {
	e.store[name] = val
	return val
}
