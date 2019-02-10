package optional

type Uint64 struct {
	isSet bool
	value uint64
}

func NewUint64(value uint64) Uint64 {
	return Uint64{
		true,
		value,
	}
}

// EmptyUint64 returns a new Uint64 that does not have a value set.
func EmptyUint64() Uint64 {
	return Uint64{
		false,
		0,
	}
}

func (i Uint64) IsSet() bool {
	return i.isSet
}

func (i Uint64) Value() uint64 {
	return i.value
}

func (i Uint64) Default(defaultValue uint64) uint64 {
	if i.isSet {
		return i.value
	}
	return defaultValue
}
