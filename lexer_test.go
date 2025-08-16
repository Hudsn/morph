package parser

import "testing"

func TestLexEquals(t *testing.T) {
	input := "="
	tests := []testCase{
		{
			tokenType:  EQUAL,
			value:      "=",
			start:      0,
			end:        1,
			rangeValue: "=",
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
