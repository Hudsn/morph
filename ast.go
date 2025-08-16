package parser

import "strings"

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
