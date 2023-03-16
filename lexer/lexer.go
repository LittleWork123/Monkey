package lexer

import (
	token2 "interpreter/token2"
)

type Lexer struct {
	input        string
	position     int  // the current position of input string(pointed to current char)
	readPosition int  // the next position of current position(pointed to the next char)
	ch           byte // the reading character
}

// To create a lexer
func New(input string) *Lexer {
	lexer := &Lexer{input: input}
	// start to read char
	lexer.readChar()
	return lexer
}

// To read char
func (l *Lexer) readChar() {
	if l.readPosition >= len(l.input) {
		l.ch = 0
	} else {
		l.ch = l.input[l.readPosition] // read next char
	}
	l.position = l.readPosition
	l.readPosition += 1
}

// Get next token2
func (l *Lexer) NextToken() token2.Token {
	var token token2.Token
	l.skipWhitespace()
	switch l.ch {
	case '=':
		if l.peekChar() == '=' {
			token.Type = token2.EQ
			token.Literal = "=="
			l.readChar()
			break
		} else {
			token = newToken(token2.ASSIGN, l.ch)
		}
		break
	case '+':
		token = newToken(token2.PLUS, l.ch)
		break
	case '-':
		token = newToken(token2.MINUS, l.ch)
		break
	case '*':
		token = newToken(token2.ASTERISK, l.ch)
		break
	case '/':
		token = newToken(token2.SLASH, l.ch)
		break
	case '!':
		if l.peekChar() == '=' {
			token.Type = token2.NOT_EQ
			token.Literal = "!="
			l.readChar()
			break
		} else {
			token = newToken(token2.BANG, l.ch)
		}
		break
	case '<':
		token = newToken(token2.LT, l.ch)
		break
	case '>':
		token = newToken(token2.GT, l.ch)
		break
	case ',':
		token = newToken(token2.COMMA, l.ch)
		break
	case ';':
		token = newToken(token2.SEMICOLON, l.ch)
		break
	case '(':
		token = newToken(token2.LPAREN, l.ch)
		break
	case ')':
		token = newToken(token2.RPAREN, l.ch)
		break
	case '{':
		token = newToken(token2.LBRACE, l.ch)
		break
	case '}':
		token = newToken(token2.RBRACE, l.ch)
		break
		// identify string type token
	case '"':
		token.Literal = l.readString()
		token.Type = token2.STRING
	case '[':
		token = newToken(token2.LBRACKET, l.ch)
		break
	case ']':
		token = newToken(token2.RBRACKET, l.ch)
		break
	// COLON
	case ':':
		token = newToken(token2.COLON, l.ch)
	default:
		if isLetter(l.ch) {
			// It has already called function readChar()
			token.Literal = l.readIdentifier()
			token.Type = token2.LookupIdent(token.Literal)
			return token
		} else if isDigit(l.ch) {
			token.Literal = l.readNumber()
			token.Type = token2.INT
			return token
		} else {
			token = newToken(token2.EOF, l.ch)
		}
	}
	// read next char
	l.readChar()
	return token
}

func newToken(tokenType token2.TokenType, ch byte) token2.Token {
	return token2.Token{
		Type:    tokenType,
		Literal: string(ch),
	}
}

// handle identifier
func (l *Lexer) readIdentifier() string {
	position := l.position
	for isLetter(l.ch) {
		l.readChar()
	}
	// the slice of string
	return l.input[position:l.position]
}

// handle the number
func (l *Lexer) readNumber() string {
	position := l.position
	for isDigit(l.ch) {
		l.readChar()
	}
	return l.input[position:l.position]
}

func isDigit(ch byte) bool {
	return '0' <= ch && ch <= '9'
}

// notion : if the identifier start with '_' is also correct
func isLetter(ch byte) bool {
	return 'a' <= ch && ch <= 'z' || 'A' <= ch && ch <= 'Z' || ch == '_'
}

// skip all white space include \t \n \r ' '
func (l *Lexer) skipWhitespace() {
	for l.ch == ' ' || l.ch == '\t' || l.ch == '\n' || l.ch == '\r' {
		l.readChar()
	}
}

// return peek char
func (l *Lexer) peekChar() byte {
	if l.readPosition >= len(l.input) {
		return 0
	} else {
		return l.input[l.readPosition]
	}
}

func (l *Lexer) readString() string {
	position := l.position + 1
	for {
		l.readChar()
		if l.ch == '"' || l.ch == 0 {
			break
		}
	}
	return l.input[position:l.position]
}
