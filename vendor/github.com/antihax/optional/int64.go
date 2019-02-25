package optional

type Int64 struct {
	isSet bool
	value int64
}

func NewInt64(value int64) Int64 {
	return Int64{
		true,
		value,
	}
}

// EmptyInt64 returns a new Int64 that does not have a value set.
func EmptyInt64() Int64 {
	return Int64{
		false,
		0,
	}
}

func (i Int64) IsSet() bool {
	return i.isSet
}

func (i Int64) Value() int64 {
	return i.value
}

func (i Int64) Default(defaultValue int64) int64 {
	if i.isSet {
		return i.value
	}
	return defaultValue
}
