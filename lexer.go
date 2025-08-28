package morph

import (
	"slices"
)

type lexer struct {
	input       []rune
	currentChar rune
	currentIdx  int
	nextIdx     int

	currentContext *lexContext
	contextStack   []*lexContext
}

type lexContext struct {
	outerHandler func() token
	depthCounter int
}

const NULLCHAR = rune(0)

func newLexer(input []rune) *lexer {
	l := &lexer{
		input:          input,
		currentIdx:     0,
		nextIdx:        0,
		currentContext: nil,
		contextStack:   []*lexContext{},
	}
	l.next()
	return l
}

func (l *lexer) next() {
	if l.nextIdx >= len(l.input) {
		l.currentChar = NULLCHAR
	} else {
		l.currentChar = l.input[l.nextIdx]
	}
	l.currentIdx = l.nextIdx
	l.nextIdx++
}

func (l *lexer) peek() rune {
	if l.nextIdx >= len(l.input) {
		return NULLCHAR
	}
	return l.input[l.nextIdx]
}

func newLexContext(callback func() token) *lexContext {
	return &lexContext{callback, 1}
}
func (l *lexer) pushNewContext(callbackFn func() token) {
	l.pushContext(newLexContext(callbackFn))
}
func (l *lexer) pushContext(c *lexContext) {
	l.contextStack = append(l.contextStack, l.currentContext)
	l.currentContext = c
}
func (l *lexer) popContext() {
	if len(l.contextStack) == 0 {
		l.currentContext = nil
		return
	}
	l.currentContext = l.contextStack[0]
	l.contextStack = l.contextStack[1:]
}

func (l *lexer) tokenize() token {
	var tok token

	l.handleWhiteSpace()
	if l.currentContext != nil && l.currentContext.depthCounter == 0 {
		nextTokHandler := l.currentContext.outerHandler
		l.popContext()
		return nextTokHandler()
	}

	switch l.currentChar {
	case NULLCHAR:
		tok = l.handleEOF()
	case '=':
		tok = l.handleEqual()
	case '.':
		tok = l.handleDot()
		if tok.tokenType == TOK_FLOAT {
			// readnumber (FLOAT processor) progresses tokens already, so we want to return early here to avoid hitting the next() call at the end of the func
			return tok
		}
	case ':':
		tok = l.handleColon()
	case '-':
		tok = l.handleMinus()
	case '!':
		tok = l.handleExclamation()
	case '\'':
		start := l.currentIdx
		l.next() // step inside quote; useful for mode switching so we can only worry about having to check for a closing ' rather than risking advancing as the first action inside the function where we could accidentally go from a closed template literal to a quote, and miss the end quote. For example: 'hello ${world}'  could pick up the string parsing again after the }, and if we called next inside the handler with it starting on ', and then skip it.
		tok = l.handleSingleQuote()
		tok.start = start
	case '{':
		return l.handleLCurly()
	case '}':
		return l.handleRCurly()
	case '$':
		return l.handleDollarSign()
	default:
		if isDigit(l.currentChar) {
			return l.readNumber()
		} else if isLetter(l.currentChar) {
			return l.readIdentifier()
		} else {
			tok = token{tokenType: TOK_ILLEGAL, start: l.currentIdx, end: l.nextIdx, value: string(l.currentChar)}
		}
	}

	l.next()

	return tok
}

// handlers

func (l *lexer) handleWhiteSpace() {
	for slices.Contains([]rune{'\r', '\n', '\t', ' '}, l.currentChar) {
		l.next()
	}
}

func (l *lexer) handleEOF() token {
	return token{
		tokenType: TOK_EOF,
		value:     string(l.currentChar),
		start:     l.currentIdx,
		end:       l.nextIdx,
	}
}

func (l *lexer) handleEqual() token {
	return token{
		tokenType: TOK_ASSIGN,
		value:     string(l.currentChar),
		start:     l.currentIdx,
		end:       l.nextIdx,
	}
}

func (l *lexer) handleDot() token {
	if isDigit(l.peek()) {
		return l.readNumber()
	}
	return token{
		tokenType: TOK_DOT,
		value:     string(l.currentChar),
		start:     l.currentIdx,
		end:       l.nextIdx,
	}
}

func (l *lexer) handleMinus() token {
	return token{
		tokenType: TOK_MINUS,
		value:     string(l.input[l.currentIdx:l.nextIdx]),
		start:     l.currentIdx,
		end:       l.nextIdx,
	}
}

func (l *lexer) handleExclamation() token {
	return token{
		tokenType: TOK_EXCLAMATION,
		value:     string(l.input[l.currentIdx:l.nextIdx]),
		start:     l.currentIdx,
		end:       l.nextIdx,
	}
}

func (l *lexer) handleColon() token {
	start := l.currentIdx
	tok := token{
		tokenType: TOK_COLON,
		start:     start,
	}
	if l.peek() == ':' {
		tok.tokenType = TOK_DOUBLE_COLON
		l.next()
	}
	tok.value = string(l.input[start:l.nextIdx])
	tok.end = l.nextIdx
	return tok
}

func (l *lexer) handleSingleQuote() token {
	start := l.currentIdx
	str := []rune{}
	tok := token{
		tokenType: TOK_STRING,
		start:     start,
	}
	for l.currentChar != '\'' && l.currentChar != NULLCHAR {
		toAdd := l.currentChar

		if l.currentChar == '\\' {

			toAdd = l.handleEscapeString('\'')
			if toAdd == NULLCHAR {
				tok.tokenType = TOK_ILLEGAL
				tok.value = "invalid escape sequence"
				l.next()
				tok.end = l.nextIdx
				return tok
			}
		}
		if l.currentChar == '$' && l.peek() == '{' {
			l.pushNewContext(l.handleSingleQuote)
			tok.end = l.currentIdx
			tok.value = string(str)
			return tok
		}
		l.next()
		str = append(str, toAdd)

	}
	if l.currentChar == NULLCHAR {
		tok.tokenType = TOK_ILLEGAL
		tok.value = "string literal not terminated"
		tok.end = l.nextIdx
		return tok
	}
	tok.value = string(str)
	tok.end = l.nextIdx
	return tok
}

func (l *lexer) handleEscapeString(quoteChar rune) rune {
	var ret rune
	switch l.peek() {
	case '\\':
		ret = '\\'
	case 't':
		ret = '\t'
	case 'n':
		ret = '\n'
	case quoteChar:
		ret = quoteChar
	case NULLCHAR:
		return NULLCHAR
	default:
		ret = NULLCHAR
	}
	l.next()
	return ret
}

func (l *lexer) handleDoubleQuote() token {
	// if l.currentChar == '\n' {
	// 		tok.tokenType = TOK_ILLEGAL
	// 		tok.value = "string literal not terminated"
	// 		tok.end = l.nextIdx
	// 		return tok
	// 	}
	return token{}
}

func (l *lexer) handleDollarSign() token {
	if l.peek() != '{' {
		return token{
			tokenType: TOK_ILLEGAL,
			value:     "invalid template expression syntax",
			start:     l.currentIdx,
			end:       l.nextIdx,
		}
	}
	tok := token{tokenType: TOK_TEMPLATE_START, start: l.currentIdx, value: "${"}
	l.next()
	tok.end = l.nextIdx
	return tok
}

func (l *lexer) handleRCurly() token {
	if l.currentContext != nil {
		l.currentContext.depthCounter = max(l.currentContext.depthCounter-1, 0)
	}
	return token{tokenType: TOK_RCURLY, value: "}", start: l.currentIdx, end: l.nextIdx}
}

func (l *lexer) handleLCurly() token {
	if l.currentContext != nil {
		l.currentContext.depthCounter++
	}
	return token{tokenType: TOK_LCURLY, value: "{", start: l.currentIdx, end: l.nextIdx}
}

// reader helpers

func (l *lexer) readIdentifier() token {
	start := l.currentIdx
	for isDigit(l.currentChar) || isLetter(l.currentChar) || isLineChar(l.currentChar) {
		l.next()
	}
	val := string(l.input[start:l.currentIdx])
	return token{
		tokenType: lookupTokenKeyword(val),
		value:     val,
		start:     start,
		end:       l.currentIdx,
	}
}

func (l *lexer) readNumber() token {
	tok := token{tokenType: TOK_INT}
	start := l.currentIdx
	encounteredDot := false
	for l.currentChar == '.' || isDigit(l.currentChar) {
		if l.currentChar == '.' {
			if !isDigit(l.peek()) || encounteredDot {
				break
			}
			tok.tokenType = TOK_FLOAT
			encounteredDot = true
		}
		l.next()
	}
	tok.start = start
	tok.end = l.currentIdx
	tok.value = string(l.input[tok.start:tok.end])
	return tok
}

func isLineChar(char rune) bool {
	return slices.Contains([]rune{'_', '-'}, char)
}
func isLetter(char rune) bool {
	return ('a' <= char && char <= 'z') || ('A' <= char && char <= 'Z')
}
func isDigit(char rune) bool {
	return '0' <= char && char <= '9'
}
