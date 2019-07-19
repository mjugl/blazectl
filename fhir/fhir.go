package fhir

import (
	"github.com/life-research/blazectl/fhir/testreport"
	"github.com/life-research/blazectl/fhir/testscript"
)

type TestScript struct {
	Name     string
	Title    string
	Variable []testscript.Variable
	Fixture  []testscript.Fixture
	Setup    *testscript.Setup
	Test     []testscript.Test
}

type TestReport struct {
	Name   string
	Status string
	Setup  *testreport.Setup
	Test   []testreport.Test
}
