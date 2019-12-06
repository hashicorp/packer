package optional

type String struct {
	isSet bool
	value string
}

func NewString(value string) String {
	return String{
		true,
		value,
	}
}

// EmptyString returns a new String that does not have a value set.
func EmptyString() String {
	return String{
		false,
		"",
	}
}

func (b String) IsSet() bool {
	return b.isSet
}

func (b String) Value() string {
	return b.value
}

func (b String) Default(defaultValue string) string {
	if b.isSet {
		return b.value
	}
	return defaultValue
}
