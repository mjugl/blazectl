package testscript

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestCompWithResult(t *testing.T) {
	if res, err := Equals.CompWithResult("0", []interface{}{0.0}); err != nil {
		t.Error(err)
	} else {
		assert.True(t, res)
	}
	if res, err := NotEquals.CompWithResult("0", []interface{}{1.0}); err != nil {
		t.Error(err)
	} else {
		assert.True(t, res)
	}
}
