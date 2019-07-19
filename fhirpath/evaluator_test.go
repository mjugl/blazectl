package fhirpath

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestEvalInvoke(t *testing.T) {
	if expr, err := Parse("a"); err != nil {
		t.Error(err)
	} else if result, err := expr.Eval([]interface{}{map[string]interface{}{"a": "b"}}); err != nil {
		t.Error(err)
	} else {
		assert.Equal(t, "b", result[0])
	}
}

func TestEvalInvokeWithType(t *testing.T) {
	if expr, err := Parse("Bundle.total"); err != nil {
		t.Error(err)
	} else if result, err := expr.Eval([]interface{}{map[string]interface{}{"resourceType": "Bundle", "total": 0}}); err != nil {
		t.Error(err)
	} else {
		assert.Equal(t, 0, result[0])
	}
}

func TestEvalInvokeStringSlice(t *testing.T) {
	if expr, err := Parse("a"); err != nil {
		t.Error(err)
	} else if result, err := expr.Eval([]interface{}{map[string]interface{}{"a": []interface{}{"b", "c"}, "d": "e"}}); err != nil {
		t.Error(err)
	} else {
		assert.Equal(t, "b", result[0])
		assert.Equal(t, "c", result[1])
	}
}

func TestEqualityExpression_Eval_number(t *testing.T) {
	if expr, err := Parse("1 = 1"); err != nil {
		t.Error(err)
	} else if result, err := expr.Eval([]interface{}{}); err != nil {
		t.Error(err)
	} else {
		assert.Equal(t, true, result[0])
	}
}

func TestEqualityExpression_Eval_string(t *testing.T) {
	if expr, err := Parse("'a' = 'a'"); err != nil {
		t.Error(err)
	} else if result, err := expr.Eval([]interface{}{}); err != nil {
		t.Error(err)
	} else {
		assert.Equal(t, true, result[0])
	}
}

func TestInequalityExpression_Eval_greaterThan(t *testing.T) {
	if expr, err := Parse("1 > 0"); err != nil {
		t.Error(err)
	} else if result, err := expr.Eval([]interface{}{}); err != nil {
		t.Error(err)
	} else {
		assert.Equal(t, true, result[0])
	}
}

func TestInequalityExpression_Eval_lessThan(t *testing.T) {
	if expr, err := Parse("1 < 0"); err != nil {
		t.Error(err)
	} else if result, err := expr.Eval([]interface{}{}); err != nil {
		t.Error(err)
	} else {
		assert.Equal(t, false, result[0])
	}
}

func TestInequalityExpression_Eval_other(t *testing.T) {
	if expr, err := Parse("a > 0"); err != nil {
		t.Error(err)
	} else if result, err := expr.Eval([]interface{}{map[string]interface{}{"a": 1}}); err != nil {
		t.Error(err)
	} else {
		assert.Equal(t, true, result[0])
	}
}

func TestFunction_Eval(t *testing.T) {
	if expr, err := Parse("a.where(b = 3).c"); err != nil {
		t.Error(err)
	} else if result, err := expr.Eval([]interface{}{
		map[string]interface{}{"a": []interface{}{
			map[string]interface{}{"b": 1, "c": 2},
			map[string]interface{}{"b": 3, "c": 4},
		}},
	}); err != nil {
		t.Error(err)
	} else {
		assert.Equal(t, 4, result[0])
	}
}

func TestLiteral_Eval(t *testing.T) {
	if expr, err := Parse("3"); err != nil {
		t.Error(err)
	} else if result, err := expr.Eval([]interface{}{}); err != nil {
		t.Error(err)
	} else {
		assert.Equal(t, 3.0, result[0])
	}
	if expr, err := Parse("'a'"); err != nil {
		t.Error(err)
	} else if result, err := expr.Eval([]interface{}{}); err != nil {
		t.Error(err)
	} else {
		assert.Equal(t, "a", result[0])
	}
}
