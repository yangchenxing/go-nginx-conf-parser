package ncparser

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"unicode"
)

type tokenType int

const (
	eof tokenType = iota
	braceOpen
	braceClose
	semicolon
	word
	comment
)

var (
	tokenTypeName = map[tokenType]string{
		eof:        "EOF",
		braceOpen:  "BRACE_OPEN",
		braceClose: "BRACE_CLOSE",
		semicolon:  "SEMICOLON",
		word:       "WORD",
		comment:    "COMMENT",
	}
)

func (t tokenType) String() string {
	return tokenTypeName[t]
}

type token struct {
	typ tokenType
	lit string
}

var (
	eofToken        = token{typ: eof}
	braceOpenToken  = token{typ: braceOpen, lit: "{"}
	braceCloseToken = token{typ: braceClose, lit: "}"}
	semicolonToken  = token{typ: semicolon, lit: ";"}
)

type scanner struct {
	r    *bufio.Reader
	line int
}

func newScanner(content []byte) *scanner {
	return &scanner{
		r: bufio.NewReader(bytes.NewBuffer(content)),
	}
}

func (s *scanner) read() (rune, error) {
	r, _, err := s.r.ReadRune()
	if r == '\n' {
		s.line++
	}
	return r, err
}

func (s *scanner) unread() {
	s.r.UnreadRune()
}

func (s *scanner) scan() token {
	s.skipWhitespace()
	r, err := s.read()
	if err == io.EOF {
		return eofToken
	}
	switch r {
	case '\'':
		return s.scanSingleQuoted()
	case '"':
		return s.scanDoubleQuoted()
	case '{':
		return braceOpenToken
	case '}':
		return braceCloseToken
	case ';':
		return semicolonToken
	case '#':
		return s.scanComment()
	}
	s.unread()
	return s.scanWord()
}

func (s *scanner) scanSingleQuoted() token {
	var buf bytes.Buffer
	quoted := false
ForLoop:
	for {
		r, err := s.read()
		if err == io.EOF {
			panic(fmt.Errorf("missing terminating \\'\\ character at line %d", s.line))
		}
		if quoted {
			switch r {
			case 'n':
				buf.WriteRune('\n')
			case 'r':
				buf.WriteRune('\r')
			case 't':
				buf.WriteRune('\t')
			case '"':
				buf.WriteRune('"')
			case '\'':
				buf.WriteRune('\'')
			case '\\':
				buf.WriteRune('\\')
			default:
				panic(fmt.Errorf("invalid quoted character: '\\%c'", r))
			}
			quoted = false
			continue
		}
		switch r {
		case '\n':
			panic(fmt.Errorf("missing terminating \\'\\ character at line %d", s.line))
		case '\\':
			quoted = true
		case '\'':
			break ForLoop
		default:
			buf.WriteRune(r)
		}
	}
	return token{typ: word, lit: buf.String()}
}

func (s *scanner) scanDoubleQuoted() token {
	var buf bytes.Buffer
	quoted := false
ForLoop:
	for {
		r, err := s.read()
		if err == io.EOF {
			panic(fmt.Errorf("missing terminating \\\"\\ character at line %d", s.line))
		}
		if quoted {
			switch r {
			case 'n':
				buf.WriteRune('\n')
			case 'r':
				buf.WriteRune('\r')
			case 't':
				buf.WriteRune('\t')
			case '"':
				buf.WriteRune('"')
			case '\'':
				buf.WriteRune('\'')
			case '\\':
				buf.WriteRune('\\')
			default:
				panic(fmt.Errorf("invalid quoted character: '\\%c'", r))
			}
			quoted = false
			continue
		}
		switch r {
		case '\n':
			panic(fmt.Errorf("missing terminating \\\"\\ character at line %d", s.line))
		case '\\':
			quoted = true
		case '"':
			break ForLoop
		default:
			buf.WriteRune(r)
		}
	}
	return token{typ: word, lit: buf.String()}
}

func (s *scanner) skipWhitespace() {
	for r, err := s.read(); err != io.EOF; r, err = s.read() {
		if !unicode.IsSpace(r) {
			s.unread()
			break
		}
	}
}

func (s *scanner) scanComment() token {
	var buf bytes.Buffer
	for {
		r, err := s.read()
		if err == io.EOF || r == '\n' {
			break
		}
		buf.WriteRune(r)
	}
	return token{typ: comment, lit: buf.String()}
}

func (s *scanner) scanWord() token {
	var buf bytes.Buffer
	for {
		r, err := s.read()
		if err == io.EOF {
			break
		}
		if unicode.IsSpace(r) {
			break
		}
		if r == '{' {
			s.unread()
			break
		}
		if r == ';' {
			s.unread()
			break
		}
		buf.WriteRune(r)
	}
	return token{typ: word, lit: buf.String()}
}
