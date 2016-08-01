package approvaltests

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"
	"runtime"
	"strings"
)

type approvalName struct {
	pc       uintptr
	fullName string
	name     string
	fileName string
	fileLine int
}

func newApprovalName(pc uintptr, f *runtime.Func) (*approvalName, error) {
	namer := &approvalName{
		pc:       pc,
		fullName: f.Name(),
	}

	namer.fileName, namer.fileLine = f.FileLine(pc)

	splits := strings.Split(namer.fullName, ".")
	namer.name = splits[len(splits)-1]

	return namer, nil
}

// Walk the call stack, and try to find the test method that was executed.
// The test method is identified by looking for the test runner, which is
// *assumed* to be common across all callers.  The test runner has a Name() of
// 'testing.tRunner'.  The method immediately previous to this is the test
// method.
func getApprovalName() (*approvalName, error) {
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
	return newApprovalName(pc[i-1], testMethod)
}

func isTestRunner(f *runtime.Func) bool {
	return f != nil && f.Name() == "testing.tRunner"
}

func (s *approvalName) compare(approvalFile, receivedFile string, reader io.Reader) error {
	received, err := ioutil.ReadAll(reader)
	if err != nil {
		return err
	}

	// Ideally, this should only be written if
	//  1. the approval file does not exist
	//  2. the results differ
	err = s.dumpReceivedTestResult(received, receivedFile)
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

	received = s.normalizeLineEndings(received)
	approved = s.normalizeLineEndings(approved)

	// The two sides are identical, nothing more to do.
	if bytes.Compare(received, approved) == 0 {
		return nil
	}

	return fmt.Errorf("failed to approved %s", s.name)
}

func (s *approvalName) normalizeLineEndings(bs []byte) []byte {
	return bytes.Replace(bs, []byte("\r\n"), []byte("\n"), -1)
}

func (s *approvalName) dumpReceivedTestResult(bs []byte, receivedFile string) error {
	err := ioutil.WriteFile(receivedFile, bs, 0644)

	return err
}

func (s *approvalName) getFileName(extWithDot string, suffix string) string {
	if !strings.HasPrefix(extWithDot, ".") {
		extWithDot = fmt.Sprintf(".%s", extWithDot)
	}

	baseName := path.Base(s.fileName)
	baseWithoutExt := baseName[:len(baseName)-len(path.Ext(s.fileName))]

	return fmt.Sprintf("%s.%s.%s%s", baseWithoutExt, s.name, suffix, extWithDot)
}

func (s *approvalName) getReceivedFile(extWithDot string) string {
	return s.getFileName(extWithDot, "received")
}

func (s *approvalName) getApprovalFile(extWithDot string) string {
	return s.getFileName(extWithDot, "approved")
}
