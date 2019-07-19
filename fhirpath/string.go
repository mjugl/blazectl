package fhirpath

import (
	"fmt"
	"strings"
)

const unknown = "<unknown>"

func (e Expression) String() string {
	return e.Left.String()
}

func (e EqualityExpression) String() string {
	if right := e.Right; right != nil {
		return fmt.Sprintf("%s %s %s", e.Left, e.Operator, right)
	}
	return e.Left.String()
}

func (e InequalityExpression) String() string {
	if right := e.Right; right != nil {
		return fmt.Sprintf("%s %s %s", e.Left, e.Operator, right)
	}
	return e.Left.String()
}

func (e UnionExpression) String() string {
	if right := e.Right; right != nil {
		return fmt.Sprintf("%s | %s", e.Left, right)
	}
	return e.Left.String()
}

func (e TypeExpression) String() string {
	if e.Type != nil {
		return fmt.Sprintf("%s %s %s", e.Expression, e.Modifier, e.Type)
	}
	return e.Expression.String()
}

func (e AdditiveExpression) String() string {
	rightExprStrings := make([]string, 0, len(e.Right))
	for _, expr := range e.Right {
		rightExprStrings = append(rightExprStrings, expr.String())
	}
	return e.Left.String() + strings.Join(rightExprStrings, ".")
}

func (e OpMultiplicativeExpression) String() string {
	return fmt.Sprintf(" %s %s", e.Operator, e.Expression)
}

func (e MultiplicativeExpression) String() string {
	rightExprStrings := make([]string, 0, len(e.Right))
	for _, expr := range e.Right {
		rightExprStrings = append(rightExprStrings, expr.String())
	}
	return e.Left.String() + strings.Join(rightExprStrings, ".")
}

func (e OpPolarityExpression) String() string {
	return fmt.Sprintf(" %s %s", e.Operator, e.Expression)
}

func (e PolarityExpression) String() string {
	if polarity := e.Polarity; polarity != nil {
		return fmt.Sprintf("%s%s", e.Polarity, e.Expression)
	}
	return e.Expression.String()
}

func (e IndexerExpression) String() string {
	if index := e.Index; index != nil {
		return fmt.Sprintf("%s.[%s]", e.Target, e.Index)
	}
	return e.Target.String()
}

func (e InvocationExpression) String() string {
	numInvocations := len(e.Invocations)
	if numInvocations > 0 {
		invocationStrs := make([]string, 0, numInvocations)
		for _, invocation := range e.Invocations {
			invocationStrs = append(invocationStrs, invocation.String())
		}
		return fmt.Sprintf("%s.%s", e.Target, strings.Join(invocationStrs, "."))
	}
	return e.Target.String()
}

func (t Term) String() string {
	if invocation := t.Invocation; invocation != nil {
		return invocation.String()
	}
	if literal := t.Literal; literal != nil {
		return literal.String()
	}
	return unknown
}

func (i Invocation) String() string {
	if identifier := i.Identifier; identifier != nil {
		return identifier.String()
	}
	if function := i.Function; function != nil {
		return function.String()
	}
	return unknown
}

func (i Identifier) String() string {
	if i.As != "" {
		return "as"
	}
	if i.Is != "" {
		return "as"
	}
	if i.Identifier != "" {
		return i.Identifier
	}
	if i.DelimitedIdentifier != "" {
		return i.DelimitedIdentifier
	}
	return unknown
}

func (f Function) String() string {
	stringParams := make([]string, 0, len(f.Params))
	for _, param := range f.Params {
		stringParams = append(stringParams, param.String())
	}
	return fmt.Sprintf("%s(%s)", f.Name, strings.Join(stringParams, ","))
}

func (l Literal) String() string {
	if boolean := l.Boolean; boolean != nil {
		return fmt.Sprint(*boolean)
	}
	if str := l.Str; str != nil {
		return *str
	}
	if number := l.Number; number != nil {
		return fmt.Sprint(*number)
	}
	return unknown
}

func (op PolarityOp) String() string {
	switch op {
	case PPlus:
		return "+"
	case PMinus:
		return "-"
	default:
		return unknown
	}
}

func (op Operator) String() string {
	switch op {
	case OpMul:
		return "*"
	case OpDiv:
		return "/"
	case OpAdd:
		return "+"
	case OpSub:
		return "-"
	case OpDivide:
		return "div"
	case OpModulo:
		return "mod"
	case OpAnd:
		return "&"
	case OpGreaterThan:
		return ">"
	case OpGreaterOrEqual:
		return ">="
	case OpLessThan:
		return "<"
	case OpLessOrEqual:
		return "<="
	case OpEqual:
		return "="
	default:
		return unknown
	}
}
