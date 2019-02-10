package optional

type Float32 struct {
	isSet bool
	value float32
}

func NewFloat32(value float32) Float32 {
	return Float32{
		true,
		value,
	}
}

// EmptyFloat32 returns a new Float32 that does not have a value set.
func EmptyFloat32() Float32 {
	return Float32{
		false,
		0,
	}
}

func (i Float32) IsSet() bool {
	return i.isSet
}

func (i Float32) Value() float32 {
	return i.value
}

func (i Float32) Default(defaultValue float32) float32 {
	if i.isSet {
		return i.value
	}
	return defaultValue
}
