package morph

import (
	"testing"
)

func TestLexDoubleQuoteEscapeError(t *testing.T) {
	input := `"my\ncool\"string\v"`
	tests := []testCase{
		{
			tokenType:  TOK_ILLEGAL,
			start:      0,
			end:        20,
			value:      "invalid escape sequence",
			rangeValue: `"my\ncool\"string\v"`,
		},
	}
	checkLexTestCase(t, input, tests)
}

func TestLexDoubleQuoteNewlineError(t *testing.T) {
	input := `"my
	string`
	tests := []testCase{
		{
			tokenType:  TOK_ILLEGAL,
			start:      0,
			end:        3,
			value:      "string literal not terminated",
			rangeValue: `"my`,
		},
	}
	checkLexTestCase(t, input, tests)
}

func TestLexDoubleQuote(t *testing.T) {
	input := `"my string"`
	tests := []testCase{
		{
			tokenType:  tok_string,
			start:      0,
			end:        11,
			value:      "my string",
			rangeValue: `"my string"`,
		},
	}
	checkLexTestCase(t, input, tests)
}

func TestLexSingleQuoteInterp(t *testing.T) {
	input := "'mystring ${myvar} nest ${'nest string ${nest_var}!'}'"
	tests := []testCase{
		{
			tokenType:  tok_string,
			start:      0,
			end:        10,
			value:      "mystring ",
			rangeValue: "'mystring ",
		},
		{
			tokenType:  TOK_TEMPLATE_START,
			start:      10,
			end:        12,
			value:      "${",
			rangeValue: "${",
		},
		{
			tokenType:  tok_ident,
			start:      12,
			end:        17,
			value:      "myvar",
			rangeValue: "myvar",
		},
		{
			tokenType:  TOK_RCURLY,
			start:      17,
			end:        18,
			value:      "}",
			rangeValue: "}",
		},
		{
			tokenType:  tok_string,
			start:      18,
			end:        24,
			value:      " nest ",
			rangeValue: " nest ",
		},
		{
			tokenType:  TOK_TEMPLATE_START,
			start:      24,
			end:        26,
			value:      "${",
			rangeValue: "${",
		},
		{
			tokenType:  tok_string,
			start:      26,
			end:        39,
			value:      "nest string ",
			rangeValue: "'nest string ",
		},
		{
			tokenType:  TOK_TEMPLATE_START,
			start:      39,
			end:        41,
			value:      "${",
			rangeValue: "${",
		},
		{
			tokenType:  tok_ident,
			start:      41,
			end:        49,
			value:      "nest_var",
			rangeValue: "nest_var",
		},
		{
			tokenType:  TOK_RCURLY,
			start:      49,
			end:        50,
			value:      "}",
			rangeValue: "}",
		},
		{
			tokenType:  tok_string,
			start:      50,
			end:        52,
			value:      "!",
			rangeValue: "!'",
		},
		{
			tokenType:  TOK_RCURLY,
			start:      52,
			end:        53,
			value:      "}",
			rangeValue: "}",
		},
		{
			tokenType:  tok_string,
			start:      53,
			end:        54,
			value:      "",
			rangeValue: "'",
		},
		{
			tokenType:  tok_eof,
			start:      0,
			end:        0,
			value:      string(nullchar),
			rangeValue: "",
		},
	}
	checkLexTestCase(t, input, tests)
	l := newLexer([]rune(input))
	for range tests {
		l.tokenize()
	}
}
func TestLexSingleQuoteEscapeError(t *testing.T) {
	input := `'mystring\v'`
	tests := []testCase{
		{
			tokenType:  TOK_ILLEGAL,
			start:      0,
			end:        12,
			value:      "invalid escape sequence",
			rangeValue: "'mystring\\v'",
		},
	}
	checkLexTestCase(t, input, tests)
}
func TestLexSingleQuoteEscape(t *testing.T) {
	input := `'mystring\n\t'`
	tests := []testCase{
		{
			tokenType:  tok_string,
			start:      0,
			end:        14,
			value:      "mystring\n\t",
			rangeValue: "'mystring\\n\\t'",
		},
	}
	checkLexTestCase(t, input, tests)
}
func TestLexSingleQuoteBase(t *testing.T) {
	input := "'mystring' 'endstring'"
	tests := []testCase{
		{
			tokenType:  tok_string,
			start:      0,
			end:        10,
			value:      "mystring",
			rangeValue: "'mystring'",
		},
		{
			tokenType:  tok_string,
			start:      11,
			end:        22,
			value:      "endstring",
			rangeValue: "'endstring'",
		},
	}
	checkLexTestCase(t, input, tests)
}

func TestLexColon(t *testing.T) {
	input := ": :: :"
	tests := []testCase{
		{
			tokenType:  TOK_COLON,
			start:      0,
			end:        1,
			value:      ":",
			rangeValue: ":",
		},
		{
			tokenType:  TOK_DOUBLE_COLON,
			start:      2,
			end:        4,
			value:      "::",
			rangeValue: "::",
		},
		{
			tokenType:  TOK_COLON,
			start:      5,
			end:        6,
			value:      ":",
			rangeValue: ":",
		},
	}
	checkLexTestCase(t, input, tests)
}

func TestLexBoolean(t *testing.T) {
	input := "true false"
	tests := []testCase{
		{
			tokenType:  TOK_TRUE,
			value:      "true",
			start:      0,
			end:        4,
			rangeValue: "true",
		},
		{
			tokenType:  TOK_FALSE,
			value:      "false",
			start:      5,
			end:        10,
			rangeValue: "false",
		},
	}
	checkLexTestCase(t, input, tests)
}

func TestLexAssign(t *testing.T) {
	input := "="
	tests := []testCase{
		{
			tokenType:  tok_assign,
			value:      "=",
			start:      0,
			end:        1,
			rangeValue: "=",
		},
		{
			tokenType:  tok_eof,
			value:      string(nullchar),
			start:      0,
			end:        0,
			rangeValue: "",
		},
	}
	checkLexTestCase(t, input, tests)
}

func TestLexDot(t *testing.T) {
	input := "."
	tests := []testCase{
		{
			tokenType:  TOK_DOT,
			value:      ".",
			start:      0,
			end:        1,
			rangeValue: ".",
		},
		{
			tokenType:  tok_eof,
			value:      string(nullchar),
			start:      0,
			end:        0,
			rangeValue: "",
		},
	}
	checkLexTestCase(t, input, tests)
}

func TestLexNumber(t *testing.T) {
	input := "123 1.23 .123"
	tests := []testCase{
		{
			tokenType:  tok_int,
			value:      "123",
			start:      0,
			end:        3,
			rangeValue: "123",
		},
		{
			tokenType:  tok_float,
			value:      "1.23",
			start:      4,
			end:        8,
			rangeValue: "1.23",
		},
		{
			tokenType:  tok_float,
			value:      ".123",
			start:      9,
			end:        13,
			rangeValue: ".123",
		},
		{
			tokenType:  tok_eof,
			value:      string(nullchar),
			start:      0,
			end:        0,
			rangeValue: "",
		},
	}
	checkLexTestCase(t, input, tests)
}

func TestLexIdent(t *testing.T) {
	input := "abc.def"
	tests := []testCase{
		{
			tokenType:  tok_ident,
			value:      "abc",
			start:      0,
			end:        3,
			rangeValue: "abc",
		},
		{
			tokenType:  TOK_DOT,
			value:      ".",
			start:      3,
			end:        4,
			rangeValue: ".",
		},
		{
			tokenType:  tok_ident,
			value:      "def",
			start:      4,
			end:        7,
			rangeValue: "def",
		},
	}
	checkLexTestCase(t, input, tests)
}

type testCase struct {
	tokenType  tokenType
	value      string
	start      int
	end        int
	rangeValue string // the literal value captured by the start and end markers; for ex in a string "asdf", it would include the quotes as well even though the value is just asdf
}

func TestLexMinus(t *testing.T) {
	input := "-5"
	tests := []testCase{
		{
			tokenType:  TOK_MINUS,
			value:      "-",
			rangeValue: "-",
			start:      0,
			end:        1,
		},
		{
			tokenType:  tok_int,
			value:      "5",
			rangeValue: "5",
			start:      1,
			end:        2,
		},
	}
	checkLexTestCase(t, input, tests)
}

func TestLexExclamation(t *testing.T) {
	input := "!5"
	tests := []testCase{
		{
			tokenType:  TOK_EXCLAMATION,
			value:      "!",
			rangeValue: "!",
			start:      0,
			end:        1,
		},
		{
			tokenType:  tok_int,
			value:      "5",
			rangeValue: "5",
			start:      1,
			end:        2,
		},
	}
	checkLexTestCase(t, input, tests)
}

//helepr

func checkLexTestCase(t *testing.T, input string, cases []testCase) {
	lexer := newLexer([]rune(input))
	for idx, tt := range cases {
		tok := lexer.tokenize()
		if tok.tokenType != tt.tokenType {
			t.Errorf("case %d: wrong token type: want=%s, got=%s", idx+1, tt.tokenType, tok.tokenType)
		}
		if tok.value != tt.value {
			t.Errorf("case %d: wrong token value: want=%s, got=%s", idx+1, tt.value, tok.value)
		}
		if tok.start != tt.start {
			t.Errorf("case %d: wrong token start index: want=%d, got=%d", idx+1, tt.start, tok.start)
		}
		if tok.end != tt.end {
			t.Errorf("case %d: wrong token end index: want=%d, got=%d", idx+1, tt.end, tok.end)
		}
		if tt.rangeValue != lexer.stringFromToken(tok) {
			t.Errorf("case %d: wrong token literal derived from range: want=%s, got=%s", idx+1, tt.rangeValue, lexer.stringFromToken(tok))
		}
	}
}
