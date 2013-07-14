package fix

// A Fixer is something that can perform a fix operation on a template.
type Fixer interface {
	// Fix takes a raw map structure input, potentially transforms it
	// in some way, and returns the new, transformed structure. The
	// Fix method is allowed to mutate the input.
	Fix(input map[string]interface{}) (map[string]interface{}, error)
}

// Fixers is the map of all available fixers, by name.
var Fixers map[string]Fixer

func init() {
	Fixers = map[string]Fixer{
		"iso-md5": new(FixerISOMD5),
	}
}
