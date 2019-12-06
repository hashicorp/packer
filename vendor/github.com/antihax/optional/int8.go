package optional

type Int8 struct {
	isSet bool
	value int8
}

func NewInt8(value int8) Int8 {
	return Int8{
		true,
		value,
	}
}

// EmptyInt8 returns a new Int8 that does not have a value set.
func EmptyInt8() Int8 {
	return Int8{
		false,
		0,
	}
}

func (i Int8) IsSet() bool {
	return i.isSet
}

func (i Int8) Value() int8 {
	return i.value
}

func (i Int8) Default(defaultValue int8) int8 {
	if i.isSet {
		return i.value
	}
	return defaultValue
}
