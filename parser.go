package ncparser

import (
	"container/list"
	"fmt"
	"io"
)

type NginxConfigureBlock []NginxConfigureCommand

type NginxConfigureCommand struct {
	Words []string
	Block NginxConfigureBlock
}

type Parser struct {
	s *Scanner
}

var (
	emptyBlock   = NginxConfigureBlock(nil)
	emptyCommand = NginxConfigureCommand{}
)

func (p *Parser) Parse(r io.Reader) (NginxConfigureBlock, error) {
	p.s = NewScanner(r)
	cmds := list.New()
ForLoop:
	for {
		token := p.s.Scan()
		switch token.Type {
		case ILLEGAL:
			return nil, token.Error
		case EOF:
			break ForLoop
		case WORD:
			cmd, err := p.scanCommand(token.Literal)
			if err != nil {
				return nil, err
			}
			cmds.PushBack(cmd)
		case COMMENT:
			continue
		default:
			return nil, fmt.Errorf("unexpected global token %s at line %d", token.Type, p.s.Line)
		}
	}
	cfg := make([]NginxConfigureCommand, cmds.Len())
	for i, cmd := 0, cmds.Front(); cmd != nil; i, cmd = i+1, cmd.Next() {
		cfg[i] = cmd.Value.(NginxConfigureCommand)
	}
	return cfg, nil
}

func (p *Parser) scanCommand(startWord string) (NginxConfigureCommand, error) {
	words := list.New()
	if startWord != "" {
		words.PushBack(startWord)
	}
	var err error
	var block NginxConfigureBlock
ForLoop:
	for {
		token := p.s.Scan()
		switch token.Type {
		case ILLEGAL:
			return emptyCommand, token.Error
		case EOF:
			return emptyCommand, fmt.Errorf("missing terminating token at line %d", p.s.Line)
		case BRACE_OPEN:
			block, err = p.scanBlock()
			if err != nil {
				return emptyCommand, err
			}
			break ForLoop
		case SEMICOLON:
			break ForLoop
		case COMMENT:
			continue
		case WORD:
			words.PushBack(token.Literal)
		default:
			return emptyCommand, fmt.Errorf("unexpected command token %s at line %d", token.Type, p.s.Line)
		}
	}
	cmd := NginxConfigureCommand{
		Words: make([]string, words.Len()),
		Block: block,
	}
	for i, word := 0, words.Front(); word != nil; i, word = i+1, word.Next() {
		cmd.Words[i] = word.Value.(string)
	}
	return cmd, nil
}

func (p *Parser) scanBlock() (NginxConfigureBlock, error) {
	cmds := list.New()
ForLoop:
	for {
		token := p.s.Scan()
		switch token.Type {
		case ILLEGAL:
			return emptyBlock, token.Error
		case EOF:
			return emptyBlock, fmt.Errorf("missing terminating token at line %d", p.s.Line)
		case BRACE_CLOSE:
			break ForLoop
		case COMMENT:
			continue
		case WORD:
			cmd, err := p.scanCommand(token.Literal)
			if err != nil {
				return emptyBlock, err
			}
			cmds.PushBack(cmd)
		default:
			return emptyBlock, fmt.Errorf("unexpected block token %s at line %d", token.Type, p.s.Line)
		}
	}
	block := make([]NginxConfigureCommand, cmds.Len())
	for i, cmd := 0, cmds.Front(); cmd != nil; i, cmd = i+1, cmd.Next() {
		block[i] = cmd.Value.(NginxConfigureCommand)
	}
	return block, nil
}
