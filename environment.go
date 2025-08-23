package morph

type environment struct {
	outer *environment
	store map[string]object
}

func newEnvironment() *environment {
	return &environment{outer: nil, store: make(map[string]object)}
}

func (e *environment) get(name string) (object, bool) {
	ret, ok := e.store[name]
	if !ok && e.outer != nil {
		ret, ok = e.outer.get(name)
	}
	return ret, ok
}

func (e environment) set(name string, val object) object {
	e.store[name] = val
	return val
}
