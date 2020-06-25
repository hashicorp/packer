package external

import (
	"encoding/json"
	"strings"
	"time"

	"github.com/pkg/errors"

	"github.com/ucloud/ucloud-sdk-go/private/protocol/http"
	"github.com/ucloud/ucloud-sdk-go/ucloud/metadata"
)

const internalBaseUrl = "http://api.service.ucloud.cn"

type AssumeRoleRequest struct {
	RoleName string
}

func LoadSTSConfig(req AssumeRoleRequest) (ConfigProvider, error) {
	httpClient := http.NewHttpClient()
	client := &metadata.DefaultClient{}
	err := client.SetHttpClient(&httpClient)
	if err != nil {
		return nil, err
	}
	return loadSTSConfig(req, client)
}

type metadataProvider interface {
	SendRequest(string) (string, error)
	SetHttpClient(http.Client) error
}

type assumeRoleData struct {
	Expiration    int
	PrivateKey    string
	ProjectID     string
	PublicKey     string
	CharacterName string
	SecurityToken string
	UHostID       string
	UPHostId      string
}

type assumeRoleResponse struct {
	RetCode int
	Message string
	Data    assumeRoleData
}

func loadSTSConfig(req AssumeRoleRequest, client metadataProvider) (ConfigProvider, error) {
	path := "/meta-data/v1/uam/security-credentials"
	if len(req.RoleName) != 0 {
		path += "/" + req.RoleName
	}

	resp, err := client.SendRequest(path)
	if err != nil {
		return nil, err
	}

	var roleResp assumeRoleResponse
	if err := json.NewDecoder(strings.NewReader(resp)).Decode(&roleResp); err != nil {
		return nil, errors.Errorf("failed to decode sts credential, %s", err)
	}

	region, err := client.SendRequest("/meta-data/latest/uhost/region")
	if err != nil {
		return nil, err
	}

	zone, err := client.SendRequest("/meta-data/latest/uhost/zone")
	if err != nil {
		return nil, err
	}

	roleData := roleResp.Data
	stsConfig := &config{
		CanExpire:     true,
		Expires:       time.Unix(int64(roleData.Expiration), 0),
		PrivateKey:    roleData.PrivateKey,
		PublicKey:     roleData.PublicKey,
		SecurityToken: roleData.SecurityToken,
		ProjectId:     roleData.ProjectID,
		Region:        region,
		Zone:          zone,
		BaseUrl:       internalBaseUrl,
	}
	return stsConfig, nil
}
