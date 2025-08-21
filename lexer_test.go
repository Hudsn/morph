package parser

import "testing"

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
			tokenType:  TOK_ASSIGN,
			value:      "=",
			start:      0,
			end:        1,
			rangeValue: "=",
		},
		{
			tokenType:  TOK_EOF,
			value:      string(NULLCHAR),
			start:      1,
			end:        2,
			rangeValue: string(NULLCHAR),
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
			tokenType:  TOK_EOF,
			value:      string(NULLCHAR),
			start:      1,
			end:        2,
			rangeValue: string(NULLCHAR),
		},
	}
	checkLexTestCase(t, input, tests)
}

func TestLexNumber(t *testing.T) {
	input := "123 1.23 .123"
	tests := []testCase{
		{
			tokenType:  TOK_INT,
			value:      "123",
			start:      0,
			end:        3,
			rangeValue: "123",
		},
		{
			tokenType:  TOK_FLOAT,
			value:      "1.23",
			start:      4,
			end:        8,
			rangeValue: "1.23",
		},
		{
			tokenType:  TOK_FLOAT,
			value:      ".123",
			start:      9,
			end:        13,
			rangeValue: ".123",
		},
		{
			tokenType:  TOK_EOF,
			value:      string(NULLCHAR),
			start:      13,
			end:        14,
			rangeValue: string(NULLCHAR),
		},
	}
	checkLexTestCase(t, input, tests)
}

func TestLexIdent(t *testing.T) {
	input := "abc.def"
	tests := []testCase{
		{
			tokenType:  TOK_IDENT,
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
			tokenType:  TOK_IDENT,
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

func TestLexMinux(t *testing.T) {
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
			tokenType:  TOK_INT,
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
			tokenType:  TOK_INT,
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
		if tt.rangeValue != string(lexer.input[tok.start:tok.end]) {
			t.Errorf("case %d: wrong token literal derived from range: want=%s, got=%s", idx+1, tt.rangeValue, string(lexer.input[tok.start:tok.end]))
		}
	}
}
