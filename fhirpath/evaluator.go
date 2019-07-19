package fhirpath

import (
	"fmt"
	"reflect"
	"strings"
	"unicode"
)

func (e *Expression) Eval(coll []interface{}) ([]interface{}, error) {
	return e.Left.Eval(coll)
}

func (e *EqualityExpression) Eval(coll []interface{}) ([]interface{}, error) {
	left, err := e.Left.Eval(coll)
	if err != nil {
		return left, err
	}
	if e.Right != nil {
		if len(left) == 0 {
			return left, nil
		}
		if len(left) > 1 {
			return left, fmt.Errorf("left side has more than one value")
		}
		right, err := e.Right.Eval(coll)
		if err != nil {
			return right, err
		}
		if len(right) == 0 {
			return right, nil
		}
		if len(right) > 1 {
			return right, fmt.Errorf("right side has more than one value")
		}
		res, err := compare(e.Operator, left[0], right[0])
		if err != nil {
			return coll, err
		}
		return []interface{}{res}, nil
	}
	return left, nil
}

func (e *InequalityExpression) Eval(coll []interface{}) ([]interface{}, error) {
	left, err := e.Left.Eval(coll)
	if err != nil {
		return left, err
	}
	if e.Right != nil {
		if len(left) == 0 {
			return left, nil
		}
		if len(left) > 1 {
			return left, fmt.Errorf("left side has more than one value")
		}
		right, err := e.Right.Eval(coll)
		if err != nil {
			return right, err
		}
		if len(right) == 0 {
			return right, nil
		}
		if len(right) > 1 {
			return right, fmt.Errorf("right side has more than one value")
		}
		res, err := compare(e.Operator, left[0], right[0])
		if err != nil {
			return coll, err
		}
		return []interface{}{res}, nil
	}
	return left, nil
}

func compare(op Operator, a interface{}, b interface{}) (bool, error) {
	switch a.(type) {
	case int:
		switch b.(type) {
		case int:
			return op.compareInt(a.(int), b.(int))
		case float64:
			return op.compareFloat64(float64(a.(int)), b.(float64))
		}
	case float64:
		switch b.(type) {
		case int:
			return op.compareFloat64(a.(float64), float64(b.(int)))
		case float64:
			return op.compareFloat64(a.(float64), b.(float64))
		}
	case string:
		switch b.(type) {
		case string:
			return op.compareString(a.(string), b.(string))
		}
	}
	return false, fmt.Errorf("uncomparable types `%s` and `%s`", reflect.TypeOf(a), reflect.TypeOf(b))
}

func (e *UnionExpression) Eval(coll []interface{}) ([]interface{}, error) {
	return e.Left.Eval(coll)
}

func (e *TypeExpression) Eval(coll []interface{}) ([]interface{}, error) {
	return e.Expression.Eval(coll)
}

func (e *AdditiveExpression) Eval(coll []interface{}) ([]interface{}, error) {
	return e.Left.Eval(coll)
}

func (e *MultiplicativeExpression) Eval(coll []interface{}) ([]interface{}, error) {
	return e.Left.Eval(coll)
}

func (e *PolarityExpression) Eval(coll []interface{}) ([]interface{}, error) {
	return e.Expression.Eval(coll)
}

func (e *IndexerExpression) Eval(coll []interface{}) ([]interface{}, error) {
	return e.Target.Eval(coll)
}

func (e *InvocationExpression) Eval(coll []interface{}) ([]interface{}, error) {
	result, err := e.Target.Eval(coll)
	if err != nil {
		return result, err
	}
	for _, invocation := range e.Invocations {
		result, err = invocation.Eval(result)
		if err != nil {
			return result, err
		}
	}
	return result, nil
}

func (t *Term) Eval(coll []interface{}) ([]interface{}, error) {
	if invocation := t.Invocation; invocation != nil {
		return invocation.Eval(coll)
	}
	if literal := t.Literal; literal != nil {
		return literal.Eval(coll)
	}
	return nil, fmt.Errorf("unsupported term `%+v`", t)
}

func (i *Invocation) Eval(coll []interface{}) ([]interface{}, error) {
	if identifier := i.Identifier; identifier != nil {
		return identifier.Eval(coll)
	}
	if function := i.Function; function != nil {
		return function.Eval(coll)
	}
	return nil, fmt.Errorf("unsupported invocation `%+v`", i)
}

func (i *Identifier) Eval(coll []interface{}) ([]interface{}, error) {
	result := make([]interface{}, 0, len(coll))

	if unicode.IsUpper([]rune(i.Identifier)[0]) {
		// Types
		for _, item := range coll {
			if mapItem, ok := item.(map[string]interface{}); ok {
				if i.Identifier == mapItem["resourceType"] {
					result = append(result, mapItem)
				}
			}
		}
	} else {
		// Normal navigation
		for _, item := range coll {
			if mapItem, ok := item.(map[string]interface{}); ok {
				if child := mapItem[i.Identifier]; child != nil {
					if children, ok := child.([]interface{}); ok {
						result = append(result, children...)
					} else {
						result = append(result, child)
					}
				}
			} else {
				return result, fmt.Errorf("unsupported navigation from `%v` with `%s`", item, i.Identifier)
			}
		}
	}
	return result, nil
}

func (f *Function) Eval(coll []interface{}) ([]interface{}, error) {
	switch name := f.Name.Identifier; name {
	case "where":
		return callWhere(f.Params, coll)
	default:
		return nil, fmt.Errorf("unsupported function: %s", name)
	}
}

func callWhere(params []Expression, coll []interface{}) ([]interface{}, error) {
	if n := len(params); n == 1 {
		criteria := params[0]

		result := make([]interface{}, 0, len(coll))
		for _, item := range coll {
			res, err := criteria.Eval([]interface{}{item})
			if err != nil {
				return nil, err
			}
			if boolRes, ok := res[0].(bool); ok {
				if boolRes {
					result = append(result, item)
				}
			} else {
				return nil, fmt.Errorf("unexpected non-bool result while calling function `where` with criteria `%+v` on item `%+v`", criteria, item)
			}
		}
		return result, nil
	} else {
		return nil, fmt.Errorf("invalid call of function `where` expected one param (criteria) but got %d", n)
	}
}

func (l *Literal) Eval(coll []interface{}) ([]interface{}, error) {
	if boolean := l.Boolean; boolean != nil {
		return []interface{}{*boolean}, nil
	}
	if str := l.Str; str != nil {
		return []interface{}{strings.Trim(*str, "'")}, nil
	}
	if number := l.Number; number != nil {
		return []interface{}{*number}, nil
	}
	return nil, fmt.Errorf("unsupported literal `%+v`", l)
}
