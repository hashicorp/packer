package optional

type Uint32 struct {
	isSet bool
	value uint32
}

func NewUint32(value uint32) Uint32 {
	return Uint32{
		true,
		value,
	}
}

// EmptyUint32 returns a new Uint32 that does not have a value set.
func EmptyUint32() Uint32 {
	return Uint32{
		false,
		0,
	}
}

func (i Uint32) IsSet() bool {
	return i.isSet
}

func (i Uint32) Value() uint32 {
	return i.value
}

func (i Uint32) Default(defaultValue uint32) uint32 {
	if i.isSet {
		return i.value
	}
	return defaultValue
}
