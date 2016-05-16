package ncparser

import (
	"bytes"
	"testing"
)

func TestParser(t *testing.T) {
	buf := bytes.NewBuffer([]byte(`
#COMMENT
# DOUBLE #COMMENT
WORD1 WORD2;
WORD3 {
    WORD4 'SQ1' "DQ\t1";
}`))
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
	var parser Parser
	block, err := parser.Parse(buf)
	if err != nil {
		t.Error("parse fail:", err.Error())
		t.FailNow()
	}
	if !equalBlock(block, expected) {
		t.Error("unequal block: expected=%v, actual=%v", expected, block)
		t.FailNow()
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
