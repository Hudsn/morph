package morph

import (
	"fmt"
	"strings"
	"time"
)

type objectType string

type object interface {
	getType() objectType
	inspect() string
	clone() object // used for deep copying so we don't allow mutates on env variables to retroactively affect other variables they're assigned to.  ex: in pseudocode "x = 1; y = x; x = 5;" y shold be equal to 1, *NOT* 5.
	isTruthy() bool
}

const (
	t_arrow objectType = "ARROW"

	t_integer objectType = "INTEGER"
	t_float   objectType = "FLOAT"
	t_boolean objectType = "BOOLEAN"
	t_string  objectType = "STRING"
	t_time    objectType = "TIME"

	t_map   objectType = "MAP"
	t_array objectType = "ARRAY"

	t_terminate objectType = "TERMINATE"
	t_null      objectType = "NULL"

	t_error objectType = "ERROR"
)

//
//object impls

type objectArray struct {
	entries []object
}

func (a *objectArray) getType() objectType { return t_array }
func (a *objectArray) inspect() string {
	stringList := []string{}
	for _, entry := range a.entries {
		stringList = append(stringList, entry.inspect())
	}

	return fmt.Sprintf("[%s]", strings.Join(stringList, ", "))
}
func (a *objectArray) clone() object {
	newArr := []object{}
	for _, entry := range a.entries {
		newArr = append(newArr, entry.clone())
	}
	return &objectArray{entries: newArr}
}
func (a *objectArray) isTruthy() bool {
	return len(a.entries) > 0
}

type objectMap struct {
	kvPairs map[string]object
}

func (m *objectMap) getType() objectType { return t_map }
func (m *objectMap) inspect() string {
	pairs := []string{}
	for key, obj := range m.kvPairs {
		pairString := fmt.Sprintf("%s: %s", key, obj.inspect())
		pairs = append(pairs, pairString)
	}
	return fmt.Sprintf("{%s}", strings.Join(pairs, ", "))
}
func (m *objectMap) clone() object {
	ret := &objectMap{
		kvPairs: make(map[string]object),
	}
	for key, obj := range m.kvPairs {
		ret.kvPairs[key] = obj.clone()
	}
	return ret
}
func (m *objectMap) isTruthy() bool { return len(m.kvPairs) > 0 }

//

type objectString struct {
	value string
}

func (s *objectString) getType() objectType { return t_string }
func (s *objectString) inspect() string     { return s.value }
func (s *objectString) clone() object       { return &objectString{value: s.value} }
func (s *objectString) isTruthy() bool      { return len(s.value) > 0 }

//

type objectInteger struct {
	value int64
}

func (i *objectInteger) getType() objectType { return t_integer }
func (i *objectInteger) inspect() string     { return fmt.Sprintf("%d", i.value) }
func (i *objectInteger) clone() object {
	return &objectInteger{value: i.value}
}
func (i *objectInteger) isTruthy() bool { return i.value != 0 }

//

type objectFloat struct {
	value float64
}

func (f *objectFloat) getType() objectType { return t_float }
func (f *objectFloat) inspect() string     { return fmt.Sprintf("%f", f.value) }
func (f *objectFloat) clone() object {
	return &objectFloat{value: f.value}
}
func (i *objectFloat) isTruthy() bool { return i.value != 0 }

//

type objectBoolean struct {
	value bool
}

func (b *objectBoolean) getType() objectType { return t_boolean }
func (b *objectBoolean) inspect() string     { return fmt.Sprintf("%t", b.value) }
func (b *objectBoolean) clone() object {
	return objectFromBoolean(b.value)
}
func (b *objectBoolean) isTruthy() bool { return b.value }

//

type objectArrowFunction struct {
	paramName  string
	statements []statement
	functions  *functionStore
}

func (af *objectArrowFunction) getType() objectType { return t_arrow }
func (af *objectArrowFunction) inspect() string {
	blockString := "{}"
	statementStringList := []string{}
	for _, stmt := range af.statements {
		statementStringList = append(statementStringList, stmt.string())
	}
	if len(blockString) > 0 {
		blockString = fmt.Sprintf("{\n\t%s\n}", strings.Join(statementStringList, "\n\t"))
	}
	return fmt.Sprintf("%s ~> %s", af.paramName, blockString)
}
func (af *objectArrowFunction) clone() object {
	return af
}
func (b *objectArrowFunction) isTruthy() bool { return false }

//

type objectNull struct{}

func (n *objectNull) getType() objectType { return t_null }
func (n *objectNull) inspect() string     { return "null" }
func (n *objectNull) clone() object {
	return n
}
func (n *objectNull) isTruthy() bool { return false }

//

// returned by builtin funcs when signaling to halt further processing and return the env values as-is.
type objectTerminate struct {
	shouldReturnNull bool
}

func (t *objectTerminate) getType() objectType { return t_terminate }
func (t *objectTerminate) inspect() string     { return "TERMINATE" }
func (t *objectTerminate) clone() object {
	return &objectTerminate{shouldReturnNull: t.shouldReturnNull}
}
func (t *objectTerminate) isTruthy() bool { return false }

//

type objectError struct {
	lineCol string
	message string
}

func (e *objectError) getType() objectType { return t_error }
func (e *objectError) inspect() string {
	if len(e.lineCol) == 0 {
		return e.message
	}
	return fmt.Sprintf("%s: %s", e.lineCol, e.message)
}
func (e *objectError) clone() object {
	return &objectError{message: e.message}
}
func (e *objectError) isTruthy() bool { return false }

//

type objectTime struct {
	value time.Time
}

func (t *objectTime) getType() objectType { return t_time }
func (t *objectTime) inspect() string {
	return t.value.Format(time.RFC3339)
}
func (t *objectTime) clone() object {
	return &objectTime{value: t.value}
}
func (t *objectTime) isTruthy() bool {
	return !t.value.IsZero()
}
