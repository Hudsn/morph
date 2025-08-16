package parser

import (
	"fmt"
)

type parser struct {
	lexer        *lexer
	currentToken token
	peekToken    token

	prevNode node

	errors []error
}

func newParser(l *lexer) *parser {
	p := &parser{
		lexer:  l,
		errors: []error{},
	}
	p.next()
	p.next()

	return p
}

func (p *parser) next() {
	p.currentToken = p.peekToken
	p.peekToken = p.lexer.tokenize()
	if p.isCurrentToken(ILLEGAL) {
		p.err("illegal token", p.currentToken.start)
	}
}

func (p *parser) parseProgram() (*program, error) {
	program := &program{statements: []statement{}}
	for !p.isCurrentToken(EOF) && !p.isCurrentToken(ILLEGAL) {
		statement := p.parseStatement()
		program.statements = append(program.statements, statement)
		p.next()
	}
	if len(p.errors) > 0 {
		return nil, p.errors[0]
	}

	return program, nil
}

func (p *parser) parseStatement() statement {
	switch p.currentToken.tokenType {

	default:

	}
}

func (p *parser) isCurrentToken(t tokenType) bool {
	return p.currentToken.tokenType == t
}
func (p *parser) isPeekToken(t tokenType) bool {
	return p.peekToken.tokenType == t
}
func (p *parser) mustNextToken(t tokenType) bool {
	if p.isPeekToken(t) {
		p.next()
		return true
	}
	msg := fmt.Sprintf("unexpected token type: %s", p.peekToken.tokenType)
	p.err(msg, p.peekToken.start)
	return false
}

func (p *parser) err(message string, position int) {
	line, col := lineAndCol(p.lexer.input, position)
	err := fmt.Errorf("parsing error at %d:%d:\n\t%s", line, col, message)
	p.errors = append(p.errors, err)
}
