package testscript

import (
	"fmt"
	"github.com/life-research/blazectl/fhir/types"
	"strings"
)

type Variable struct {
	Name         string
	DefaultValue string
	Description  string
	Expression   string
	HeaderField  string
	Hint         string
	Path         string
	SourceId     string
}

type Fixture struct {
	Id         string
	Autocreate bool
	Autodelete bool
	Resource   *types.Reference
}

type Setup struct {
	Action []Action
}

type Test struct {
	Name        string
	Description string
	Action      []Action
}

type Teardown struct {
	Action []Action
}

type Action struct {
	Operation *Operation
	Assert    *Assert
}

type Operation struct {
	Type             *types.Coding
	Resource         string
	Label            string
	Description      string
	Accept           string
	ContentType      string
	Designation      int
	EncodeRequestUrl bool
	Method           string
	Origin           int
	Params           string
	RequestId        string
	ResponseId       string
	SourceId         string
	TargetId         string
	Url              string
}

type Assert struct {
	Label       string
	Description string
	Expression  string
	Operator    AssertionOperator
	Path        string
	Resource    string
	Response    *AssertionResponse
	Value       string
	WarningOnly bool
}

type AssertionOperator int

const (
	Equals AssertionOperator = iota
	NotEquals
	In
	NotIn
	GreaterThan
	LessThan
	Empty
	NotEmpty
	Contains
	NotContains
	Eval
)

func (op *AssertionOperator) UnmarshalJSON(json []byte) error {
	s := strings.Trim(string(json), `"`)
	switch s {
	case "equals":
		*op = Equals
	case "notEquals":
		*op = NotEquals
	case "in":
		*op = In
	case "notIn":
		*op = NotIn
	case "greaterThan":
		*op = GreaterThan
	case "lessThan":
		*op = LessThan
	case "empty":
		*op = Empty
	case "notEmpty":
		*op = NotEmpty
	case "contains":
		*op = Contains
	case "notContains":
		*op = NotContains
	case "eval":
		*op = Eval
	default:
		return fmt.Errorf("unknown assertion response type: %s", s)
	}
	return nil
}

func (op AssertionOperator) String() string {
	switch op {
	case Equals:
		return "equals"
	case NotEquals:
		return "notEquals"
	case In:
		return "in"
	case NotIn:
		return "notIn"
	case GreaterThan:
		return "greaterThan"
	case LessThan:
		return "lessThan"
	case Empty:
		return "empty"
	case NotEmpty:
		return "notEmpty"
	case Contains:
		return "vontains"
	case NotContains:
		return "notContains"
	case Eval:
		return "eval"
	default:
		return "<unknown>"
	}
}

func (op AssertionOperator) CompWithResult(value string, result []interface{}) (bool, error) {
	switch op {
	case Equals:
		return len(result) == 1 && value == fmt.Sprint(result[0]), nil
	case NotEquals:
		return len(result) == 1 && value != fmt.Sprint(result[0]), nil
	default:
		return false, fmt.Errorf("unsupported assertion operator `%s`", op)
	}
}

func (op AssertionOperator) Comp(a interface{}, b interface{}) (bool, error) {
	switch op {
	case Equals:
		return a == b, nil
	case NotEquals:
		return a != b, nil
	default:
		return false, fmt.Errorf("unsupported assertion operator `%s`", op)
	}
}

func (op AssertionOperator) IntComp(i1 int, i2 int) (bool, error) {
	switch op {
	case Equals:
		return i1 == i2, nil
	case NotEquals:
		return i1 != i2, nil
	case GreaterThan:
		return i1 > i2, nil
	case LessThan:
		return i1 < i2, nil
	default:
		return false, fmt.Errorf("unsupported assertion operator `%s`", op)
	}
}

type AssertionResponse int

const (
	Okay               AssertionResponse = 200
	Created            AssertionResponse = 201
	NoContent          AssertionResponse = 204
	NotModified        AssertionResponse = 304
	Bad                AssertionResponse = 400
	Forbidden          AssertionResponse = 403
	NotFound           AssertionResponse = 404
	MethodNotAllowed   AssertionResponse = 405
	Conflict           AssertionResponse = 409
	Gone               AssertionResponse = 410
	PreconditionFailed AssertionResponse = 412
	Unprocessable      AssertionResponse = 422
)

func (t *AssertionResponse) UnmarshalJSON(json []byte) error {
	s := strings.Trim(string(json), `"`)
	switch s {
	case "okay":
		*t = Okay
	case "created":
		*t = Created
	case "noContent":
		*t = NoContent
	case "notModified":
		*t = NotModified
	case "bad":
		*t = Bad
	case "forbidden":
		*t = Forbidden
	case "notFound":
		*t = NotFound
	case "methodNotAllowed":
		*t = MethodNotAllowed
	case "conflict":
		*t = Conflict
	case "gone":
		*t = Gone
	case "preconditionFailed":
		*t = PreconditionFailed
	case "unprocessable":
		*t = Unprocessable
	default:
		return fmt.Errorf("unknown assertion response type: %s", s)
	}
	return nil
}

// Code returns the HTTP status code of the AssertionResponse
func (t AssertionResponse) Code() int {
	return int(t)
}

func (t AssertionResponse) String() string {
	switch t {
	case Okay:
		return "Okay"
	case Created:
		return "Created"
	case NoContent:
		return "NoContent"
	case NotModified:
		return "NotModified"
	case Bad:
		return "Bad"
	case Forbidden:
		return "Forbidden"
	case NotFound:
		return "NotFound"
	case MethodNotAllowed:
		return "MethodNotAllowed"
	case Conflict:
		return "Conflict"
	case Gone:
		return "Gone"
	case PreconditionFailed:
		return "PreconditionFailed"
	case Unprocessable:
		return "Unprocessable"
	default:
		return "<unknown>"
	}
}
