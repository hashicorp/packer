package opc

// String returns a pointer to a string
func String(v string) *string {
	return &v
}

// Int returns a pointer to an int
func Int(v int) *int {
	return &v
}
