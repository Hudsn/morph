package morph

import "fmt"

type objectType string

type object interface {
	getType() objectType
	inspect() string
}

const (
	T_INTEGER objectType = "INTEGER"
	T_FLOAT   objectType = "FLOAT"
	T_BOOLEAN objectType = "BOOLEAN"

	T_ERROR   objectType = "ERROR"
	T_SIGTERM objectType = "SIGTERM"
	T_NULL    objectType = "NULL"
)

//

type objectInteger struct {
	value int64
}

func (i *objectInteger) getType() objectType { return T_INTEGER }
func (i *objectInteger) inspect() string     { return fmt.Sprintf("%d", i.value) }

//

type objectFloat struct {
	value float64
}

func (f *objectFloat) getType() objectType { return T_FLOAT }
func (f *objectFloat) inspect() string     { return fmt.Sprintf("%f", f.value) }

//

type objectBoolean struct {
	value bool
}

func (b *objectBoolean) getType() objectType { return T_BOOLEAN }
func (b *objectBoolean) inspect() string     { return fmt.Sprintf("%t", b.value) }

//

type objectNull struct{}

func (n *objectNull) getType() objectType { return T_NULL }
func (n *objectNull) inspect() string     { return "null" }

//

type objectError struct {
	message string
}

func (e *objectError) getType() objectType { return T_ERROR }
func (e *objectError) inspect() string     { return "ERROR: " + e.message }

//

// returned by builtin funcs when signaling to halt further processing and return the env values as-is.
type objectSigterm struct{}

func (s *objectSigterm) getType() objectType { return T_SIGTERM }
func (s *objectSigterm) inspect() string     { return "SIGTERM" }
