package morph

import (
	"fmt"
	"strings"
)

type objectType string

type object interface {
	getType() objectType
	inspect() string
}

const (
	T_INTEGER objectType = "INTEGER"
	T_FLOAT   objectType = "FLOAT"
	T_BOOLEAN objectType = "BOOLEAN"

	T_MAP objectType = "MAP"

	T_ERROR objectType = "ERROR"
	T_TERM  objectType = "TERMINATE"
	T_NULL  objectType = "NULL"
)

//
//object impls

type objectMapPair struct {
	key   string
	value object
}

type objectMap struct {
	kvPairs map[string]objectMapPair
}

func (m *objectMap) getType() objectType { return T_MAP }
func (m *objectMap) inspect() string {
	pairs := []string{}
	for _, pair := range m.kvPairs {
		pairString := fmt.Sprintf("%s: %s", pair.key, pair.value.inspect())
		pairs = append(pairs, pairString)
	}
	return fmt.Sprintf("{%s}", strings.Join(pairs, ", "))
}

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
type objectTerm struct{}

func (t *objectTerm) getType() objectType { return T_TERM }
func (t *objectTerm) inspect() string     { return "TERMINATE" }
