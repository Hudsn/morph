package parser

import "strings"

type token string

const (
	IDENT token = "IDENT"

	INT token = "INT"

	DOT token = "."

	EQUAL token = "="

	//keywords
	WHEN token = "WHEN"
)

var keywordMap = map[string]token{
	"when": WHEN,
}

func lookupTokenKeyword(ident string) token {
	ident = strings.ToLower(ident)
	if ret, ok := keywordMap[ident]; ok {
		return ret
	}
	return IDENT
}
