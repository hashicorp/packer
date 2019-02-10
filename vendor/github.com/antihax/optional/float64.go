package optional

type Float64 struct {
	isSet bool
	value float64
}

func NewFloat64(value float64) Float64 {
	return Float64{
		true,
		value,
	}
}

// EmptyFloat64 returns a new Float64 that does not have a value set.
func EmptyFloat64() Float64 {
	return Float64{
		false,
		0,
	}
}

func (i Float64) IsSet() bool {
	return i.isSet
}

func (i Float64) Value() float64 {
	return i.value
}

func (i Float64) Default(defaultValue float64) float64 {
	if i.isSet {
		return i.value
	}
	return defaultValue
}
