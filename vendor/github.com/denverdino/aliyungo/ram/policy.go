package ram

type PolicyRequest struct {
	PolicyName     string
	PolicyType     string
	Description    string
	PolicyDocument string
	SetAsDefault   string
	VersionId      string
}
type PolicyListResponse struct {
	RamCommonResponse
	Policies struct {
		Policy []Policy
	}
}

type PolicyResponse struct {
	RamCommonResponse
	Policy Policy
}

type PolicyQueryRequest struct {
	PolicyType string
	Marker     string
	MaxItems   int8
}

type PolicyQueryResponse struct {
	IsTruncated bool
	Marker      string
	Policies    struct {
		Policy []Policy
	}
}

type PolicyVersionResponse struct {
	RamCommonResponse
	IsDefaultVersion bool
	VersionId        string
	CreateDate       string
	PolicyDocument   string
}

type AttachPolicyRequest struct {
	PolicyRequest
	UserName string
}

type AttachPolicyToRoleRequest struct {
	PolicyRequest
	RoleName string
}

func (client *RamClient) CreatePolicy(policyReq PolicyRequest) (PolicyResponse, error) {
	var resp PolicyResponse
	err := client.Invoke("CreatePolicy", policyReq, &resp)
	if err != nil {
		return PolicyResponse{}, err
	}
	return resp, nil
}

func (client *RamClient) GetPolicy(policyReq PolicyRequest) (PolicyResponse, error) {
	var resp PolicyResponse
	err := client.Invoke("GetPolicy", policyReq, &resp)
	if err != nil {
		return PolicyResponse{}, err
	}
	return resp, nil
}

func (client *RamClient) DeletePolicy(policyReq PolicyRequest) (RamCommonResponse, error) {
	var resp RamCommonResponse
	err := client.Invoke("DeletePolicy", policyReq, &resp)
	if err != nil {
		return RamCommonResponse{}, err
	}
	return resp, nil
}

func (client *RamClient) ListPolicies(policyQuery PolicyQueryRequest) (PolicyQueryResponse, error) {
	var resp PolicyQueryResponse
	err := client.Invoke("ListPolicies", policyQuery, &resp)
	if err != nil {
		return PolicyQueryResponse{}, err
	}
	return resp, nil
}

func (client *RamClient) CreatePolicyVersion(policyReq PolicyRequest) (PolicyVersionResponse, error) {
	var resp PolicyVersionResponse
	err := client.Invoke("CreatePolicyVersion", policyReq, &resp)
	if err != nil {
		return PolicyVersionResponse{}, err
	}
	return resp, nil
}

func (client *RamClient) GetPolicyVersion(policyReq PolicyRequest) (PolicyVersionResponse, error) {
	var resp PolicyVersionResponse
	err := client.Invoke("GetPolicyVersion", policyReq, &resp)
	if err != nil {
		return PolicyVersionResponse{}, err
	}
	return resp, nil
}

func (client *RamClient) DeletePolicyVersion(policyReq PolicyRequest) (RamCommonResponse, error) {
	var resp RamCommonResponse
	err := client.Invoke("DeletePolicyVersion", policyReq, &resp)
	if err != nil {
		return RamCommonResponse{}, err
	}
	return resp, nil
}

func (client *RamClient) ListPolicyVersions(policyReq PolicyRequest) (PolicyVersionResponse, error) {
	var resp PolicyVersionResponse
	err := client.Invoke("ListPolicyVersions", policyReq, &resp)
	if err != nil {
		return PolicyVersionResponse{}, err
	}
	return resp, nil
}

//TODO
func (client *RamClient) SetDefaultPolicyVersion() {}

func (client *RamClient) AttachPolicyToUser(attachPolicyRequest AttachPolicyRequest) (RamCommonResponse, error) {
	var resp RamCommonResponse
	err := client.Invoke("AttachPolicyToUser", attachPolicyRequest, &resp)
	if err != nil {
		return RamCommonResponse{}, err
	}
	return resp, nil
}

func (client *RamClient) DetachPolicyFromUser(attachPolicyRequest AttachPolicyRequest) (RamCommonResponse, error) {
	var resp RamCommonResponse
	err := client.Invoke("DetachPolicyFromUser", attachPolicyRequest, &resp)
	if err != nil {
		return RamCommonResponse{}, err
	}
	return resp, nil
}

//TODO
func (client *RamClient) ListEnitiesForPolicy() {}

func (client *RamClient) ListPoliciesForUser(userQuery UserQueryRequest) (PolicyListResponse, error) {
	var resp PolicyListResponse
	err := client.Invoke("ListPoliciesForUser", userQuery, &resp)
	if err != nil {
		return PolicyListResponse{}, err
	}
	return resp, nil
}

//
//Role related
//
func (client *RamClient) AttachPolicyToRole(attachPolicyRequest AttachPolicyToRoleRequest) (RamCommonResponse, error) {
	var resp RamCommonResponse
	err := client.Invoke("AttachPolicyToRole", attachPolicyRequest, &resp)
	if err != nil {
		return RamCommonResponse{}, err
	}
	return resp, nil
}

func (client *RamClient) DetachPolicyFromRole(attachPolicyRequest AttachPolicyToRoleRequest) (RamCommonResponse, error) {
	var resp RamCommonResponse
	err := client.Invoke("DetachPolicyFromRole", attachPolicyRequest, &resp)
	if err != nil {
		return RamCommonResponse{}, err
	}
	return resp, nil
}

func (client *RamClient) ListPoliciesForRole(roleQuery RoleQueryRequest) (PolicyListResponse, error) {
	var resp PolicyListResponse
	err := client.Invoke("ListPoliciesForRole", roleQuery, &resp)
	if err != nil {
		return PolicyListResponse{}, err
	}
	return resp, nil
}

//
//Group related
//
//TODO
//
func (client *RamClient) ListPoliciesForGroup() {}
