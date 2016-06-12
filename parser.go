package ncparser

import (
	"container/list"
	"fmt"
	"sync"
)

// NginxConfigureBlock represent a block in nginx configure file.
// The content of a nginx configure file should be a block.
type NginxConfigureBlock []NginxConfigureCommand

// NginxConfigureCommand represenct a command in nginx configure file.
type NginxConfigureCommand struct {
	// Words compose the command
	Words []string

	// Block follow the command
	Block NginxConfigureBlock
}

type parser struct {
	sync.Mutex
	*scanner
}

var (
	emptyBlock   = NginxConfigureBlock(nil)
	emptyCommand = NginxConfigureCommand{}
)

// Parse the content of nginx configure file into NginxConfigureBlock
func Parse(content []byte) (blk NginxConfigureBlock, err error) {
	var p parser
	return p.parse(content)
}

func (p *parser) parse(content []byte) (blk NginxConfigureBlock, err error) {
	p.Lock()
	defer p.Unlock()
	defer func() {
		if e := recover(); e != nil {
			err = e.(error)
		}
	}()
	p.scanner = newScanner(content)
	cmds := list.New()
ForLoop:
	for {
		token := p.scan()
		switch token.typ {
		case eof:
			break ForLoop
		case word:
			cmd, err := p.scanCommand(token.lit)
			if err != nil {
				return nil, err
			}
			cmds.PushBack(cmd)
		case comment:
			continue
		default:
			return nil, fmt.Errorf("unexpected global token %s at line %d", token.typ, p.line)
		}
	}
	cfg := make([]NginxConfigureCommand, cmds.Len())
	for i, cmd := 0, cmds.Front(); cmd != nil; i, cmd = i+1, cmd.Next() {
		cfg[i] = cmd.Value.(NginxConfigureCommand)
	}
	return cfg, nil
}

func (p *parser) scanCommand(startWord string) (NginxConfigureCommand, error) {
	words := list.New()
	if startWord != "" {
		words.PushBack(startWord)
	}
	var err error
	var block NginxConfigureBlock
ForLoop:
	for {
		token := p.scan()
		switch token.typ {
		case eof:
			return emptyCommand, fmt.Errorf("missing terminating token at line %d", p.line)
		case braceOpen:
			block, err = p.scanBlock()
			if err != nil {
				return emptyCommand, err
			}
			break ForLoop
		case semicolon:
			break ForLoop
		case comment:
			continue
		case word:
			words.PushBack(token.lit)
		default:
			return emptyCommand, fmt.Errorf("unexpected command token %s at line %d", token.typ, p.line)
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

func (p *parser) scanBlock() (NginxConfigureBlock, error) {
	cmds := list.New()
ForLoop:
	for {
		token := p.scan()
		switch token.typ {
		case eof:
			return emptyBlock, fmt.Errorf("missing terminating token at line %d", p.line)
		case braceClose:
			break ForLoop
		case comment:
			continue
		case word:
			cmd, err := p.scanCommand(token.lit)
			if err != nil {
				return emptyBlock, err
			}
			cmds.PushBack(cmd)
		default:
			return emptyBlock, fmt.Errorf("unexpected block token %s at line %d", token.typ, p.line)
		}
	}
	block := make([]NginxConfigureCommand, cmds.Len())
	for i, cmd := 0, cmds.Front(); cmd != nil; i, cmd = i+1, cmd.Next() {
		block[i] = cmd.Value.(NginxConfigureCommand)
	}
	return block, nil
}
