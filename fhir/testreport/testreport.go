package testreport

type Setup struct {
	Action []Action
}

func (s Setup) Failed() bool {
	if l := len(s.Action); l > 0 {
		return s.Action[l-1].Failed()
	}
	return false
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
	Operation *ActionResult
	Assert    *ActionResult
}

func (a Action) Failed() bool {
	if operation := a.Operation; operation != nil {
		return operation.Failed()
	} else if assert := a.Assert; assert != nil {
		return assert.Failed()
	}
	return false
}

type ActionResult struct {
	Result  ActionResultCode
	Message string
	Detail  string
}

func (r ActionResult) Failed() bool {
	switch r.Result {
	case Fail, Error:
		return true
	}
	return false
}

type ActionResultCode int

const (
	Pass ActionResultCode = iota
	Skip
	Fail
	Warning
	Error
)

func (c ActionResultCode) String() string {
	switch c {
	case Pass:
		return "PASS"
	case Skip:
		return "SKIP"
	case Fail:
		return "FAIL"
	case Warning:
		return "WARNING"
	case Error:
		return "ERROR"
	default:
		return "<Unknown>"
	}
}
