package optional

type Uint8 struct {
	isSet bool
	value uint8
}

func NewUint8(value uint8) Uint8 {
	return Uint8{
		true,
		value,
	}
}

// EmptyUint8 returns a new Uint8 that does not have a value set.
func EmptyUint8() Uint8 {
	return Uint8{
		false,
		0,
	}
}

func (i Uint8) IsSet() bool {
	return i.isSet
}

func (i Uint8) Value() uint8 {
	return i.value
}

func (i Uint8) Default(defaultValue uint8) uint8 {
	if i.isSet {
		return i.value
	}
	return defaultValue
}
