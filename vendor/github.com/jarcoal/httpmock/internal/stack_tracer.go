package internal

import (
	"bytes"
	"fmt"
	"net/http"
	"runtime"
	"strings"
)

type StackTracer struct {
	CustomFn func(...interface{})
	Err      error
}

func (n StackTracer) Error() string {
	if n.Err == nil {
		return ""
	}
	return n.Err.Error()
}

// CheckStackTracer checks for specific error returned by
// NewNotFoundResponder function or Trace Responder method.
func CheckStackTracer(req *http.Request, err error) error {
	if nf, ok := err.(StackTracer); ok {
		if nf.CustomFn != nil {
			pc := make([]uintptr, 128)
			npc := runtime.Callers(2, pc)
			pc = pc[:npc]

			var mesg bytes.Buffer
			var netHTTPBegin, netHTTPEnd bool

			// Start recording at first net/http call if any...
			for {
				frames := runtime.CallersFrames(pc)

				var lastFn string
				for {
					frame, more := frames.Next()

					if !netHTTPEnd {
						if netHTTPBegin {
							netHTTPEnd = !strings.HasPrefix(frame.Function, "net/http.")
						} else {
							netHTTPBegin = strings.HasPrefix(frame.Function, "net/http.")
						}
					}

					if netHTTPEnd {
						if lastFn != "" {
							if mesg.Len() == 0 {
								if nf.Err != nil {
									mesg.WriteString(nf.Err.Error())
								} else {
									fmt.Fprintf(&mesg, "%s %s", req.Method, req.URL)
								}
								mesg.WriteString("\nCalled from ")
							} else {
								mesg.WriteString("\n  ")
							}
							fmt.Fprintf(&mesg, "%s()\n    at %s:%d", lastFn, frame.File, frame.Line)
						}
					}
					lastFn = frame.Function

					if !more {
						break
					}
				}

				// At least one net/http frame found
				if mesg.Len() > 0 {
					break
				}
				netHTTPEnd = true // retry without looking at net/http frames
			}

			nf.CustomFn(mesg.String())
		}
		err = nf.Err
	}
	return err
}
