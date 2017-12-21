package main

import (
	"bytes"
	"fmt"
	"strconv"
	"unicode"
)

func main() {

}

func evaluate(expression string) int {
	runes := make([]rune, 0, len(expression))
	for _, r := range expression {
		runes = append(runes, r)
	}
	expr, _ := Parse(runes)
	return expr.Value(NewContext())
}

func Parse(expression []rune) (expr Expression, size int) {
	switch {
	case unicode.IsDigit(expression[0]) || expression[0] == '-':
		return ParseInt(expression)
	case unicode.IsLetter(expression[0]):
		return ParseId(expression)
	case string(expression[:5]) == "(add ":
		return ParseAdd(expression)
	case string(expression[:6]) == "(mult ":
		return ParseMult(expression)
	case string(expression[:5]) == "(let ":
		return ParseLet(expression)
	}
	panic(fmt.Errorf("wrong expression: %q", string(expression)))
}

func FirstNumFrom(text []rune) string {
	var res bytes.Buffer
	if text[0] == '-' {
		res.WriteRune(text[0])
		text = text[1:]
	}
	for len(text) > 0 && unicode.IsDigit(text[0]) {
		res.WriteRune(text[0])
		text = text[1:]
	}
	return res.String()
}

func ParseInt(text []rune) (expr Int, size int) {
	res64, err := strconv.ParseInt(FirstNumFrom(text), 10, 32)
	if err != nil {
		panic(err)
	}
	res := int(res64)
	return Int(res), len(strconv.Itoa(res))
}
func ParseAdd(text []rune) (expr AddExpr, size int) {
	startFirstExpr := text[5:]
	firstExpr, firstLen := Parse(startFirstExpr)
	startSecondExpr := startFirstExpr[firstLen+1:] // +1 for space between
	secondExpr, secondLen := Parse(startSecondExpr)
	return AddExpr{firstExpr, secondExpr}, firstLen + secondLen + 5 + 1 + 1 // 5 <- prefix='(add ', 1 <- suffix=')' 1 - for space between arguments
}
func ParseMult(text []rune) (expr MultExpr, size int) {
	defer func() {
		fmt.Printf("%s -> %d\n", string(text), size)
	}()
	startFirstExpr := text[6:] // Start after suffix ='(mult'
	firstExpr, firstLen := Parse(startFirstExpr)
	startSecondExpr := startFirstExpr[firstLen+1:] // +1 for space between
	secondExpr, secondLen := Parse(startSecondExpr)
	return MultExpr{firstExpr, secondExpr}, firstLen + secondLen + 6 + 1 + 1 // 5 <- prefix='(mult ', 1 <- suffix=')'
}
func ParseLet(t []rune) (expr LetExpr, size int) {
	text := t[5:] // Start after suffic='(let '
	var lastExpr Expression
	lastExpr = Int(0)
	pairs := make([]LetPair, 0, len(text)>>2)
	for len(text) > 0 {
		if !unicode.IsLetter(text[0]) {
			exp, expSize := Parse(text)
			lastExpr = exp
			text = text[expSize:]
			break
		}
		id, idSize := ParseId(text)
		text = text[idSize:]
		if text[0] == ')' {
			lastExpr = id
			break
		}
		text = text[1:]
		e, eSize := Parse(text)
		pairs = append(pairs, LetPair{string(id), e})
		text = text[eSize+1:] // +1 for space after expression
	}
	return LetExpr{pairs, lastExpr}, len(t) - len(text) + 1 // +1 for suffix=')'

}
func ParseId(text []rune) (expr Id, size int) {
	var res bytes.Buffer
	for len(text) > 0 && unicode.IsLetter(text[0]) || unicode.IsDigit(text[0]) {
		res.WriteRune(text[0])
		text = text[1:]
	}
	id := res.String()
	if len(id) == 0 {
		panic(fmt.Errorf("Wrong id: %q", string(text)))
	}
	return Id(id), len(id)
}

// Calculation types
type Context map[string]int

func NewContext() Context {
	return make(map[string]int)
}
func (c Context) Copy() Context {
	n := make(map[string]int)
	for k, v := range c {
		n[k] = v
	}
	return n
}

type Expression interface {
	Value(c Context) int
}

type Int int

func (i Int) Value(c Context) int {
	return int(i)
}

type LetPair struct {
	Id   string
	Expr Expression
}

type LetExpr struct {
	Definitions []LetPair
	Expr        Expression
}

func (l LetExpr) Value(c Context) int {
	currentContext := c.Copy()
	for _, v := range l.Definitions {
		currentContext[v.Id] = v.Expr.Value(currentContext)
	}
	return l.Expr.Value(currentContext)
}

type AddExpr struct {
	FirstExpr  Expression
	SecondExpr Expression
}

func (a AddExpr) Value(c Context) int {
	return a.FirstExpr.Value(c) + a.SecondExpr.Value(c)
}

type MultExpr struct {
	FirstExpr  Expression
	SecondExpr Expression
}

func (m MultExpr) Value(c Context) int {
	return m.FirstExpr.Value(c) * m.SecondExpr.Value(c)
}

type Id string

func (i Id) Value(c Context) int {
	return c[string(i)]
}
