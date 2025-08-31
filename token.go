package morph

import "fmt"

type token struct {
	tokenType tokenType
	value     string

	start int
	end   int
}

type tokenType string

const (
	tok_ident tokenType = "IDENT"

	tok_int             tokenType = "INT"
	tok_float           tokenType = "FLOAT"
	tok_string          tokenType = "STRING"
	tok_template_string tokenType = "TEMPLATE_STRING"

	// separators
	tok_assign       tokenType = "="
	TOK_DOT          tokenType = "."
	TOK_COLON        tokenType = ":"
	TOK_DOUBLE_COLON tokenType = "::"

	// containers
	TOK_TEMPLATE_START tokenType = "${"
	TOK_LPAREN         tokenType = "("
	TOK_RPAREN         tokenType = ")"
	TOK_LCURLY         tokenType = "{"
	TOK_RCURLY         tokenType = "}"
	TOK_LSQUARE        tokenType = "["
	TOK_RSQUARE        tokenType = "]"

	// operations
	TOK_EXCLAMATION tokenType = "!"
	TOK_PLUS        tokenType = "+"
	TOK_MINUS       tokenType = "-"
	TOK_ASTERISK    tokenType = "*"
	TOK_SLASH       tokenType = "/"
	TOK_MOD         tokenType = "%"
	TOK_PIPE        tokenType = "|"

	// (in)equality checks
	TOK_EQUAL     tokenType = "=="
	TOK_NOT_EQUAL tokenType = "!="
	TOK_LT        tokenType = "<"
	TOK_LTEQ      tokenType = "<="
	TOK_GT        tokenType = ">"
	TOK_GTEQ      tokenType = ">="

	//keywords
	TOK_WHEN  tokenType = "WHEN"
	TOK_SET   tokenType = "SET"
	TOK_TRUE  tokenType = "TRUE"
	TOK_FALSE tokenType = "FALSE"

	tok_eof     tokenType = "EOF"
	TOK_ILLEGAL tokenType = "ILLEGAL"
)

var keywordMap = map[string]tokenType{
	"when":  TOK_WHEN,
	"WHEN":  TOK_WHEN,
	"set":   TOK_SET,
	"SET":   TOK_SET,
	"true":  TOK_TRUE,
	"false": TOK_FALSE,
}

func lookupTokenKeyword(ident string) tokenType {
	if ret, ok := keywordMap[ident]; ok {
		return ret
	}
	return tok_ident
}

func lineColString(line int, col int) string {
	return fmt.Sprintf("%d:%d", line, col)
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
