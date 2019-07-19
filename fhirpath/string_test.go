package fhirpath

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestLiteral_String(t *testing.T) {
	b := true
	assert.Equal(t, "true", Literal{Boolean: &b}.String())
	b = false
	assert.Equal(t, "false", Literal{Boolean: &b}.String())
	s := "'a'"
	assert.Equal(t, "'a'", Literal{Str: &s}.String())
	n := 1.0
	assert.Equal(t, "1", Literal{Number: &n}.String())
	n = 1.1
	assert.Equal(t, "1.1", Literal{Number: &n}.String())
}

func TestPolarityOp_String(t *testing.T) {
	assert.Equal(t, "+", PPlus.String())
	assert.Equal(t, "-", PMinus.String())
}
