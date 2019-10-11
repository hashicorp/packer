package optional

type Int struct {
	isSet bool
	value int
}

func NewInt(value int) Int {
	return Int{
		true,
		value,
	}
}

// EmptyInt returns a new Int that does not have a value set.
func EmptyInt() Int {
	return Int{
		false,
		0,
	}
}

func (i Int) IsSet() bool {
	return i.isSet
}

func (i Int) Value() int {
	return i.value
}

func (i Int) Default(defaultValue int) int {
	if i.isSet {
		return i.value
	}
	return defaultValue
}
