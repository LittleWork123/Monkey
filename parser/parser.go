package parser

import (
	"fmt"
	"interpreter/ast"
	"interpreter/lexer"
	"interpreter/token2"
	"strconv"
)

type Parser struct {
	l *lexer.Lexer

	curToken  token2.Token
	peekToken token2.Token
	errors    []string

	// prefix function and infix function
	prefixParseFns map[token2.TokenType]prefixParseFn
	infixParseFns  map[token2.TokenType]infixParseFn
}

const (
	_ int = iota // set increment number
	LOWEST
	EQUALS      //==
	LESSGREATER // > or <
	SUM         // +
	PRODUCT     // *
	PREFIX      //-X or +X
	CALL        // myFunction(X)
	INDEX
)

type (
	prefixParseFn func() ast.Expression
	infixParseFn  func(ast.Expression) ast.Expression
)

var precedences = map[token2.TokenType]int{
	token2.EQ:       EQUALS,
	token2.NOT_EQ:   EQUALS,
	token2.LT:       LESSGREATER,
	token2.GT:       LESSGREATER,
	token2.PLUS:     SUM,
	token2.MINUS:    SUM,
	token2.SLASH:    PRODUCT,
	token2.ASTERISK: PRODUCT,
	token2.LPAREN:   CALL,
	token2.LBRACKET: INDEX,
}

func (p *Parser) peekPrecedence() int {
	if p, ok := precedences[p.peekToken.Type]; ok {
		return p
	}
	return LOWEST
}

func (p *Parser) curPrecedence() int {
	if p, ok := precedences[p.curToken.Type]; ok {
		return p
	}
	return LOWEST
}
func New(l *lexer.Lexer) *Parser {
	p := &Parser{
		l:      l,
		errors: []string{},
	}

	p.prefixParseFns = make(map[token2.TokenType]prefixParseFn)
	p.registerPrefix(token2.IDENT, p.parseIdentifier)
	p.registerPrefix(token2.INT, p.parseIntegerLiteral)
	p.registerPrefix(token2.BANG, p.parsePrefixExpression)
	p.registerPrefix(token2.MINUS, p.parsePrefixExpression)
	p.registerPrefix(token2.TRUE, p.parseBoolean)
	p.registerPrefix(token2.FALSE, p.parseBoolean)
	p.registerPrefix(token2.LPAREN, p.parseGroupedExpression)
	p.registerPrefix(token2.IF, p.parseIfExpression)
	p.registerPrefix(token2.FUNCTION, p.parseFunctionLiteral)
	p.registerPrefix(token2.STRING, p.parseStringLiteral)
	p.registerPrefix(token2.LBRACKET, p.parseArrayLiteral)
	p.registerPrefix(token2.LBRACE, p.parseHashLiteral)

	p.infixParseFns = make(map[token2.TokenType]infixParseFn)
	p.registerInfix(token2.PLUS, p.parseInfixExpression)
	p.registerInfix(token2.MINUS, p.parseInfixExpression)
	p.registerInfix(token2.ASTERISK, p.parseInfixExpression)
	p.registerInfix(token2.SLASH, p.parseInfixExpression)
	p.registerInfix(token2.EQ, p.parseInfixExpression)
	p.registerInfix(token2.NOT_EQ, p.parseInfixExpression)
	p.registerInfix(token2.GT, p.parseInfixExpression)
	p.registerInfix(token2.LT, p.parseInfixExpression)
	p.registerInfix(token2.LPAREN, p.parseCallExpression)
	p.registerInfix(token2.LBRACKET, p.parseIndexExpression)
	// read two token to initialize curToken and peekToken
	p.nextToken()
	p.nextToken()
	return p
}

func (p *Parser) parseStringLiteral() ast.Expression {
	return &ast.StringLiteral{
		Token: p.curToken,
		Value: p.curToken.Literal,
	}
}
func (p *Parser) parseInfixExpression(left ast.Expression) ast.Expression {
	expression := &ast.InfixExpression{
		Token:    p.curToken,
		Operator: p.curToken.Literal,
		Left:     left,
	}
	precedence := p.curPrecedence()
	p.nextToken()
	expression.Right = p.parseExpression(precedence)
	return expression
}

func (p *Parser) parsePrefixExpression() ast.Expression {
	expression := &ast.PrefixExpression{
		Token:    p.curToken,
		Operator: p.curToken.Literal,
	}
	p.nextToken()
	expression.Right = p.parseExpression(PREFIX)
	return expression
}
func (p *Parser) parseIdentifier() ast.Expression {
	return &ast.Identifier{
		Token: p.curToken,
		Value: p.curToken.Literal,
	}
}

func (p *Parser) parseIntegerLiteral() ast.Expression {
	literal := &ast.IntegerLiteral{Token: p.curToken}
	value, err := strconv.ParseInt(p.curToken.Literal, 0, 64)
	if err != nil {
		msg := fmt.Sprintf("could not parse %q as integer", p.curToken.Literal)
		p.errors = append(p.errors, msg)
		return nil
	}
	literal.Value = value
	return literal
}

func (p *Parser) nextToken() {
	p.curToken = p.peekToken
	p.peekToken = p.l.NextToken()
}

func (p *Parser) ParseProgram() *ast.Program {
	program := &ast.Program{}
	program.Statements = []ast.Statement{}
	for !p.curTokenIs(token2.EOF) {
		statement := p.parseStatement()
		if statement != nil {
			program.Statements = append(program.Statements, statement)
		}
		p.nextToken()
	}
	return program
}

func (p *Parser) parseStatement() ast.Statement {
	switch p.curToken.Type {
	case token2.LET:
		return p.parseLetStatement()
	case token2.RETURN:
		return p.parseReturnStatement()
	default:
		return p.parseExpressionStatement()
	}
}

// parse let statement
func (p *Parser) parseLetStatement() *ast.LetStatement {
	statement := &ast.LetStatement{Token: p.curToken}
	if !p.expectPeek(token2.IDENT) {
		return nil
	}
	statement.Name = &ast.Identifier{
		Token: p.curToken,
		Value: p.curToken.Literal,
	}
	if !p.expectPeek(token2.ASSIGN) {
		return nil
	}
	p.nextToken()
	statement.Value = p.parseExpression(LOWEST)
	for !p.curTokenIs(token2.SEMICOLON) {
		p.nextToken()
	}
	return statement
}

func (p *Parser) parseReturnStatement() *ast.ReturnStatement {
	statement := &ast.ReturnStatement{
		Token: p.curToken,
	}
	p.nextToken()

	statement.ReturnValue = p.parseExpression(LOWEST)

	for !p.curTokenIs(token2.SEMICOLON) {
		p.nextToken()
	}
	return statement
}
func (p *Parser) expectPeek(t token2.TokenType) bool {
	if p.peekTokenIs(t) {
		p.nextToken()
		return true
	} else {
		p.peekError(t)
		return false
	}
}

func (p *Parser) curTokenIs(t token2.TokenType) bool {
	return p.curToken.Type == t
}

func (p *Parser) peekTokenIs(t token2.TokenType) bool {
	return p.peekToken.Type == t
}

// Errors process the illegal statement
func (p *Parser) Errors() []string {
	return p.errors
}

func (p *Parser) peekError(t token2.TokenType) {
	msg := fmt.Sprintf("expected next token to be %s, got %s instead",
		t, p.peekToken.Type)
	p.errors = append(p.errors, msg)
}

// fin in prefix function
func (p *Parser) registerPrefix(tokenType token2.TokenType, fn prefixParseFn) {
	p.prefixParseFns[tokenType] = fn
}

// fill in infix function
func (p *Parser) registerInfix(tokenType token2.TokenType, fn infixParseFn) {
	p.infixParseFns[tokenType] = fn
}

func (p *Parser) parseExpressionStatement() *ast.ExpressionStatement {
	stmt := &ast.ExpressionStatement{Token: p.curToken}
	stmt.Expression = p.parseExpression(LOWEST)
	if p.peekTokenIs(token2.SEMICOLON) {
		p.nextToken()
	}
	return stmt
}

func (p *Parser) parseExpression(precedence int) ast.Expression {
	prefix := p.prefixParseFns[p.curToken.Type]
	if prefix == nil {
		p.noPrefixParseFnError(p.curToken.Type)
		return nil
	}
	leftExp := prefix()
	// This loop won`t end until execute when if-statement is false
	for !p.peekTokenIs(token2.SEMICOLON) && precedence < p.peekPrecedence() {
		infix := p.infixParseFns[p.peekToken.Type]
		if infix == nil {
			return leftExp
		}
		p.nextToken()
		leftExp = infix(leftExp)
	}
	return leftExp
}

func (p *Parser) noPrefixParseFnError(t token2.TokenType) {
	msg := fmt.Sprintf("no prefix parse function for %s found", t)
	p.errors = append(p.errors, msg)
}

func (p *Parser) parseBoolean() ast.Expression {
	return &ast.Boolean{
		Token: p.curToken,
		Value: p.curTokenIs(token2.TRUE),
	}
}

func (p *Parser) parseGroupedExpression() ast.Expression {
	p.nextToken()
	exp := p.parseExpression(LOWEST)
	if !p.expectPeek(token2.RPAREN) {
		return nil
	}
	return exp
}

func (p *Parser) parseIfExpression() ast.Expression {
	expression := &ast.IfExpression{Token: p.curToken}
	if !p.expectPeek(token2.LPAREN) {
		return nil
	}
	p.nextToken()
	expression.Condition = p.parseExpression(LOWEST)
	if !p.expectPeek(token2.RPAREN) {
		return nil
	}
	if !p.expectPeek(token2.LBRACE) {
		return nil
	}
	expression.Consequence = p.parseBlockStatement()

	if p.peekTokenIs(token2.ELSE) {
		p.nextToken()
		if !p.expectPeek(token2.LBRACE) {
			return nil
		}
		expression.Alternative = p.parseBlockStatement()
	}
	return expression
}

func (p *Parser) parseBlockStatement() *ast.BlockStatement {
	block := &ast.BlockStatement{Token: p.curToken}
	block.Statements = []ast.Statement{}
	p.nextToken()
	if !p.curTokenIs(token2.RBRACE) && !p.curTokenIs(token2.EOF) {
		stmt := p.parseStatement()
		if stmt != nil {
			block.Statements = append(block.Statements, stmt)
		}
		p.nextToken()
	}
	return block
}

func (p *Parser) parseFunctionLiteral() ast.Expression {
	lit := &ast.FunctionLiteral{Token: p.curToken}
	if !p.expectPeek(token2.LPAREN) {
		return nil
	}
	lit.Parameters = p.parseFunctionParameters()
	if !p.expectPeek(token2.LBRACE) {
		return nil
	}
	lit.Body = p.parseBlockStatement()
	return lit
}

func (p *Parser) parseFunctionParameters() []*ast.Identifier {
	identifiers := []*ast.Identifier{}
	if p.peekTokenIs(token2.RPAREN) {
		p.nextToken()
		return identifiers
	}
	p.nextToken()
	ident := &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}
	identifiers = append(identifiers, ident)
	for p.peekTokenIs(token2.COMMA) {
		p.nextToken()
		p.nextToken()
		ident := &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}
		identifiers = append(identifiers, ident)
	}
	if !p.expectPeek(token2.RPAREN) {
		return nil
	}
	return identifiers
}

func (p *Parser) parseCallExpression(function ast.Expression) ast.Expression {
	exp := &ast.CallExpression{Token: p.curToken, Function: function}
	exp.Arguments = p.parseExpressionList(token2.RPAREN)
	return exp
}

func (p *Parser) parseArrayLiteral() ast.Expression {
	array := &ast.ArrayLiteral{Token: p.curToken}
	array.Elements = p.parseExpressionList(token2.RBRACKET)
	return array
}

func (p *Parser) parseExpressionList(end token2.TokenType) []ast.Expression {
	list := []ast.Expression{}
	if p.peekTokenIs(end) {
		p.nextToken()
		return list
	}
	p.nextToken()
	expression := p.parseExpression(LOWEST)
	list = append(list, expression)
	for p.peekTokenIs(token2.COMMA) {
		p.nextToken()
		p.nextToken()
		expression = p.parseExpression(LOWEST)
		list = append(list, expression)
	}
	if !p.expectPeek(end) {
		return nil
	}
	return list
}

func (p *Parser) parseIndexExpression(left ast.Expression) ast.Expression {
	indexExpression := &ast.IndexExpression{Token: p.curToken}
	indexExpression.Left = left
	if !p.curTokenIs(token2.LBRACKET) {
		return nil
	}
	p.nextToken()
	indexExpression.Index = p.parseExpression(LOWEST)
	if !p.expectPeek(token2.RBRACKET) {
		return nil
	}
	return indexExpression
}

func (p *Parser) parseHashLiteral() ast.Expression {
	hashLiteral := &ast.HashLiteral{Token: p.curToken}
	hashLiteral.Pairs = make(map[ast.Expression]ast.Expression)
	if !p.curTokenIs(token2.LBRACE) {
		return nil
	}
	for !p.peekTokenIs(token2.RBRACE) {
		p.nextToken()
		key := p.parseExpression(LOWEST)
		if !p.expectPeek(token2.COLON) {
			return nil
		}
		p.nextToken()
		value := p.parseExpression(LOWEST)
		hashLiteral.Pairs[key] = value
		//ingenious method &&
		if !p.peekTokenIs(token2.RBRACE) && !p.expectPeek(token2.COMMA) {
			return nil
		}
	}
	if !p.expectPeek(token2.RBRACE) {
		return nil
	}
	return hashLiteral
}
