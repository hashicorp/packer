package common

type Credential struct {
	SecretId  string
	SecretKey string
}

func NewCredential(secretId, secretKey string) *Credential {
	return &Credential{
		SecretId:  secretId,
		SecretKey: secretKey,
	}
}

func (c *Credential) GetCredentialParams() map[string]string {
	return map[string]string{
		"SecretId": c.SecretId,
	}
}

type TokenCredential struct {
	SecretId  string
	SecretKey string
	Token     string
}

func NewTokenCredential(secretId, secretKey, token string) *TokenCredential {
	return &TokenCredential{
		SecretId:  secretId,
		SecretKey: secretKey,
		Token:     token,
	}
}

func (c *TokenCredential) GetCredentialParams() map[string]string {
	return map[string]string{
		"SecretId": c.SecretId,
		"Token":    c.Token,
	}
}
