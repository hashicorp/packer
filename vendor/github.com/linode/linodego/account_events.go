package linodego

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/linode/linodego/internal/duration"
	"github.com/linode/linodego/internal/parseabletime"
	"github.com/linode/linodego/pkg/errors"
)

// Event represents an action taken on the Account.
type Event struct {
	// The unique ID of this Event.
	ID int `json:"id"`

	// Current status of the Event, Enum: "failed" "finished" "notification" "scheduled" "started"
	Status EventStatus `json:"status"`

	// The action that caused this Event. New actions may be added in the future.
	Action EventAction `json:"action"`

	// A percentage estimating the amount of time remaining for an Event. Returns null for notification events.
	PercentComplete int `json:"percent_complete"`

	// The rate of completion of the Event. Only some Events will return rate; for example, migration and resize Events.
	Rate *string `json:"rate"`

	// If this Event has been read.
	Read bool `json:"read"`

	// If this Event has been seen.
	Seen bool `json:"seen"`

	// The estimated time remaining until the completion of this Event. This value is only returned for in-progress events.
	TimeRemaining *int `json:"-"`

	// The username of the User who caused the Event.
	Username string `json:"username"`

	// Detailed information about the Event's entity, including ID, type, label, and URL used to access it.
	Entity *EventEntity `json:"entity"`

	// Detailed information about the Event's secondary or related entity, including ID, type, label, and URL used to access it.
	SecondaryEntity *EventEntity `json:"secondary_entity"`

	// When this Event was created.
	Created *time.Time `json:"-"`
}

// EventAction constants start with Action and include all known Linode API Event Actions.
type EventAction string

// EventAction constants represent the actions that cause an Event. New actions may be added in the future.
const (
	ActionAccountUpdate            EventAction = "account_update"
	ActionAccountSettingsUpdate    EventAction = "account_settings_update"
	ActionBackupsEnable            EventAction = "backups_enable"
	ActionBackupsCancel            EventAction = "backups_cancel"
	ActionBackupsRestore           EventAction = "backups_restore"
	ActionCommunityQuestionReply   EventAction = "community_question_reply"
	ActionCommunityLike            EventAction = "community_like"
	ActionCreateCardUpdated        EventAction = "credit_card_updated"
	ActionDiskCreate               EventAction = "disk_create"
	ActionDiskDelete               EventAction = "disk_delete"
	ActionDiskUpdate               EventAction = "disk_update"
	ActionDiskDuplicate            EventAction = "disk_duplicate"
	ActionDiskImagize              EventAction = "disk_imagize"
	ActionDiskResize               EventAction = "disk_resize"
	ActionDNSRecordCreate          EventAction = "dns_record_create"
	ActionDNSRecordDelete          EventAction = "dns_record_delete"
	ActionDNSRecordUpdate          EventAction = "dns_record_update"
	ActionDNSZoneCreate            EventAction = "dns_zone_create"
	ActionDNSZoneDelete            EventAction = "dns_zone_delete"
	ActionDNSZoneUpdate            EventAction = "dns_zone_update"
	ActionFirewallCreate           EventAction = "firewall_create"
	ActionFirewallDelete           EventAction = "firewall_delete"
	ActionFirewallDisable          EventAction = "firewall_disable"
	ActionFirewallEnable           EventAction = "firewall_enable"
	ActionFirewallUpdate           EventAction = "firewall_update"
	ActionFirewallDeviceAdd        EventAction = "firewall_device_add"
	ActionFirewallDeviceRemove     EventAction = "firewall_device_remove"
	ActionHostReboot               EventAction = "host_reboot"
	ActionImageDelete              EventAction = "image_delete"
	ActionImageUpdate              EventAction = "image_update"
	ActionLassieReboot             EventAction = "lassie_reboot"
	ActionLinodeAddIP              EventAction = "linode_addip"
	ActionLinodeBoot               EventAction = "linode_boot"
	ActionLinodeClone              EventAction = "linode_clone"
	ActionLinodeCreate             EventAction = "linode_create"
	ActionLinodeDelete             EventAction = "linode_delete"
	ActionLinodeUpdate             EventAction = "linode_update"
	ActionLinodeDeleteIP           EventAction = "linode_deleteip"
	ActionLinodeMigrate            EventAction = "linode_migrate"
	ActionLinodeMutate             EventAction = "linode_mutate"
	ActionLinodeMutateCreate       EventAction = "linode_mutate_create"
	ActionLinodeReboot             EventAction = "linode_reboot"
	ActionLinodeRebuild            EventAction = "linode_rebuild"
	ActionLinodeResize             EventAction = "linode_resize"
	ActionLinodeResizeCreate       EventAction = "linode_resize_create"
	ActionLinodeShutdown           EventAction = "linode_shutdown"
	ActionLinodeSnapshot           EventAction = "linode_snapshot"
	ActionLinodeConfigCreate       EventAction = "linode_config_create"
	ActionLinodeConfigDelete       EventAction = "linode_config_delete"
	ActionLinodeConfigUpdate       EventAction = "linode_config_update"
	ActionLishBoot                 EventAction = "lish_boot"
	ActionLKENodeCreate            EventAction = "lke_node_create"
	ActionLongviewClientCreate     EventAction = "longviewclient_create"
	ActionLongviewClientDelete     EventAction = "longviewclient_delete"
	ActionLongviewClientUpdate     EventAction = "longviewclient_update"
	ActionManagedDisabled          EventAction = "managed_disabled"
	ActionManagedEnabled           EventAction = "managed_enabled"
	ActionManagedServiceCreate     EventAction = "managed_service_create"
	ActionManagedServiceDelete     EventAction = "managed_service_delete"
	ActionNodebalancerCreate       EventAction = "nodebalancer_create"
	ActionNodebalancerDelete       EventAction = "nodebalancer_delete"
	ActionNodebalancerUpdate       EventAction = "nodebalancer_update"
	ActionNodebalancerConfigCreate EventAction = "nodebalancer_config_create"
	ActionNodebalancerConfigDelete EventAction = "nodebalancer_config_delete"
	ActionNodebalancerConfigUpdate EventAction = "nodebalancer_config_update"
	ActionPasswordReset            EventAction = "password_reset"
	ActionPaymentSubmitted         EventAction = "payment_submitted"
	ActionStackScriptCreate        EventAction = "stackscript_create"
	ActionStackScriptDelete        EventAction = "stackscript_delete"
	ActionStackScriptUpdate        EventAction = "stackscript_update"
	ActionStackScriptPublicize     EventAction = "stackscript_publicize"
	ActionStackScriptRevise        EventAction = "stackscript_revise"
	ActionTFADisabled              EventAction = "tfa_disabled"
	ActionTFAEnabled               EventAction = "tfa_enabled"
	ActionTicketAttachmentUpload   EventAction = "ticket_attachment_upload"
	ActionTicketCreate             EventAction = "ticket_create"
	ActionTicketUpdate             EventAction = "ticket_update"
	ActionVolumeAttach             EventAction = "volume_attach"
	ActionVolumeClone              EventAction = "volume_clone"
	ActionVolumeCreate             EventAction = "volume_create"
	ActionVolumeDelte              EventAction = "volume_delete"
	ActionVolumeUpdate             EventAction = "volume_update"
	ActionVolumeDetach             EventAction = "volume_detach"
	ActionVolumeResize             EventAction = "volume_resize"
)

// EntityType constants start with Entity and include Linode API Event Entity Types
type EntityType string

// EntityType contants are the entities an Event can be related to.
const (
	EntityLinode       EntityType = "linode"
	EntityDisk         EntityType = "disk"
	EntityDomain       EntityType = "domain"
	EntityFirewall     EntityType = "firewall"
	EntityNodebalancer EntityType = "nodebalancer"
)

// EventStatus constants start with Event and include Linode API Event Status values
type EventStatus string

// EventStatus constants reflect the current status of an Event
const (
	EventFailed       EventStatus = "failed"
	EventFinished     EventStatus = "finished"
	EventNotification EventStatus = "notification"
	EventScheduled    EventStatus = "scheduled"
	EventStarted      EventStatus = "started"
)

// EventEntity provides detailed information about the Event's
// associated entity, including ID, Type, Label, and a URL that
// can be used to access it.
type EventEntity struct {
	// ID may be a string or int, it depends on the EntityType
	ID    interface{} `json:"id"`
	Label string      `json:"label"`
	Type  EntityType  `json:"type"`
	URL   string      `json:"url"`
}

// EventsPagedResponse represents a paginated Events API response
type EventsPagedResponse struct {
	*PageOptions
	Data []Event `json:"data"`
}

// endpoint gets the endpoint URL for Event
func (EventsPagedResponse) endpoint(c *Client) string {
	endpoint, err := c.Events.Endpoint()
	if err != nil {
		panic(err)
	}

	return endpoint
}

// UnmarshalJSON implements the json.Unmarshaler interface
func (i *Event) UnmarshalJSON(b []byte) error {
	type Mask Event

	p := struct {
		*Mask
		Created       *parseabletime.ParseableTime `json:"created"`
		TimeRemaining json.RawMessage              `json:"time_remaining"`
	}{
		Mask: (*Mask)(i),
	}

	if err := json.Unmarshal(b, &p); err != nil {
		return err
	}

	i.Created = (*time.Time)(p.Created)
	i.TimeRemaining = duration.UnmarshalTimeRemaining(p.TimeRemaining)

	return nil
}

// endpointWithID gets the endpoint URL for a specific Event
func (i Event) endpointWithID(c *Client) string {
	endpoint, err := c.Events.Endpoint()
	if err != nil {
		panic(err)
	}

	endpoint = fmt.Sprintf("%s/%d", endpoint, i.ID)

	return endpoint
}

// appendData appends Events when processing paginated Event responses
func (resp *EventsPagedResponse) appendData(r *EventsPagedResponse) {
	resp.Data = append(resp.Data, r.Data...)
}

// ListEvents gets a collection of Event objects representing actions taken
// on the Account. The Events returned depend on the token grants and the grants
// of the associated user.
func (c *Client) ListEvents(ctx context.Context, opts *ListOptions) ([]Event, error) {
	response := EventsPagedResponse{}
	err := c.listHelper(ctx, &response, opts)

	if err != nil {
		return nil, err
	}

	return response.Data, nil
}

// GetEvent gets the Event with the Event ID
func (c *Client) GetEvent(ctx context.Context, id int) (*Event, error) {
	e, err := c.Events.Endpoint()
	if err != nil {
		return nil, err
	}

	e = fmt.Sprintf("%s/%d", e, id)
	r, err := c.R(ctx).SetResult(&Event{}).Get(e)

	if err != nil {
		return nil, err
	}

	return r.Result().(*Event), nil
}

// MarkEventRead marks a single Event as read.
func (c *Client) MarkEventRead(ctx context.Context, event *Event) error {
	e := event.endpointWithID(c)
	e = fmt.Sprintf("%s/read", e)

	_, err := errors.CoupleAPIErrors(c.R(ctx).Post(e))

	return err
}

// MarkEventsSeen marks all Events up to and including this Event by ID as seen.
func (c *Client) MarkEventsSeen(ctx context.Context, event *Event) error {
	e := event.endpointWithID(c)
	e = fmt.Sprintf("%s/seen", e)

	_, err := errors.CoupleAPIErrors(c.R(ctx).Post(e))

	return err
}
