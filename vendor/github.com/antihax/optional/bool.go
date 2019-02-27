package optional

type Bool struct {
	isSet bool
	value bool
}

func NewBool(value bool) Bool {
	return Bool{
		true,
		value,
	}
}

// EmptyBool returns a new Bool that does not have a value set.
func EmptyBool() Bool {
	return Bool{
		false,
		false,
	}
}

func (b Bool) IsSet() bool {
	return b.isSet
}

func (b Bool) Value() bool {
	return b.value
}

func (b Bool) Default(defaultValue bool) bool {
	if b.isSet {
		return b.value
	}
	return defaultValue
}
