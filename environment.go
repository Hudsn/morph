package morph

type environment struct {
	store     map[string]object
	functions *functionStore
}

func newEnvironment(fstore *functionStore) *environment {
	return &environment{functions: fstore, store: make(map[string]object)}
}

func (e *environment) get(name string) (object, bool) {
	ret, ok := e.store[name]
	return ret, ok
}

func (e environment) set(name string, val object) object {
	e.store[name] = val
	return val
}
