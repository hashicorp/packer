package hcloud

// String returns a pointer to the passed string s.
func String(s string) *string { return &s }

// Int returns a pointer to the passed integer i.
func Int(i int) *int { return &i }

// Bool returns a pointer to the passed bool b.
func Bool(b bool) *bool { return &b }
