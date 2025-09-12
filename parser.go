package morph

import (
	"fmt"
	"strconv"
)

const (
	_ int = iota
	lowest
	assign
	binary_or  // ||
	binary_and // &&
	equality   // includes inequality like !=, <= >, etc...
	sum
	product
	prefix  // !true or -1.234
	highest // dot expression like myobj.myattr, index like myarr[0], fn call like myfunc(arg)
)

var precedenceMap = map[tokenType]int{
	tok_equal:      equality,
	tok_not_equal:  equality,
	tok_lt:         equality,
	tok_lteq:       equality,
	tok_gt:         equality,
	tok_gteq:       equality,
	tok_binary_and: binary_and,
	tok_binary_or:  binary_or,
	tok_minus:      sum,
	tok_plus:       sum,
	tok_asterisk:   product,
	tok_slash:      product,
	tok_mod:        product,
	tok_lsquare:    highest,
	tok_dot:        highest,
	tok_lparen:     highest,
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
	p.registerPrefixFunc(tok_minus, p.parsePrefixExpression)
	p.registerPrefixFunc(tok_exclamation, p.parsePrefixExpression)
	p.registerPrefixFunc(tok_ident, p.parseIdentiferExpression)
	p.registerPrefixFunc(tok_int, p.parseIntegerLiteral)
	p.registerPrefixFunc(tok_float, p.parseFloatLiteral)
	p.registerPrefixFunc(tok_true, p.parseBooleanLiteral)
	p.registerPrefixFunc(tok_false, p.parseBooleanLiteral)
	p.registerPrefixFunc(tok_string, p.parseStringLiteral)
	p.registerPrefixFunc(tok_template_string, p.parseTemplateExpression)
	p.registerPrefixFunc(tok_lparen, p.parseGroupedExpression)
	p.registerPrefixFunc(tok_lsquare, p.parseArrayLiteral)
	p.registerPrefixFunc(tok_lcurly, p.parseMapLiteral)

	p.registerInfixFunc(tok_dot, p.parsePathExpression)
	p.registerInfixFunc(tok_plus, p.parseInfixExpression)
	p.registerInfixFunc(tok_minus, p.parseInfixExpression)
	p.registerInfixFunc(tok_asterisk, p.parseInfixExpression)
	p.registerInfixFunc(tok_slash, p.parseInfixExpression)
	p.registerInfixFunc(tok_mod, p.parseInfixExpression)
	p.registerInfixFunc(tok_equal, p.parseInfixExpression)
	p.registerInfixFunc(tok_not_equal, p.parseInfixExpression)
	p.registerInfixFunc(tok_lt, p.parseInfixExpression)
	p.registerInfixFunc(tok_lteq, p.parseInfixExpression)
	p.registerInfixFunc(tok_gt, p.parseInfixExpression)
	p.registerInfixFunc(tok_gteq, p.parseInfixExpression)
	p.registerInfixFunc(tok_binary_and, p.parseInfixExpression)
	p.registerInfixFunc(tok_binary_or, p.parseInfixExpression)
	p.registerInfixFunc(tok_lsquare, p.parseIndexExpression)
	p.registerInfixFunc(tok_lparen, p.parseCallExpression)
}

func (p *parser) next() {
	p.currentToken = p.peekToken
	p.peekToken = p.lexer.tokenize()
	if p.isCurrentToken(tok_illegal) {
		p.err("illegal token", p.currentToken.start)
	}
}

func (p *parser) parseProgram() (*program, error) {
	program := &program{statements: []statement{}}
	for !p.isCurrentToken(tok_eof) && !p.isCurrentToken(tok_illegal) {
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
	case tok_set:
		return p.parseSetStatement()
	case tok_when:
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

func (p *parser) parseInfixExpression(left expression) expression {
	ret := &infixExpression{tok: p.currentToken, left: left, operator: p.currentToken.value}
	precedence := lookupPrecedence(p.currentToken.tokenType)
	p.next()
	ret.right = p.parseExpression(precedence)
	return ret
}

func (p *parser) parsePrefixExpression() expression {
	ret := &prefixExpression{tok: p.currentToken, operator: p.currentToken.value}
	p.next()
	ret.right = p.parseExpression(prefix)
	return ret
}

// specific expression parsers

// fn call
func (p *parser) parseCallExpression(left expression) expression {
	funcName, ok := left.(assignable)
	if !ok {
		p.err("invalid function name", left.position().start)
		return nil
	}
	ret := &callExpression{tok: p.currentToken, name: funcName}
	ret.arguments = p.parseExpressionList(tok_rparen)
	ret.endPos = p.currentToken.end
	return ret
}

func (p *parser) parseMapLiteral() expression {
	ret := &mapLiteral{tok: p.currentToken}
	ret.pairs = make(map[string]expression)
	for !p.isPeekToken(tok_rcurly) {
		p.next()
		key := p.parseExpression(lowest)
		strNode, ok := key.(*stringLiteral)
		if !ok {
			p.err("map key expression must be a string literal", key.position().start)
			return nil
		}
		if !p.mustNextToken(tok_colon) {
			return nil
		}
		p.next()
		ret.pairs[strNode.value] = p.parseExpression(lowest)
		if !p.isPeekToken(tok_rcurly) && !p.mustNextToken(tok_comma) {
			return nil
		}
	}
	if !p.mustNextToken(tok_rcurly) {
		return nil
	}
	ret.endPos = p.currentToken.end
	return ret
}

// arrays
func (p *parser) parseArrayLiteral() expression {
	ret := &arrayLiteral{tok: p.currentToken}
	ret.entries = p.parseExpressionList(tok_rsquare)
	ret.endPos = p.currentToken.end
	return ret
}

func (p *parser) parseExpressionList(endTok tokenType) []expression {
	if p.isPeekToken(endTok) {
		p.next()
		return []expression{}
	}
	p.next() // to first item

	ret := []expression{p.parseExpression(lowest)}

	for p.isPeekToken(tok_comma) {
		p.next() // to comma
		p.next() // to next expression
		ret = append(ret, p.parseExpression(lowest))
	}

	if !p.mustNextToken(endTok) {
		return nil
	}

	return ret
}

//

func (p *parser) parseIndexExpression(left expression) expression {
	leftExpr, ok := left.(assignable)
	if !ok {
		p.err("invalid index expression. cannot call index on a non-identifier", left.position().start)
		return nil
	}
	ret := &indexExpression{tok: p.currentToken, left: leftExpr}
	if p.isPeekToken(tok_rsquare) {
		p.err("invalid index expression", p.currentToken.start)
		return nil
	}
	p.next()

	ret.index = p.parseExpression(lowest)
	p.mustNextToken(tok_rsquare)
	ret.endPos = p.currentToken.end

	return ret
}

//

func (p *parser) parseGroupedExpression() expression {
	p.next()
	exp := p.parseExpression(lowest)
	if !p.mustNextToken(tok_rparen) {
		return nil
	}
	return exp
}

//

func (p *parser) parseTemplateExpression() expression {
	if !p.mustCurrentToken(tok_template_string) {
		return nil
	}
	ret := &templateExpression{tok: p.currentToken, parts: []expression{}}

	ret.parts = append(ret.parts, &stringLiteral{tok: p.currentToken, value: p.currentToken.value})

	for p.isPeekToken(tok_template_string) || p.isPeekToken(tok_template_start) {
		p.next()
		switch p.currentToken.tokenType {
		case tok_template_string:
			ret.parts = append(ret.parts, &stringLiteral{tok: p.currentToken, value: p.currentToken.value})
		case tok_template_start:
			if toAdd, gotExpr := p.parseTemplateInnerExpression(); gotExpr {
				ret.parts = append(ret.parts, toAdd)
			}
		}
	}
	return ret
}

func (p *parser) parseTemplateInnerExpression() (expression, bool) {
	p.next()
	if p.isCurrentToken(tok_rcurly) {
		return nil, false
	}
	toAdd := p.parseExpression(lowest)
	p.mustNextToken(tok_rcurly)
	return toAdd, true
}

func (p *parser) parsePathExpression(left expression) expression {
	ret := &pathExpression{tok: p.currentToken}
	precedence := lookupPrecedence(p.currentToken.tokenType)
	leftPart, ok := left.(pathPartExpression)
	if !ok {
		p.err(fmt.Sprintf("invalid path expression: %s", left.string()), left.position().start)
		return nil
	}
	ret.left = leftPart
	p.next()

	itemCandidate := p.parseExpression(precedence)
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
		return nil
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
		return nil
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
	potentialTarget := p.parseExpression(lowest)
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
	ret.value = p.parseExpression(lowest)
	return ret
}

func (p *parser) parseWhenStatement() *whenStatement {
	ret := &whenStatement{tok: p.currentToken}
	p.next() // to expr
	ret.condition = p.parseExpression(lowest)
	if !p.mustNextToken(tok_double_colon) {
		return nil
	}
	p.next()
	ret.consequence = p.parseStatement()
	return ret
}

func (p *parser) parseExpressionStatement() *expressionStatement {
	ret := &expressionStatement{tok: p.currentToken}
	ret.expression = p.parseExpression(lowest)
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
func (p *parser) mustCurrentToken(t tokenType) bool {
	if p.isCurrentToken(t) {
		return true
	}
	msg := fmt.Sprintf("unexpected token type. expected=%q got=%q", t, p.currentToken.tokenType)
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
	return lowest
}
