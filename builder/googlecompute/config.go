//go:generate struct-markdown
//go:generate mapstructure-to-hcl2 -type Config,CustomerEncryptionKey

package googlecompute

import (
	"errors"
	"fmt"
	"os"
	"regexp"
	"time"

	"github.com/hashicorp/packer/common"
	"github.com/hashicorp/packer/common/uuid"
	"github.com/hashicorp/packer/helper/communicator"
	"github.com/hashicorp/packer/helper/config"
	"github.com/hashicorp/packer/packer"
	"github.com/hashicorp/packer/template/interpolate"
	"golang.org/x/oauth2/jwt"
	compute "google.golang.org/api/compute/v1"
)

// used for ImageName and ImageFamily
var validImageName = regexp.MustCompile(`^[a-z]([-a-z0-9]{0,61}[a-z0-9])?$`)

// Config is the configuration structure for the GCE builder. It stores
// both the publicly settable state as well as the privately generated
// state of the config object.
type Config struct {
	common.PackerConfig `mapstructure:",squash"`
	Comm                communicator.Config `mapstructure:",squash"`

	// The JSON file containing your account credentials. Not required if you
	// run Packer on a GCE instance with a service account. Instructions for
	// creating the file or using service accounts are above.
	AccountFile string `mapstructure:"account_file" required:"false"`
	// The project ID that will be used to launch instances and store images.
	ProjectId string `mapstructure:"project_id" required:"true"`
	// Full or partial URL of the guest accelerator type. GPU accelerators can
	// only be used with `"on_host_maintenance": "TERMINATE"` option set.
	// Example:
	// `"projects/project_id/zones/europe-west1-b/acceleratorTypes/nvidia-tesla-k80"`
	AcceleratorType string `mapstructure:"accelerator_type" required:"false"`
	// Number of guest accelerator cards to add to the launched instance.
	AcceleratorCount int64 `mapstructure:"accelerator_count" required:"false"`
	// The name of a pre-allocated static external IP address. Note, must be
	// the name and not the actual IP address.
	Address string `mapstructure:"address" required:"false"`
	// If true, the default service account will not be used if
	// service_account_email is not specified. Set this value to true and omit
	// service_account_email to provision a VM with no service account.
	DisableDefaultServiceAccount bool `mapstructure:"disable_default_service_account" required:"false"`
	// The name of the disk, if unset the instance name will be used.
	DiskName string `mapstructure:"disk_name" required:"false"`
	// The size of the disk in GB. This defaults to 10, which is 10GB.
	DiskSizeGb int64 `mapstructure:"disk_size" required:"false"`
	// Type of disk used to back your instance, like pd-ssd or pd-standard.
	// Defaults to pd-standard.
	DiskType string `mapstructure:"disk_type" required:"false"`
	// The unique name of the resulting image. Defaults to
	// "packer-{{timestamp}}".
	ImageName string `mapstructure:"image_name" required:"false"`
	// The description of the resulting image.
	ImageDescription string `mapstructure:"image_description" required:"false"`
	// Image encryption key to apply to the created image. Possible values:
	// * kmsKeyName -  The name of the encryption key that is stored in Google Cloud KMS.
	// * RawKey: - A 256-bit customer-supplied encryption key, encodes in RFC 4648 base64.
	//
	// example:
	//
	//  ``` json
	//  {
	//     "kmsKeyName": "projects/${project}/locations/${region}/keyRings/computeEngine/cryptoKeys/computeEngine/cryptoKeyVersions/4"
	//  }
	//  ```
	ImageEncryptionKey *CustomerEncryptionKey `mapstructure:"image_encryption_key" required:"false"`
	// The name of the image family to which the resulting image belongs. You
	// can create disks by specifying an image family instead of a specific
	// image name. The image family always returns its latest image that is not
	// deprecated.
	ImageFamily string `mapstructure:"image_family" required:"false"`
	// Key/value pair labels to apply to the created image.
	ImageLabels map[string]string `mapstructure:"image_labels" required:"false"`
	// Licenses to apply to the created image.
	ImageLicenses []string `mapstructure:"image_licenses" required:"false"`
	// A name to give the launched instance. Beware that this must be unique.
	// Defaults to "packer-{{uuid}}".
	InstanceName string `mapstructure:"instance_name" required:"false"`
	// Key/value pair labels to apply to the launched instance.
	Labels map[string]string `mapstructure:"labels" required:"false"`
	// The machine type. Defaults to "n1-standard-1".
	MachineType string `mapstructure:"machine_type" required:"false"`
	// Metadata applied to the launched instance.
	Metadata map[string]string `mapstructure:"metadata" required:"false"`
	// Metadata applied to the launched instance. Values are files.
	MetadataFiles map[string]string `mapstructure:"metadata_files"`
	// A Minimum CPU Platform for VM Instance. Availability and default CPU
	// platforms vary across zones, based on the hardware available in each GCP
	// zone.
	// [Details](https://cloud.google.com/compute/docs/instances/specify-min-cpu-platform)
	MinCpuPlatform string `mapstructure:"min_cpu_platform" required:"false"`
	// The Google Compute network id or URL to use for the launched instance.
	// Defaults to "default". If the value is not a URL, it will be
	// interpolated to
	// projects/((network_project_id))/global/networks/((network)). This value
	// is not required if a subnet is specified.
	Network string `mapstructure:"network" required:"false"`
	// The project ID for the network and subnetwork to use for launched
	// instance. Defaults to project_id.
	NetworkProjectId string `mapstructure:"network_project_id" required:"false"`
	// If true, the instance will not have an external IP. use_internal_ip must
	// be true if this property is true.
	OmitExternalIP bool `mapstructure:"omit_external_ip" required:"false"`
	// Sets Host Maintenance Option. Valid choices are `MIGRATE` and
	// `TERMINATE`. Please see [GCE Instance Scheduling
	// Options](https://cloud.google.com/compute/docs/instances/setting-instance-scheduling-options),
	// as not all machine\_types support `MIGRATE` (i.e. machines with GPUs).
	// If preemptible is true this can only be `TERMINATE`. If preemptible is
	// false, it defaults to `MIGRATE`
	OnHostMaintenance string `mapstructure:"on_host_maintenance" required:"false"`
	// If true, launch a preemptible instance.
	Preemptible bool `mapstructure:"preemptible" required:"false"`
	// The time to wait for instance state changes. Defaults to "5m".
	StateTimeout time.Duration `mapstructure:"state_timeout" required:"false"`
	// The region in which to launch the instance. Defaults to the region
	// hosting the specified zone.
	Region string `mapstructure:"region" required:"false"`
	// The service account scopes for launched
	// instance. Defaults to:
	//
	// ``` json
	// [
	//   "https://www.googleapis.com/auth/userinfo.email",
	//   "https://www.googleapis.com/auth/compute",
	//   "https://www.googleapis.com/auth/devstorage.full_control"
	// ]
	// ```
	Scopes []string `mapstructure:"scopes" required:"false"`
	// The service account to be used for launched instance. Defaults to the
	// project's default service account unless disable_default_service_account
	// is true.
	ServiceAccountEmail string `mapstructure:"service_account_email" required:"false"`
	// The source image to use to create the new image from. You can also
	// specify source_image_family instead. If both source_image and
	// source_image_family are specified, source_image takes precedence.
	// Example: "debian-8-jessie-v20161027"
	SourceImage string `mapstructure:"source_image" required:"true"`
	// The source image family to use to create the new image from. The image
	// family always returns its latest image that is not deprecated. Example:
	// "debian-8".
	SourceImageFamily string `mapstructure:"source_image_family" required:"true"`
	// The project ID of the project containing the source image.
	SourceImageProjectId string `mapstructure:"source_image_project_id" required:"false"`
	// The path to a startup script to run on the VM from which the image will
	// be made.
	StartupScriptFile string `mapstructure:"startup_script_file" required:"false"`
	// The Google Compute subnetwork id or URL to use for the launched
	// instance. Only required if the network has been created with custom
	// subnetting. Note, the region of the subnetwork must match the region or
	// zone in which the VM is launched. If the value is not a URL, it will be
	// interpolated to
	// projects/((network_project_id))/regions/((region))/subnetworks/((subnetwork))
	Subnetwork string `mapstructure:"subnetwork" required:"false"`
	// Assign network tags to apply firewall rules to VM instance.
	Tags []string `mapstructure:"tags" required:"false"`
	// If true, use the instance's internal IP instead of its external IP
	// during building.
	UseInternalIP bool `mapstructure:"use_internal_ip" required:"false"`
	// Can be set instead of account_file. If set, this builder will use
	// HashiCorp Vault to generate an Oauth token for authenticating against
	// Google's cloud. The value should be the path of the token generator
	// within vault.
	// For information on how to configure your Vault + GCP engine to produce
	// Oauth tokens, see https://www.vaultproject.io/docs/auth/gcp.html
	// You must have the environment variables VAULT_ADDR and VAULT_TOKEN set,
	// along with any other relevant variables for accessing your vault
	// instance. For more information, see the Vault docs:
	// https://www.vaultproject.io/docs/commands/#environment-variables
	// Example:`"vault_gcp_oauth_engine": "gcp/token/my-project-editor",`
	VaultGCPOauthEngine string `mapstructure:"vault_gcp_oauth_engine"`
	// The zone in which to launch the instance used to create the image.
	// Example: "us-central1-a"
	Zone string `mapstructure:"zone" required:"true"`

	account            *jwt.Config
	imageAlreadyExists bool
	ctx                interpolate.Context
}

func (c *Config) Prepare(raws ...interface{}) ([]string, error) {
	c.ctx.Funcs = TemplateFuncs
	err := config.Decode(c, &config.DecodeOpts{
		Interpolate:        true,
		InterpolateContext: &c.ctx,
		InterpolateFilter: &interpolate.RenderFilter{
			Exclude: []string{
				"run_command",
			},
		},
	}, raws...)
	if err != nil {
		return nil, err
	}

	var errs *packer.MultiError

	// Set defaults.
	if c.Network == "" && c.Subnetwork == "" {
		c.Network = "default"
	}

	if c.NetworkProjectId == "" {
		c.NetworkProjectId = c.ProjectId
	}

	if c.DiskSizeGb == 0 {
		c.DiskSizeGb = 10
	}

	if c.DiskType == "" {
		c.DiskType = "pd-standard"
	}

	if c.ImageDescription == "" {
		c.ImageDescription = "Created by Packer"
	}

	if c.OnHostMaintenance == "MIGRATE" && c.Preemptible {
		errs = packer.MultiErrorAppend(errs,
			errors.New("on_host_maintenance must be TERMINATE when using preemptible instances."))
	}
	// Setting OnHostMaintenance Correct Defaults
	//   "MIGRATE" : Possible and default if Preemptible is false
	//   "TERMINATE": Required if Preemptible is true
	if c.Preemptible {
		c.OnHostMaintenance = "TERMINATE"
	} else {
		if c.OnHostMaintenance == "" {
			c.OnHostMaintenance = "MIGRATE"
		}
	}

	// Make sure user sets a valid value for on_host_maintenance option
	if !(c.OnHostMaintenance == "MIGRATE" || c.OnHostMaintenance == "TERMINATE") {
		errs = packer.MultiErrorAppend(errs,
			errors.New("on_host_maintenance must be one of MIGRATE or TERMINATE."))
	}

	if c.ImageName == "" {
		img, err := interpolate.Render("packer-{{timestamp}}", nil)
		if err != nil {
			errs = packer.MultiErrorAppend(errs,
				fmt.Errorf("Unable to parse image name: %s ", err))
		} else {
			c.ImageName = img
		}
	}

	// used for ImageName and ImageFamily
	imageErrorText := "Invalid image %s %q: The first character must be a lowercase letter, and all following characters must be a dash, lowercase letter, or digit, except the last character, which cannot be a dash"

	if len(c.ImageName) > 63 {
		errs = packer.MultiErrorAppend(errs,
			errors.New("Invalid image name: Must not be longer than 63 characters"))
	}

	if !validImageName.MatchString(c.ImageName) {
		errs = packer.MultiErrorAppend(errs, errors.New(fmt.Sprintf(imageErrorText, "name", c.ImageName)))
	}

	if len(c.ImageFamily) > 63 {
		errs = packer.MultiErrorAppend(errs,
			errors.New("Invalid image family: Must not be longer than 63 characters"))
	}

	if c.ImageFamily != "" {
		if !validImageName.MatchString(c.ImageFamily) {
			errs = packer.MultiErrorAppend(errs, errors.New(fmt.Sprintf(imageErrorText, "family", c.ImageFamily)))
		}
	}

	if c.InstanceName == "" {
		c.InstanceName = fmt.Sprintf("packer-%s", uuid.TimeOrderedUUID())
	}

	if c.DiskName == "" {
		c.DiskName = c.InstanceName
	}

	if c.MachineType == "" {
		c.MachineType = "n1-standard-1"
	}

	if c.StateTimeout == 0 {
		c.StateTimeout = 5 * time.Minute
	}

	if es := c.Comm.Prepare(&c.ctx); len(es) > 0 {
		errs = packer.MultiErrorAppend(errs, es...)
	}

	// Process required parameters.
	if c.ProjectId == "" {
		errs = packer.MultiErrorAppend(
			errs, errors.New("a project_id must be specified"))
	}

	if c.Scopes == nil {
		c.Scopes = []string{
			"https://www.googleapis.com/auth/userinfo.email",
			"https://www.googleapis.com/auth/compute",
			"https://www.googleapis.com/auth/devstorage.full_control",
		}
	}

	if c.SourceImage == "" && c.SourceImageFamily == "" {
		errs = packer.MultiErrorAppend(
			errs, errors.New("a source_image or source_image_family must be specified"))
	}

	if c.Zone == "" {
		errs = packer.MultiErrorAppend(
			errs, errors.New("a zone must be specified"))
	}
	if c.Region == "" && len(c.Zone) > 2 {
		// get region from Zone
		region := c.Zone[:len(c.Zone)-2]
		c.Region = region
	}

	// Authenticating via an account file
	if c.AccountFile != "" {
		if c.VaultGCPOauthEngine != "" {
			errs = packer.MultiErrorAppend(errs, fmt.Errorf("You cannot "+
				"specify both account_file and vault_gcp_oauth_engine."))
		}
		cfg, err := ProcessAccountFile(c.AccountFile)
		if err != nil {
			errs = packer.MultiErrorAppend(errs, err)
		}
		c.account = cfg
	}

	if c.OmitExternalIP && c.Address != "" {
		errs = packer.MultiErrorAppend(fmt.Errorf("you can not specify an external address when 'omit_external_ip' is true"))
	}

	if c.OmitExternalIP && !c.UseInternalIP {
		errs = packer.MultiErrorAppend(fmt.Errorf("'use_internal_ip' must be true if 'omit_external_ip' is true"))
	}

	if c.AcceleratorCount > 0 && len(c.AcceleratorType) == 0 {
		errs = packer.MultiErrorAppend(fmt.Errorf("'accelerator_type' must be set when 'accelerator_count' is more than 0"))
	}

	if c.AcceleratorCount > 0 && c.OnHostMaintenance != "TERMINATE" {
		errs = packer.MultiErrorAppend(fmt.Errorf("'on_host_maintenance' must be set to 'TERMINATE' when 'accelerator_count' is more than 0"))
	}

	// If DisableDefaultServiceAccount is provided, don't allow a value for ServiceAccountEmail
	if c.DisableDefaultServiceAccount && c.ServiceAccountEmail != "" {
		errs = packer.MultiErrorAppend(fmt.Errorf("you may not specify a 'service_account_email' when 'disable_default_service_account' is true"))
	}

	if c.StartupScriptFile != "" {
		if _, err := os.Stat(c.StartupScriptFile); err != nil {
			errs = packer.MultiErrorAppend(
				errs, fmt.Errorf("startup_script_file: %v", err))
		}
	}

	// Check for any errors.
	if errs != nil && len(errs.Errors) > 0 {
		return nil, errs
	}

	return nil, nil
}

type CustomerEncryptionKey struct {
	// KmsKeyName: The name of the encryption key that is stored in Google
	// Cloud KMS.
	KmsKeyName string `json:"kmsKeyName,omitempty"`

	// RawKey: Specifies a 256-bit customer-supplied encryption key, encoded
	// in RFC 4648 base64 to either encrypt or decrypt this resource.
	RawKey string `json:"rawKey,omitempty"`
}

func (k *CustomerEncryptionKey) ComputeType() *compute.CustomerEncryptionKey {
	if k == nil {
		return nil
	}
	return &compute.CustomerEncryptionKey{
		KmsKeyName: k.KmsKeyName,
		RawKey:     k.RawKey,
	}
}
