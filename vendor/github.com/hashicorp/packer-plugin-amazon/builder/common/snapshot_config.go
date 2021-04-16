//go:generate struct-markdown

package common

import "github.com/hashicorp/packer-plugin-sdk/template/config"

// SnapshotConfig is for common configuration related to creating AMIs.
type SnapshotConfig struct {
	// Key/value pair tags to apply to snapshot. They will override AMI tags if
	// already applied to snapshot. This is a [template
	// engine](/docs/templates/legacy_json_templates/engine), see [Build template
	// data](#build-template-data) for more information.
	SnapshotTags map[string]string `mapstructure:"snapshot_tags" required:"false"`
	// Same as [`snapshot_tags`](#snapshot_tags) but defined as a singular
	// repeatable block containing a `key` and a `value` field. In HCL2 mode the
	// [`dynamic_block`](/docs/templates/hcl_templates/expressions#dynamic-blocks)
	// will allow you to create those programatically.
	SnapshotTag config.KeyValues `mapstructure:"snapshot_tag" required:"false"`
	// A list of account IDs that have
	// access to create volumes from the snapshot(s). By default no additional
	// users other than the user creating the AMI has permissions to create
	// volumes from the backing snapshot(s).
	SnapshotUsers []string `mapstructure:"snapshot_users" required:"false"`
	// A list of groups that have access to
	// create volumes from the snapshot(s). By default no groups have permission
	// to create volumes from the snapshot(s). all will make the snapshot
	// publicly accessible.
	SnapshotGroups []string `mapstructure:"snapshot_groups" required:"false"`
}
