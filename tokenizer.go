package main

import (
	"errors"
	"unicode/utf8"
)

const (
	RuneBom   = 0xfeff
	RuneEof   = 0
	RuneError = 65533
)

type Pos struct {
	offset int
	line   int
	column int
}

type TokenKind struct {
	str string
}

type Token struct {
	Pos
	kind TokenKind
	text string
}

type Tokenizer struct {
	Pos
	data              string
	r                 rune
	w                 int
	currentLineOffset int
	insertSemicolon   bool
}

func NewTokenizer(data string) Tokenizer {
	t := Tokenizer{
		Pos:  Pos{line: 1},
		data: data,
	}
	t.Next()
	if t.r == RuneBom {
		t.Next()
	}
	return t
}

func (t *Tokenizer) Next() rune {
	if t.offset >= len(t.data) {
		return RuneEof
	} else {
		t.offset += t.w
		t.r, t.w = utf8.DecodeRuneInString(t.data[t.offset:])
		t.column = t.offset - t.currentLineOffset
		if t.offset >= len(t.data) {
			return RuneEof
		}
	}
	return t.r
}

func (t *Tokenizer) skipWhiteSpace(newLine bool) {
loop:
	for t.offset < len(t.data) {
		switch t.r {
		case ' ', '\t', '\v', '\f', '\r':
			t.Next()
		case '\n':
			if newLine {
				break loop
			}
			t.line++
			t.currentLineOffset = t.offset
			t.column = 1
			t.Next()
		default:
			switch t.r {
			case 0x2028, 0x2029, 0xFEFF:
				t.Next()
				continue loop
			}
			break loop
		}
	}
}

func isLetter(r rune) bool {
	return ('a' <= r && r <= 'z') || ('A' <= r && r <= 'Z')
}
func isDigit(r rune) bool {
	return '0' <= r && r <= '9'
}

func isAny(r rune, runes ...rune) bool {
	for _, v := range runes {
		if r == v {
			return true
		}
	}
	return false
}

func (t *Tokenizer) Token() (token Token, err error) {
	t.skipWhiteSpace(t.insertSemicolon)

	token.Pos = t.Pos
	token.kind = Invalid

	var currRune rune = t.r
	t.Next()
	switch {
	case currRune == RuneError:
		// TODO: actually make this work.
		token.kind = EOF

	case currRune == RuneEof:
		token.kind = EOF
		err = errors.New("EOF")

	case isAny(currRune, '\n', ';'):
		t.insertSemicolon = false
		token.text = "\n"
		token.kind = Semicolon
		t.line++
		t.currentLineOffset = t.offset
		t.column = 1
		return
	case isLetter(currRune):
		token.kind = Identifier
		for t.offset < len(t.data) {
			switch {
			case isLetter(t.r) || isDigit(t.r) || t.r == '_':
				t.Next()
				continue
			}
			break
		}
		for _, v := range []TokenKind{
			Foreign,
			Import,
			Fn,
			TypeDef,
			Package,
			Use,
		} {
			if t.data[token.offset:t.offset] == v.str {
				token.kind = v
				break
			}
		}
	case currRune == '/':
		if t.r == '/' {
			token.kind = Comment
			for t.offset < len(t.data) {
				if t.r == '\n' {
					break
				}
				t.Next()
			}
		}
	case currRune == '(':
		token.kind = LeftParen
	case currRune == ')':
		token.kind = RightParen
	case currRune == '{':
		token.kind = LeftBrace
	case currRune == '}':
		token.kind = RightBrace
	case currRune == ',':
		token.kind = Comma
	case currRune == ':':
		token.kind = Colon
	case currRune == '.':
		token.kind = Dot
	case currRune == '"':
		token.kind = String
		for t.offset < len(t.data) {
			switch {
			case t.r == '"':
				t.Next()
				break
			default:
				t.Next()
				continue
			}
			break
		}
	default:
		err = errors.New("invalid character")
	}
	text := t.data[token.offset:t.offset]
	switch token.kind {
	case String:
		text = text[1 : len(text)-1]
	case EOF, Semicolon:
		t.insertSemicolon = false
	case Identifier, LeftParen, RightParen:
		t.insertSemicolon = true
	default:
		t.insertSemicolon = false
	}
	token.text = text
	return
}

var (
	Invalid = TokenKind{"invalid"}
	EOF     = TokenKind{"EOF"}
	Comment = TokenKind{"comment"}

	Identifier = TokenKind{"identifier"}
	String     = TokenKind{"string"}

	// Keywords
	Package = TokenKind{"package"}
	Foreign = TokenKind{"foreign"}
	TypeDef = TokenKind{"type"}
	Import  = TokenKind{"import"}
	Use     = TokenKind{"use"}
	Fn      = TokenKind{"fn"}
	Struct  = TokenKind{"struct"}

	// Separators
	LeftParen  = TokenKind{"("}
	RightParen = TokenKind{")"}
	LeftBrace  = TokenKind{"{"}
	RightBrace = TokenKind{"}"}
	Comma      = TokenKind{","}
	Colon      = TokenKind{":"}
	Semicolon  = TokenKind{";"}
	Dot        = TokenKind{"."}
)
