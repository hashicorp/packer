package optional

type Byte struct {
	isSet bool
	value byte
}

func NewByte(value byte) Byte {
	return Byte{
		true,
		value,
	}
}

// EmptyByte returns a new Byte that does not have a value set.
func EmptyByte() Byte {
	return Byte{
		false,
		0,
	}
}

func (b Byte) IsSet() bool {
	return b.isSet
}

func (b Byte) Value() byte {
	return b.value
}

func (b Byte) Default(defaultValue byte) byte {
	if b.isSet {
		return b.value
	}
	return defaultValue
}
