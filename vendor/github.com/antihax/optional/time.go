package optional

import "time"

type Time struct {
	isSet bool
	value time.Time
}

func NewTime(value time.Time) Time {
	return Time{
		true,
		value,
	}
}

// EmptyTime returns a new Time that does not have a value set.
func EmptyTime() Time {
	return Time{
		false,
		time.Time{},
	}
}

func (b Time) IsSet() bool {
	return b.isSet
}

func (b Time) Value() time.Time {
	return b.value
}

func (b Time) Default(defaultValue time.Time) time.Time {
	if b.isSet {
		return b.value
	}
	return defaultValue
}
