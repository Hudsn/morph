package morph

import (
	"github.com/hudsn/morph/lang"
)

type morph struct {
	program       *lang.Program
	functionStore *lang.FunctionStore
}

type Opt func(*morph)

func WithFunctionStore(fstore *lang.FunctionStore) func(*morph) {
	return func(m *morph) {
		m.functionStore = fstore
	}
}

func New(input string, opts ...Opt) (*morph, error) {
	m := &morph{
		functionStore: lang.DefaultFunctionStore(),
	}

	for _, fn := range opts {
		fn(m)
	}

	program, err := lang.NewProgram(input, m.functionStore)
	if err != nil {
		return nil, err
	}
	m.program = program

	return m, nil
}

func (m *morph) Exec(inputData []byte) ([]byte, error) {
	return m.program.Run(inputData)
}
