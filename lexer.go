package parser

type lexer struct {
	input       []rune
	currentChar rune
	currentIdx  int
	nextIdx     int
}

const NULLCHAR = rune(0)

func newLexer(input []rune) *lexer {
	l := &lexer{
		input:      input,
		currentIdx: 0,
		nextIdx:    0,
	}
	l.scanNext()
	return l
}

func (l *lexer) scanNext() {
	if l.nextIdx >= len(l.input) {
		l.currentChar = NULLCHAR
	} else {
		l.currentChar = l.input[l.nextIdx]
	}
	l.currentIdx = l.nextIdx
	l.nextIdx++
}

func (l *lexer) scanToken() token {
	var tok token

	l.handleWhiteSpace()

	// ...

	return tok
}

func (l *lexer) handleWhiteSpace() {

}
