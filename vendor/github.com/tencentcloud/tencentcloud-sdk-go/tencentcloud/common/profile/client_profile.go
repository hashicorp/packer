package profile

type ClientProfile struct {
	HttpProfile *HttpProfile
	SignMethod  string
}

func NewClientProfile() *ClientProfile {
	return &ClientProfile{
		HttpProfile: NewHttpProfile(),
		SignMethod:  "HmacSHA256",
	}
}
