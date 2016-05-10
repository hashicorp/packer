package common

import "testing"

func TestShouldDumpPublicValues(t *testing.T) {
	type S struct {
		MyString string
		myString string
	}

	data := &S{
		MyString: "value1",
	}

	dumps := make([]string, 0, 2)
	DumpConfig(data, func(s string) { dumps = append(dumps, s) })

	if len(dumps) != 1 {
		t.Fatalf("Expected len(dumps) to be 1, but got %d", len(dumps))

	}
	if dumps[0] != "MyString=value1" {
		t.Errorf("Expected dumps[0] to be 'MyString=value1', but got %s", dumps[0])
	}
}

func TestShouldOnlyDumpStrings(t *testing.T) {
	type S struct {
		MyString        string
		MyInt           int
		MyFloat32       float32
		MyStringPointer *string
	}

	s := "value1"
	data := &S{
		MyString:        s,
		MyInt:           1,
		MyFloat32:       2.0,
		MyStringPointer: &s,
	}

	dumps := make([]string, 0, 4)
	DumpConfig(data, func(s string) { dumps = append(dumps, s) })

	if len(dumps) != 1 {
		t.Fatalf("Expected len(dumps) to be 1, but got %d", len(dumps))

	}
	if dumps[0] != "MyString=value1" {
		t.Errorf("Expected dumps[0] to be 'MyString=value1', but got %s", dumps[0])
	}
}

func TestDumpShouldMaskSensitiveFieldValues(t *testing.T) {
	type S struct {
		MyString        string
		MySecret        string
		MySecretValue   string
		MyPassword      string
		MyPasswordValue string
	}

	data := &S{
		MyString:        "my-string",
		MySecret:        "s3cr3t",
		MySecretValue:   "s3cr3t-value",
		MyPassword:      "p@ssw0rd",
		MyPasswordValue: "p@ssw0rd-value",
	}

	dumps := make([]string, 0, 5)
	DumpConfig(data, func(s string) { dumps = append(dumps, s) })

	if len(dumps) != 5 {
		t.Fatalf("Expected len(dumps) to be 5, but got %d", len(dumps))

	}
	if dumps[0] != "MyString=my-string" {
		t.Errorf("Expected dumps[0] to be 'MyString=my-string', but got %s", dumps[0])
	}
	if dumps[1] != "MySecret=******" {
		t.Errorf("Expected dumps[1] to be 'MySecret=******', but got %s", dumps[1])
	}
	if dumps[2] != "MySecretValue=************" {
		t.Errorf("Expected dumps[2] to be 'MySecret=************', but got %s", dumps[2])
	}
	if dumps[3] != "MyPassword=********" {
		t.Errorf("Expected dumps[3] to be 'MyPassword=******** but got %s", dumps[3])
	}
	if dumps[4] != "MyPasswordValue=**************" {
		t.Errorf("Expected dumps[4] to be 'MyPasswordValue=**************' but got %s", dumps[4])
	}
}

func TestDumpConfigShouldDumpTopLevelValuesOnly(t *testing.T) {
	type N struct {
		NestedString string
	}

	type S struct {
		Nested1  N
		MyString string
		Nested2  N
	}

	data := &S{
		Nested1: N{
			NestedString: "nested-string1",
		},
		MyString: "my-string",
		Nested2: N{
			NestedString: "nested-string2",
		},
	}

	dumps := make([]string, 0, 1)
	DumpConfig(data, func(s string) { dumps = append(dumps, s) })

	if len(dumps) != 1 {
		t.Fatalf("Expected len(dumps) to be 1, but got %d", len(dumps))

	}
	if dumps[0] != "MyString=my-string" {
		t.Errorf("Expected dumps[0] to be 'MyString=my-string', but got %s", dumps[0])
	}
}
