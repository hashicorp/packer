package optional

type Complex128 struct {
	isSet bool
	value complex128
}

func NewComplex128(value complex128) Complex128 {
	return Complex128{
		true,
		value,
	}
}

// EmptyComplex128 returns a new Complex128 that does not have a value set.
func EmptyComplex128() Complex128 {
	return Complex128{
		false,
		0,
	}
}

func (i Complex128) IsSet() bool {
	return i.isSet
}

func (i Complex128) Value() complex128 {
	return i.value
}

func (i Complex128) Default(defaultValue complex128) complex128 {
	if i.isSet {
		return i.value
	}
	return defaultValue
}
