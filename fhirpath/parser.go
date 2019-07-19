package fhirpath

import (
	"fmt"
	"github.com/alecthomas/participle"
	"github.com/alecthomas/participle/lexer"
	"github.com/alecthomas/participle/lexer/ebnf"
)

// Parses a string in to a FHIRPath expression
func Parse(s string) (Expression, error) {
	expr := Expression{}
	err := Parser.ParseString(s, &expr)
	if err != nil {
		return expr, err
	}
	return expr, nil
}

type Expression struct {
	Left EqualityExpression `@@`
}

type EqualityExpression struct {
	Left     InequalityExpression  `@@`
	Operator Operator              `[ @EqualOp `
	Right    *InequalityExpression `@@ ]`
}

type InequalityExpression struct {
	Left     UnionExpression  `@@`
	Operator Operator         `[ @CmpOp `
	Right    *UnionExpression `@@ ]`
}

type UnionExpression struct {
	Left  TypeExpression  `@@`
	Right *TypeExpression `[ "|" @@ ]`
}

type TypeExpression struct {
	Expression AdditiveExpression `@@`
	Modifier   string             `[ @("as" | "is") `
	Type       *TypeSpecifier     `@@ ]`
}

type AdditiveExpression struct {
	Left  MultiplicativeExpression     `@@`
	Right []OpMultiplicativeExpression `{ @@ }`
}

type OpMultiplicativeExpression struct {
	Operator   Operator                 `@("+" | "-" | "&")`
	Expression MultiplicativeExpression `@@`
}

type MultiplicativeExpression struct {
	Left  PolarityExpression     `@@`
	Right []OpPolarityExpression `{ @@ }`
}

type Operator int

const (
	OpMul Operator = iota
	OpDiv
	OpAdd
	OpSub
	OpDivide
	OpModulo
	OpAnd
	OpGreaterThan
	OpGreaterOrEqual
	OpLessThan
	OpLessOrEqual
	OpEqual
)

func (op *Operator) Capture(s []string) error {
	switch s[0] {
	case "+":
		*op = OpAdd
	case "-":
		*op = OpSub
	case "*":
		*op = OpMul
	case "/":
		*op = OpDiv
	case "div":
		*op = OpDivide
	case "mod":
		*op = OpModulo
	case "&":
		*op = OpAnd
	case "<=":
		*op = OpLessOrEqual
	case "<":
		*op = OpLessThan
	case ">":
		*op = OpGreaterThan
	case ">=":
		*op = OpGreaterOrEqual
	case "=":
		*op = OpEqual
	default:
		return fmt.Errorf("unexpected operator: %s", s[0])
	}
	return nil
}

func (op Operator) compareInt(a int, b int) (bool, error) {
	switch op {
	case OpLessOrEqual:
		return a <= b, nil
	case OpLessThan:
		return a < b, nil
	case OpGreaterThan:
		return a > b, nil
	case OpGreaterOrEqual:
		return a >= b, nil
	case OpEqual:
		return a == b, nil
	default:
		return false, fmt.Errorf("inappropriate comparison operator `%s` used for numbers", op)
	}
}

func (op Operator) compareFloat64(a float64, b float64) (bool, error) {
	switch op {
	case OpLessOrEqual:
		return a <= b, nil
	case OpLessThan:
		return a < b, nil
	case OpGreaterThan:
		return a > b, nil
	case OpGreaterOrEqual:
		return a >= b, nil
	case OpEqual:
		return a == b, nil
	default:
		return false, fmt.Errorf("inappropriate comparison operator `%s` used for numbers", op)
	}
}

func (op Operator) compareString(a string, b string) (bool, error) {
	switch op {
	case OpEqual:
		return a == b, nil
	default:
		return false, fmt.Errorf("inappropriate comparison operator `%s` used for strings", op)
	}
}

type OpPolarityExpression struct {
	Operator   Operator           `@("*" | "/" | "div" | "mod")`
	Expression PolarityExpression `@@`
}

type PolarityOp int

const (
	PPlus PolarityOp = iota
	PMinus
)

func (o *PolarityOp) Capture(s []string) error {
	switch s[0] {
	case "+":
		*o = PPlus
	case "-":
		*o = PMinus
	default:
		return fmt.Errorf("unexpected polarity operator: |%s|", s[0])
	}
	return nil
}

type PolarityExpression struct {
	Polarity   *PolarityOp       `@("+" | "-")?`
	Expression IndexerExpression `@@`
}

type IndexerExpression struct {
	Target InvocationExpression  `@@`
	Index  *InvocationExpression `("[" @@ "]")?`
}

type InvocationExpression struct {
	Target      Term         `@@`
	Invocations []Invocation `{ "." @@ }`
}

type Term struct {
	Literal    *Literal    `@@`
	Invocation *Invocation `|@@`
}

type Literal struct {
	Boolean *bool    `@Boolean`
	Str     *string  `|@String`
	Number  *float64 `|@Number`
}

type Invocation struct {
	Function   *Function   `@@`
	Identifier *Identifier `|@@`
}

type TypeSpecifier struct {
	QualifiedIdentifier QualifiedIdentifier `@@`
}

type QualifiedIdentifier struct {
	Identifiers []Identifier `@@ { "." @@ }`
}

type Identifier struct {
	As                  string `@"as"`
	Is                  string `|@"is"`
	Identifier          string `|@Identifier`
	DelimitedIdentifier string `|@DelimitedIdentifier`
}

type Function struct {
	Name   Identifier   `@@`
	Params []Expression `"(" [ @@ ] { "," @@ } ")"`
}

var (
	fhirPathLexer = lexer.Must(ebnf.New(`
    Boolean = "true" | "false" .
    Identifier = (alpha | "_") { "_" | alpha | digit } .
    DelimitedIdentifier = "\u0060" { "\u0000"…"\uffff"-"\u0060" } "\u0060" .
    String = "'" { "\u0000"…"\uffff"-"'" } "'" .
    Number = digit { digit } [ [ "." ] digit { digit } ] . 
    Whitespace = " " | "\t" | "\n" | "\r" .
    Punct = "." | "[" | "]"  | "(" | ")" .
    Operator = "+" | "-" | "&" | "*" | "/" | "div" | "mod" | "|" .
    CmpOp = "<" [ "=" ] | ">" [ "=" ] .
    EqualOp = "=" | "~" | "!" ( "=" | "~" ) .

    alpha = "a"…"z" | "A"…"Z" .
    digit = "0"…"9" .
`))

	Parser = participle.MustBuild(&Expression{},
		participle.Lexer(fhirPathLexer),
		participle.Elide("Whitespace"),
	)
)
