package parser

import "strings"

type token struct {
	tokenType tokenType
	value     string

	start int
	end   int
}

type tokenType string

const (
	TOK_IDENT tokenType = "IDENT"

	TOK_INT   tokenType = "INT"
	TOK_FLOAT tokenType = "FLOAT"

	TOK_DOT tokenType = "."

	TOK_EQUAL       tokenType = "="
	TOK_MINUS       tokenType = "-"
	TOK_EXCLAMATION tokenType = "!"

	//keywords
	TOK_WHEN tokenType = "WHEN"

	TOK_EOF     tokenType = "EOF"
	TOK_ILLEGAL tokenType = "ILLEGAL"
)

var keywordMap = map[string]tokenType{
	"when": TOK_WHEN,
}

func lookupTokenKeyword(ident string) tokenType {
	ident = strings.ToLower(ident)
	if ret, ok := keywordMap[ident]; ok {
		return ret
	}
	return TOK_IDENT
}

func lineAndCol(input []rune, targetIdx int) (int, int) {
	line := 1
	col := 1
	for _, r := range input[:targetIdx] {
		switch r {
		case '\n': // reset if newline
			line++
			col = 1
		default:
			col++
		}
	}
	return line, col
}
