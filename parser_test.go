package ncparser

import (
	"testing"
)

func TestParser(t *testing.T) {
	content := []byte(`
#COMMENT
# DOUBLE #COMMENT
WORD1 WORD2;
WORD3 {
    WORD4 'SQ1' "DQ\t1" #COMMENT
    ;
    #COMMENT
}`)
	expected := NginxConfigureBlock([]NginxConfigureCommand{
		{
			Words: []string{"WORD1", "WORD2"},
		},
		{
			Words: []string{"WORD3"},
			Block: NginxConfigureBlock([]NginxConfigureCommand{
				{
					Words: []string{"WORD4", "SQ1", "DQ\t1"},
				},
			}),
		},
	})
	block, err := Parse(content)
	if err != nil {
		t.Error("parse fail:", err.Error())
		t.FailNow()
	}
	if !equalBlock(block, expected) {
		t.Errorf("unequal block: expected=%v, actual=%v\n", expected, block)
		t.FailNow()
	}
}

func TestParseUnterminatedCommand(t *testing.T) {
	content := []byte(`WORD  `)
	block, err := Parse(content)
	if err == nil {
		t.Error("unexpected parse result:", block)
	}
}

func TestParseUnterminatedBlock(t *testing.T) {
	content := []byte(`WORD { `)
	block, err := Parse(content)
	if err == nil {
		t.Error("unexpected parse result:", block)
	}
}

func TestParseInvalidTokenInCommand(t *testing.T) {
	content := []byte(`WORD };`)
	block, err := Parse(content)
	if err == nil {
		t.Error("unexpected parse result:", block)
	}
}

func TestParseInvalidTokenInBlock(t *testing.T) {
	content := []byte(`WORD { {`)
	block, err := Parse(content)
	if err == nil {
		t.Error("unexpected parse result:", block)
	}
}

func TestParseInvalidCommandInBlock(t *testing.T) {
	content := []byte(`WORD { WORD`)
	block, err := Parse(content)
	if err == nil {
		t.Error("unexpected parse result:", block)
	}
}

func TestParseInvalidFirstToken(t *testing.T) {
	content := []byte(`}`)
	block, err := Parse(content)
	if err == nil {
		t.Error("unexpected parse result:", block)
	}
}

func TestParseWithScannerFailure(t *testing.T) {
	content := []byte(`"WORD\|"`)
	block, err := Parse(content)
	if err == nil {
		t.Error("unexpected parse result:", block)
	}
}

func equalBlock(a, b NginxConfigureBlock) bool {
	if len(a) != len(b) {
		return false
	}
	for i, c := range a {
		d := b[i]
		if len(c.Words) != len(d.Words) {
			return false
		}
		for j, w := range c.Words {
			if w != d.Words[j] {
				return false
			}
		}
		if !equalBlock(c.Block, d.Block) {
			return false
		}
	}
	return true
}
