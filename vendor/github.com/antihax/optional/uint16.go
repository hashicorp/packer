package optional

type Uint16 struct {
	isSet bool
	value uint16
}

func NewUint16(value uint16) Uint16 {
	return Uint16{
		true,
		value,
	}
}

// EmptyUint16 returns a new Uint16 that does not have a value set.
func EmptyUint16() Uint16 {
	return Uint16{
		false,
		0,
	}
}

func (i Uint16) IsSet() bool {
	return i.isSet
}

func (i Uint16) Value() uint16 {
	return i.value
}

func (i Uint16) Default(defaultValue uint16) uint16 {
	if i.isSet {
		return i.value
	}
	return defaultValue
}
