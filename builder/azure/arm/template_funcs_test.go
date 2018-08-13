package arm

import "testing"

func TestTemplateCleanImageName(t *testing.T) {
	vals := []struct {
		origName string
		expected string
	}{
		// test that valid name is unchanged
		{
			origName: "abcde-012345xyz",
			expected: "abcde-012345xyz",
		},
		// test that colons are converted to hyphens
		{
			origName: "abcde-012345v1.0:0",
			expected: "abcde-012345v1.0-0",
		},
		// Name starting with number is not valid, but not in scope of this
		// function to correct
		{
			origName: "012345v1.0:0",
			expected: "012345v1.0-0",
		},
		// Name over 80 chars is not valid, but not corrected by this function.
		{
			origName: "l012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789",
			expected: "l012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789",
		},
		// Name cannot end in a -Name over 80 chars is not valid, but not corrected by this function.
		{
			origName: "abcde-:_",
			expected: "abcde",
		},
		// Lost of special characters
		{
			origName: "My()./-_:&^ $%[]#'@name",
			expected: "My--.--_-----------name",
		},
	}

	for _, v := range vals {
		name := templateCleanImageName(v.origName)
		if name != v.expected {
			t.Fatalf("template names do not match: expected %s got %s\n", v.expected, name)
		}
	}
}
