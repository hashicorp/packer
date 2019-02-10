package optional

type Int16 struct {
	isSet bool
	value int16
}

func NewInt16(value int16) Int16 {
	return Int16{
		true,
		value,
	}
}

// EmptyInt16 returns a new Int16 that does not have a value set.
func EmptyInt16() Int16 {
	return Int16{
		false,
		0,
	}
}

func (i Int16) IsSet() bool {
	return i.isSet
}

func (i Int16) Value() int16 {
	return i.value
}

func (i Int16) Default(defaultValue int16) int16 {
	if i.isSet {
		return i.value
	}
	return defaultValue
}
