package optional

type Complex64 struct {
	isSet bool
	value complex64
}

func NewComplex64(value complex64) Complex64 {
	return Complex64{
		true,
		value,
	}
}

// EmptyComplex64 returns a new Complex64 that does not have a value set.
func EmptyComplex64() Complex64 {
	return Complex64{
		false,
		0,
	}
}

func (i Complex64) IsSet() bool {
	return i.isSet
}

func (i Complex64) Value() complex64 {
	return i.value
}

func (i Complex64) Default(defaultValue complex64) complex64 {
	if i.isSet {
		return i.value
	}
	return defaultValue
}
