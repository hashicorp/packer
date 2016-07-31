package winrm

import (
	"errors"

	. "gopkg.in/check.v1"
)

func (s *WinRMSuite) TestError(c *C) {

	err := errWinrm{
		message: "Some test error",
	}
	same := errors.New("Some test error")
	func(err, same error) {
		t, ok := err.(errWinrm)
		c.Assert(ok, Equals, true)
		c.Assert(t.Error(), Equals, same.Error())
	}(err, same)
}
