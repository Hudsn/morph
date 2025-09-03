package morph

import (
	"fmt"
	"strings"
)

type node interface {
	token() token
	string() string
	position() position
}

type position struct {
	start int
	end   int
}

type statement interface {
	node
	statementNode()
}

type expression interface {
	node
	expressionNode()
}

type program struct {
	tok        token
	statements []statement
}

func (p *program) token() token {
	return p.tok
}
func (p *program) string() string {
	strs := []string{}
	for _, entry := range p.statements {
		strs = append(strs, entry.string())
	}
	return strings.Join(strs, "\n")
}
func (p *program) position() position {
	ret := position{start: 0, end: 0}
	if len(p.statements) == 0 {
		return ret
	}
	ret.start = p.statements[0].position().start
	ret.end = p.statements[len(p.statements)-1].position().end
	return ret
}

// statements
//

type setStatement struct {
	tok    token
	target assignable
	value  expression
}

type assignable interface {
	expression
	toAssignPath() *assignPath
}

type assignStepType string

const (
	assign_step_env     assignStepType = "ENV"
	assign_step_map_key assignStepType = "MAPKEY"
	assign_step_invalid assignStepType = "INVALID" // things like index expressions. eg: we don't want to be able to directly assign indexes like item[99] = "abc"
)

type assignPath struct {
	stepType assignStepType
	partName string
	next     *assignPath
}

func (s *setStatement) statementNode() {}
func (s *setStatement) token() token   { return s.tok }
func (s *setStatement) string() string {
	ret := ""
	if s.target != nil {
		ret = s.target.string()
	}
	return fmt.Sprintf("%s = %s", ret, s.value.string())
}
func (s *setStatement) position() position {
	return position{
		start: s.target.position().start,
		end:   s.value.position().end,
	}
}

//

type whenStatement struct {
	tok         token
	condition   expression
	consequence statement
}

func (ws *whenStatement) statementNode() {}
func (ws *whenStatement) token() token   { return ws.tok }
func (ws *whenStatement) string() string {
	return fmt.Sprintf("%s %s :: %s", ws.tok.value, ws.condition.string(), ws.consequence.string())
}
func (ws *whenStatement) position() position {
	return position{
		start: ws.tok.start,
		end:   ws.consequence.position().end,
	}
}

//

type expressionStatement struct {
	tok        token
	expression expression
}

func (es *expressionStatement) statementNode() {}
func (es *expressionStatement) string() string {
	if es.expression != nil {
		return es.expression.string()
	}
	return ""
}
func (es *expressionStatement) token() token { return es.tok }
func (es *expressionStatement) position() position {
	return position{
		start: es.expression.position().start,
		end:   es.expression.position().end,
	}
}

// expressions
//

type prefixExpression struct {
	tok      token
	operator string
	right    expression
}

func (pe *prefixExpression) expressionNode() {}
func (pe *prefixExpression) token() token    { return pe.tok }
func (pe *prefixExpression) string() string {
	return fmt.Sprintf("(%s%s)", pe.operator, pe.right.string())
}
func (pe *prefixExpression) position() position {
	return position{
		start: pe.tok.start,
		end:   pe.right.position().end,
	}
}

//

type infixExpression struct {
	tok      token
	left     expression
	operator string
	right    expression
}

func (ie *infixExpression) expressionNode() {}
func (ie *infixExpression) token() token    { return ie.tok }
func (ie *infixExpression) string() string {
	return fmt.Sprintf("(%s %s %s)", ie.left.string(), ie.operator, ie.right.string())
}
func (ie *infixExpression) position() position {
	return position{
		start: ie.left.position().start,
		end:   ie.right.position().end,
	}
}

//

type arrayLiteral struct {
	tok     token
	entries []expression
	endPos  int
}

func (a *arrayLiteral) expressionNode() {}
func (a *arrayLiteral) token() token    { return a.tok }
func (a *arrayLiteral) string() string {
	stringItems := []string{}
	for _, entry := range a.entries {
		stringItems = append(stringItems, entry.string())
	}
	return fmt.Sprintf("[%s]", strings.Join(stringItems, ", "))
}
func (a *arrayLiteral) position() position {
	return position{
		start: a.tok.start,
		end:   a.endPos,
	}
}

//

type indexExpression struct {
	tok    token
	left   *identifierExpression
	index  expression
	endPos int
}

func (ie *indexExpression) expressionNode() {}
func (ie *indexExpression) pathPartNode()   {}
func (ie *indexExpression) token() token    { return ie.tok }
func (ie *indexExpression) string() string {
	return fmt.Sprintf("(%s[%s])", ie.left.string(), ie.index.string())
}
func (ie *indexExpression) position() position {
	return position{
		start: ie.left.position().start,
		end:   ie.endPos,
	}
}

//

type identifierExpression struct {
	tok   token
	value string
}

func (ie *identifierExpression) expressionNode() {}
func (ie *identifierExpression) pathPartNode()   {}
func (ie *identifierExpression) token() token    { return ie.tok }
func (ie *identifierExpression) string() string  { return ie.value }
func (ie *identifierExpression) position() position {
	return position{
		start: ie.tok.start,
		end:   ie.tok.end,
	}
}
func (ie *identifierExpression) toAssignPath() *assignPath {
	return &assignPath{stepType: assign_step_env, partName: ie.value, next: nil}
}

//

// nodes that can be part of a valid path to fetch data. ex: ident, string, indexExpression, or another pathexpression
type pathPartExpression interface {
	expression
	pathPartNode()
}

type pathExpression struct {
	left      pathPartExpression
	tok       token
	attribute pathPartExpression
}

func (pe *pathExpression) expressionNode() {}
func (pe *pathExpression) pathPartNode()   {}
func (pe *pathExpression) token() token    { return pe.tok }
func (pe *pathExpression) position() position {
	return position{
		start: pe.left.position().start,
		end:   pe.attribute.position().end,
	}
}
func (pe *pathExpression) string() string {
	return fmt.Sprintf("%s.%s", pe.left.string(), pe.attribute.string())
}
func (pe *pathExpression) toAssignPath() *assignPath {
	return pe.toAssignPathRe(pe, nil)
}
func (pe *pathExpression) toAssignPathRe(current pathPartExpression, next *assignPath) *assignPath {
	ret := &assignPath{next: next}
	switch v := current.(type) {
	case *identifierExpression:
		ret.stepType = assign_step_env
		ret.partName = v.value
	case *pathExpression:
		ret.stepType, ret.partName = handlePathStepAttribute(v.attribute)
		return pe.toAssignPathRe(v.left, ret)
	default:
		ret.stepType = assign_step_invalid
		ret.partName = ""
	}
	return ret
}
func handlePathStepAttribute(attr pathPartExpression) (assignStepType, string) {
	switch v := attr.(type) {
	case *identifierExpression:
		return assign_step_map_key, v.value
	default:
		return assign_step_invalid, ""
	}
}

//

type integerLiteral struct {
	tok   token
	value int64
}

func (il *integerLiteral) expressionNode() {}
func (il *integerLiteral) token() token    { return il.tok }
func (il *integerLiteral) string() string  { return il.tok.value }
func (il *integerLiteral) position() position {
	return position{
		start: il.tok.start,
		end:   il.tok.end,
	}
}

//

type floatLiteral struct {
	tok   token
	value float64
}

func (fl *floatLiteral) expressionNode() {}
func (fl *floatLiteral) token() token    { return fl.tok }
func (fl *floatLiteral) string() string  { return fl.tok.value }
func (fl *floatLiteral) position() position {
	return position{
		start: fl.tok.start,
		end:   fl.tok.end,
	}
}

//

type booleanLiteral struct {
	tok   token
	value bool
}

func (bl *booleanLiteral) expressionNode() {}
func (bl *booleanLiteral) token() token    { return bl.tok }
func (bl *booleanLiteral) string() string  { return bl.tok.value }
func (bl *booleanLiteral) position() position {
	return position{
		start: bl.tok.start,
		end:   bl.tok.end,
	}
}

//

type stringLiteral struct {
	tok   token
	value string
}

func (sl *stringLiteral) expressionNode() {}
func (sl *stringLiteral) pathPartNode()   {}
func (sl *stringLiteral) token() token    { return sl.tok }
func (sl *stringLiteral) string() string  { return fmt.Sprintf("\"%s\"", sl.value) }
func (sl *stringLiteral) position() position {
	return position{
		start: sl.tok.start,
		end:   sl.tok.end,
	}
}

//

type templateExpression struct {
	tok   token
	parts []expression
}

func (te *templateExpression) expressionNode() {}
func (te *templateExpression) token() token    { return te.tok }
func (te *templateExpression) position() position {
	if len(te.parts) == 0 {
		return position{
			start: te.tok.start,
			end:   te.tok.end,
		}
	}
	return position{
		start: te.parts[0].position().start,
		end:   te.parts[len(te.parts)-1].position().end,
	}
}
func (te *templateExpression) string() string {
	stringParts := []string{}
	for _, entry := range te.parts {
		stringParts = append(stringParts, entry.string())
	}
	return strings.Join(stringParts, "")
}
