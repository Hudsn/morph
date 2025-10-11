package morph

type FunctionStore struct {
}

type InstanceStore struct {
}

// function entries contain documentation information AND runnable instances of functions
type FunctionEntry struct {
	Name      string
	Namespace string
}

func (fe *FunctionEntry) instance() *functionInstance {
	// todo
	return nil
}

// function instance is a runable instance of a function
type functionInstance struct {
	name string
	args []functionInstanceArg
}

type FunctionArg struct {
	Name string
	Type PublicType
}

type functionInstanceArg struct {
}
