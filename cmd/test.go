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

package cmd

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/PaesslerAG/jsonpath"
	"github.com/life-research/blazectl/fhir"
	"github.com/life-research/blazectl/fhir/testreport"
	"github.com/life-research/blazectl/fhir/testscript"
	"github.com/life-research/blazectl/fhirpath"
	"github.com/logrusorgru/aurora"
	"github.com/spf13/cobra"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// testCmd represents the test command
var testCmd = &cobra.Command{
	Use:       "test",
	Short:     "Execute test scripts",
	Long:      `Executes all test scripts located in a directory.`,
	ValidArgs: []string{"directory"},
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) < 1 {
			return errors.New("requires a directory argument")
		}
		return checkDir(args[0])
	},
	Run: func(cmd *cobra.Command, args []string) {
		dir := args[0]

		files, err := ioutil.ReadDir(dir)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		fmt.Printf("Found %d files\n", len(files))

		client := &fhir.Client{Base: server}

		for _, file := range files {
			testScript, err := readTestScript(filepath.Join(dir, file.Name()))
			if err != nil {
				switch err.(type) {
				case fhir.InvalidResourceTypeError:
				default:
					fmt.Println(err)
					os.Exit(1)
				}
			} else {
				report, err := executeTestScript(client, dir, testScript)
				if err != nil {
					fmt.Println(err)
					os.Exit(1)
				}
				printReport(report)
			}
		}
	},
}

func readTestScript(filename string) (fhir.TestScript, error) {
	file, err := os.Open(filename)
	if err != nil {
		return fhir.TestScript{}, err
	}
	defer file.Close()
	return fhir.ReadTestScript(file)
}

func executeTestScript(client *fhir.Client, dir string, testScript fhir.TestScript) (fhir.TestReport, error) {
	report := fhir.TestReport{Name: reportName(testScript)}
	fixtures, err := loadFixtures(dir, testScript.Fixture)
	if err != nil {
		return report, err
	}
	variables, err := prepareVariables(testScript.Variable)
	if err != nil {
		return report, err
	}
	ctx := context{client: client, fixtures: fixtures, variables: variables}
	if setup := testScript.Setup; setup != nil {
		setupReport := executeSetup(ctx, *setup)
		report.Setup = &setupReport
	}
	if report.Setup.Failed() {
		return report, nil
	}
	for _, test := range testScript.Test {
		report.Test = append(report.Test, executeTest(ctx, test))
	}
	return report, nil
}

func reportName(script fhir.TestScript) string {
	if title := script.Title; title != "" {
		return title
	}
	return script.Name
}

func loadFixtures(dir string, fixtures []testscript.Fixture) (fixtures, error) {
	result := make(map[string]fixture)
	for _, fix := range fixtures {
		if id := fix.Id; id != "" {
			if resourceRef := fix.Resource; resourceRef != nil {
				filename := filepath.Join(dir, resourceRef.Reference)
				info, err := os.Stat(filename)
				if err != nil {
					return nil, err
				}
				if info.IsDir() {
					return nil, fmt.Errorf("fixture `%s` resolved to a directory", resourceRef.Reference)
				}
				resource, err := readFixture(filename)
				if err != nil {
					return nil, err
				}
				result[id] = fixture{resource: resource}
			}
		}
	}
	return result, nil
}

func readFixture(filename string) (map[string]interface{}, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	return fhir.ReadGeneric(file)
}

func prepareVariables(variables []testscript.Variable) (variables, error) {
	result := make(map[string]variable)
	for _, v := range variables {
		if name := v.Name; name != "" {
			if v.Expression != "" {
				expr, err := fhirpath.Parse(v.Expression)
				if err != nil {
					return result, fmt.Errorf("error while parsing FHIRPath expression `%s` of vaiable `%s`: %s", v.Expression, name, err)
				}
				result[name] = variable{expression: expr, sourceId: v.SourceId}
			}
		}
	}
	return result, nil
}

func executeSetup(ctx context, setup testscript.Setup) testreport.Setup {
	report := testreport.Setup{}
	for _, action := range setup.Action {
		var actionReport testreport.Action
		ctx, actionReport = executeAction(ctx, action)
		report.Action = append(report.Action, actionReport)
		if actionReport.Failed() {
			break
		}
	}
	return report
}

func executeTest(ctx context, test testscript.Test) testreport.Test {
	report := testreport.Test{Name: test.Name, Description: test.Description}
	for _, action := range test.Action {
		var actionReport testreport.Action
		ctx, actionReport = executeAction(ctx, action)
		report.Action = append(report.Action, actionReport)
		if actionReport.Failed() {
			break
		}
	}
	return report
}

func executeAction(ctx context, action testscript.Action) (context, testreport.Action) {
	if operation := action.Operation; operation != nil {
		ctx, actionResult := executeOperation(ctx, *operation)
		return ctx, testreport.Action{
			Operation: &actionResult,
		}
	} else if assert := action.Assert; assert != nil {
		actionResult := executeAssert(ctx, *assert)
		return ctx, testreport.Action{
			Assert: &actionResult,
		}
	}
	panic("TestScript action without operation or assert")
}

func executeOperation(ctx context, op testscript.Operation) (context, testreport.ActionResult) {
	switch code := op.Type.Code; code {
	case "history":
		return executeHistoryOperation(ctx, op)
	case "create":
		return executeCreateOperation(ctx, op)
	default:
		return ctx, errorf("unsupported operation type `%s`", code)
	}
}

func executeHistoryOperation(ctx context, op testscript.Operation) (context, testreport.ActionResult) {
	req, err := newHistoryRequest(ctx, op)
	if err != nil {
		return ctx, errorf("Error while creating system history request: %s", err)
	}
	resp, err := ctx.client.Do(req)
	if err != nil {
		return ctx, errorf("Error while performing system history request: %s", err)
	}
	defer resp.Body.Close()

	resource, err := fhir.ReadGeneric(resp.Body)
	if err != nil {
		return ctx, errorf("Error while reading a system history response: %s", err)
	}
	if responseId := op.ResponseId; responseId != "" {
		ctx.fixtures[responseId] = fixture{resp.Header, resource}
	}
	ctx.lastResult = &operationResult{req, resp, resource}
	return ctx, pass("Successful system history request")
}

func newHistoryRequest(ctx context, operation testscript.Operation) (*http.Request, error) {
	url, err := historyRequestUrl(ctx, operation)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Accept", "application/fhir+json")
	return req, nil
}

func historyRequestUrl(ctx context, operation testscript.Operation) (string, error) {
	if url := operation.Url; url != "" {
		replacement, err := replaceVariables(ctx, url)
		if err != nil {
			return "", fmt.Errorf("error while replacing variables in operation.url `%s`: %s", url, err)
		}
		return replacement, nil
	}

	if params := operation.Params; params != "" {
		replacement, err := replaceVariables(ctx, params)
		if err != nil {
			return "", fmt.Errorf("error while replacing variables in operation.params `%s`: %s", params, err)
		}
		return historyBaseUrl(ctx.client, operation) + replacement, nil
	}

	return historyBaseUrl(ctx.client, operation) + "/_history", nil
}

func historyBaseUrl(client *fhir.Client, operation testscript.Operation) string {
	if resource := operation.Resource; resource != "" {
		return client.Base + "/" + resource
	}
	return client.Base
}

func executeCreateOperation(ctx context, op testscript.Operation) (context, testreport.ActionResult) {
	if sourceId := op.SourceId; sourceId != "" {
		if resource := ctx.fixtures[sourceId].resource; resource != nil {
			if resourceType, ok := resource["resourceType"].(string); ok {
				payload, err := json.Marshal(resource)
				if err != nil {
					return ctx, errorf("Error while preparing create request: %s", err)
				}
				req, err := ctx.client.NewCreateRequest(resourceType, bytes.NewReader(payload))
				if err != nil {
					return ctx, errorf("Error while creating create request: %s", err)
				}
				resp, err := ctx.client.Do(req)
				if err != nil {
					return ctx, errorf("Error while performing create request: %s", err)
				}
				defer resp.Body.Close()

				resource, err := fhir.ReadGeneric(resp.Body)
				if err != nil {
					return ctx, errorf("Error while reading create response: %s", err)
				}
				ctx.lastResult = &operationResult{req, resp, resource}
				return ctx, passf("Successful create %s request", resourceType)
			} else {
				return ctx, errorf("fixture `%s` has no resource type", sourceId)
			}
		} else {
			return ctx, errorf("fixture `%s` not found", sourceId)
		}
	} else {
		return ctx, errorf("create operation without source id")
	}
}

var varRegExp = regexp.MustCompile("\\${([^}]+)}")

func FindAllVarNames(s string) []string {
	allMatches := varRegExp.FindAllStringSubmatch(s, -1)
	subMatches := make([]string, 0, len(allMatches))
	for _, matches := range allMatches {
		subMatches = append(subMatches, matches[1])
	}
	return subMatches
}

// Replaces variables in string `s` with data from fixtures and variables
func replaceVariables(ctx context, s string) (string, error) {
	replacements := make(map[string]string)
	for _, name := range FindAllVarNames(s) {
		if variable, ok := ctx.variables[name]; ok {
			if sourceId := variable.sourceId; sourceId != "" {
				if resource := ctx.fixtures[sourceId].resource; resource != nil {
					result, err := variable.expression.Eval([]interface{}{resource})
					if err != nil {
						return "", fmt.Errorf("error while evaluating FHIRPath expression `%s`: %s", variable.expression, err)
					}
					if len(result) != 1 {
						return "", fmt.Errorf("FHIRPath expression `%s` didn't evaluate to a single value. Instead %d values were returned", variable.expression, len(result))
					}
					replacements["${"+name+"}"] = fmt.Sprint(result[0])
				} else {
					return "", fmt.Errorf("missing resource with sourceId `%s` in variable `%s`", sourceId, name)
				}
			} else {
				return "", fmt.Errorf("variable `%s` without sourceId", name)
			}
		} else {
			return "", fmt.Errorf("missing Variable `%s`", name)
		}
	}

	return varRegExp.ReplaceAllStringFunc(s, func(name string) string {
		return replacements[name]
	}), nil
}

func executeAssert(ctx context, assert testscript.Assert) testreport.ActionResult {
	op := assert.Operator

	if expectedResourceType := assert.Resource; expectedResourceType != "" {
		if resource := ctx.lastResult.resource; resource != nil {
			if actualResourceType, ok := resource["resourceType"].(string); ok {
				if res, err := op.Comp(actualResourceType, expectedResourceType); err != nil {
					return errorf("Error while comparing: %s", err)
				} else if !res {
					return failf("Expect response resource type of `%s` but was `%s`.", expectedResourceType, actualResourceType)
				}
			} else {
				return failf("Expect response resource type of `%s` but the response isn't a resource.", expectedResourceType)
			}
		} else {
			return failf("Expect response resource type of `%s` but no response resource was available.", expectedResourceType)
		}
	}

	if expectedResponse := assert.Response; expectedResponse != nil {
		if res, err := op.IntComp(expectedResponse.Code(), ctx.lastResult.response.StatusCode); err != nil {
			return errorf("Error while comparing: %s", err)
		} else if !res {
			return failf(statusCodeFormat(op), expectedResponse, ctx.lastResult.response.Status)
		}
	}

	// assert by expression
	if expression := assert.Expression; expression != "" {
		expr, err := fhirpath.Parse(expression)
		if err != nil {
			return errorf("Error while parsing FHIRPath expression: `%s`", expression)
		}
		if resource := ctx.lastResult.resource; resource != nil {
			result, err := expr.Eval([]interface{}{resource})
			if err != nil {
				return errorf("Error while evaluating FHIRPath expression `%s`: %s", expression, err)
			}
			if len(result) != 1 {
				return errorf("FHIRPath expression `%s` didn't evaluate to a single value. Instead %d values were returned.", expression, len(result))
			}
			if res, ok := result[0].(bool); ok {
				if !res {
					return failf("FHIRPath expression `%s` didn't evaluate to true.", expression)
				}
			} else {
				return errorf("FHIRPath expression `%s` didn't evaluate to a boolean value. Instead the value was `%v`.", expression, result[0])
			}
		} else {
			return errorf("Given FHIRPath expression but response resource is missing.")
		}
	}

	// assert by path
	if path := assert.Path; path != "" {
		if resource := ctx.lastResult.resource; resource != nil {
			if value := assert.Value; value != "" {
				value, err := replaceVariables(ctx, value)
				if err != nil {
					return errorf("error while replacing variables in value of `%s`: %s", assert.Description, err)
				}
				result, err := jsonpath.Get(path, resource)
				if err != nil {
					return errorf("Error while evaluating JSONPath expression `%s`: %s", path, err)
				}
				if value != fmt.Sprint(result) {
					return failf("Expected `%s` but got `%+v`", value, result)
				}
			} else {
				return errorf("Given JSONPath expression but expected value is missing.")
			}
		} else {
			return errorf("Given JSONPath expression but response resource is missing.")
		}
	}

	return pass(assert.Description)
}

func statusCodeFormat(op testscript.AssertionOperator) string {
	switch op {
	case testscript.Equals:
		return "Expect response status code equal to `%s` but was `%s`."
	case testscript.NotEquals:
		return "Expect response status code not equal to `%s` but was `%s`."
	default:
		return "Unknown comparison between `%s` and `%s`."
	}
}

func pass(msg string) testreport.ActionResult {
	return testreport.ActionResult{Result: testreport.Pass, Message: msg}
}

func passf(format string, a ...interface{}) testreport.ActionResult {
	return testreport.ActionResult{
		Result:  testreport.Pass,
		Message: fmt.Sprintf(format, a...),
	}
}

func failf(format string, a ...interface{}) testreport.ActionResult {
	return actionResult(testreport.Fail, format, a...)
}

func errorf(format string, a ...interface{}) testreport.ActionResult {
	return actionResult(testreport.Error, format, a...)
}

func actionResult(code testreport.ActionResultCode, format string, a ...interface{}) testreport.ActionResult {
	return testreport.ActionResult{
		Result:  code,
		Message: fmt.Sprintf(format, a...),
	}
}

var keyValueFormat = "%-11s : %s\n"
var divider = strings.Repeat("-", 72)

func printReport(report fhir.TestReport) {
	fmt.Println()
	fmt.Printf(keyValueFormat, aurora.Bold("Report"), aurora.Bold(report.Name))
	fmt.Println(aurora.Bold(divider))
	if setup := report.Setup; setup != nil {
		fmt.Println("Setup")
		fmt.Println(divider)
		printActions(setup.Action)
		fmt.Println(divider)
	}
	testName := func(test testreport.Test) string {
		if name := test.Name; name != "" {
			return name
		}
		return "<unknown>"
	}
	for _, test := range report.Test {
		fmt.Println("Test")
		fmt.Println(divider)
		fmt.Printf(keyValueFormat, "Name", testName(test))
		if desc := test.Description; desc != "" {
			fmt.Printf(keyValueFormat, "Description", desc)
		}
		fmt.Println(divider)
		printActions(test.Action)
		fmt.Println(divider)
	}
}

func printActions(actions []testreport.Action) {
	for _, action := range actions {
		if result := action.Operation; result != nil {
			printActionResult(result)
		} else if result := action.Assert; result != nil {
			printActionResult(result)
		}
	}
}

func printActionResult(result *testreport.ActionResult) {
	fmt.Printf(keyValueFormat, aurora.Bold(colorResultCode(result.Result)), result.Message)
}

func colorResultCode(result testreport.ActionResultCode) aurora.Value {
	switch result {
	case testreport.Pass:
		return aurora.Green(result)
	case testreport.Skip:
		return aurora.Blue(result)
	case testreport.Fail:
		return aurora.Red(result)
	case testreport.Warning:
		return aurora.Yellow(result)
	case testreport.Error:
		return aurora.Red(result)
	default:
		return aurora.Black(result)
	}
}

func init() {
	rootCmd.AddCommand(testCmd)
}

type context struct {
	client     *fhir.Client
	fixtures   fixtures
	variables  variables
	lastResult *operationResult
}

type fixtures map[string]fixture

type fixture struct {
	headers  http.Header
	resource map[string]interface{}
}

type variables map[string]variable

type variable struct {
	expression fhirpath.Expression
	sourceId   string
}

type operationResult struct {
	request  *http.Request
	response *http.Response
	resource map[string]interface{}
}
