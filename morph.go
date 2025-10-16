package morph

import (
	"encoding/json"
	"errors"
)

type morph struct {
	program       *program
	functionStore *functionStore
}

type Opt func(*morph)

func WithFunctionStore(fstore *functionStore) func(*morph) {
	return func(m *morph) {
		m.functionStore = fstore
	}
}

func New(input string, opts ...Opt) (*morph, error) {
	inputRunes := []rune(input)
	l := newLexer(inputRunes)
	p := newParser(l)
	program, err := p.parseProgram()
	if err != nil {
		return nil, err
	}

	m := &morph{
		program:       program,
		functionStore: NewDefaultFunctionStore(),
	}

	for _, fn := range opts {
		fn(m)
	}

	return m, nil
}

func (m *morph) ToAny(inputData []byte) (interface{}, error) {
	env := newEnvironment(m.functionStore)
	env.set("@in", convertBytesToObject(inputData))
	res := m.program.eval(env)
	if isObjectErr(res) {
		return nil, errors.New(res.inspect())
	}
	res, ok := env.get("@out")
	if !ok {
		return nil, nil
	}
	return convertObjectToNative(res)
}

func (m *morph) ToJSON(inputData []byte) ([]byte, error) {
	out, err := m.ToAny(inputData)
	if err != nil {
		return nil, err
	}
	return json.Marshal(out)
}
