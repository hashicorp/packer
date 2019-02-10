package optional

type Uint struct {
	isSet bool
	value uint
}

func NewUint(value uint) Uint {
	return Uint{
		true,
		value,
	}
}

// EmptyUint returns a new Uint that does not have a value set.
func EmptyUint() Uint {
	return Uint{
		false,
		0,
	}
}

func (i Uint) IsSet() bool {
	return i.isSet
}

func (i Uint) Value() uint {
	return i.value
}

func (i Uint) Default(defaultValue uint) uint {
	if i.isSet {
		return i.value
	}
	return defaultValue
}
