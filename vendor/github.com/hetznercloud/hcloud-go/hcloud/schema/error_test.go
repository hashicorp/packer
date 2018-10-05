package schema

import (
	"encoding/json"
	"testing"
)

func TestError(t *testing.T) {
	t.Run("UnmarshalJSON", func(t *testing.T) {
		data := []byte(`{
			"code": "invalid_input",
			"message": "invalid input",
			"details": {
				"fields": [
					{
						"name": "broken_field",
						"messages": ["is required"]
					}
				]
			}
		}`)

		e := &Error{}
		err := json.Unmarshal(data, e)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if e.Code != "invalid_input" {
			t.Errorf("unexpected Code: %v", e.Code)
		}
		if e.Message != "invalid input" {
			t.Errorf("unexpected Message: %v", e.Message)
		}
		if e.Details == nil {
			t.Fatalf("unexpected Details: %v", e.Details)
		}
		d, ok := e.Details.(ErrorDetailsInvalidInput)
		if !ok {
			t.Fatalf("unexpected Details type (should be ErrorDetailsInvalidInput): %v", e.Details)
		}
		if len(d.Fields) != 1 {
			t.Fatalf("unexpected Details.Fields length (should be 1): %v", d.Fields)
		}
		if d.Fields[0].Name != "broken_field" {
			t.Errorf("unexpected Details.Fields[0].Name: %v", d.Fields[0].Name)
		}
		if len(d.Fields[0].Messages) != 1 {
			t.Fatalf("unexpected Details.Fields[0].Messages length (should be 1): %v", d.Fields[0].Messages)
		}
		if d.Fields[0].Messages[0] != "is required" {
			t.Errorf("unexpected Details.Fields[0].Messages[0]: %v", d.Fields[0].Messages[0])
		}
	})
}
