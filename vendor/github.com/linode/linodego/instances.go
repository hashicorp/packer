package linodego

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"time"

	"github.com/linode/linodego/internal/parseabletime"
)

/*
 * https://developers.linode.com/v4/reference/endpoints/linode/instances
 */

// InstanceStatus constants start with Instance and include Linode API Instance Status values
type InstanceStatus string

// InstanceStatus constants reflect the current status of an Instance
const (
	InstanceBooting      InstanceStatus = "booting"
	InstanceRunning      InstanceStatus = "running"
	InstanceOffline      InstanceStatus = "offline"
	InstanceShuttingDown InstanceStatus = "shutting_down"
	InstanceRebooting    InstanceStatus = "rebooting"
	InstanceProvisioning InstanceStatus = "provisioning"
	InstanceDeleting     InstanceStatus = "deleting"
	InstanceMigrating    InstanceStatus = "migrating"
	InstanceRebuilding   InstanceStatus = "rebuilding"
	InstanceCloning      InstanceStatus = "cloning"
	InstanceRestoring    InstanceStatus = "restoring"
	InstanceResizing     InstanceStatus = "resizing"
)

// Instance represents a linode object
type Instance struct {
	ID              int             `json:"id"`
	Created         *time.Time      `json:"-"`
	Updated         *time.Time      `json:"-"`
	Region          string          `json:"region"`
	Alerts          *InstanceAlert  `json:"alerts"`
	Backups         *InstanceBackup `json:"backups"`
	Image           string          `json:"image"`
	Group           string          `json:"group"`
	IPv4            []*net.IP       `json:"ipv4"`
	IPv6            string          `json:"ipv6"`
	Label           string          `json:"label"`
	Type            string          `json:"type"`
	Status          InstanceStatus  `json:"status"`
	Hypervisor      string          `json:"hypervisor"`
	Specs           *InstanceSpec   `json:"specs"`
	WatchdogEnabled bool            `json:"watchdog_enabled"`
	Tags            []string        `json:"tags"`
}

// InstanceSpec represents a linode spec
type InstanceSpec struct {
	Disk     int `json:"disk"`
	Memory   int `json:"memory"`
	VCPUs    int `json:"vcpus"`
	Transfer int `json:"transfer"`
}

// InstanceAlert represents a metric alert
type InstanceAlert struct {
	CPU           int `json:"cpu"`
	IO            int `json:"io"`
	NetworkIn     int `json:"network_in"`
	NetworkOut    int `json:"network_out"`
	TransferQuota int `json:"transfer_quota"`
}

// InstanceBackup represents backup settings for an instance
type InstanceBackup struct {
	Enabled  bool `json:"enabled"`
	Schedule struct {
		Day    string `json:"day,omitempty"`
		Window string `json:"window,omitempty"`
	}
}

// InstanceTransfer pool stats for a Linode Instance during the current billing month
type InstanceTransfer struct {
	// Bytes of transfer this instance has consumed
	Used int `json:"used"`

	// GB of billable transfer this instance has consumed
	Billable int `json:"billable"`

	// GB of transfer this instance adds to the Transfer pool
	Quota int `json:"quota"`
}

// InstanceCreateOptions require only Region and Type
type InstanceCreateOptions struct {
	Region          string            `json:"region"`
	Type            string            `json:"type"`
	Label           string            `json:"label,omitempty"`
	Group           string            `json:"group,omitempty"`
	RootPass        string            `json:"root_pass,omitempty"`
	AuthorizedKeys  []string          `json:"authorized_keys,omitempty"`
	AuthorizedUsers []string          `json:"authorized_users,omitempty"`
	StackScriptID   int               `json:"stackscript_id,omitempty"`
	StackScriptData map[string]string `json:"stackscript_data,omitempty"`
	BackupID        int               `json:"backup_id,omitempty"`
	Image           string            `json:"image,omitempty"`
	BackupsEnabled  bool              `json:"backups_enabled,omitempty"`
	PrivateIP       bool              `json:"private_ip,omitempty"`
	Tags            []string          `json:"tags,omitempty"`

	// Creation fields that need to be set explicitly false, "", or 0 use pointers
	SwapSize *int  `json:"swap_size,omitempty"`
	Booted   *bool `json:"booted,omitempty"`
}

// InstanceUpdateOptions is an options struct used when Updating an Instance
type InstanceUpdateOptions struct {
	Label           string          `json:"label,omitempty"`
	Group           string          `json:"group,omitempty"`
	Backups         *InstanceBackup `json:"backups,omitempty"`
	Alerts          *InstanceAlert  `json:"alerts,omitempty"`
	WatchdogEnabled *bool           `json:"watchdog_enabled,omitempty"`
	Tags            *[]string       `json:"tags,omitempty"`
}

// UnmarshalJSON implements the json.Unmarshaler interface
func (i *Instance) UnmarshalJSON(b []byte) error {
	type Mask Instance

	p := struct {
		*Mask
		Created *parseabletime.ParseableTime `json:"created"`
		Updated *parseabletime.ParseableTime `json:"updated"`
	}{
		Mask: (*Mask)(i),
	}

	if err := json.Unmarshal(b, &p); err != nil {
		return err
	}

	i.Created = (*time.Time)(p.Created)
	i.Updated = (*time.Time)(p.Updated)

	return nil
}

// GetUpdateOptions converts an Instance to InstanceUpdateOptions for use in UpdateInstance
func (i *Instance) GetUpdateOptions() InstanceUpdateOptions {
	return InstanceUpdateOptions{
		Label:           i.Label,
		Group:           i.Group,
		Backups:         i.Backups,
		Alerts:          i.Alerts,
		WatchdogEnabled: &i.WatchdogEnabled,
		Tags:            &i.Tags,
	}
}

// InstanceCloneOptions is an options struct sent when Cloning an Instance
type InstanceCloneOptions struct {
	Region string `json:"region,omitempty"`
	Type   string `json:"type,omitempty"`

	// LinodeID is an optional existing instance to use as the target of the clone
	LinodeID       int    `json:"linode_id,omitempty"`
	Label          string `json:"label,omitempty"`
	Group          string `json:"group,omitempty"`
	BackupsEnabled bool   `json:"backups_enabled"`
	Disks          []int  `json:"disks,omitempty"`
	Configs        []int  `json:"configs,omitempty"`
}

// InstanceResizeOptions is an options struct used when resizing an instance
type InstanceResizeOptions struct {
	Type string `json:"type"`

	// When enabled, an instance resize will also resize a data disk if the instance has no more than one data disk and one swap disk
	AllowAutoDiskResize *bool `json:"allow_auto_disk_resize,omitempty"`
}

// InstancesPagedResponse represents a linode API response for listing
type InstancesPagedResponse struct {
	*PageOptions
	Data []Instance `json:"data"`
}

// endpoint gets the endpoint URL for Instance
func (InstancesPagedResponse) endpoint(c *Client) string {
	endpoint, err := c.Instances.Endpoint()
	if err != nil {
		panic(err)
	}
	return endpoint
}

// appendData appends Instances when processing paginated Instance responses
func (resp *InstancesPagedResponse) appendData(r *InstancesPagedResponse) {
	resp.Data = append(resp.Data, r.Data...)
}

// ListInstances lists linode instances
func (c *Client) ListInstances(ctx context.Context, opts *ListOptions) ([]Instance, error) {
	response := InstancesPagedResponse{}
	err := c.listHelper(ctx, &response, opts)

	if err != nil {
		return nil, err
	}
	return response.Data, nil
}

// GetInstance gets the instance with the provided ID
func (c *Client) GetInstance(ctx context.Context, linodeID int) (*Instance, error) {
	e, err := c.Instances.Endpoint()
	if err != nil {
		return nil, err
	}
	e = fmt.Sprintf("%s/%d", e, linodeID)
	r, err := coupleAPIErrors(c.R(ctx).
		SetResult(Instance{}).
		Get(e))
	if err != nil {
		return nil, err
	}
	return r.Result().(*Instance), nil
}

// GetInstanceTransfer gets the instance with the provided ID
func (c *Client) GetInstanceTransfer(ctx context.Context, linodeID int) (*InstanceTransfer, error) {
	e, err := c.Instances.Endpoint()
	if err != nil {
		return nil, err
	}
	e = fmt.Sprintf("%s/%d/transfer", e, linodeID)
	r, err := coupleAPIErrors(c.R(ctx).
		SetResult(InstanceTransfer{}).
		Get(e))
	if err != nil {
		return nil, err
	}
	return r.Result().(*InstanceTransfer), nil
}

// CreateInstance creates a Linode instance
func (c *Client) CreateInstance(ctx context.Context, instance InstanceCreateOptions) (*Instance, error) {
	var body string
	e, err := c.Instances.Endpoint()
	if err != nil {
		return nil, err
	}

	req := c.R(ctx).SetResult(&Instance{})

	if bodyData, err := json.Marshal(instance); err == nil {
		body = string(bodyData)
	} else {
		return nil, NewError(err)
	}

	r, err := coupleAPIErrors(req.
		SetBody(body).
		Post(e))

	if err != nil {
		return nil, err
	}
	return r.Result().(*Instance), nil
}

// UpdateInstance creates a Linode instance
func (c *Client) UpdateInstance(ctx context.Context, id int, instance InstanceUpdateOptions) (*Instance, error) {
	var body string
	e, err := c.Instances.Endpoint()
	if err != nil {
		return nil, err
	}
	e = fmt.Sprintf("%s/%d", e, id)

	req := c.R(ctx).SetResult(&Instance{})

	if bodyData, err := json.Marshal(instance); err == nil {
		body = string(bodyData)
	} else {
		return nil, NewError(err)
	}

	r, err := coupleAPIErrors(req.
		SetBody(body).
		Put(e))

	if err != nil {
		return nil, err
	}
	return r.Result().(*Instance), nil
}

// RenameInstance renames an Instance
func (c *Client) RenameInstance(ctx context.Context, linodeID int, label string) (*Instance, error) {
	return c.UpdateInstance(ctx, linodeID, InstanceUpdateOptions{Label: label})
}

// DeleteInstance deletes a Linode instance
func (c *Client) DeleteInstance(ctx context.Context, id int) error {
	e, err := c.Instances.Endpoint()
	if err != nil {
		return err
	}
	e = fmt.Sprintf("%s/%d", e, id)

	_, err = coupleAPIErrors(c.R(ctx).Delete(e))
	return err
}

// BootInstance will boot a Linode instance
// A configID of 0 will cause Linode to choose the last/best config
func (c *Client) BootInstance(ctx context.Context, id int, configID int) error {
	bodyStr := ""

	if configID != 0 {
		bodyMap := map[string]int{"config_id": configID}
		bodyJSON, err := json.Marshal(bodyMap)
		if err != nil {
			return NewError(err)
		}
		bodyStr = string(bodyJSON)
	}

	e, err := c.Instances.Endpoint()
	if err != nil {
		return err
	}

	e = fmt.Sprintf("%s/%d/boot", e, id)
	_, err = coupleAPIErrors(c.R(ctx).
		SetBody(bodyStr).
		Post(e))

	return err
}

// CloneInstance clone an existing Instances Disks and Configuration profiles to another Linode Instance
func (c *Client) CloneInstance(ctx context.Context, id int, options InstanceCloneOptions) (*Instance, error) {
	var body string
	e, err := c.Instances.Endpoint()
	if err != nil {
		return nil, err
	}
	e = fmt.Sprintf("%s/%d/clone", e, id)

	req := c.R(ctx).SetResult(&Instance{})

	if bodyData, err := json.Marshal(options); err == nil {
		body = string(bodyData)
	} else {
		return nil, NewError(err)
	}

	r, err := coupleAPIErrors(req.
		SetBody(body).
		Post(e))

	if err != nil {
		return nil, err
	}

	return r.Result().(*Instance), nil
}

// RebootInstance reboots a Linode instance
// A configID of 0 will cause Linode to choose the last/best config
func (c *Client) RebootInstance(ctx context.Context, id int, configID int) error {
	bodyStr := "{}"

	if configID != 0 {
		bodyMap := map[string]int{"config_id": configID}
		bodyJSON, err := json.Marshal(bodyMap)
		if err != nil {
			return NewError(err)
		}
		bodyStr = string(bodyJSON)
	}

	e, err := c.Instances.Endpoint()
	if err != nil {
		return err
	}

	e = fmt.Sprintf("%s/%d/reboot", e, id)

	_, err = coupleAPIErrors(c.R(ctx).
		SetBody(bodyStr).
		Post(e))

	return err
}

// InstanceRebuildOptions is a struct representing the options to send to the rebuild linode endpoint
type InstanceRebuildOptions struct {
	Image           string            `json:"image,omitempty"`
	RootPass        string            `json:"root_pass,omitempty"`
	AuthorizedKeys  []string          `json:"authorized_keys,omitempty"`
	AuthorizedUsers []string          `json:"authorized_users,omitempty"`
	StackScriptID   int               `json:"stackscript_id,omitempty"`
	StackScriptData map[string]string `json:"stackscript_data,omitempty"`
	Booted          *bool             `json:"booted,omitempty"`
}

// RebuildInstance Deletes all Disks and Configs on this Linode,
// then deploys a new Image to this Linode with the given attributes.
func (c *Client) RebuildInstance(ctx context.Context, id int, opts InstanceRebuildOptions) (*Instance, error) {
	o, err := json.Marshal(opts)
	if err != nil {
		return nil, NewError(err)
	}
	b := string(o)
	e, err := c.Instances.Endpoint()
	if err != nil {
		return nil, err
	}
	e = fmt.Sprintf("%s/%d/rebuild", e, id)
	r, err := coupleAPIErrors(c.R(ctx).
		SetBody(b).
		SetResult(&Instance{}).
		Post(e))
	if err != nil {
		return nil, err
	}
	return r.Result().(*Instance), nil
}

// InstanceRescueOptions fields are those accepted by RescueInstance
type InstanceRescueOptions struct {
	Devices InstanceConfigDeviceMap `json:"devices"`
}

// RescueInstance reboots an instance into a safe environment for performing many system recovery and disk management tasks.
// Rescue Mode is based on the Finnix recovery distribution, a self-contained and bootable Linux distribution.
// You can also use Rescue Mode for tasks other than disaster recovery, such as formatting disks to use different filesystems,
// copying data between disks, and downloading files from a disk via SSH and SFTP.
func (c *Client) RescueInstance(ctx context.Context, id int, opts InstanceRescueOptions) error {
	o, err := json.Marshal(opts)
	if err != nil {
		return NewError(err)
	}
	b := string(o)
	e, err := c.Instances.Endpoint()
	if err != nil {
		return err
	}
	e = fmt.Sprintf("%s/%d/rescue", e, id)

	_, err = coupleAPIErrors(c.R(ctx).
		SetBody(b).
		Post(e))

	return err
}

// ResizeInstance resizes an instance to new Linode type
func (c *Client) ResizeInstance(ctx context.Context, id int, opts InstanceResizeOptions) error {
	o, err := json.Marshal(opts)
	if err != nil {
		return NewError(err)
	}
	body := string(o)
	e, err := c.Instances.Endpoint()
	if err != nil {
		return err
	}
	e = fmt.Sprintf("%s/%d/resize", e, id)

	_, err = coupleAPIErrors(c.R(ctx).
		SetBody(body).
		Post(e))

	return err
}

// ShutdownInstance - Shutdown an instance
func (c *Client) ShutdownInstance(ctx context.Context, id int) error {
	return c.simpleInstanceAction(ctx, "shutdown", id)
}

// MutateInstance Upgrades a Linode to its next generation.
func (c *Client) MutateInstance(ctx context.Context, id int) error {
	return c.simpleInstanceAction(ctx, "mutate", id)
}

// MigrateInstance - Migrate an instance
func (c *Client) MigrateInstance(ctx context.Context, id int) error {
	return c.simpleInstanceAction(ctx, "migrate", id)
}

// simpleInstanceAction is a helper for Instance actions that take no parameters
// and return empty responses `{}` unless they return a standard error
func (c *Client) simpleInstanceAction(ctx context.Context, action string, id int) error {
	e, err := c.Instances.Endpoint()
	if err != nil {
		return err
	}
	e = fmt.Sprintf("%s/%d/%s", e, id, action)
	_, err = coupleAPIErrors(c.R(ctx).Post(e))
	return err
}
