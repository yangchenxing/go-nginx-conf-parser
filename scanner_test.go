package ncparser

import (
	"testing"
)

func TestScanner(t *testing.T) {
	content := []byte(`
#COMMENT
# DOUBLE #COMMENT
WORD1 WORD2;
WORD3 {
    WORD4 'SQ\t\r\n\'\"\\1' "DQ\t\r\n\'\"\\1";
}`)
	scanner := newScanner(content)
	expectedTokens := []token{
		{typ: comment, lit: "COMMENT"},
		{typ: comment, lit: " DOUBLE #COMMENT"},
		{typ: word, lit: "WORD1"},
		{typ: word, lit: "WORD2"},
		semicolonToken,
		{typ: word, lit: "WORD3"},
		braceOpenToken,
		{typ: word, lit: "WORD4"},
		{typ: word, lit: "SQ\t\r\n'\"\\1"},
		{typ: word, lit: "DQ\t\r\n'\"\\1"},
		semicolonToken,
		braceCloseToken,
	}
	for i, expectedToken := range expectedTokens {
		token := scanner.scan()
		if token != expectedToken {
			t.Errorf("unexpected token: i=%d, expected=%q, actual=%q\n", i, expectedToken, token)
			t.FailNow()
		}
	}
	if token := scanner.scan(); token.typ != eof {
		t.Errorf("unexpected token: expected=%q, actual=%q\n", eofToken, token)
	}
}

func TestScanUnterminatedSingleQuotedString1(t *testing.T) {
	content := []byte(`'WORD2`)
	scanner := newScanner(content)
	var tok token
	defer func() {
		if err := recover(); err == nil {
			t.Error("unexpected scan result:", tok)
		}
	}()
	tok = scanner.scan()
}

func TestScanUnterminatedSingleQuotedString2(t *testing.T) {
	content := []byte("'WORD2\n")
	scanner := newScanner(content)
	var tok token
	defer func() {
		if err := recover(); err == nil {
			t.Error("unexpected scan result:", tok)
		}
	}()
	tok = scanner.scan()
}

func TestScanInvalidQuotedCharInSingleQuotedString(t *testing.T) {
	content := []byte(`'WORD2\/'`)
	scanner := newScanner(content)
	var tok token
	defer func() {
		if err := recover(); err == nil {
			t.Error("unexpected scan result:", tok)
		}
	}()
	tok = scanner.scan()
}

func TestScanUnterminatedDoubleQuotedString1(t *testing.T) {
	content := []byte(`"WORD2`)
	scanner := newScanner(content)
	var tok token
	defer func() {
		if err := recover(); err == nil {
			t.Error("unexpected scan result:", tok)
		}
	}()
	tok = scanner.scan()
}

func TestScanUnterminatedDoubleQuotedString2(t *testing.T) {
	content := []byte("\"WORD2\n")
	scanner := newScanner(content)
	var tok token
	defer func() {
		if err := recover(); err == nil {
			t.Error("unexpected scan result:", tok)
		}
	}()
	tok = scanner.scan()
}

func TestScanInvalidQuotedCharInDoubleQuotedString(t *testing.T) {
	content := []byte(`"WORD2\/"`)
	scanner := newScanner(content)
	var tok token
	defer func() {
		if err := recover(); err == nil {
			t.Error("unexpected scan result:", tok)
		}
	}()
	tok = scanner.scan()
}

func TestScanLastWord(t *testing.T) {
	content := []byte("WORD")
	scanner := newScanner(content)
	defer func() {
		if err := recover(); err != nil {
			t.Error("scan fail:", err.(error).Error())
		}
	}()
	tok := scanner.scan()
	if tok.lit != "WORD" {
		t.Error("unexpected result:", tok)
	}
}
