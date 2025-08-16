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
	IDENT tokenType = "IDENT"

	INT   tokenType = "INT"
	FLOAT tokenType = "FLOAT"

	DOT tokenType = "."

	EQUAL tokenType = "="

	//keywords
	WHEN tokenType = "WHEN"

	EOF     tokenType = "EOF"
	ILLEGAL tokenType = "ILLEGAL"
)

var keywordMap = map[string]tokenType{
	"when": WHEN,
}

func lookupTokenKeyword(ident string) tokenType {
	ident = strings.ToLower(ident)
	if ret, ok := keywordMap[ident]; ok {
		return ret
	}
	return IDENT
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
