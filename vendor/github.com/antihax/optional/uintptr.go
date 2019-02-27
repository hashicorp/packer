package optional

type Uintptr struct {
	isSet bool
	value uintptr
}

func NewUintptr(value uintptr) Uintptr {
	return Uintptr{
		true,
		value,
	}
}

// EmptyUintptr returns a new Uintptr that does not have a value set.
func EmptyUintptr() Uintptr {
	return Uintptr{
		false,
		0,
	}
}

func (i Uintptr) IsSet() bool {
	return i.isSet
}

func (i Uintptr) Value() uintptr {
	return i.value
}

func (i Uintptr) Default(defaultValue uintptr) uintptr {
	if i.isSet {
		return i.value
	}
	return defaultValue
}
