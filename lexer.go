package morph

import (
	"slices"
	"strings"
)

type lexer struct {
	input       []rune
	currentChar rune
	currentIdx  int
	nextIdx     int
	isEnd       bool

	context *lexContext
}

const nullchar = rune(0)

func newLexer(input []rune) *lexer {
	l := &lexer{
		input:      input,
		currentIdx: 0,
		nextIdx:    0,
		context:    defaultLexContext,
	}
	l.next()
	return l
}

func (l *lexer) next() {
	if l.nextIdx >= len(l.input) || l.isEnd {
		l.currentChar = nullchar
		l.isEnd = true
	} else {
		l.currentChar = l.input[l.nextIdx]
	}
	// NOTE: use of max to safely handle progression to avoid oob errors.
	l.currentIdx = min(l.nextIdx, len(l.input)-1)
	l.nextIdx = min(l.nextIdx+1, len(l.input))
}

func (l *lexer) peek() rune {
	if l.nextIdx >= len(l.input) {
		return nullchar
	}
	return l.input[l.nextIdx]
}

func (l *lexer) tokenize() token {
	var tok token

	if l.context.outer != nil && l.context.depthCounter == 0 {
		handlerFn := lexerContextTypeToHandlerFn(l, l.context.contextType)
		l.context = l.context.outer
		tok = handlerFn()
		return tok
	}

	l.handleWhiteSpace() // must come after context handling since string contexts want to include whitespace rather than "eat" it

	switch l.currentChar {
	case nullchar:
		tok = l.handleEOF()
	case '=':
		tok = l.handleEqual()
	case '<':
		tok = l.handleLT()
	case '>':
		tok = l.handleGT()
	case ',':
		tok = token{tokenType: tok_comma, start: l.currentIdx, end: l.nextIdx, value: string(l.currentChar), lineCol: l.lineColString(l.currentIdx)}
	case '.':
		tok = l.handleDot()
		if tok.tokenType == tok_float {
			// readnumber (the FLOAT processor) will have progressed tokens already, so we want to return early here to avoid hitting an erroneous extra next() call at the end of the func
			return tok
		}
	case ':':
		tok = l.handleColon()
	case '+':
		tok = token{tokenType: tok_plus, start: l.currentIdx, end: l.nextIdx, value: string(l.currentChar), lineCol: l.lineColString(l.currentIdx)}
	case '-':
		tok = token{tokenType: tok_minus, start: l.currentIdx, end: l.nextIdx, value: string(l.currentChar), lineCol: l.lineColString(l.currentIdx)}
	case '*':
		tok = token{tokenType: tok_asterisk, start: l.currentIdx, end: l.nextIdx, value: string(l.currentChar), lineCol: l.lineColString(l.currentIdx)}
	case '/':
		if l.peek() != '/' {
			tok = token{tokenType: tok_slash, start: l.currentIdx, end: l.nextIdx, value: string(l.currentChar), lineCol: l.lineColString(l.currentIdx)}
		} else { // handle comment
			l.next()
			for l.peek() != '\n' && l.peek() != nullchar {
				l.next()
			}
			l.next()
			return l.tokenize()
		}
	case '%':
		tok = token{tokenType: tok_mod, start: l.currentIdx, end: l.nextIdx, value: string(l.currentChar), lineCol: l.lineColString(l.currentIdx)}
	case '!':
		tok = l.handleExclamation()
	case '\'':
		start := l.currentIdx
		l.next() // step inside quote; useful for mode switching so we can only worry about having to check for a closing ' rather than risking advancing as the first action inside the function where we could accidentally go from a closed template literal to a quote, and miss the end quote. For example: 'hello ${world}'  could pick up the string parsing again after the }, and if we called next inside the handler with it starting on ', and then skip it.
		tok = l.handleSingleQuote()
		tok.start = start
		tok.lineCol = l.lineColString(start)
		return tok
	case '&':
		tok = l.handleAmpersand()
	case '|':
		tok = l.handlePipe()
	case '"':
		start := l.currentIdx
		l.next()
		tok = l.handleDoubleQuote()
		tok.start = start
		tok.lineCol = l.lineColString(start)
	case '{':
		tok = l.handleLCurly()
	case '}':
		tok = l.handleRCurly()
	case '[':
		tok = token{tokenType: tok_lsquare, value: "[", start: l.currentIdx, end: l.nextIdx, lineCol: l.lineColString(l.currentIdx)}
	case ']':
		tok = token{tokenType: tok_rsquare, value: "]", start: l.currentIdx, end: l.nextIdx, lineCol: l.lineColString(l.currentIdx)}
	case '(':
		tok = token{tokenType: tok_lparen, value: "(", start: l.currentIdx, end: l.nextIdx, lineCol: l.lineColString(l.currentIdx)}
	case ')':
		tok = token{tokenType: tok_rparen, value: ")", start: l.currentIdx, end: l.nextIdx, lineCol: l.lineColString(l.currentIdx)}
	case '$':
		tok = l.handleDollarSign()
	case '~':
		tok = l.handleTilde()
	default:
		if isDigit(l.currentChar) {
			tok = l.readNumber()
			return tok
		} else if isLetter(l.currentChar) || l.currentChar == '@' {
			tok = l.readIdentifier()
			return tok
		} else {
			tok = token{tokenType: tok_illegal, start: l.currentIdx, end: l.nextIdx, value: string(l.currentChar)}
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
		tokenType: tok_eof,
		value:     string(l.currentChar),
		start:     0, // just set to 0 at this point since we're done anyway
		end:       0,
		lineCol:   l.lineColString(len(l.input)),
	}
}

func (l *lexer) handleTilde() token {
	if l.peek() != '>' {
		return token{
			tokenType: tok_illegal,
			value:     "unexpected input character",
			start:     l.currentIdx,
			end:       l.nextIdx,
			lineCol:   l.lineColString(l.currentIdx),
		}
	}
	start := l.currentIdx
	l.next()
	return token{
		tokenType: tok_arrow,
		value:     string(l.input[start:l.nextIdx]),
		start:     start,
		end:       l.nextIdx,
		lineCol:   l.lineColString(start),
	}
}

func (l *lexer) handleAmpersand() token {
	if l.peek() != '&' {
		return token{
			tokenType: tok_illegal,
			value:     "invalid ampersand use",
			start:     l.currentIdx,
			end:       l.nextIdx,
			lineCol:   l.lineColString(l.currentIdx),
		}
	}
	start := l.currentIdx
	l.next()
	return token{
		tokenType: tok_binary_and,
		value:     string(l.input[start:l.nextIdx]),
		start:     start,
		end:       l.nextIdx,
		lineCol:   l.lineColString(start),
	}
}

func (l *lexer) handlePipe() token {
	start := l.currentIdx
	if l.peek() == '|' {
		l.next()
		return token{
			tokenType: tok_binary_or,
			start:     start,
			end:       l.nextIdx,
			value:     string(l.input[start:l.nextIdx]),
			lineCol:   l.lineColString(start),
		}
	}
	return token{
		tokenType: tok_pipe,
		start:     start,
		end:       l.nextIdx,
		value:     string(l.input[start:l.nextIdx]),
		lineCol:   l.lineColString(start),
	}
}

func (l *lexer) handleLT() token {
	start := l.currentIdx
	if l.peek() == '=' {
		l.next()
		return token{
			tokenType: tok_lteq,
			start:     start,
			end:       l.nextIdx,
			value:     string(l.input[start:l.nextIdx]),
			lineCol:   l.lineColString(start),
		}
	}
	return token{
		tokenType: tok_lt,
		start:     start,
		end:       l.nextIdx,
		value:     string(l.input[start:l.nextIdx]),
		lineCol:   l.lineColString(start),
	}
}

func (l *lexer) handleGT() token {
	start := l.currentIdx
	if l.peek() == '=' {
		l.next()
		return token{
			tokenType: tok_gteq,
			start:     start,
			end:       l.nextIdx,
			value:     string(l.input[start:l.nextIdx]),
			lineCol:   l.lineColString(start),
		}
	}
	return token{
		tokenType: tok_gt,
		start:     start,
		end:       l.nextIdx,
		value:     string(l.input[start:l.nextIdx]),
		lineCol:   l.lineColString(start),
	}
}

func (l *lexer) handleEqual() token {
	if l.peek() == '=' {
		start := l.currentIdx
		l.next()
		return token{
			tokenType: tok_equal,
			value:     string(l.input[start:l.nextIdx]),
			start:     start,
			end:       l.nextIdx,
			lineCol:   l.lineColString(start),
		}
	}
	return token{
		tokenType: tok_assign,
		value:     string(l.currentChar),
		start:     l.currentIdx,
		end:       l.nextIdx,
		lineCol:   l.lineColString(l.currentIdx),
	}
}

func (l *lexer) handleDot() token {
	if isDigit(l.peek()) {
		return l.readNumber()
	}
	return token{
		tokenType: tok_dot,
		value:     string(l.currentChar),
		start:     l.currentIdx,
		end:       l.nextIdx,
		lineCol:   l.lineColString(l.currentIdx),
	}
}

func (l *lexer) handleExclamation() token {
	if l.peek() == '=' {
		start := l.currentIdx
		l.next()
		return token{
			tokenType: tok_not_equal,
			value:     string(l.input[start:l.nextIdx]),
			start:     start,
			end:       l.nextIdx,
			lineCol:   l.lineColString(start),
		}
	}
	return token{
		tokenType: tok_exclamation,
		value:     string(l.input[l.currentIdx:l.nextIdx]),
		start:     l.currentIdx,
		end:       l.nextIdx,
		lineCol:   l.lineColString(l.currentIdx),
	}
}

func (l *lexer) handleColon() token {
	start := l.currentIdx
	tok := token{
		tokenType: tok_colon,
		start:     start,
		lineCol:   l.lineColString(start),
	}
	if l.peek() == ':' {
		tok.tokenType = tok_double_colon
		l.next()
	}
	tok.value = string(l.input[start:l.nextIdx])
	tok.end = l.nextIdx
	return tok
}

func (l *lexer) handleDoubleQuote() token {
	start := l.currentIdx
	tok := token{
		tokenType: tok_string,
		start:     start,
		lineCol:   l.lineColString(start),
	}
	str := []rune{}
	for l.currentChar != '"' {
		if l.currentChar == nullchar || l.currentChar == '\n' {
			tok.tokenType = tok_illegal
			tok.end = l.currentIdx
			tok.value = "string literal not terminated"
			return tok
		}
		toAdd := l.currentChar
		if l.currentChar == '\\' {
			toAdd = l.handleEscapeString('"')
			if toAdd == nullchar {
				tok.tokenType = tok_illegal
				tok.value = "invalid escape sequence"
				l.next()
				tok.end = l.nextIdx
				return tok
			}
		}
		l.next()
		str = append(str, toAdd)
	}
	tok.value = string(str)
	tok.end = l.nextIdx
	// l.next()
	return tok
}

func (l *lexer) handleSingleQuote() token {
	start := l.currentIdx
	str := []rune{}
	tok := token{
		tokenType: tok_template_string,
		start:     start,
		lineCol:   l.lineColString(start),
	}
	for l.currentChar != '\'' && l.currentChar != nullchar {
		toAdd := l.currentChar

		if l.currentChar == '\\' {

			toAdd = l.handleEscapeString('\'')
			if toAdd == nullchar {
				tok.tokenType = tok_illegal
				tok.value = "invalid escape sequence"
				l.next()
				tok.end = l.nextIdx
				return tok
			}
		}

		if l.currentChar == '$' && l.peek() == '{' {
			l.newEnclosedContext(lex_ctx_single_quote, tok_lcurly, tok_rcurly, 1)
			tok.end = l.currentIdx
			tok.value = string(str)
			return tok
		}
		l.next()
		str = append(str, toAdd)

	}
	if l.currentChar == nullchar {
		tok.tokenType = tok_illegal
		tok.value = "string literal not terminated"
		tok.end = l.nextIdx
		return tok
	}
	tok.value = string(str)
	tok.end = l.nextIdx
	l.next()
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
	default:
		ret = nullchar
	}
	l.next()
	return ret
}

func (l *lexer) handleDollarSign() token {
	if l.peek() != '{' {
		return token{
			tokenType: tok_illegal,
			value:     "invalid template expression syntax",
			start:     l.currentIdx,
			end:       l.nextIdx,
			lineCol:   l.lineColString(l.currentIdx),
		}
	}
	tok := token{tokenType: tok_template_start, start: l.currentIdx, value: "${", lineCol: l.lineColString(l.currentIdx)}
	l.next()
	tok.end = l.nextIdx
	return tok
}

func (l *lexer) handleRCurly() token {
	l.maybeIncrDecrContext(tok_rcurly)
	return token{tokenType: tok_rcurly, value: "}", start: l.currentIdx, end: l.nextIdx, lineCol: l.lineColString(l.currentIdx)}
}

func (l *lexer) handleLCurly() token {
	l.maybeIncrDecrContext(tok_lcurly)
	return token{tokenType: tok_lcurly, value: "{", start: l.currentIdx, end: l.nextIdx, lineCol: l.lineColString(l.currentIdx)}
}

// formatting helprs

func (l *lexer) stringFromToken(t token) string {
	if l.isEnd && t.tokenType == tok_eof {
		return ""
	}
	return string(l.input[t.start:t.end])
}

func (l *lexer) lineColString(targetIdx int) string {
	return lineColString(lineAndCol(l.input, targetIdx))
}

// reader helpers

func (l *lexer) readIdentifier() token {
	start := l.currentIdx
	if l.currentChar == '@' {
		l.next()
	}
	if !isLetter(l.currentChar) {
		return token{
			tokenType: tok_illegal,
			value:     "identifiers that start with @ must be @in or @out",
			start:     start,
			end:       l.currentIdx,
			lineCol:   l.lineColString(start),
		}
	}
	for isDigit(l.currentChar) || isLetter(l.currentChar) || isLineChar(l.currentChar) {
		l.next()
	}
	endIdx := l.currentIdx
	if l.isEnd {
		endIdx = l.nextIdx
	}
	val := string(l.input[start:endIdx])
	if strings.HasPrefix(val, "@") && !slices.Contains([]string{"@in", "@out"}, val) {
		return token{
			tokenType: tok_illegal,
			value:     "identifiers that start with @ must be @in or @out",
			start:     start,
			end:       endIdx,
			lineCol:   l.lineColString(start),
		}
	}
	return token{
		tokenType: lookupTokenKeyword(val),
		value:     val,
		start:     start,
		end:       endIdx,
		lineCol:   l.lineColString(start),
	}
}

func (l *lexer) readNumber() token {
	tok := token{tokenType: tok_int}
	start := l.currentIdx
	encounteredDot := false
	for l.currentChar == '.' || isDigit(l.currentChar) {
		if l.currentChar == '.' {
			if !isDigit(l.peek()) || encounteredDot {
				break
			}
			tok.tokenType = tok_float
			encounteredDot = true
		}
		l.next()
	}
	tok.start = start
	tok.end = l.currentIdx
	if l.isEnd {
		tok.end = l.nextIdx
	}
	tok.lineCol = l.lineColString(start)
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

// context helpers

type lexContextType string

const (
	lex_ctx_single_quote lexContextType = "SINGLE_QUOTE"
	lex_ctx_default      lexContextType = "DEFAULT"
)

func lexerContextTypeToHandlerFn(l *lexer, t lexContextType) func() token {
	switch t {
	case lex_ctx_single_quote:
		return l.handleSingleQuote
	default:
		return l.tokenize
	}
}

type lexContext struct {
	outer            *lexContext // nil means we're at the top level
	contextType      lexContextType
	incrTriggerToken tokenType
	decrTriggerToken tokenType
	depthCounter     int
}

var defaultLexContext *lexContext = &lexContext{
	outer:       nil,
	contextType: lex_ctx_default,
}

func (l *lexer) newEnclosedContext(t lexContextType, incr tokenType, decr tokenType, initCounter int) {
	new := &lexContext{
		outer:            l.context,
		contextType:      t,
		incrTriggerToken: incr,
		decrTriggerToken: decr,
		depthCounter:     initCounter,
	}
	l.context = new
}
func (l *lexer) maybeIncrDecrContext(tt tokenType) {
	if tt == l.context.decrTriggerToken {
		l.context.depthCounter = max(l.context.depthCounter-1, 0)
	}
	if tt == l.context.incrTriggerToken {
		l.context.depthCounter++
	}
}
