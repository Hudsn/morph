package morph

import "strings"

type token struct {
	tokenType tokenType
	value     string

	start   int
	end     int
	lineCol string
}

type tokenType string

const (
	tok_ident tokenType = "IDENT"

	tok_int             tokenType = "INT"
	tok_float           tokenType = "FLOAT"
	tok_string          tokenType = "STRING"
	tok_template_string tokenType = "TEMPLATE_STRING"

	// separators
	tok_arrow        tokenType = "~>"
	tok_assign       tokenType = "="
	tok_dot          tokenType = "."
	tok_colon        tokenType = ":"
	tok_double_colon tokenType = "::"
	tok_comma        tokenType = ","

	// containers
	tok_template_start tokenType = "${"
	tok_lparen         tokenType = "("
	tok_rparen         tokenType = ")"
	tok_lcurly         tokenType = "{"
	tok_rcurly         tokenType = "}"
	tok_lsquare        tokenType = "["
	tok_rsquare        tokenType = "]"

	// operations
	tok_exclamation tokenType = "!"
	tok_plus        tokenType = "+"
	tok_minus       tokenType = "-"
	tok_asterisk    tokenType = "*"
	tok_slash       tokenType = "/"
	tok_mod         tokenType = "%"
	tok_binary_and  tokenType = "&&"
	tok_binary_or   tokenType = "||"
	tok_pipe        tokenType = "|"

	// (in)equality checks
	tok_equal     tokenType = "=="
	tok_not_equal tokenType = "!="
	tok_lt        tokenType = "<"
	tok_lteq      tokenType = "<="
	tok_gt        tokenType = ">"
	tok_gteq      tokenType = ">="

	//keywords
	tok_if    tokenType = "IF"
	tok_when  tokenType = "WHEN"
	tok_set   tokenType = "SET"
	tok_true  tokenType = "TRUE"
	tok_false tokenType = "FALSE"
	tok_null  tokenType = "NULL"

	tok_eof     tokenType = "EOF"
	tok_illegal tokenType = "ILLEGAL"
)

var keywordMap = map[string]tokenType{
	"if":    tok_if,
	"set":   tok_set,
	"true":  tok_true,
	"false": tok_false,
	"null":  tok_null,
}

func lookupTokenKeyword(ident string) tokenType {
	ident = strings.ToLower(ident)
	if ret, ok := keywordMap[ident]; ok {
		return ret
	}
	return tok_ident
}
