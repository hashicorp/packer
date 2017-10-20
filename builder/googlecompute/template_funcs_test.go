package googlecompute

import "testing"

func Test_templateCleanImageName(t *testing.T) {
	vals := []struct {
		origName string
		expected string
	}{
		// test that valid name is unchanged
		{
			origName: "abcde-012345xyz",
			expected: "abcde-012345xyz",
		},

		//test that capital letters are converted to lowercase
		{
			origName: "ABCDE-012345xyz",
			expected: "abcde-012345xyz",
		},
		// test that periods and colons are converted to hyphens
		{
			origName: "abcde-012345v1.0:0",
			expected: "abcde-012345v1-0-0",
		},
		// Name starting with number is not valid, but not in scope of this
		// function to correct
		{
			origName: "012345v1.0:0",
			expected: "012345v1-0-0",
		},
		// Name over 64 chars is not valid, but not corrected by this function.
		{
			origName: "loooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooong",
			expected: "loooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooong",
		},
	}

	for _, v := range vals {
		name := templateCleanImageName(v.origName)
		if name != v.expected {
			t.Fatalf("template names do not match: expected %s got %s\n", v.expected, name)
		}
	}
}
