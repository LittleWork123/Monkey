package lexer

import (
	"interpreter/token2"
	"testing"
)

// test next token whether read correctly
func TestNextToken(t *testing.T) {
	input := `
	let five = 5;
	let ten = 10;
	let add = fn(x,y){
		x + y;
	};
	
	let result = add(five,ten);
	!-/*5;
	 <  >   ;
	if return true
	else false
	==
 	!=
	"footbar"
`
	tests := []struct {
		expectedType    token2.TokenType
		expectedLiteral string
	}{
		// let five = 5;
		{token2.LET, "let"},
		{token2.IDENT, "five"},
		{token2.ASSIGN, "="},
		{token2.INT, "5"},
		{token2.SEMICOLON, ";"},

		// let ten = 10;
		{token2.LET, "let"},
		{token2.IDENT, "ten"},
		{token2.ASSIGN, "="},
		{token2.INT, "10"},
		{token2.SEMICOLON, ";"},

		// let add = fn(x,y){x+y;};
		{token2.LET, "let"},
		{token2.IDENT, "add"},
		{token2.ASSIGN, "="},
		{token2.FUNCTION, "fn"},
		{token2.LPAREN, "("},
		{token2.IDENT, "x"},
		{token2.COMMA, ","},
		{token2.IDENT, "y"},
		{token2.RPAREN, ")"},
		{token2.LBRACE, "{"},
		{token2.IDENT, "x"},
		{token2.PLUS, "+"},
		{token2.IDENT, "y"},
		{token2.SEMICOLON, ";"},
		{token2.RBRACE, "}"},
		{token2.SEMICOLON, ";"},

		//let result = add(five,ten);
		{token2.LET, "let"},
		{token2.IDENT, "result"},
		{token2.ASSIGN, "="},
		{token2.IDENT, "add"},
		{token2.LPAREN, "("},
		{token2.IDENT, "five"},
		{token2.COMMA, ","},
		{token2.IDENT, "ten"},
		{token2.RPAREN, ")"},
		{token2.SEMICOLON, ";"},

		{token2.BANG, "!"},
		{token2.MINUS, "-"},
		{token2.SLASH, "/"},
		{token2.ASTERISK, "*"},
		{token2.INT, "5"},
		{token2.SEMICOLON, ";"},

		{token2.LT, "<"},
		{token2.GT, ">"},
		{token2.SEMICOLON, ";"},

		// 	if return true
		//	else false
		{token2.IF, "if"},
		{token2.RETURN, "return"},
		{token2.TRUE, "true"},
		{token2.ELSE, "else"},
		{token2.FALSE, "false"},

		{token2.EQ, "=="},
		{token2.NOT_EQ, "!="},
		{token2.STRING, "footbar"},
		{token2.EOF, "\x00"}, // change here

	}

	l := New(input)
	for i, tt := range tests {
		token := l.NextToken()
		if token.Type != tt.expectedType {
			t.Fatalf("tests[%d] - tokentype wrong. expected=%q, got=%q",
				i, tt.expectedType, token.Type)
		}

		if token.Literal != tt.expectedLiteral {
			t.Fatalf("tests[%d] - literal wrong. expected=%q, got=%q",
				i, tt.expectedLiteral, token.Literal)
		}
	}
}
