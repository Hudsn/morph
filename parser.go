package morph

import (
	"fmt"
	"strconv"
)

const (
	_ int = iota
	LOWEST
	ASSIGN
	EQUALITY
	SUM
	PRODUCT
	PREFIX
	PATH
)

var precedenceMap = map[tokenType]int{
	TOK_MINUS:    SUM,
	TOK_PLUS:     SUM,
	TOK_ASTERISK: PRODUCT,
	TOK_SLASH:    PRODUCT,
	TOK_MOD:      PRODUCT,
	TOK_DOT:      PATH,
}

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

func (p *parser) registerFuncs() {
	p.registerPrefixFunc(TOK_MINUS, p.parsePrefixExpression)
	p.registerPrefixFunc(TOK_EXCLAMATION, p.parsePrefixExpression)
	p.registerPrefixFunc(tok_ident, p.parseIdentiferExpression)
	p.registerPrefixFunc(tok_int, p.parseIntegerLiteral)
	p.registerPrefixFunc(tok_float, p.parseFloatLiteral)
	p.registerPrefixFunc(TOK_TRUE, p.parseBooleanLiteral)
	p.registerPrefixFunc(TOK_FALSE, p.parseBooleanLiteral)
	p.registerPrefixFunc(tok_string, p.parseStringLiteral)
	p.registerPrefixFunc(tok_template_string, p.parseTemplateExpression)

	p.registerInfixFunc(TOK_DOT, p.parsePathExpression)
	// p.registerInfixFunc(TOK_TEMPLATE_START, p.parseTemplateExpression)

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
	for !p.isCurrentToken(tok_eof) && !p.isCurrentToken(TOK_ILLEGAL) {
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
	case TOK_SET:
		return p.parseSetStatement()
	case TOK_WHEN:
		return p.parseWhenStatement()
	default:
		return p.parseExpressionStatement()
	}
}

func (p *parser) parseExpression(precedence int) expression {
	prefixFn := p.prefixFuncMap[p.currentToken.tokenType]
	if prefixFn == nil {
		msg := fmt.Sprintf("unexpected sequence: %s", p.lexer.stringFromToken(p.currentToken))
		errPos := p.currentToken.start
		if p.isCurrentToken(tok_eof) {
			msg = "unexpected EOF"
			errPos = p.lexer.nextIdx
		}
		p.err(msg, errPos)
		return nil
	}
	leftExp := prefixFn()

	for precedence < p.peekPrecedence() {
		infixFn := p.infixFuncMap[p.peekToken.tokenType]
		if infixFn == nil {
			return leftExp
		}
		p.next()
		leftExp = infixFn(leftExp)
	}
	return leftExp
}

func (p *parser) parsePrefixExpression() expression {
	ret := &prefixExpression{tok: p.currentToken, operator: p.currentToken.value}
	p.next()
	ret.right = p.parseExpression(PREFIX)
	return ret
}

// specific expression parsers

func (p *parser) parseTemplateExpression() expression {
	ret := &templateExpression{tok: p.currentToken, parts: []expression{}}
	p.next()

	return ret
}

func (p *parser) parsePathExpression(left expression) expression {
	ret := &pathExpression{tok: p.currentToken}
	leftPart, ok := left.(pathPartExpression)
	if !ok {
		p.err(fmt.Sprintf("invalid path expression: %s", left.string()), left.position().start)
		return nil
	}
	ret.left = leftPart
	p.next()
	itemCandidate := p.parseExpression(PATH)
	item, ok := itemCandidate.(pathPartExpression)
	if !ok {
		p.err(fmt.Sprintf("invalid path expression: %s", itemCandidate.string()), itemCandidate.position().start)
		return nil
	}
	ret.attribute = item
	return ret
}

func (p *parser) parseIdentiferExpression() expression {
	return &identifierExpression{tok: p.currentToken, value: p.currentToken.value}
}

func (p *parser) parseIntegerLiteral() expression {
	ret := &integerLiteral{tok: p.currentToken}

	num, err := strconv.ParseInt(p.currentToken.value, 10, 64)
	if err != nil {
		msg := fmt.Sprintf("invalid integer: %s", p.currentToken.value)
		p.err(msg, p.currentToken.start)
	}
	ret.value = num
	return ret
}

func (p *parser) parseFloatLiteral() expression {
	ret := &floatLiteral{tok: p.currentToken}

	num, err := strconv.ParseFloat(p.currentToken.value, 64)
	if err != nil {
		msg := fmt.Sprintf("invalid float: %s", p.currentToken.value)
		p.err(msg, p.currentToken.start)
	}
	ret.value = num
	return ret
}

func (p *parser) parseBooleanLiteral() expression {
	ret := &booleanLiteral{tok: p.currentToken}
	switch p.currentToken.value {
	case "true":
		ret.value = true
	case "false":
		ret.value = false
	default:
		msg := fmt.Sprintf("invalid boolean: %s", p.currentToken.value)
		p.err(msg, p.currentToken.start)
		return nil
	}
	return ret
}

func (p *parser) parseStringLiteral() expression {
	return &stringLiteral{tok: p.currentToken, value: p.currentToken.value}
}

//

// specific statement parsers

func (p *parser) parseSetStatement() *setStatement {
	ret := &setStatement{tok: p.currentToken}
	start := p.currentToken.start
	if !p.mustNextToken(tok_ident) { // ident is fine here since paths always start with ident
		return nil
	}
	potentialTarget := p.parseExpression(LOWEST)
	target, ok := potentialTarget.(assignable)
	if !ok {
		sequence := p.rawStringFromStartEnd(start, p.currentToken.end)
		msg := fmt.Sprintf("SET statement should be followed by an assignable expression. instead got: %s", sequence)
		p.err(msg, p.currentToken.start)
		return nil
	}

	ret.target = target
	if !p.mustNextToken(tok_assign) {
		return nil
	}
	p.next()
	ret.value = p.parseExpression(LOWEST)
	return ret
}

func (p *parser) parseWhenStatement() *whenStatement {
	ret := &whenStatement{tok: p.currentToken}
	p.next() // to expr
	ret.condition = p.parseExpression(LOWEST)
	if !p.mustNextToken(TOK_DOUBLE_COLON) {
		return nil
	}
	p.next()
	ret.consequence = p.parseStatement()
	return ret
}

func (p *parser) parseExpressionStatement() *expressionStatement {
	ret := &expressionStatement{tok: p.currentToken}
	ret.expression = p.parseExpression(LOWEST)
	return ret
}

// helpers

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
	msg := fmt.Sprintf("unexpected token type. expected=%q got=%q", t, p.peekToken.tokenType)
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

func (p *parser) registerPrefixFunc(t tokenType, fn prefixFunc) {
	p.prefixFuncMap[t] = fn
}
func (p *parser) registerInfixFunc(t tokenType, fn infixFunc) {
	p.infixFuncMap[t] = fn
}

func (p *parser) peekPrecedence() int {
	return lookupPrecedence(p.peekToken.tokenType)
}

func lookupPrecedence(t tokenType) int {
	if precedence, ok := precedenceMap[t]; ok {
		return precedence
	}
	return LOWEST
}
