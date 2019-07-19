package fhir

import (
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestTestScript_UnmarshalJSON_wrong_type(t *testing.T) {
	b := []byte(`{"resourceType": "Patient"}`)
	var s TestScript
	err := json.Unmarshal(b, &s)
	if assert.Error(t, err) {
		assert.Equal(t, InvalidResourceTypeError{"TestScript", "Patient"}, err)
	}
}

func TestTestScript_UnmarshalJSON_success(t *testing.T) {
	b := []byte(`{"resourceType": "TestScript", "name": "name-140946"}`)
	var s TestScript
	if err := json.Unmarshal(b, &s); err != nil {
		t.Error(err)
	} else {
		assert.Equal(t, "name-140946", s.Name)
	}
}
