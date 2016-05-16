package ncparser

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"unicode"
)

type TokenType int

const (
	ILLEGAL TokenType = iota
	EOF
	BRACE_OPEN
	BRACE_CLOSE
	SEMICOLON
	WORD
	COMMENT
)

var (
	tokenTypeName = map[TokenType]string{
		ILLEGAL:     "ILLEGAL",
		EOF:         "EOF",
		BRACE_OPEN:  "BRACE_OPEN",
		BRACE_CLOSE: "BRACE_CLOSE",
		SEMICOLON:   "SEMICOLON",
		WORD:        "WORD",
		COMMENT:     "COMMENT",
	}
)

type Token struct {
	Type    TokenType
	Literal string
	Error   error
}

func (t TokenType) String() string {
	return tokenTypeName[t]
}

func (t Token) String() string {
	switch t.Type {
	case EOF, BRACE_OPEN, BRACE_CLOSE, SEMICOLON:
		return t.Type.String()
	case ILLEGAL:
		return fmt.Sprintf("%s[%s]", t.Type, t.Error.Error())
	}
	return fmt.Sprintf("%s[%s]", t.Type, t.Literal)
}

func newErrorToken(err error) Token {
	return Token{Type: ILLEGAL, Error: err}
}

func newWordToken(word string) Token {
	return Token{Type: WORD, Literal: word}
}

func newCommentToken(comment string) Token {
	return Token{Type: COMMENT, Literal: comment}
}

var (
	eofToken        = Token{Type: EOF}
	braceOpenToken  = Token{Type: BRACE_OPEN}
	braceCloseToken = Token{Type: BRACE_CLOSE}
	semicolonToken  = Token{Type: SEMICOLON}
)

type Scanner struct {
	r    *bufio.Reader
	Line int
}

func NewScanner(r io.Reader) *Scanner {
	return &Scanner{
		r: bufio.NewReader(r),
	}
}

func (s *Scanner) read() (rune, error) {
	r, _, err := s.r.ReadRune()
	if r == '\n' {
		s.Line++
	}
	return r, err
}

func (s *Scanner) unread() {
	s.r.UnreadRune()
}

func (s *Scanner) Scan() Token {
	if err := s.skipWhitespace(); err != nil {
		return newErrorToken(err)
	}
	r, err := s.read()
	if err == io.EOF {
		return eofToken
	} else if err != nil {
		return newErrorToken(err)
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

func (s *Scanner) scanSingleQuoted() Token {
	var buf bytes.Buffer
	quoted := false
ForLoop:
	for {
		r, err := s.read()
		if err == io.EOF {
			return newErrorToken(fmt.Errorf("missing terminating \\'\\ character at line %d", s.Line))
		} else if err != nil {
			return newErrorToken(err)
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
			default:
				return newErrorToken(fmt.Errorf("invalid quoted character: '\\%c'", r))
			}
			quoted = false
			continue
		}
		switch r {
		case '\n':
			return newErrorToken(fmt.Errorf("missing terminating \\'\\ character at line %d", s.Line))
		case '\\':
			quoted = true
		case '\'':
			break ForLoop
		default:
			buf.WriteRune(r)
		}
	}
	return newWordToken(buf.String())
}

func (s *Scanner) scanDoubleQuoted() Token {
	var buf bytes.Buffer
	quoted := false
ForLoop:
	for {
		r, err := s.read()
		if err == io.EOF {
			return newErrorToken(fmt.Errorf("missing terminating \\\"\\ character at line %d", s.Line))
		} else if err != nil {
			return newErrorToken(err)
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
			default:
				return newErrorToken(fmt.Errorf("invalid quoted character: '\\%c'", r))
			}
			quoted = false
			continue
		}
		switch r {
		case '\n':
			return newErrorToken(fmt.Errorf("missing terminating \\\"\\ character at line %d", s.Line))
		case '\\':
			quoted = true
		case '"':
			break ForLoop
		default:
			buf.WriteRune(r)
		}
	}
	return newWordToken(buf.String())
}

func (s *Scanner) skipWhitespace() error {
	for {
		r, err := s.read()
		if err == io.EOF {
			return nil
		} else if err != nil {
			return err
		} else if !unicode.IsSpace(r) {
			s.unread()
			break
		}
	}
	return nil
}

func (s *Scanner) scanComment() Token {
	var buf bytes.Buffer
	for {
		r, err := s.read()
		if err == io.EOF || r == '\n' {
			break
		}
		if err != nil {
			return newErrorToken(err)
		}
		buf.WriteRune(r)
	}
	return newCommentToken(buf.String())
}

func (s *Scanner) scanWord() Token {
	var buf bytes.Buffer
	for {
		r, err := s.read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return newErrorToken(err)
		}
		if unicode.IsSpace(r) {
			break
		}
		if r == ';' {
			s.unread()
			break
		}
		buf.WriteRune(r)
	}
	return newWordToken(buf.String())
}
