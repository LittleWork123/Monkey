/*
	This package is to define token2 unit
*/
package token2

type TokenType string

type Token struct {
	Type    TokenType // token2 type
	Literal string    // literal notation
}

// all token2 type
const (
	STRING  = "STRING"  // string type
	ILLEGAL = "ILLEGAL" // unknown token2
	EOF     = "EOF"     // the end of file

	// identifier + literal
	IDENT = "IDENT" // add, foobar, x, y, ...
	INT   = "INT"   // 1,2,3,4,5,6

	// operator
	ASSIGN   = "="
	PLUS     = "+"
	MINUS    = "-"
	BANG     = "!"
	ASTERISK = "*"
	SLASH    = "/"

	LT     = "<"
	GT     = ">"
	EQ     = "=="
	NOT_EQ = "!="

	// separator
	COMMA     = ","
	SEMICOLON = ";"

	LPAREN = "("
	RPAREN = ")"
	LBRACE = "{"
	RBRACE = "}"
	// to support array and array[index]
	LBRACKET = "["
	RBRACKET = "]"
	COLON    = ":"
	// keyword
	FUNCTION = "FUNCTION"
	LET      = "LET"
	TRUE     = "TRUE"
	FALSE    = "FALSE"
	IF       = "IF"
	ELSE     = "ELSE"
	RETURN   = "RETURN"
)

var keywords = map[string]TokenType{
	"fn":     FUNCTION,
	"let":    LET,
	"true":   TRUE,
	"false":  FALSE,
	"else":   ELSE,
	"if":     IF,
	"return": RETURN,
}

// find function mapping in keyword
func LookupIdent(ident string) TokenType {
	if token, ok := keywords[ident]; ok {
		return token
	}
	return IDENT
}
