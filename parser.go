package parser

import (
	"fmt"
)

const (
	_ int = iota
	LOWEST
	ASSIGN
	WHEN
	EQUALITY
	SUM
	PRODUCT
	DOT_CALL_INDEX
)

type prefixFunc func() expression
type infixFunc func(expression) expression

type parser struct {
	lexer        *lexer
	currentToken token
	peekToken    token

	prefixFuncMap map[tokenType]prefixFunc
	infixFuncMap  map[tokenType]infixFunc

	errors []error
}

func newParser(l *lexer) *parser {
	p := &parser{
		lexer:         l,
		errors:        []error{},
		prefixFuncMap: map[tokenType]prefixFunc{},
		infixFuncMap:  map[tokenType]infixFunc{},
	}
	p.registerFuncs()
	p.next()
	p.next()
	return p
}

func (p *parser) next() {
	p.currentToken = p.peekToken
	p.peekToken = p.lexer.tokenize()
	if p.isCurrentToken(TOK_ILLEGAL) {
		p.err("illegal token", p.currentToken.start)
	}
}

func (p *parser) parseProgram() (*program, error) {
	program := &program{statements: []statement{}}
	for !p.isCurrentToken(TOK_EOF) && !p.isCurrentToken(TOK_ILLEGAL) {
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
		return p.parseExpressionStatement()
	}
}

func (p *parser) parseExpression(precedence int) expression {
	prefixFn := p.prefixFuncMap[p.currentToken.tokenType]
	if prefixFn == nil {
		msg := fmt.Sprintf("unexpected sequence: %s", p.rawStringFromStartEnd(p.currentToken.start, p.currentToken.end))
		p.err(msg, p.currentToken.start)
		return nil
	}
	leftExp := prefixFn()

	for precedence < p.peekPrecedence() {
		infixFn := p.infixFuncMap[p.currentToken.tokenType]
		if infixFn == nil {
			return leftExp
		}
		p.next()
		leftExp = infixFn(leftExp)
	}

	return leftExp
}

func (p *parser) parseExpressionStatement() *expressionStatement {
	ret := &expressionStatement{tok: p.currentToken}
	ret.expression = p.parseExpression(LOWEST)

	return ret
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

func (p *parser) rawStringFromStartEnd(start, end int) string {
	return string(p.lexer.input[start:end])
}

func (p *parser) registerFuncs() {

}

func (p *parser) registerPrefixFunc(t tokenType, fn prefixFunc) {
	p.prefixFuncMap[t] = fn
}
func (p *parser) registerInfixFunc(t tokenType, fn infixFunc) {
	p.infixFuncMap[t] = fn
}

func (p *parser) peekPrecedence() int {
	return lookupPrecedence(p.peekToken.tokenType)
}

var precedenceMap = map[tokenType]int{}

func lookupPrecedence(t tokenType) int {
	if precedence, ok := precedenceMap[t]; ok {
		return precedence
	}
	return LOWEST
}
