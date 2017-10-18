package googlecompute

import "testing"

func Test_templateCleanImageName(t *testing.T) {
	vals := []struct {
		origName string
		expected string
	}{
		{
			origName: "abcde-012345xyz",
			expected: "abcde-012345xyz",
		},
		{
			origName: "ABCDE-012345xyz",
			expected: "abcde-012345xyz",
		},
		{
			origName: "abcde-012345v1.0.0",
			expected: "abcde-012345v1-0-0",
		},
		{
			origName: "a123456789012345678901234567890123456789012345678901234567890123456789",
			expected: "a12345678901234567890123456789012345678901234567890123456789012",
		},
		{
			origName: "01234567890123456789012345678901234567890123456789012345678901.",
			expected: "a1234567890123456789012345678901234567890123456789012345678901a",
		},
		{
			origName: "01234567890123456789012345678901234567890123456789012345678901-",
			expected: "a1234567890123456789012345678901234567890123456789012345678901a",
		},
	}

	for _, v := range vals {
		name := templateCleanImageName(v.origName)
		if name != v.expected {
			t.Fatalf("template names do not match: expected %s got %s\n", v.expected, name)
		}
	}
}
