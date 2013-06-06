package ssh

// An implementation of ssh.ClientPassword so that you can use a static
// string password for the password to ClientAuthPassword.
type Password string

func (p Password) Password(user string) (string, error) {
	return string(p), nil
}
