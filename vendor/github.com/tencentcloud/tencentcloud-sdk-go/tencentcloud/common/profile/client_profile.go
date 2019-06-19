package profile

type ClientProfile struct {
	HttpProfile     *HttpProfile
	SignMethod      string
	UnsignedPayload bool
}

func NewClientProfile() *ClientProfile {
	return &ClientProfile{
		HttpProfile:     NewHttpProfile(),
		SignMethod:      "TC3-HMAC-SHA256",
		UnsignedPayload: false,
	}
}
