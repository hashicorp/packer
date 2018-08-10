package sed 

import (
	"fmt"
	"regexp"
)

// conditions are what I'm calling the '1,10' in
// commands ike '1,10 d'.  They are the line numbers,
// regexps, and '$' that you can use to control when
// commands execute.

type condition interface {
	isMet(svm *vm) bool
}

// -----------------------------------------------------
type numbercond int // for matching line number conditions

func (n numbercond) isMet(svm *vm) bool {
	return svm.lineno == int(n)
}

// -----------------------------------------------------
type eofcond struct{} // for matching the condition '$'

func (_ eofcond) isMet(svm *vm) bool {
	return svm.lastl
}

// -----------------------------------------------------
type regexpcond struct {
	re *regexp.Regexp // for matching regexp conditions
}

func (r *regexpcond) isMet(svm *vm) (answer bool) {
	return r.re.MatchString(svm.pat)
}

func newRECondition(s string, loc *location) (*regexpcond, error) {
	re, err := regexp.Compile(s)
	if err != nil {
		err = fmt.Errorf("Regexp Error: %s %v", err.Error(), loc)
	}
	return &regexpcond{re}, err
}
