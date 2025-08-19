package parser

import (
	"fmt"
	"strings"
)

type node interface {
	token() token
	string() string
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

type prefixExpression struct {
	tok      token
	operator string
	right    expression
}

func (pe *prefixExpression) expressionNode() {}
func (pe *prefixExpression) token() token    { return pe.tok }
func (pe *prefixExpression) string() string {
	return fmt.Sprintf("%s%s", pe.operator, pe.right.string())
}
