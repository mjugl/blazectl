package cmd

import (
	"github.com/life-research/blazectl/fhirpath"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestFindAllVarNames(t *testing.T) {
	assert.Equal(t, []string{"a"}, FindAllVarNames("${a}"))
	assert.Equal(t, []string{"a", "b"}, FindAllVarNames("${a}${b}"))
}

func TestReplaceVariables(t *testing.T) {
	resource := make(map[string]interface{})
	resource["a"] = "b"

	fixtures := make(map[string]fixture)
	fixtures["F"] = fixture{resource: resource}

	variables := make(map[string]variable)
	expr, err := fhirpath.Parse("a")
	if err != nil {
		t.Error(err)
		return
	}
	variables["V"] = variable{expression: expr, sourceId: "F"}

	replacement, err := replaceVariables(context{fixtures: fixtures, variables: variables}, "${V}")
	if err != nil {
		t.Error(err)
		return
	}
	assert.Equal(t, "b", replacement)
}
