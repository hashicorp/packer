package optional

type Rune struct {
	isSet bool
	value rune
}

func NewRune(value rune) Rune {
	return Rune{
		true,
		value,
	}
}

// EmptyRune returns a new Rune that does not have a value set.
func EmptyRune() Rune {
	return Rune{
		false,
		0,
	}
}

func (b Rune) IsSet() bool {
	return b.isSet
}

func (b Rune) Value() rune {
	return b.value
}

func (b Rune) Default(defaultValue rune) rune {
	if b.isSet {
		return b.value
	}
	return defaultValue
}
