package ast

import (
	"interpreter/token2"
	"testing"
)

// input = `let myVar = anotherVar`
func TestString(i *testing.T) {
	program := &Program{
		Statements: []Statement{
			&LetStatement{
				Token: token2.Token{
					Type:    token2.LET,
					Literal: "let",
				},
				Name: &Identifier{
					Token: token2.Token{
						Type:    token2.IDENT,
						Literal: "myVar",
					},
					Value: "myVar",
				},
				Value: &Identifier{
					Token: token2.Token{
						Type:    token2.IDENT,
						Literal: "anotherVar",
					},
					Value: "anotherVar",
				},
			},
		},
	}
	if program.String() != "let myVar = anotherVar;" {
		i.Errorf("prgram.String() wrong. got=%q", program.String())
	}
}
