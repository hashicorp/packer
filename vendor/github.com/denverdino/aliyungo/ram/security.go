package ram

//TODO implement ram api about security
/*
	SetAccountAlias()
	GetAccountAlias()
	ClearAccountAlias()
	SetPasswordPolicy()
	GetPasswordPolicy()
*/
type AccountAliasResponse struct {
	RamCommonResponse
	AccountAlias string
}

type PasswordPolicyResponse struct {
	RamCommonResponse
	PasswordPolicy
}

type PasswordPolicyRequest struct {
	PasswordPolicy
}

func (client *RamClient) SetAccountAlias(accountalias AccountAlias) (RamCommonResponse, error) {
	return RamCommonResponse{}, nil
}

func (client *RamClient) GetAccountAlias() (AccountAliasResponse, error) {
	return AccountAliasResponse{}, nil
}
func (client *RamClient) ClearAccountAlias() (RamCommonResponse, error) {
	return RamCommonResponse{}, nil
}
func (client *RamClient) SetPasswordPolicy(passwordPolicy PasswordPolicyRequest) (PasswordPolicyResponse, error) {
	return PasswordPolicyResponse{}, nil
}
func (client *RamClient) GetPasswordPolicy(accountAlias AccountAlias) (PasswordPolicyResponse, error) {
	return PasswordPolicyResponse{}, nil
}
