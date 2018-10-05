package hcloud

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/url"
	"strconv"
	"time"

	"github.com/hetznercloud/hcloud-go/hcloud/schema"
)

// Server represents a server in the Hetzner Cloud.
type Server struct {
	ID              int
	Name            string
	Status          ServerStatus
	Created         time.Time
	PublicNet       ServerPublicNet
	ServerType      *ServerType
	Datacenter      *Datacenter
	IncludedTraffic uint64
	OutgoingTraffic uint64
	IngoingTraffic  uint64
	BackupWindow    string
	RescueEnabled   bool
	Locked          bool
	ISO             *ISO
	Image           *Image
	Protection      ServerProtection
	Labels          map[string]string
}

// ServerProtection represents the protection level of a server.
type ServerProtection struct {
	Delete, Rebuild bool
}

// ServerStatus specifies a server's status.
type ServerStatus string

const (
	// ServerStatusInitializing is the status when a server is initializing.
	ServerStatusInitializing ServerStatus = "initializing"

	// ServerStatusOff is the status when a server is off.
	ServerStatusOff ServerStatus = "off"

	// ServerStatusRunning is the status when a server is running.
	ServerStatusRunning ServerStatus = "running"
)

// ServerPublicNet represents a server's public network.
type ServerPublicNet struct {
	IPv4        ServerPublicNetIPv4
	IPv6        ServerPublicNetIPv6
	FloatingIPs []*FloatingIP
}

// ServerPublicNetIPv4 represents a server's public IPv4 address.
type ServerPublicNetIPv4 struct {
	IP      net.IP
	Blocked bool
	DNSPtr  string
}

// ServerPublicNetIPv6 represents a server's public IPv6 network and address.
type ServerPublicNetIPv6 struct {
	IP      net.IP
	Network *net.IPNet
	Blocked bool
	DNSPtr  map[string]string
}

// DNSPtrForIP returns the reverse dns pointer of the ip address.
func (s *ServerPublicNetIPv6) DNSPtrForIP(ip net.IP) string {
	return s.DNSPtr[ip.String()]
}

// ServerRescueType represents rescue types.
type ServerRescueType string

// List of rescue types.
const (
	ServerRescueTypeLinux32   ServerRescueType = "linux32"
	ServerRescueTypeLinux64   ServerRescueType = "linux64"
	ServerRescueTypeFreeBSD64 ServerRescueType = "freebsd64"
)

// ServerClient is a client for the servers API.
type ServerClient struct {
	client *Client
}

// GetByID retrieves a server by its ID.
func (c *ServerClient) GetByID(ctx context.Context, id int) (*Server, *Response, error) {
	req, err := c.client.NewRequest(ctx, "GET", fmt.Sprintf("/servers/%d", id), nil)
	if err != nil {
		return nil, nil, err
	}

	var body schema.ServerGetResponse
	resp, err := c.client.Do(req, &body)
	if err != nil {
		if IsError(err, ErrorCodeNotFound) {
			return nil, resp, nil
		}
		return nil, nil, err
	}
	return ServerFromSchema(body.Server), resp, nil
}

// GetByName retreives a server by its name.
func (c *ServerClient) GetByName(ctx context.Context, name string) (*Server, *Response, error) {
	path := "/servers?name=" + url.QueryEscape(name)
	req, err := c.client.NewRequest(ctx, "GET", path, nil)
	if err != nil {
		return nil, nil, err
	}

	var body schema.ServerListResponse
	resp, err := c.client.Do(req, &body)
	if err != nil {
		return nil, nil, err
	}

	if len(body.Servers) == 0 {
		return nil, resp, nil
	}
	return ServerFromSchema(body.Servers[0]), resp, nil
}

// Get retrieves a server by its ID if the input can be parsed as an integer, otherwise it retrieves a server by its name.
func (c *ServerClient) Get(ctx context.Context, idOrName string) (*Server, *Response, error) {
	if id, err := strconv.Atoi(idOrName); err == nil {
		return c.GetByID(ctx, int(id))
	}
	return c.GetByName(ctx, idOrName)
}

// ServerListOpts specifies options for listing servers.
type ServerListOpts struct {
	ListOpts
}

// List returns a list of servers for a specific page.
func (c *ServerClient) List(ctx context.Context, opts ServerListOpts) ([]*Server, *Response, error) {
	path := "/servers?" + valuesForListOpts(opts.ListOpts).Encode()
	req, err := c.client.NewRequest(ctx, "GET", path, nil)
	if err != nil {
		return nil, nil, err
	}

	var body schema.ServerListResponse
	resp, err := c.client.Do(req, &body)
	if err != nil {
		return nil, nil, err
	}
	servers := make([]*Server, 0, len(body.Servers))
	for _, s := range body.Servers {
		servers = append(servers, ServerFromSchema(s))
	}
	return servers, resp, nil
}

// All returns all servers.
func (c *ServerClient) All(ctx context.Context) ([]*Server, error) {
	return c.AllWithOpts(ctx, ServerListOpts{ListOpts{PerPage: 50}})
}

// AllWithOpts returns all servers for the given options.
func (c *ServerClient) AllWithOpts(ctx context.Context, opts ServerListOpts) ([]*Server, error) {
	allServers := []*Server{}

	_, err := c.client.all(func(page int) (*Response, error) {
		opts.Page = page
		servers, resp, err := c.List(ctx, opts)
		if err != nil {
			return resp, err
		}
		allServers = append(allServers, servers...)
		return resp, nil
	})
	if err != nil {
		return nil, err
	}

	return allServers, nil
}

// ServerCreateOpts specifies options for creating a new server.
type ServerCreateOpts struct {
	Name             string
	ServerType       *ServerType
	Image            *Image
	SSHKeys          []*SSHKey
	Location         *Location
	Datacenter       *Datacenter
	UserData         string
	StartAfterCreate *bool
	Labels           map[string]string
}

// Validate checks if options are valid.
func (o ServerCreateOpts) Validate() error {
	if o.Name == "" {
		return errors.New("missing name")
	}
	if o.ServerType == nil || (o.ServerType.ID == 0 && o.ServerType.Name == "") {
		return errors.New("missing server type")
	}
	if o.Image == nil || (o.Image.ID == 0 && o.Image.Name == "") {
		return errors.New("missing image")
	}
	if o.Location != nil && o.Datacenter != nil {
		return errors.New("location and datacenter are mutually exclusive")
	}
	return nil
}

// ServerCreateResult is the result of a create server call.
type ServerCreateResult struct {
	Server       *Server
	Action       *Action
	RootPassword string
}

// Create creates a new server.
func (c *ServerClient) Create(ctx context.Context, opts ServerCreateOpts) (ServerCreateResult, *Response, error) {
	if err := opts.Validate(); err != nil {
		return ServerCreateResult{}, nil, err
	}

	var reqBody schema.ServerCreateRequest
	reqBody.UserData = opts.UserData
	reqBody.Name = opts.Name
	reqBody.StartAfterCreate = opts.StartAfterCreate
	if opts.ServerType.ID != 0 {
		reqBody.ServerType = opts.ServerType.ID
	} else if opts.ServerType.Name != "" {
		reqBody.ServerType = opts.ServerType.Name
	}
	if opts.Image.ID != 0 {
		reqBody.Image = opts.Image.ID
	} else if opts.Image.Name != "" {
		reqBody.Image = opts.Image.Name
	}
	if opts.Labels != nil {
		reqBody.Labels = &opts.Labels
	}
	for _, sshKey := range opts.SSHKeys {
		reqBody.SSHKeys = append(reqBody.SSHKeys, sshKey.ID)
	}
	if opts.Location != nil {
		if opts.Location.ID != 0 {
			reqBody.Location = strconv.Itoa(opts.Location.ID)
		} else {
			reqBody.Location = opts.Location.Name
		}
	}
	if opts.Datacenter != nil {
		if opts.Datacenter.ID != 0 {
			reqBody.Datacenter = strconv.Itoa(opts.Datacenter.ID)
		} else {
			reqBody.Datacenter = opts.Datacenter.Name
		}
	}
	reqBodyData, err := json.Marshal(reqBody)
	if err != nil {
		return ServerCreateResult{}, nil, err
	}

	req, err := c.client.NewRequest(ctx, "POST", "/servers", bytes.NewReader(reqBodyData))
	if err != nil {
		return ServerCreateResult{}, nil, err
	}

	var respBody schema.ServerCreateResponse
	resp, err := c.client.Do(req, &respBody)
	if err != nil {
		return ServerCreateResult{}, resp, err
	}
	result := ServerCreateResult{
		Server: ServerFromSchema(respBody.Server),
		Action: ActionFromSchema(respBody.Action),
	}
	if respBody.RootPassword != nil {
		result.RootPassword = *respBody.RootPassword
	}
	return result, resp, nil
}

// Delete deletes a server.
func (c *ServerClient) Delete(ctx context.Context, server *Server) (*Response, error) {
	req, err := c.client.NewRequest(ctx, "DELETE", fmt.Sprintf("/servers/%d", server.ID), nil)
	if err != nil {
		return nil, err
	}
	return c.client.Do(req, nil)
}

// ServerUpdateOpts specifies options for updating a server.
type ServerUpdateOpts struct {
	Name   string
	Labels map[string]string
}

// Update updates a server.
func (c *ServerClient) Update(ctx context.Context, server *Server, opts ServerUpdateOpts) (*Server, *Response, error) {
	reqBody := schema.ServerUpdateRequest{
		Name: opts.Name,
	}
	if opts.Labels != nil {
		reqBody.Labels = &opts.Labels
	}
	reqBodyData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, nil, err
	}

	path := fmt.Sprintf("/servers/%d", server.ID)
	req, err := c.client.NewRequest(ctx, "PUT", path, bytes.NewReader(reqBodyData))
	if err != nil {
		return nil, nil, err
	}

	respBody := schema.ServerUpdateResponse{}
	resp, err := c.client.Do(req, &respBody)
	if err != nil {
		return nil, resp, err
	}
	return ServerFromSchema(respBody.Server), resp, nil
}

// Poweron starts a server.
func (c *ServerClient) Poweron(ctx context.Context, server *Server) (*Action, *Response, error) {
	path := fmt.Sprintf("/servers/%d/actions/poweron", server.ID)
	req, err := c.client.NewRequest(ctx, "POST", path, nil)
	if err != nil {
		return nil, nil, err
	}

	respBody := schema.ServerActionPoweronResponse{}
	resp, err := c.client.Do(req, &respBody)
	if err != nil {
		return nil, resp, err
	}
	return ActionFromSchema(respBody.Action), resp, nil
}

// Reboot reboots a server.
func (c *ServerClient) Reboot(ctx context.Context, server *Server) (*Action, *Response, error) {
	path := fmt.Sprintf("/servers/%d/actions/reboot", server.ID)
	req, err := c.client.NewRequest(ctx, "POST", path, nil)
	if err != nil {
		return nil, nil, err
	}

	respBody := schema.ServerActionRebootResponse{}
	resp, err := c.client.Do(req, &respBody)
	if err != nil {
		return nil, resp, err
	}
	return ActionFromSchema(respBody.Action), resp, nil
}

// Reset resets a server.
func (c *ServerClient) Reset(ctx context.Context, server *Server) (*Action, *Response, error) {
	path := fmt.Sprintf("/servers/%d/actions/reset", server.ID)
	req, err := c.client.NewRequest(ctx, "POST", path, nil)
	if err != nil {
		return nil, nil, err
	}

	respBody := schema.ServerActionResetResponse{}
	resp, err := c.client.Do(req, &respBody)
	if err != nil {
		return nil, resp, err
	}
	return ActionFromSchema(respBody.Action), resp, nil
}

// Shutdown shuts down a server.
func (c *ServerClient) Shutdown(ctx context.Context, server *Server) (*Action, *Response, error) {
	path := fmt.Sprintf("/servers/%d/actions/shutdown", server.ID)
	req, err := c.client.NewRequest(ctx, "POST", path, nil)
	if err != nil {
		return nil, nil, err
	}

	respBody := schema.ServerActionShutdownResponse{}
	resp, err := c.client.Do(req, &respBody)
	if err != nil {
		return nil, resp, err
	}
	return ActionFromSchema(respBody.Action), resp, nil
}

// Poweroff stops a server.
func (c *ServerClient) Poweroff(ctx context.Context, server *Server) (*Action, *Response, error) {
	path := fmt.Sprintf("/servers/%d/actions/poweroff", server.ID)
	req, err := c.client.NewRequest(ctx, "POST", path, nil)
	if err != nil {
		return nil, nil, err
	}

	respBody := schema.ServerActionPoweroffResponse{}
	resp, err := c.client.Do(req, &respBody)
	if err != nil {
		return nil, resp, err
	}
	return ActionFromSchema(respBody.Action), resp, nil
}

// ServerResetPasswordResult is the result of resetting a server's password.
type ServerResetPasswordResult struct {
	Action       *Action
	RootPassword string
}

// ResetPassword resets a server's password.
func (c *ServerClient) ResetPassword(ctx context.Context, server *Server) (ServerResetPasswordResult, *Response, error) {
	path := fmt.Sprintf("/servers/%d/actions/reset_password", server.ID)
	req, err := c.client.NewRequest(ctx, "POST", path, nil)
	if err != nil {
		return ServerResetPasswordResult{}, nil, err
	}

	respBody := schema.ServerActionResetPasswordResponse{}
	resp, err := c.client.Do(req, &respBody)
	if err != nil {
		return ServerResetPasswordResult{}, resp, err
	}
	return ServerResetPasswordResult{
		Action:       ActionFromSchema(respBody.Action),
		RootPassword: respBody.RootPassword,
	}, resp, nil
}

// ServerCreateImageOpts specifies options for creating an image from a server.
type ServerCreateImageOpts struct {
	Type        ImageType
	Description *string
	Labels      map[string]string
}

// Validate checks if options are valid.
func (o ServerCreateImageOpts) Validate() error {
	switch o.Type {
	case ImageTypeSnapshot, ImageTypeBackup:
		break
	case "":
		break
	default:
		return errors.New("invalid type")
	}

	return nil
}

// ServerCreateImageResult is the result of creating an image from a server.
type ServerCreateImageResult struct {
	Action *Action
	Image  *Image
}

// CreateImage creates an image from a server.
func (c *ServerClient) CreateImage(ctx context.Context, server *Server, opts *ServerCreateImageOpts) (ServerCreateImageResult, *Response, error) {
	var reqBody schema.ServerActionCreateImageRequest
	if opts != nil {
		if err := opts.Validate(); err != nil {
			return ServerCreateImageResult{}, nil, fmt.Errorf("invalid options: %s", err)
		}
		if opts.Description != nil {
			reqBody.Description = opts.Description
		}
		if opts.Type != "" {
			reqBody.Type = String(string(opts.Type))
		}
		if opts.Labels != nil {
			reqBody.Labels = &opts.Labels
		}
	}
	reqBodyData, err := json.Marshal(reqBody)
	if err != nil {
		return ServerCreateImageResult{}, nil, err
	}

	path := fmt.Sprintf("/servers/%d/actions/create_image", server.ID)
	req, err := c.client.NewRequest(ctx, "POST", path, bytes.NewReader(reqBodyData))
	if err != nil {
		return ServerCreateImageResult{}, nil, err
	}

	respBody := schema.ServerActionCreateImageResponse{}
	resp, err := c.client.Do(req, &respBody)
	if err != nil {
		return ServerCreateImageResult{}, resp, err
	}
	return ServerCreateImageResult{
		Action: ActionFromSchema(respBody.Action),
		Image:  ImageFromSchema(respBody.Image),
	}, resp, nil
}

// ServerEnableRescueOpts specifies options for enabling rescue mode for a server.
type ServerEnableRescueOpts struct {
	Type    ServerRescueType
	SSHKeys []*SSHKey
}

// ServerEnableRescueResult is the result of enabling rescue mode for a server.
type ServerEnableRescueResult struct {
	Action       *Action
	RootPassword string
}

// EnableRescue enables rescue mode for a server.
func (c *ServerClient) EnableRescue(ctx context.Context, server *Server, opts ServerEnableRescueOpts) (ServerEnableRescueResult, *Response, error) {
	reqBody := schema.ServerActionEnableRescueRequest{
		Type: String(string(opts.Type)),
	}
	for _, sshKey := range opts.SSHKeys {
		reqBody.SSHKeys = append(reqBody.SSHKeys, sshKey.ID)
	}
	reqBodyData, err := json.Marshal(reqBody)
	if err != nil {
		return ServerEnableRescueResult{}, nil, err
	}

	path := fmt.Sprintf("/servers/%d/actions/enable_rescue", server.ID)
	req, err := c.client.NewRequest(ctx, "POST", path, bytes.NewReader(reqBodyData))
	if err != nil {
		return ServerEnableRescueResult{}, nil, err
	}

	respBody := schema.ServerActionEnableRescueResponse{}
	resp, err := c.client.Do(req, &respBody)
	if err != nil {
		return ServerEnableRescueResult{}, resp, err
	}
	result := ServerEnableRescueResult{
		Action:       ActionFromSchema(respBody.Action),
		RootPassword: respBody.RootPassword,
	}
	return result, resp, nil
}

// DisableRescue disables rescue mode for a server.
func (c *ServerClient) DisableRescue(ctx context.Context, server *Server) (*Action, *Response, error) {
	path := fmt.Sprintf("/servers/%d/actions/disable_rescue", server.ID)
	req, err := c.client.NewRequest(ctx, "POST", path, nil)
	if err != nil {
		return nil, nil, err
	}

	respBody := schema.ServerActionDisableRescueResponse{}
	resp, err := c.client.Do(req, &respBody)
	if err != nil {
		return nil, resp, err
	}
	return ActionFromSchema(respBody.Action), resp, nil
}

// ServerRebuildOpts specifies options for rebuilding a server.
type ServerRebuildOpts struct {
	Image *Image
}

// Rebuild rebuilds a server.
func (c *ServerClient) Rebuild(ctx context.Context, server *Server, opts ServerRebuildOpts) (*Action, *Response, error) {
	reqBody := schema.ServerActionRebuildRequest{}
	if opts.Image.ID != 0 {
		reqBody.Image = opts.Image.ID
	} else {
		reqBody.Image = opts.Image.Name
	}
	reqBodyData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, nil, err
	}

	path := fmt.Sprintf("/servers/%d/actions/rebuild", server.ID)
	req, err := c.client.NewRequest(ctx, "POST", path, bytes.NewReader(reqBodyData))
	if err != nil {
		return nil, nil, err
	}

	respBody := schema.ServerActionRebuildResponse{}
	resp, err := c.client.Do(req, &respBody)
	if err != nil {
		return nil, resp, err
	}
	return ActionFromSchema(respBody.Action), resp, nil
}

// AttachISO attaches an ISO to a server.
func (c *ServerClient) AttachISO(ctx context.Context, server *Server, iso *ISO) (*Action, *Response, error) {
	reqBody := schema.ServerActionAttachISORequest{}
	if iso.ID != 0 {
		reqBody.ISO = iso.ID
	} else {
		reqBody.ISO = iso.Name
	}
	reqBodyData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, nil, err
	}

	path := fmt.Sprintf("/servers/%d/actions/attach_iso", server.ID)
	req, err := c.client.NewRequest(ctx, "POST", path, bytes.NewReader(reqBodyData))
	if err != nil {
		return nil, nil, err
	}

	respBody := schema.ServerActionAttachISOResponse{}
	resp, err := c.client.Do(req, &respBody)
	if err != nil {
		return nil, resp, err
	}
	return ActionFromSchema(respBody.Action), resp, nil
}

// DetachISO detaches the currently attached ISO from a server.
func (c *ServerClient) DetachISO(ctx context.Context, server *Server) (*Action, *Response, error) {
	path := fmt.Sprintf("/servers/%d/actions/detach_iso", server.ID)
	req, err := c.client.NewRequest(ctx, "POST", path, nil)
	if err != nil {
		return nil, nil, err
	}

	respBody := schema.ServerActionDetachISOResponse{}
	resp, err := c.client.Do(req, &respBody)
	if err != nil {
		return nil, resp, err
	}
	return ActionFromSchema(respBody.Action), resp, nil
}

// EnableBackup enables backup for a server. Pass in an empty backup window to let the
// API pick a window for you. See the API documentation at docs.hetzner.cloud for a list
// of valid backup windows.
func (c *ServerClient) EnableBackup(ctx context.Context, server *Server, window string) (*Action, *Response, error) {
	reqBody := schema.ServerActionEnableBackupRequest{}
	if window != "" {
		reqBody.BackupWindow = String(window)
	}
	reqBodyData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, nil, err
	}

	path := fmt.Sprintf("/servers/%d/actions/enable_backup", server.ID)
	req, err := c.client.NewRequest(ctx, "POST", path, bytes.NewReader(reqBodyData))
	if err != nil {
		return nil, nil, err
	}

	respBody := schema.ServerActionEnableBackupResponse{}
	resp, err := c.client.Do(req, &respBody)
	if err != nil {
		return nil, resp, err
	}
	return ActionFromSchema(respBody.Action), resp, nil
}

// DisableBackup disables backup for a server.
func (c *ServerClient) DisableBackup(ctx context.Context, server *Server) (*Action, *Response, error) {
	path := fmt.Sprintf("/servers/%d/actions/disable_backup", server.ID)
	req, err := c.client.NewRequest(ctx, "POST", path, nil)
	if err != nil {
		return nil, nil, err
	}

	respBody := schema.ServerActionDisableBackupResponse{}
	resp, err := c.client.Do(req, &respBody)
	if err != nil {
		return nil, resp, err
	}
	return ActionFromSchema(respBody.Action), resp, nil
}

// ServerChangeTypeOpts specifies options for changing a server's type.
type ServerChangeTypeOpts struct {
	ServerType  *ServerType // new server type
	UpgradeDisk bool        // whether disk should be upgraded
}

// ChangeType changes a server's type.
func (c *ServerClient) ChangeType(ctx context.Context, server *Server, opts ServerChangeTypeOpts) (*Action, *Response, error) {
	reqBody := schema.ServerActionChangeTypeRequest{
		UpgradeDisk: opts.UpgradeDisk,
	}
	if opts.ServerType.ID != 0 {
		reqBody.ServerType = opts.ServerType.ID
	} else {
		reqBody.ServerType = opts.ServerType.Name
	}
	reqBodyData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, nil, err
	}

	path := fmt.Sprintf("/servers/%d/actions/change_type", server.ID)
	req, err := c.client.NewRequest(ctx, "POST", path, bytes.NewReader(reqBodyData))
	if err != nil {
		return nil, nil, err
	}

	respBody := schema.ServerActionChangeTypeResponse{}
	resp, err := c.client.Do(req, &respBody)
	if err != nil {
		return nil, resp, err
	}
	return ActionFromSchema(respBody.Action), resp, nil
}

// ChangeDNSPtr changes or resets the reverse DNS pointer for a server IP address.
// Pass a nil ptr to reset the reverse DNS pointer to its default value.
func (c *ServerClient) ChangeDNSPtr(ctx context.Context, server *Server, ip string, ptr *string) (*Action, *Response, error) {
	reqBody := schema.ServerActionChangeDNSPtrRequest{
		IP:     ip,
		DNSPtr: ptr,
	}
	reqBodyData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, nil, err
	}

	path := fmt.Sprintf("/servers/%d/actions/change_dns_ptr", server.ID)
	req, err := c.client.NewRequest(ctx, "POST", path, bytes.NewReader(reqBodyData))
	if err != nil {
		return nil, nil, err
	}

	respBody := schema.ServerActionChangeDNSPtrResponse{}
	resp, err := c.client.Do(req, &respBody)
	if err != nil {
		return nil, resp, err
	}
	return ActionFromSchema(respBody.Action), resp, nil
}

// ServerChangeProtectionOpts specifies options for changing the resource protection level of a server.
type ServerChangeProtectionOpts struct {
	Rebuild *bool
	Delete  *bool
}

// ChangeProtection changes the resource protection level of a server.
func (c *ServerClient) ChangeProtection(ctx context.Context, image *Server, opts ServerChangeProtectionOpts) (*Action, *Response, error) {
	reqBody := schema.ServerActionChangeProtectionRequest{
		Rebuild: opts.Rebuild,
		Delete:  opts.Delete,
	}
	reqBodyData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, nil, err
	}

	path := fmt.Sprintf("/servers/%d/actions/change_protection", image.ID)
	req, err := c.client.NewRequest(ctx, "POST", path, bytes.NewReader(reqBodyData))
	if err != nil {
		return nil, nil, err
	}

	respBody := schema.ServerActionChangeProtectionResponse{}
	resp, err := c.client.Do(req, &respBody)
	if err != nil {
		return nil, resp, err
	}
	return ActionFromSchema(respBody.Action), resp, err
}
