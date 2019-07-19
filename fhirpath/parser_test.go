// Copyright © 2019 Alexander Kiel <alexander.kiel@life.uni-leipzig.de>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package fhirpath

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestParseTermExpressionWithBooleanLiteral(t *testing.T) {
	if expr, err := Parse("true"); err != nil {
		t.Error(err)
	} else {
		assert.Equal(t, true, *expr.Left.Left.Left.Left.Expression.Left.Left.Expression.Target.Target.Literal.Boolean)
	}
}

func TestParseTermExpressionWithMemberInvocation(t *testing.T) {
	expr := Expression{}
	if err := Parser.ParseString("a", &expr); err != nil {
		t.Error(err)
	} else {
		assert.Equal(t, "a", expr.Left.Left.Left.Left.Expression.Left.Left.Expression.Target.Target.Invocation.Identifier.Identifier)
	}
}

func TestParseTermExpressionWithFunctionInvocation(t *testing.T) {
	expr := Expression{}
	if err := Parser.ParseString("a()", &expr); err != nil {
		t.Error(err)
	} else {
		assert.Equal(t, "a", expr.Left.Left.Left.Left.Expression.Left.Left.Expression.Target.Target.Invocation.Function.Name.Identifier)
	}
}

func TestParseTermExpressionWithFunctionInvocationAndOneParam(t *testing.T) {
	expr := Expression{}
	if err := Parser.ParseString("a(1)", &expr); err != nil {
		t.Error(err)
	} else {
		assert.Equal(t, "a", expr.Left.Left.Left.Left.Expression.Left.Left.Expression.Target.Target.Invocation.Function.Name.Identifier)
		assert.Equal(t, 1.0, *expr.Left.Left.Left.Left.Expression.Left.Left.Expression.Target.Target.Invocation.Function.Params[0].Left.Left.Left.Left.Expression.Left.Left.Expression.Target.Target.Literal.Number)
	}
}

func TestParseInvocationExpression(t *testing.T) {
	expr := Expression{}
	if err := Parser.ParseString("a.b", &expr); err != nil {
		t.Error(err)
	} else {
		assert.Equal(t, "a", expr.Left.Left.Left.Left.Expression.Left.Left.Expression.Target.Target.Invocation.Identifier.Identifier)
		assert.Equal(t, "b", expr.Left.Left.Left.Left.Expression.Left.Left.Expression.Target.Invocations[0].Identifier.Identifier)
	}
}

func TestParseIndexerExpression(t *testing.T) {
	expr := Expression{}
	if err := Parser.ParseString("a[0]", &expr); err != nil {
		t.Error(err)
	} else {
		assert.Equal(t, "a", expr.Left.Left.Left.Left.Expression.Left.Left.Expression.Target.Target.Invocation.Identifier.Identifier)
		assert.Equal(t, 0.0, *expr.Left.Left.Left.Left.Expression.Left.Left.Expression.Index.Target.Literal.Number)
	}
}

func TestParsePolarityExpressionNothing(t *testing.T) {
	expr := Expression{}
	if err := Parser.ParseString("a", &expr); err != nil {
		t.Error(err)
	} else {
		assert.Nil(t, expr.Left.Left.Left.Left.Expression.Left.Left.Polarity)
		assert.Equal(t, "a", expr.Left.Left.Left.Left.Expression.Left.Left.Expression.Target.Target.Invocation.Identifier.Identifier)
	}
}

func TestParsePolarityExpressionMinus(t *testing.T) {
	expr := Expression{}
	if err := Parser.ParseString("-a", &expr); err != nil {
		t.Error(err)
	} else {
		assert.Equal(t, PMinus, *expr.Left.Left.Left.Left.Expression.Left.Left.Polarity)
		assert.Equal(t, "a", expr.Left.Left.Left.Left.Expression.Left.Left.Expression.Target.Target.Invocation.Identifier.Identifier)
	}
}

func TestParseMultiplicativeExpression(t *testing.T) {
	expr := Expression{}
	if err := Parser.ParseString("a*b", &expr); err != nil {
		t.Error(err)
	} else {
		assert.Equal(t, "a", expr.Left.Left.Left.Left.Expression.Left.Left.Expression.Target.Target.Invocation.Identifier.Identifier)
		assert.Equal(t, OpMul, expr.Left.Left.Left.Left.Expression.Left.Right[0].Operator)
		assert.Equal(t, "b", expr.Left.Left.Left.Left.Expression.Left.Right[0].Expression.Expression.Target.Target.Invocation.Identifier.Identifier)
	}
}

func TestParseAdditiveExpression(t *testing.T) {
	expr := Expression{}
	if err := Parser.ParseString("a+b", &expr); err != nil {
		t.Error(err)
	} else {
		assert.Equal(t, "a", expr.Left.Left.Left.Left.Expression.Left.Left.Expression.Target.Target.Invocation.Identifier.Identifier)
		assert.Equal(t, OpAdd, expr.Left.Left.Left.Left.Expression.Right[0].Operator)
		assert.Equal(t, "b", expr.Left.Left.Left.Left.Expression.Right[0].Expression.Left.Expression.Target.Target.Invocation.Identifier.Identifier)
	}
}

func TestParseTypeExpression(t *testing.T) {
	expr := Expression{}
	if err := Parser.ParseString("a as b", &expr); err != nil {
		t.Error(err)
	} else {
		assert.Equal(t, "a", expr.Left.Left.Left.Left.Expression.Left.Left.Expression.Target.Target.Invocation.Identifier.Identifier)
		assert.Equal(t, "as", expr.Left.Left.Left.Left.Modifier)
		assert.Equal(t, "b", expr.Left.Left.Left.Left.Type.QualifiedIdentifier.Identifiers[0].Identifier)
	}
}

func TestParseUnionExpression(t *testing.T) {
	expr := Expression{}
	if err := Parser.ParseString("a | b", &expr); err != nil {
		t.Error(err)
	} else {
		assert.Equal(t, "a", expr.Left.Left.Left.Left.Expression.Left.Left.Expression.Target.Target.Invocation.Identifier.Identifier)
		assert.Equal(t, "b", expr.Left.Left.Left.Right.Expression.Left.Left.Expression.Target.Target.Invocation.Identifier.Identifier)
	}
}

func TestParseInequalityExpression(t *testing.T) {
	expr := Expression{}
	if err := Parser.ParseString("a > 0", &expr); err != nil {
		t.Error(err)
	} else {
		assert.Equal(t, "a", expr.Left.Left.Left.Left.Expression.Left.Left.Expression.Target.Target.Invocation.Identifier.Identifier)
		assert.Equal(t, OpGreaterThan, expr.Left.Left.Operator)
		assert.Equal(t, 0.0, *expr.Left.Left.Right.Left.Expression.Left.Left.Expression.Target.Target.Literal.Number)
	}
}
