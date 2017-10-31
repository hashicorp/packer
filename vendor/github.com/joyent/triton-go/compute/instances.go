package compute

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/hashicorp/errwrap"
	"github.com/joyent/triton-go/client"
)

type InstancesClient struct {
	client *client.Client
}

const (
	CNSTagDisable    = "triton.cns.disable"
	CNSTagReversePTR = "triton.cns.reverse_ptr"
	CNSTagServices   = "triton.cns.services"
)

// InstanceCNS is a container for the CNS-specific attributes.  In the API these
// values are embedded within a Instance's Tags attribute, however they are
// exposed to the caller as their native types.
type InstanceCNS struct {
	Disable    bool
	ReversePTR string
	Services   []string
}

type Instance struct {
	ID              string                 `json:"id"`
	Name            string                 `json:"name"`
	Type            string                 `json:"type"`
	Brand           string                 `json:"brand"`
	State           string                 `json:"state"`
	Image           string                 `json:"image"`
	Memory          int                    `json:"memory"`
	Disk            int                    `json:"disk"`
	Metadata        map[string]string      `json:"metadata"`
	Tags            map[string]interface{} `json:"tags"`
	Created         time.Time              `json:"created"`
	Updated         time.Time              `json:"updated"`
	Docker          bool                   `json:"docker"`
	IPs             []string               `json:"ips"`
	Networks        []string               `json:"networks"`
	PrimaryIP       string                 `json:"primaryIp"`
	FirewallEnabled bool                   `json:"firewall_enabled"`
	ComputeNode     string                 `json:"compute_node"`
	Package         string                 `json:"package"`
	DomainNames     []string               `json:"dns_names"`
	CNS             InstanceCNS
}

// _Instance is a private facade over Instance that handles the necessary API
// overrides from VMAPI's machine endpoint(s).
type _Instance struct {
	Instance
	Tags map[string]interface{} `json:"tags"`
}

type NIC struct {
	IP      string `json:"ip"`
	MAC     string `json:"mac"`
	Primary bool   `json:"primary"`
	Netmask string `json:"netmask"`
	Gateway string `json:"gateway"`
	State   string `json:"state"`
	Network string `json:"network"`
}

type GetInstanceInput struct {
	ID string
}

func (gmi *GetInstanceInput) Validate() error {
	if gmi.ID == "" {
		return fmt.Errorf("machine ID can not be empty")
	}

	return nil
}

func (c *InstancesClient) Get(ctx context.Context, input *GetInstanceInput) (*Instance, error) {
	if err := input.Validate(); err != nil {
		return nil, errwrap.Wrapf("unable to get machine: {{err}}", err)
	}

	path := fmt.Sprintf("/%s/machines/%s", c.client.AccountName, input.ID)
	reqInputs := client.RequestInput{
		Method: http.MethodGet,
		Path:   path,
	}
	response, err := c.client.ExecuteRequestRaw(ctx, reqInputs)
	if response != nil {
		defer response.Body.Close()
	}
	if response == nil || response.StatusCode == http.StatusNotFound || response.StatusCode == http.StatusGone {
		return nil, &client.TritonError{
			StatusCode: response.StatusCode,
			Code:       "ResourceNotFound",
		}
	}
	if err != nil {
		return nil, errwrap.Wrapf("Error executing Get request: {{err}}",
			c.client.DecodeError(response.StatusCode, response.Body))
	}

	var result *_Instance
	decoder := json.NewDecoder(response.Body)
	if err = decoder.Decode(&result); err != nil {
		return nil, errwrap.Wrapf("Error decoding Get response: {{err}}", err)
	}

	native, err := result.toNative()
	if err != nil {
		return nil, errwrap.Wrapf("unable to convert API response for instances to native type: {{err}}", err)
	}

	return native, nil
}

type ListInstancesInput struct {
	Brand       string
	Alias       string
	Name        string
	Image       string
	State       string
	Memory      uint16
	Limit       uint16
	Offset      uint16
	Tags        []string // query by arbitrary tags prefixed with "tag."
	Tombstone   bool
	Docker      bool
	Credentials bool
}

func (c *InstancesClient) List(ctx context.Context, input *ListInstancesInput) ([]*Instance, error) {
	path := fmt.Sprintf("/%s/machines", c.client.AccountName)

	query := &url.Values{}
	if input.Brand != "" {
		query.Set("brand", input.Brand)
	}
	if input.Name != "" {
		query.Set("name", input.Name)
	}
	if input.Image != "" {
		query.Set("image", input.Image)
	}
	if input.State != "" {
		query.Set("state", input.State)
	}
	if input.Memory >= 1 && input.Memory <= 1000 {
		query.Set("memory", fmt.Sprintf("%d", input.Memory))
	}
	if input.Limit >= 1 {
		query.Set("limit", fmt.Sprintf("%d", input.Limit))
	}
	if input.Offset >= 0 {
		query.Set("offset", fmt.Sprintf("%d", input.Offset))
	}
	if input.Tombstone {
		query.Set("tombstone", "true")
	}
	if input.Docker {
		query.Set("docker", "true")
	}
	if input.Credentials {
		query.Set("credentials", "true")
	}

	reqInputs := client.RequestInput{
		Method: http.MethodGet,
		Path:   path,
		Query:  query,
	}
	respReader, err := c.client.ExecuteRequest(ctx, reqInputs)
	if err != nil {
		return nil, errwrap.Wrapf("Error executing List request: {{err}}", err)
	}

	var results []*_Instance
	decoder := json.NewDecoder(respReader)
	if err = decoder.Decode(&results); err != nil {
		return nil, errwrap.Wrapf("Error decoding List response: {{err}}", err)
	}

	machines := make([]*Instance, 0, len(results))
	for _, machineAPI := range results {
		native, err := machineAPI.toNative()
		if err != nil {
			return nil, errwrap.Wrapf("unable to convert API response for instances to native type: {{err}}", err)
		}
		machines = append(machines, native)
	}

	return machines, nil
}

type CreateInstanceInput struct {
	Name            string
	Package         string
	Image           string
	Networks        []string
	Affinity        []string
	LocalityStrict  bool
	LocalityNear    []string
	LocalityFar     []string
	Metadata        map[string]string
	Tags            map[string]string
	FirewallEnabled bool
	CNS             InstanceCNS
}

func (input *CreateInstanceInput) toAPI() (map[string]interface{}, error) {
	const numExtraParams = 8
	result := make(map[string]interface{}, numExtraParams+len(input.Metadata)+len(input.Tags))

	result["firewall_enabled"] = input.FirewallEnabled

	if input.Name != "" {
		result["name"] = input.Name
	}

	if input.Package != "" {
		result["package"] = input.Package
	}

	if input.Image != "" {
		result["image"] = input.Image
	}

	if len(input.Networks) > 0 {
		result["networks"] = input.Networks
	}

	// validate that affinity and locality are not included together
	hasAffinity := len(input.Affinity) > 0
	hasLocality := len(input.LocalityNear) > 0 || len(input.LocalityFar) > 0
	if hasAffinity && hasLocality {
		return nil, fmt.Errorf("Cannot include both Affinity and Locality")
	}

	// affinity takes precendence over locality regardless
	if len(input.Affinity) > 0 {
		result["affinity"] = input.Affinity
	} else {
		locality := struct {
			Strict bool     `json:"strict"`
			Near   []string `json:"near,omitempty"`
			Far    []string `json:"far,omitempty"`
		}{
			Strict: input.LocalityStrict,
			Near:   input.LocalityNear,
			Far:    input.LocalityFar,
		}
		result["locality"] = locality
	}

	for key, value := range input.Tags {
		result[fmt.Sprintf("tag.%s", key)] = value
	}

	// NOTE(justinwr): CNSTagServices needs to be a tag if available. No other
	// CNS tags will be handled at this time.
	input.CNS.toTags(result)
	if val, found := result[CNSTagServices]; found {
		result["tag."+CNSTagServices] = val
		delete(result, CNSTagServices)
	}

	for key, value := range input.Metadata {
		result[fmt.Sprintf("metadata.%s", key)] = value
	}

	return result, nil
}

func (c *InstancesClient) Create(ctx context.Context, input *CreateInstanceInput) (*Instance, error) {
	path := fmt.Sprintf("/%s/machines", c.client.AccountName)
	body, err := input.toAPI()
	if err != nil {
		return nil, errwrap.Wrapf("Error preparing Create request: {{err}}", err)
	}

	reqInputs := client.RequestInput{
		Method: http.MethodPost,
		Path:   path,
		Body:   body,
	}
	respReader, err := c.client.ExecuteRequest(ctx, reqInputs)
	if respReader != nil {
		defer respReader.Close()
	}
	if err != nil {
		return nil, errwrap.Wrapf("Error executing Create request: {{err}}", err)
	}

	var result *Instance
	decoder := json.NewDecoder(respReader)
	if err = decoder.Decode(&result); err != nil {
		return nil, errwrap.Wrapf("Error decoding Create response: {{err}}", err)
	}

	return result, nil
}

type DeleteInstanceInput struct {
	ID string
}

func (c *InstancesClient) Delete(ctx context.Context, input *DeleteInstanceInput) error {
	path := fmt.Sprintf("/%s/machines/%s", c.client.AccountName, input.ID)
	reqInputs := client.RequestInput{
		Method: http.MethodDelete,
		Path:   path,
	}
	response, err := c.client.ExecuteRequestRaw(ctx, reqInputs)
	if response == nil {
		return fmt.Errorf("Delete request has empty response")
	}
	if response.Body != nil {
		defer response.Body.Close()
	}
	if response.StatusCode == http.StatusNotFound || response.StatusCode == http.StatusGone {
		return nil
	}
	if err != nil {
		return errwrap.Wrapf("Error executing Delete request: {{err}}",
			c.client.DecodeError(response.StatusCode, response.Body))
	}

	return nil
}

type DeleteTagsInput struct {
	ID string
}

func (c *InstancesClient) DeleteTags(ctx context.Context, input *DeleteTagsInput) error {
	path := fmt.Sprintf("/%s/machines/%s/tags", c.client.AccountName, input.ID)
	reqInputs := client.RequestInput{
		Method: http.MethodDelete,
		Path:   path,
	}
	response, err := c.client.ExecuteRequestRaw(ctx, reqInputs)
	if response == nil {
		return fmt.Errorf("DeleteTags request has empty response")
	}
	if response.Body != nil {
		defer response.Body.Close()
	}
	if response.StatusCode == http.StatusNotFound {
		return nil
	}
	if err != nil {
		return errwrap.Wrapf("Error executing DeleteTags request: {{err}}",
			c.client.DecodeError(response.StatusCode, response.Body))
	}

	return nil
}

type DeleteTagInput struct {
	ID  string
	Key string
}

func (c *InstancesClient) DeleteTag(ctx context.Context, input *DeleteTagInput) error {
	path := fmt.Sprintf("/%s/machines/%s/tags/%s", c.client.AccountName, input.ID, input.Key)
	reqInputs := client.RequestInput{
		Method: http.MethodDelete,
		Path:   path,
	}
	response, err := c.client.ExecuteRequestRaw(ctx, reqInputs)
	if response == nil {
		return fmt.Errorf("DeleteTag request has empty response")
	}
	if response.Body != nil {
		defer response.Body.Close()
	}
	if response.StatusCode == http.StatusNotFound {
		return nil
	}
	if err != nil {
		return errwrap.Wrapf("Error executing DeleteTag request: {{err}}",
			c.client.DecodeError(response.StatusCode, response.Body))
	}

	return nil
}

type RenameInstanceInput struct {
	ID   string
	Name string
}

func (c *InstancesClient) Rename(ctx context.Context, input *RenameInstanceInput) error {
	path := fmt.Sprintf("/%s/machines/%s", c.client.AccountName, input.ID)

	params := &url.Values{}
	params.Set("action", "rename")
	params.Set("name", input.Name)

	reqInputs := client.RequestInput{
		Method: http.MethodPost,
		Path:   path,
		Query:  params,
	}
	respReader, err := c.client.ExecuteRequestURIParams(ctx, reqInputs)
	if respReader != nil {
		defer respReader.Close()
	}
	if err != nil {
		return errwrap.Wrapf("Error executing Rename request: {{err}}", err)
	}

	return nil
}

type ReplaceTagsInput struct {
	ID   string
	Tags map[string]string
	CNS  InstanceCNS
}

// toAPI is used to join Tags and CNS tags into the same JSON object before
// sending an API request to the API gateway.
func (input ReplaceTagsInput) toAPI() map[string]interface{} {
	result := map[string]interface{}{}
	for key, value := range input.Tags {
		result[key] = value
	}
	input.CNS.toTags(result)
	return result
}

func (c *InstancesClient) ReplaceTags(ctx context.Context, input *ReplaceTagsInput) error {
	path := fmt.Sprintf("/%s/machines/%s/tags", c.client.AccountName, input.ID)
	reqInputs := client.RequestInput{
		Method: http.MethodPut,
		Path:   path,
		Body:   input.toAPI(),
	}
	respReader, err := c.client.ExecuteRequest(ctx, reqInputs)
	if respReader != nil {
		defer respReader.Close()
	}
	if err != nil {
		return errwrap.Wrapf("Error executing ReplaceTags request: {{err}}", err)
	}

	return nil
}

type AddTagsInput struct {
	ID   string
	Tags map[string]string
}

func (c *InstancesClient) AddTags(ctx context.Context, input *AddTagsInput) error {
	path := fmt.Sprintf("/%s/machines/%s/tags", c.client.AccountName, input.ID)
	reqInputs := client.RequestInput{
		Method: http.MethodPost,
		Path:   path,
		Body:   input.Tags,
	}
	respReader, err := c.client.ExecuteRequest(ctx, reqInputs)
	if respReader != nil {
		defer respReader.Close()
	}
	if err != nil {
		return errwrap.Wrapf("Error executing AddTags request: {{err}}", err)
	}

	return nil
}

type GetTagInput struct {
	ID  string
	Key string
}

func (c *InstancesClient) GetTag(ctx context.Context, input *GetTagInput) (string, error) {
	path := fmt.Sprintf("/%s/machines/%s/tags/%s", c.client.AccountName, input.ID, input.Key)
	reqInputs := client.RequestInput{
		Method: http.MethodGet,
		Path:   path,
	}
	respReader, err := c.client.ExecuteRequest(ctx, reqInputs)
	if respReader != nil {
		defer respReader.Close()
	}
	if err != nil {
		return "", errwrap.Wrapf("Error executing GetTag request: {{err}}", err)
	}

	var result string
	decoder := json.NewDecoder(respReader)
	if err = decoder.Decode(&result); err != nil {
		return "", errwrap.Wrapf("Error decoding GetTag response: {{err}}", err)
	}

	return result, nil
}

type ListTagsInput struct {
	ID string
}

func (c *InstancesClient) ListTags(ctx context.Context, input *ListTagsInput) (map[string]interface{}, error) {
	path := fmt.Sprintf("/%s/machines/%s/tags", c.client.AccountName, input.ID)
	reqInputs := client.RequestInput{
		Method: http.MethodGet,
		Path:   path,
	}
	respReader, err := c.client.ExecuteRequest(ctx, reqInputs)
	if respReader != nil {
		defer respReader.Close()
	}
	if err != nil {
		return nil, errwrap.Wrapf("Error executing ListTags request: {{err}}", err)
	}

	var result map[string]interface{}
	decoder := json.NewDecoder(respReader)
	if err = decoder.Decode(&result); err != nil {
		return nil, errwrap.Wrapf("Error decoding ListTags response: {{err}}", err)
	}

	_, tags := tagsExtractMeta(result)
	return tags, nil
}

type GetMetadataInput struct {
	ID  string
	Key string
}

// GetMetadata returns a single metadata entry associated with an instance.
func (c *InstancesClient) GetMetadata(ctx context.Context, input *GetMetadataInput) (string, error) {
	if input.Key == "" {
		return "", fmt.Errorf("Missing metadata Key from input: %s", input.Key)
	}

	path := fmt.Sprintf("/%s/machines/%s/metadata/%s", c.client.AccountName, input.ID, input.Key)
	reqInputs := client.RequestInput{
		Method: http.MethodGet,
		Path:   path,
	}
	response, err := c.client.ExecuteRequestRaw(ctx, reqInputs)
	if response != nil {
		defer response.Body.Close()
	}
	if response.StatusCode == http.StatusNotFound || response.StatusCode == http.StatusGone {
		return "", &client.TritonError{
			StatusCode: response.StatusCode,
			Code:       "ResourceNotFound",
		}
	}
	if err != nil {
		return "", errwrap.Wrapf("Error executing Get request: {{err}}",
			c.client.DecodeError(response.StatusCode, response.Body))
	}

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return "", errwrap.Wrapf("Error unwrapping request body: {{err}}",
			c.client.DecodeError(response.StatusCode, response.Body))
	}

	return fmt.Sprintf("%s", body), nil
}

type ListMetadataInput struct {
	ID          string
	Credentials bool
}

func (c *InstancesClient) ListMetadata(ctx context.Context, input *ListMetadataInput) (map[string]string, error) {
	path := fmt.Sprintf("/%s/machines/%s/metadata", c.client.AccountName, input.ID)

	query := &url.Values{}
	if input.Credentials {
		query.Set("credentials", "true")
	}

	reqInputs := client.RequestInput{
		Method: http.MethodGet,
		Path:   path,
		Query:  query,
	}
	respReader, err := c.client.ExecuteRequest(ctx, reqInputs)
	if respReader != nil {
		defer respReader.Close()
	}
	if err != nil {
		return nil, errwrap.Wrapf("Error executing ListMetadata request: {{err}}", err)
	}

	var result map[string]string
	decoder := json.NewDecoder(respReader)
	if err = decoder.Decode(&result); err != nil {
		return nil, errwrap.Wrapf("Error decoding ListMetadata response: {{err}}", err)
	}

	return result, nil
}

type UpdateMetadataInput struct {
	ID       string
	Metadata map[string]string
}

func (c *InstancesClient) UpdateMetadata(ctx context.Context, input *UpdateMetadataInput) (map[string]string, error) {
	path := fmt.Sprintf("/%s/machines/%s/metadata", c.client.AccountName, input.ID)
	reqInputs := client.RequestInput{
		Method: http.MethodPost,
		Path:   path,
		Body:   input.Metadata,
	}
	respReader, err := c.client.ExecuteRequest(ctx, reqInputs)
	if respReader != nil {
		defer respReader.Close()
	}
	if err != nil {
		return nil, errwrap.Wrapf("Error executing UpdateMetadata request: {{err}}", err)
	}

	var result map[string]string
	decoder := json.NewDecoder(respReader)
	if err = decoder.Decode(&result); err != nil {
		return nil, errwrap.Wrapf("Error decoding UpdateMetadata response: {{err}}", err)
	}

	return result, nil
}

type DeleteMetadataInput struct {
	ID  string
	Key string
}

// DeleteMetadata deletes a single metadata key from an instance
func (c *InstancesClient) DeleteMetadata(ctx context.Context, input *DeleteMetadataInput) error {
	if input.Key == "" {
		return fmt.Errorf("Missing metadata Key from input: %s", input.Key)
	}

	path := fmt.Sprintf("/%s/machines/%s/metadata/%s", c.client.AccountName, input.ID, input.Key)
	reqInputs := client.RequestInput{
		Method: http.MethodDelete,
		Path:   path,
	}
	respReader, err := c.client.ExecuteRequest(ctx, reqInputs)
	if respReader != nil {
		defer respReader.Close()
	}
	if err != nil {
		return errwrap.Wrapf("Error executing DeleteMetadata request: {{err}}", err)
	}

	return nil
}

type DeleteAllMetadataInput struct {
	ID string
}

// DeleteAllMetadata deletes all metadata keys from this instance
func (c *InstancesClient) DeleteAllMetadata(ctx context.Context, input *DeleteAllMetadataInput) error {
	path := fmt.Sprintf("/%s/machines/%s/metadata", c.client.AccountName, input.ID)
	reqInputs := client.RequestInput{
		Method: http.MethodDelete,
		Path:   path,
	}
	respReader, err := c.client.ExecuteRequest(ctx, reqInputs)
	if respReader != nil {
		defer respReader.Close()
	}
	if err != nil {
		return errwrap.Wrapf("Error executing DeleteAllMetadata request: {{err}}", err)
	}

	return nil
}

type ResizeInstanceInput struct {
	ID      string
	Package string
}

func (c *InstancesClient) Resize(ctx context.Context, input *ResizeInstanceInput) error {
	path := fmt.Sprintf("/%s/machines/%s", c.client.AccountName, input.ID)

	params := &url.Values{}
	params.Set("action", "resize")
	params.Set("package", input.Package)

	reqInputs := client.RequestInput{
		Method: http.MethodPost,
		Path:   path,
		Query:  params,
	}
	respReader, err := c.client.ExecuteRequestURIParams(ctx, reqInputs)
	if respReader != nil {
		defer respReader.Close()
	}
	if err != nil {
		return errwrap.Wrapf("Error executing Resize request: {{err}}", err)
	}

	return nil
}

type EnableFirewallInput struct {
	ID string
}

func (c *InstancesClient) EnableFirewall(ctx context.Context, input *EnableFirewallInput) error {
	path := fmt.Sprintf("/%s/machines/%s", c.client.AccountName, input.ID)

	params := &url.Values{}
	params.Set("action", "enable_firewall")

	reqInputs := client.RequestInput{
		Method: http.MethodPost,
		Path:   path,
		Query:  params,
	}
	respReader, err := c.client.ExecuteRequestURIParams(ctx, reqInputs)
	if respReader != nil {
		defer respReader.Close()
	}
	if err != nil {
		return errwrap.Wrapf("Error executing EnableFirewall request: {{err}}", err)
	}

	return nil
}

type DisableFirewallInput struct {
	ID string
}

func (c *InstancesClient) DisableFirewall(ctx context.Context, input *DisableFirewallInput) error {
	path := fmt.Sprintf("/%s/machines/%s", c.client.AccountName, input.ID)

	params := &url.Values{}
	params.Set("action", "disable_firewall")

	reqInputs := client.RequestInput{
		Method: http.MethodPost,
		Path:   path,
		Query:  params,
	}
	respReader, err := c.client.ExecuteRequestURIParams(ctx, reqInputs)
	if respReader != nil {
		defer respReader.Close()
	}
	if err != nil {
		return errwrap.Wrapf("Error executing DisableFirewall request: {{err}}", err)
	}

	return nil
}

type ListNICsInput struct {
	InstanceID string
}

func (c *InstancesClient) ListNICs(ctx context.Context, input *ListNICsInput) ([]*NIC, error) {
	path := fmt.Sprintf("/%s/machines/%s/nics", c.client.AccountName, input.InstanceID)
	reqInputs := client.RequestInput{
		Method: http.MethodGet,
		Path:   path,
	}
	respReader, err := c.client.ExecuteRequest(ctx, reqInputs)
	if respReader != nil {
		defer respReader.Close()
	}
	if err != nil {
		return nil, errwrap.Wrapf("Error executing ListNICs request: {{err}}", err)
	}

	var result []*NIC
	decoder := json.NewDecoder(respReader)
	if err = decoder.Decode(&result); err != nil {
		return nil, errwrap.Wrapf("Error decoding ListNICs response: {{err}}", err)
	}

	return result, nil
}

type GetNICInput struct {
	InstanceID string
	MAC        string
}

func (c *InstancesClient) GetNIC(ctx context.Context, input *GetNICInput) (*NIC, error) {
	mac := strings.Replace(input.MAC, ":", "", -1)
	path := fmt.Sprintf("/%s/machines/%s/nics/%s", c.client.AccountName, input.InstanceID, mac)
	reqInputs := client.RequestInput{
		Method: http.MethodGet,
		Path:   path,
	}
	response, err := c.client.ExecuteRequestRaw(ctx, reqInputs)
	if response != nil {
		defer response.Body.Close()
	}
	switch response.StatusCode {
	case http.StatusNotFound:
		return nil, &client.TritonError{
			StatusCode: response.StatusCode,
			Code:       "ResourceNotFound",
		}
	}
	if err != nil {
		return nil, errwrap.Wrapf("Error executing GetNIC request: {{err}}", err)
	}

	var result *NIC
	decoder := json.NewDecoder(response.Body)
	if err = decoder.Decode(&result); err != nil {
		return nil, errwrap.Wrapf("Error decoding ListNICs response: {{err}}", err)
	}

	return result, nil
}

type AddNICInput struct {
	InstanceID string `json:"-"`
	Network    string `json:"network"`
}

// AddNIC asynchronously adds a NIC to a given instance.  If a NIC for a given
// network already exists, a ResourceFound error will be returned.  The status
// of the addition of a NIC can be polled by calling GetNIC()'s and testing NIC
// until its state is set to "running".  Only one NIC per network may exist.
// Warning: this operation causes the instance to restart.
func (c *InstancesClient) AddNIC(ctx context.Context, input *AddNICInput) (*NIC, error) {
	path := fmt.Sprintf("/%s/machines/%s/nics", c.client.AccountName, input.InstanceID)
	reqInputs := client.RequestInput{
		Method: http.MethodPost,
		Path:   path,
		Body:   input,
	}
	response, err := c.client.ExecuteRequestRaw(ctx, reqInputs)
	if response != nil {
		defer response.Body.Close()
	}
	switch response.StatusCode {
	case http.StatusFound:
		return nil, &client.TritonError{
			StatusCode: response.StatusCode,
			Code:       "ResourceFound",
			Message:    response.Header.Get("Location"),
		}
	}
	if err != nil {
		return nil, errwrap.Wrapf("Error executing AddNIC request: {{err}}", err)
	}

	var result *NIC
	decoder := json.NewDecoder(response.Body)
	if err = decoder.Decode(&result); err != nil {
		return nil, errwrap.Wrapf("Error decoding AddNIC response: {{err}}", err)
	}

	return result, nil
}

type RemoveNICInput struct {
	InstanceID string
	MAC        string
}

// RemoveNIC removes a given NIC from a machine asynchronously.  The status of
// the removal can be polled via GetNIC().  When GetNIC() returns a 404, the NIC
// has been removed from the instance.  Warning: this operation causes the
// machine to restart.
func (c *InstancesClient) RemoveNIC(ctx context.Context, input *RemoveNICInput) error {
	mac := strings.Replace(input.MAC, ":", "", -1)
	path := fmt.Sprintf("/%s/machines/%s/nics/%s", c.client.AccountName, input.InstanceID, mac)
	reqInputs := client.RequestInput{
		Method: http.MethodDelete,
		Path:   path,
	}
	response, err := c.client.ExecuteRequestRaw(ctx, reqInputs)
	if response != nil {
		defer response.Body.Close()
	}
	switch response.StatusCode {
	case http.StatusNotFound:
		return &client.TritonError{
			StatusCode: response.StatusCode,
			Code:       "ResourceNotFound",
		}
	}
	if err != nil {
		return errwrap.Wrapf("Error executing RemoveNIC request: {{err}}", err)
	}

	return nil
}

type StopInstanceInput struct {
	InstanceID string
}

func (c *InstancesClient) Stop(ctx context.Context, input *StopInstanceInput) error {
	path := fmt.Sprintf("/%s/machines/%s", c.client.AccountName, input.InstanceID)

	params := &url.Values{}
	params.Set("action", "stop")

	reqInputs := client.RequestInput{
		Method: http.MethodPost,
		Path:   path,
		Query:  params,
	}
	respReader, err := c.client.ExecuteRequestURIParams(ctx, reqInputs)
	if respReader != nil {
		defer respReader.Close()
	}
	if err != nil {
		return errwrap.Wrapf("Error executing Stop request: {{err}}", err)
	}

	return nil
}

type StartInstanceInput struct {
	InstanceID string
}

func (c *InstancesClient) Start(ctx context.Context, input *StartInstanceInput) error {
	path := fmt.Sprintf("/%s/machines/%s", c.client.AccountName, input.InstanceID)

	params := &url.Values{}
	params.Set("action", "start")

	reqInputs := client.RequestInput{
		Method: http.MethodPost,
		Path:   path,
		Query:  params,
	}
	respReader, err := c.client.ExecuteRequestURIParams(ctx, reqInputs)
	if respReader != nil {
		defer respReader.Close()
	}
	if err != nil {
		return errwrap.Wrapf("Error executing Start request: {{err}}", err)
	}

	return nil
}

var reservedInstanceCNSTags = map[string]struct{}{
	CNSTagDisable:    {},
	CNSTagReversePTR: {},
	CNSTagServices:   {},
}

// tagsExtractMeta() extracts all of the misc parameters from Tags and returns a
// clean CNS and Tags struct.
func tagsExtractMeta(tags map[string]interface{}) (InstanceCNS, map[string]interface{}) {
	nativeCNS := InstanceCNS{}
	nativeTags := make(map[string]interface{}, len(tags))
	for k, raw := range tags {
		if _, found := reservedInstanceCNSTags[k]; found {
			switch k {
			case CNSTagDisable:
				b := raw.(bool)
				nativeCNS.Disable = b
			case CNSTagReversePTR:
				s := raw.(string)
				nativeCNS.ReversePTR = s
			case CNSTagServices:
				nativeCNS.Services = strings.Split(raw.(string), ",")
			default:
				// TODO(seanc@): should assert, logic fail
			}
		} else {
			nativeTags[k] = raw
		}
	}

	return nativeCNS, nativeTags
}

// toNative() exports a given _Instance (API representation) to its native object
// format.
func (api *_Instance) toNative() (*Instance, error) {
	m := Instance(api.Instance)
	m.CNS, m.Tags = tagsExtractMeta(api.Tags)
	return &m, nil
}

// toTags() injects its state information into a Tags map suitable for use to
// submit an API call to the vmapi machine endpoint
func (cns *InstanceCNS) toTags(m map[string]interface{}) {
	if cns.Disable {
		// NOTE(justinwr): The JSON encoder and API require the CNSTagDisable
		// attribute to be an actual boolean, not a bool string.
		m[CNSTagDisable] = cns.Disable
	}
	if cns.ReversePTR != "" {
		m[CNSTagReversePTR] = cns.ReversePTR
	}
	if len(cns.Services) > 0 {
		m[CNSTagServices] = strings.Join(cns.Services, ",")
	}
}
