package fhir

import (
	"encoding/json"
	"fmt"
	"github.com/life-research/blazectl/fhir/testscript"
)

type InvalidResourceTypeError struct {
	Expected string
	Actual   string
}

func (e InvalidResourceTypeError) Error() string {
	return fmt.Sprintf("expected resource of type `%s` but was `%s`", e.Expected, e.Actual)
}

type jsonTestScript struct {
	ResourceType string
	Name         string
	Title        string
	Variable     []testscript.Variable
	Fixture      []testscript.Fixture
	Setup        *testscript.Setup
	Test         []testscript.Test
}

func (s *TestScript) UnmarshalJSON(b []byte) error {
	var j jsonTestScript
	if err := json.Unmarshal(b, &j); err != nil {
		return err
	}
	if j.ResourceType != "TestScript" {
		return InvalidResourceTypeError{"TestScript", j.ResourceType}
	}
	*s = TestScript{
		Name:     j.Name,
		Title:    j.Title,
		Variable: j.Variable,
		Fixture:  j.Fixture,
		Setup:    j.Setup,
		Test:     j.Test,
	}
	return nil
}
