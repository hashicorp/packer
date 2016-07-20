# ApprovalTests.go

ApprovalTests for go

[![Build Status](https://travis-ci.org/approvals/go-approval-tests.png?branch=master)](https://travis-ci.org/approvals/go-approval-tests)

# Golden master Verification Library
ApprovalTests allows for easy testing of larger objects, strings and anything else that can be saved to a file (images, sounds, csv,  etc...)

#Examples
##In Project
Note: ApprovalTests uses approvaltests to test itself. Therefore there are many examples in the code itself.

 * [approvals_test.go](approvals_test.go)

##JSON
VerifyJSONBytes - Simple Formatting for easy comparison. Also uses the .json file extension 

```go
func TestVerifyJSON(t *testing.T) {
	jsonb := []byte("{ \"foo\": \"bar\", \"age\": 42, \"bark\": \"woof\" }")
	VerifyJSONBytes(t, jsonb)
}
```
Matches file: approvals_test.TestVerifyJSON.received.json

```json
{
  "age": 42,
  "bark": "woof",
  "foo": "bar"
}
```

##Reporters
ApprovalTests becomes *much* more powerful with reporters. Reporters launch programs on failure to help you understand, fix and approve results.

You can make your own easily, [here's an example](reporters/beyond_compare.go)
You can also declare which one to use. Either at the 
### Method level
```go
r := UseReporter(reporters.NewIntelliJ())
defer r.Close()
```
### Test Level
```go
func TestMain(m *testing.M) {
	r := UseReporter(reporters.NewBeyondCompareReporter())
	defer r.Close()

	m.Run()
}
```
