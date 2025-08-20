package parser

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

// things like an identifier or path. ex: myIdent or myobj.mypath
type assignable interface {
	node
	assignableNode()
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
	// TODO
}

//

type identifierExpression struct {
	tok   token
	value string
}

func (ie *identifierExpression) expressionNode() {}
func (ie *identifierExpression) assignableNode() {}
func (ie *identifierExpression) token() token    { return ie.tok }
func (ie *identifierExpression) string() string  { return ie.value }
func (ie *identifierExpression) position() position {
	return position{
		start: ie.tok.start,
		end:   ie.tok.end,
	}
}

//

type pathExpression struct {
	tok  token
	left expression
	item expression
}

func (pe *pathExpression) expressionNode() {}
func (pe *pathExpression) assignableNode() {}
func (pe *pathExpression) token() token    { return pe.tok }
func (pe *pathExpression) string() string {
	return fmt.Sprintf("%s.%s", pe.left.string(), pe.item.string())
}
func (pe *pathExpression) position() position {
	return position{
		start: pe.left.position().start,
		end:   pe.item.position().end,
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
		start: bl.position().start,
		end:   bl.position().end,
	}
}
