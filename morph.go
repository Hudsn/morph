package morph

import (
	"encoding/json"
	"errors"
	"fmt"
)

type morph struct {
	program          *program
	functionStoreOld *functionStore
	functionStore    *FunctionStore
}

type Opt func(*morph)

func WithFunctionStoreOld(fstore *functionStore) func(*morph) {
	return func(m *morph) {
		m.functionStoreOld = fstore
	}
}
func WithFunctionStore(fstore *FunctionStore) func(*morph) {
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
		program:          program,
		functionStoreOld: NewDefaultFunctionStore(),
	}

	for _, fn := range opts {
		fn(m)
	}
	return m, nil
}

func (m *morph) ToAny(inputData []byte) (interface{}, error) {
	inputObject := convertBytesToObject(inputData)
	if isObjectErr(inputObject) {
		fmt.Println("SDKJSDFKJ")
		return nil, objectToError(inputObject)
	}
	env := newEnvironment(m.functionStoreOld)
	if m.functionStore != nil {
		env.functionStore = m.functionStore
	}
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
