package optional

type Int32 struct {
	isSet bool
	value int32
}

func NewInt32(value int32) Int32 {
	return Int32{
		true,
		value,
	}
}

// EmptyInt32 returns a new Int32 that does not have a value set.
func EmptyInt32() Int32 {
	return Int32{
		false,
		0,
	}
}

func (i Int32) IsSet() bool {
	return i.isSet
}

func (i Int32) Value() int32 {
	return i.value
}

func (i Int32) Default(defaultValue int32) int32 {
	if i.isSet {
		return i.value
	}
	return defaultValue
}
