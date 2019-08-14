package vultr

import (
	"fmt"

	"github.com/hashicorp/packer/version"
	"github.com/vultr/govultr"
)

func newVultrClient(apiKey string) *govultr.Client {
	client := govultr.NewClient(nil, apiKey)
	userAgent := fmt.Sprintf("Packer/%s", version.FormattedVersion())
	client.SetUserAgent(userAgent)
	return client
}
