package optional

// Optional represents a generic optional type, stored as an interface{}.
type Interface struct {
	isSet bool
	value interface{}
}

func NewInterface(value interface{}) Interface {
	return Interface{
		true,
		value,
	}
}

// EmptyInterface returns a new Interface that does not have a value set.
func EmptyInterface() Interface {
	return Interface{
		false,
		nil,
	}
}

func (b Interface) IsSet() bool {
	return b.isSet
}

func (b Interface) Value() interface{} {
	return b.value
}

func (b Interface) Default(defaultValue interface{}) interface{} {
	if b.isSet {
		return b.value
	}
	return defaultValue
}
