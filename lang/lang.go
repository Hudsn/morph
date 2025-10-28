package lang

import (
	"encoding/json"
	"errors"
)

// public helpers for running a morph program.

type Program struct {
	inner         *program
	functionStore *FunctionStore
}

func NewProgram(programInput string, funcStore *FunctionStore) (*Program, error) {
	l := newLexer([]rune(programInput))
	p := newParser(l)
	program, err := p.parseProgram()
	if err != nil {
		return nil, err
	}
	return &Program{
		inner:         program,
		functionStore: funcStore,
	}, nil
}

func (p *Program) Run(inputData []byte) ([]byte, error) {
	inputObject := convertBytesToObject(inputData)
	if isObjectErr(inputObject) {
		return nil, objectToError(inputObject)
	}
	env := newEnvironment(p.functionStore)
	if p.functionStore != nil {
		env.functionStore = p.functionStore
	}
	env.set("@in", convertBytesToObject(inputData))
	res := p.inner.eval(env)
	if isObjectErr(res) {
		return nil, errors.New(res.inspect())
	}
	res, ok := env.get("@out")
	if !ok {
		return []byte("null"), nil
	}
	outputIface, err := convertObjectToNative(res)
	if err != nil {
		return nil, err
	}
	return json.Marshal(outputIface)

}
