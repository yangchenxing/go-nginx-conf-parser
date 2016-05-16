package ncparser

import (
	"bytes"
	"testing"
)

func TestScanner(t *testing.T) {
	buf := bytes.NewBuffer([]byte(`
#COMMENT
# DOUBLE #COMMENT
WORD1 WORD2;
WORD3 {
    WORD4 'SQ1' "DQ\t1";
}`))
	scanner := NewScanner(buf)
	expectedTokens := []Token{
		{Type: COMMENT, Literal: "COMMENT"},
		{Type: COMMENT, Literal: " DOUBLE #COMMENT"},
		{Type: WORD, Literal: "WORD1"},
		{Type: WORD, Literal: "WORD2"},
		semicolonToken,
		{Type: WORD, Literal: "WORD3"},
		braceOpenToken,
		{Type: WORD, Literal: "WORD4"},
		{Type: WORD, Literal: "SQ1"},
		{Type: WORD, Literal: "DQ\t1"},
		semicolonToken,
		braceCloseToken,
	}
	for i, expectedToken := range expectedTokens {
		token := scanner.Scan()
		if token != expectedToken {
			t.Errorf("unexpected token: i=%d, expected=%q, actual=%q", i, expectedToken, token)
			t.FailNow()
		}
	}
	if token := scanner.Scan(); token.Type != EOF {
		t.Error("unexpected token: expected=%q, actual=%q", eofToken, token)
	}
}
