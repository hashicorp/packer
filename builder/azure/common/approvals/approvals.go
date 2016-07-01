package approvals

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"runtime"
	"strings"
	"testing"
)

type testState struct {
	pc       uintptr
	fullName string
	name     string
	fileName string
	fileLine int
}

func Verify(t *testing.T, reader io.Reader) error {
	state, err := findTestMethod()
	if err != nil {
		return err
	}

	return state.compare(state.getApprovalFile(), reader)
}

func (s *testState) compare(approvalFile string, reader io.Reader) error {
	received, err := ioutil.ReadAll(reader)
	if err != nil {
		return err
	}

	// Ideally, this should only be written if
	//  1. the approval file does not exist
	//  2. the results differ
	err = s.dumpReceivedTestResult(received)
	if err != nil {
		return err
	}

	fh, err := os.Open(approvalFile)
	if err != nil {
		return err
	}
	defer fh.Close()

	approved, err := ioutil.ReadAll(fh)
	if err != nil {
		return err
	}

	// The two sides are identical, nothing more to do.
	if bytes.Compare(received, approved) == 0 {
		return nil
	}

	return fmt.Errorf("failed to approved %s", s.name)
}

func (s *testState) dumpReceivedTestResult(bs []byte) error {
	fn := s.getReceivedFile()
	err := ioutil.WriteFile(fn, bs, 0644)

	return err
}

func (s *testState) getReceivedFile() string {
	return fmt.Sprintf("%s.received.txt", s.name)
}

func (s *testState) getApprovalFile() string {
	return fmt.Sprintf("%s.approved.txt", s.name)
}

func newTestState(pc uintptr, f *runtime.Func) (*testState, error) {
	state := &testState{
		pc:       pc,
		fullName: f.Name(),
	}

	state.fileName, state.fileLine = f.FileLine(pc)

	splits := strings.Split(state.fullName, ".")
	state.name = splits[len(splits)-1]

	return state, nil
}

// Walk the call stack, and try to find the test method that was executed.
// The test method is identified by looking for the test runner, which is
// *assumed* to be common across all callers.  The test runner has a Name() of
// 'testing.tRunner'.  The method immediately previous to this is the test
// method.
func findTestMethod() (*testState, error) {
	pc := make([]uintptr, 100)
	count := runtime.Callers(0, pc)

	i := 0
	var lastFunc *runtime.Func

	for ; i < count; i++ {
		lastFunc = runtime.FuncForPC(pc[i])
		if isTestRunner(lastFunc) {
			break
		}
	}

	if i == 0 || !isTestRunner(lastFunc) {
		return nil, fmt.Errorf("approvals: could not find the test method")
	}

	testMethod := runtime.FuncForPC(pc[i-1])
	return newTestState(pc[i-1], testMethod)
}

func isTestRunner(f *runtime.Func) bool {
	return f != nil && f.Name() == "testing.tRunner"
}
