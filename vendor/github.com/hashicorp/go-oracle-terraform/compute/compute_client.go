package compute

import (
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/hashicorp/go-oracle-terraform/client"
	"github.com/hashicorp/go-oracle-terraform/opc"
)

const CMP_ACME = "/Compute-%s"
const CMP_USERNAME = "/Compute-%s/%s"
const CMP_QUALIFIED_NAME = "%s/%s"

// Client represents an authenticated compute client, with compute credentials and an api client.
type ComputeClient struct {
	client       *client.Client
	authCookie   *http.Cookie
	cookieIssued time.Time
}

func NewComputeClient(c *opc.Config) (*ComputeClient, error) {
	computeClient := &ComputeClient{}
	client, err := client.NewClient(c)
	if err != nil {
		return nil, err
	}
	computeClient.client = client

	if err := computeClient.getAuthenticationCookie(); err != nil {
		return nil, err
	}

	return computeClient, nil
}

func (c *ComputeClient) executeRequest(method, path string, body interface{}) (*http.Response, error) {
	reqBody, err := c.client.MarshallRequestBody(body)
	if err != nil {
		return nil, err
	}

	req, err := c.client.BuildRequestBody(method, path, reqBody)
	if err != nil {
		return nil, err
	}

	debugReqString := fmt.Sprintf("HTTP %s Req (%s)", method, path)
	if body != nil {
		req.Header.Set("Content-Type", "application/oracle-compute-v3+json")
		// Don't leak credentials in STDERR
		if path != "/authenticate/" {
			debugReqString = fmt.Sprintf("%s:\n %+v", debugReqString, string(reqBody))
		}
	}
	// Log the request before the authentication cookie, so as not to leak credentials
	c.client.DebugLogString(debugReqString)
	// If we have an authentication cookie, let's authenticate, refreshing cookie if need be
	if c.authCookie != nil {
		if time.Since(c.cookieIssued).Minutes() > 25 {
			c.authCookie = nil
			if err := c.getAuthenticationCookie(); err != nil {
				return nil, err
			}
		}
		req.AddCookie(c.authCookie)
	}

	resp, err := c.client.ExecuteRequest(req)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (c *ComputeClient) getACME() string {
	return fmt.Sprintf(CMP_ACME, *c.client.IdentityDomain)
}

func (c *ComputeClient) getUserName() string {
	return fmt.Sprintf(CMP_USERNAME, *c.client.IdentityDomain, *c.client.UserName)
}

func (c *ComputeClient) getQualifiedACMEName(name string) string {
	if name == "" {
		return ""
	}
	if strings.HasPrefix(name, "/Compute-") && len(strings.Split(name, "/")) == 1 {
		return name
	}
	return fmt.Sprintf(CMP_QUALIFIED_NAME, c.getACME(), name)
}

// From compute_client
// GetObjectName returns the fully-qualified name of an OPC object, e.g. /identity-domain/user@email/{name}
func (c *ComputeClient) getQualifiedName(name string) string {
	if name == "" {
		return ""
	}
	if strings.HasPrefix(name, "/oracle") || strings.HasPrefix(name, "/Compute-") {
		return name
	}
	return fmt.Sprintf(CMP_QUALIFIED_NAME, c.getUserName(), name)
}

func (c *ComputeClient) getObjectPath(root, name string) string {
	return fmt.Sprintf("%s%s", root, c.getQualifiedName(name))
}

// GetUnqualifiedName returns the unqualified name of an OPC object, e.g. the {name} part of /identity-domain/user@email/{name}
func (c *ComputeClient) getUnqualifiedName(name string) string {
	if name == "" {
		return name
	}
	if strings.HasPrefix(name, "/oracle") {
		return name
	}
	if !strings.Contains(name, "/") {
		return name
	}

	nameParts := strings.Split(name, "/")
	return strings.Join(nameParts[3:], "/")
}

func (c *ComputeClient) unqualify(names ...*string) {
	for _, name := range names {
		*name = c.getUnqualifiedName(*name)
	}
}

func (c *ComputeClient) unqualifyUrl(url *string) {
	var validID = regexp.MustCompile(`(\/(Compute[^\/\s]+))(\/[^\/\s]+)(\/[^\/\s]+)`)
	name := validID.FindString(*url)
	*url = c.getUnqualifiedName(name)
}

func (c *ComputeClient) getQualifiedList(list []string) []string {
	for i, name := range list {
		list[i] = c.getQualifiedName(name)
	}
	return list
}

func (c *ComputeClient) getUnqualifiedList(list []string) []string {
	for i, name := range list {
		list[i] = c.getUnqualifiedName(name)
	}
	return list
}

func (c *ComputeClient) getQualifiedListName(name string) string {
	nameParts := strings.Split(name, ":")
	listType := nameParts[0]
	listName := nameParts[1]
	return fmt.Sprintf("%s:%s", listType, c.getQualifiedName(listName))
}

func (c *ComputeClient) unqualifyListName(qualifiedName string) string {
	nameParts := strings.Split(qualifiedName, ":")
	listType := nameParts[0]
	listName := nameParts[1]
	return fmt.Sprintf("%s:%s", listType, c.getUnqualifiedName(listName))
}
