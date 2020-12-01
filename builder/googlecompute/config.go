//go:generate struct-markdown
//go:generate mapstructure-to-hcl2 -type Config,CustomerEncryptionKey

package googlecompute

import (
	"errors"
	"fmt"
	"os"
	"regexp"
	"runtime"
	"time"

	"github.com/hashicorp/packer/helper/communicator"
	"github.com/hashicorp/packer/packer-plugin-sdk/common"
	packersdk "github.com/hashicorp/packer/packer-plugin-sdk/packer"
	"github.com/hashicorp/packer/packer-plugin-sdk/template/config"
	"github.com/hashicorp/packer/packer-plugin-sdk/template/interpolate"
	"github.com/hashicorp/packer/packer-plugin-sdk/uuid"
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
	// This allows service account impersonation as per the [docs](https://cloud.google.com/iam/docs/impersonating-service-accounts).
	ImpersonateServiceAccount string `mapstructure:"impersonate_service_account" required:"false"`
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
	// Create a Shielded VM image with Secure Boot enabled. It helps ensure that
	// the system only runs authentic software by verifying the digital signature
	// of all boot components, and halting the boot process if signature verification
	// fails. [Details](https://cloud.google.com/security/shielded-cloud/shielded-vm)
	EnableSecureBoot bool `mapstructure:"enable_secure_boot" required:"false"`
	// Create a Shielded VM image with virtual trusted platform module
	// Measured Boot enabled. A vTPM is a virtualized trusted platform module,
	// which is a specialized computer chip you can use to protect objects,
	// like keys and certificates, that you use to authenticate access to your
	// system. [Details](https://cloud.google.com/security/shielded-cloud/shielded-vm)
	EnableVtpm bool `mapstructure:"enable_vtpm" required:"false"`
	// Integrity monitoring helps you understand and make decisions about the
	// state of your VM instances. Note: integrity monitoring relies on having
	// vTPM enabled. [Details](https://cloud.google.com/security/shielded-cloud/shielded-vm)
	EnableIntegrityMonitoring bool `mapstructure:"enable_integrity_monitoring" required:"false"`
	// Whether to use an IAP proxy.
	IAPConfig `mapstructure:",squash"`
	// Skip creating the image. Useful for setting to `true` during a build test stage. Defaults to `false`.
	SkipCreateImage bool `mapstructure:"skip_create_image" required:"false"`
	// The unique name of the resulting image. Defaults to
	// `packer-{{timestamp}}`.
	ImageName string `mapstructure:"image_name" required:"false"`
	// The description of the resulting image.
	ImageDescription string `mapstructure:"image_description" required:"false"`
	// Image encryption key to apply to the created image. Possible values:
	// * kmsKeyName -  The name of the encryption key that is stored in Google Cloud KMS.
	// * RawKey: - A 256-bit customer-supplied encryption key, encodes in RFC 4648 base64.
	//
	// examples:
	//
	//  ```json
	//  {
	//     "kmsKeyName": "projects/${project}/locations/${region}/keyRings/computeEngine/cryptoKeys/computeEngine/cryptoKeyVersions/4"
	//  }
	//  ```
	//
	//  ```hcl
	//   image_encryption_key {
	//     kmsKeyName = "projects/${var.project}/locations/${var.region}/keyRings/computeEngine/cryptoKeys/computeEngine/cryptoKeyVersions/4"
	//   }
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
	// Storage location, either regional or multi-regional, where snapshot
	// content is to be stored and only accepts 1 value. Always defaults to a nearby regional or multi-regional
	// location.
	//
	// multi-regional example:
	//
	//  ```json
	//  {
	//     "image_storage_locations": ["us"]
	//  }
	//  ```
	// regional example:
	//
	//  ```json
	//  {
	//     "image_storage_locations": ["us-east1"]
	//  }
	//  ```
	ImageStorageLocations []string `mapstructure:"image_storage_locations" required:"false"`
	// A name to give the launched instance. Beware that this must be unique.
	// Defaults to `packer-{{uuid}}`.
	InstanceName string `mapstructure:"instance_name" required:"false"`
	// Key/value pair labels to apply to the launched instance.
	Labels map[string]string `mapstructure:"labels" required:"false"`
	// The machine type. Defaults to "n1-standard-1".
	MachineType string `mapstructure:"machine_type" required:"false"`
	// Metadata applied to the launched instance.
	// All metadata configuration values are expected to be of type string.
	// Google metadata options that take a value of `TRUE` or `FALSE` should be
	// set as a string (i.e  `"TRUE"` `"FALSE"` or `"true"` `"false"`).
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
	// `projects/((network_project_id))/global/networks/((network))`. This value
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
	// ```json
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
	// A list of project IDs to search for the source image. Packer will search the first
	// project ID in the list first, and fall back to the next in the list, until it finds the source image.
	SourceImageProjectId []string `mapstructure:"source_image_project_id" required:"false"`
	// The path to a startup script to run on the launched instance from which the image will
	// be made. When set, the contents of the startup script file will be added to the instance metadata
	// under the `"startup_script"` metadata property. See [Providing startup script contents directly](https://cloud.google.com/compute/docs/startupscript#providing_startup_script_contents_directly) for more details.
	//
	// When using `startup_script_file` the following rules apply:
	// - The contents of the script file will overwrite the value of the `"startup_script"` metadata property at runtime.
	// - The contents of the script file will be wrapped in Packer's startup script wrapper, unless `wrap_startup_script` is disabled. See `wrap_startup_script` for more details.
	// - Not supported by Windows instances. See [Startup Scripts for Windows](https://cloud.google.com/compute/docs/startupscript#providing_a_startup_script_for_windows_instances) for more details.
	StartupScriptFile string `mapstructure:"startup_script_file" required:"false"`
	// For backwards compatibility this option defaults to `"true"` in the future it will default to `"false"`.
	// If "true", the contents of `startup_script_file` or `"startup_script"` in the instance metadata
	// is wrapped in a Packer specific script that tracks the execution and completion of the provided
	// startup script. The wrapper ensures that the builder will not continue until the startup script has been executed.
	// - The use of the wrapped script file requires that the user or service account
	// running the build has the compute.instance.Metadata role.
	WrapStartupScriptFile config.Trilean `mapstructure:"wrap_startup_script" required:"false"`
	// The Google Compute subnetwork id or URL to use for the launched
	// instance. Only required if the network has been created with custom
	// subnetting. Note, the region of the subnetwork must match the region or
	// zone in which the VM is launched. If the value is not a URL, it will be
	// interpolated to
	// `projects/((network_project_id))/regions/((region))/subnetworks/((subnetwork))`
	Subnetwork string `mapstructure:"subnetwork" required:"false"`
	// Assign network tags to apply firewall rules to VM instance.
	Tags []string `mapstructure:"tags" required:"false"`
	// If true, use the instance's internal IP instead of its external IP
	// during building.
	UseInternalIP bool `mapstructure:"use_internal_ip" required:"false"`
	// If true, OSLogin will be used to manage SSH access to the compute instance by
	// dynamically importing a temporary SSH key to the Google account's login profile,
	// and setting the `enable-oslogin` to `TRUE` in the instance metadata.
	// Optionally, `use_os_login` can be used with an existing `ssh_username` and `ssh_private_key_file`
	// if a SSH key has already been added to the Google account's login profile - See [Adding SSH Keys](https://cloud.google.com/compute/docs/instances/managing-instance-access#add_oslogin_keys).
	//
	// SSH keys can be added to an individual user account
	//
	//```shell-session
	// $ gcloud compute os-login ssh-keys add --key-file=/home/user/.ssh/my-key.pub
	//
	// $ gcloud compute os-login describe-profile
	//PosixAccounts:
	//- accountId: <project-id>
	//  gid: '34567890754'
	//  homeDirectory: /home/user_example_com
	//  ...
	//  primary: true
	//  uid: '2504818925'
	//  username: /home/user_example_com
	//sshPublicKeys:
	//  000000000000000000000000000000000000000000000000000000000000000a:
	//    fingerprint: 000000000000000000000000000000000000000000000000000000000000000a
	//```
	//
	// Or SSH keys can be added to an associated service account
	//```shell-session
	// $ gcloud auth activate-service-account --key-file=<path to service account credentials file (e.g account.json)>
	// $ gcloud compute os-login ssh-keys add --key-file=/home/user/.ssh/my-key.pub
	//
	// $ gcloud compute os-login describe-profile
	//PosixAccounts:
	//- accountId: <project-id>
	//  gid: '34567890754'
	//  homeDirectory: /home/sa_000000000000000000000
	//  ...
	//  primary: true
	//  uid: '2504818925'
	//  username: sa_000000000000000000000
	//sshPublicKeys:
	//  000000000000000000000000000000000000000000000000000000000000000a:
	//    fingerprint: 000000000000000000000000000000000000000000000000000000000000000a
	//```
	UseOSLogin bool `mapstructure:"use_os_login" required:"false"`
	// Can be set instead of account_file. If set, this builder will use
	// HashiCorp Vault to generate an Oauth token for authenticating against
	// Google Cloud. The value should be the path of the token generator
	// within vault.
	// For information on how to configure your Vault + GCP engine to produce
	// Oauth tokens, see https://www.vaultproject.io/docs/auth/gcp
	// You must have the environment variables VAULT_ADDR and VAULT_TOKEN set,
	// along with any other relevant variables for accessing your vault
	// instance. For more information, see the Vault docs:
	// https://www.vaultproject.io/docs/commands/#environment-variables
	// Example:`"vault_gcp_oauth_engine": "gcp/token/my-project-editor",`
	VaultGCPOauthEngine string `mapstructure:"vault_gcp_oauth_engine"`
	// The zone in which to launch the instance used to create the image.
	// Example: "us-central1-a"
	Zone string `mapstructure:"zone" required:"true"`

	account            *ServiceAccount
	imageAlreadyExists bool
	ctx                interpolate.Context
}

func (c *Config) Prepare(raws ...interface{}) ([]string, error) {
	c.ctx.Funcs = TemplateFuncs
	err := config.Decode(c, &config.DecodeOpts{
		PluginType:         BuilderId,
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

	var errs *packersdk.MultiError

	// Set defaults.
	if c.Network == "" && c.Subnetwork == "" {
		c.Network = "default"
	}

	if c.NetworkProjectId == "" {
		c.NetworkProjectId = c.ProjectId
	}

	if c.DiskSizeGb == 0 {
		c.DiskSizeGb = 20
	}

	if c.DiskType == "" {
		c.DiskType = "pd-standard"
	}

	// Disabling the vTPM also disables integrity monitoring, because integrity
	// monitoring relies on data gathered by Measured Boot.
	if !c.EnableVtpm {
		if c.EnableIntegrityMonitoring {
			errs = packersdk.MultiErrorAppend(errs,
				errors.New("You cannot enable Integrity Monitoring when vTPM is disabled."))
		}
	}

	if c.ImageDescription == "" {
		c.ImageDescription = "Created by Packer"
	}

	if c.OnHostMaintenance == "MIGRATE" && c.Preemptible {
		errs = packersdk.MultiErrorAppend(errs,
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
		errs = packersdk.MultiErrorAppend(errs,
			errors.New("on_host_maintenance must be one of MIGRATE or TERMINATE."))
	}

	if c.ImageName == "" {
		img, err := interpolate.Render("packer-{{timestamp}}", nil)
		if err != nil {
			errs = packersdk.MultiErrorAppend(errs,
				fmt.Errorf("Unable to parse image name: %s ", err))
		} else {
			c.ImageName = img
		}
	}

	// used for ImageName and ImageFamily
	imageErrorText := "Invalid image %s %q: The first character must be a lowercase letter, and all following characters must be a dash, lowercase letter, or digit, except the last character, which cannot be a dash"

	if len(c.ImageName) > 63 {
		errs = packersdk.MultiErrorAppend(errs,
			errors.New("Invalid image name: Must not be longer than 63 characters"))
	}

	if !validImageName.MatchString(c.ImageName) {
		errs = packersdk.MultiErrorAppend(errs, fmt.Errorf(imageErrorText, "name", c.ImageName))
	}

	if len(c.ImageFamily) > 63 {
		errs = packersdk.MultiErrorAppend(errs,
			errors.New("Invalid image family: Must not be longer than 63 characters"))
	}

	if c.ImageFamily != "" {
		if !validImageName.MatchString(c.ImageFamily) {
			errs = packersdk.MultiErrorAppend(errs, fmt.Errorf(imageErrorText, "family", c.ImageFamily))
		}
	}

	if len(c.ImageStorageLocations) > 1 {
		errs = packersdk.MultiErrorAppend(errs,
			errors.New("Invalid image storage locations: Must not have more than 1 region"))
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

	// Set up communicator
	if es := c.Comm.Prepare(&c.ctx); len(es) > 0 {
		errs = packersdk.MultiErrorAppend(errs, es...)
	}

	// set defaults for IAP
	if c.IAPConfig.IAPHashBang == "" {
		if runtime.GOOS == "windows" {
			c.IAPConfig.IAPHashBang = ""
		} else {
			c.IAPConfig.IAPHashBang = "/bin/sh"
		}
	}
	if c.IAPConfig.IAPExt == "" {
		if runtime.GOOS == "windows" {
			c.IAPConfig.IAPExt = ".cmd"
		}
	}
	if c.IAPConfig.IAPTunnelLaunchWait == 0 {
		if c.Comm.Type == "winrm" {
			// when starting up, WinRM can cause the tunnel to take 30 seconds
			// before timing out
			c.IAPConfig.IAPTunnelLaunchWait = 40
		} else {
			c.IAPConfig.IAPTunnelLaunchWait = 30
		}
	}

	// Configure IAP: Update SSH config to use localhost proxy instead
	if c.IAPConfig.IAP {
		if !SupportsIAPTunnel(&c.Comm) {
			err := fmt.Errorf("Error: IAP tunnel is not implemented for %s communicator", c.Comm.Type)
			errs = packersdk.MultiErrorAppend(errs, err)
		}
		// These configuration values are copied early to the generic host parameter when configuring
		// StepConnect. As such they must be set now. Ideally we would handle this as part of
		// ApplyIAPTunnel and set them during StepStartTunnel but that means defering when the
		// CommHost function reads the value from the configuration, perhaps pass in b.config.Comm
		// instead of b.config.Comm.Host()?
		c.Comm.SSHHost = "localhost"
		c.Comm.WinRMHost = "localhost"
	}

	// Process required parameters.
	if c.ProjectId == "" {
		errs = packersdk.MultiErrorAppend(
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
		errs = packersdk.MultiErrorAppend(
			errs, errors.New("a source_image or source_image_family must be specified"))
	}

	if c.Zone == "" {
		errs = packersdk.MultiErrorAppend(
			errs, errors.New("a zone must be specified"))
	}
	if c.Region == "" && len(c.Zone) > 2 {
		// get region from Zone
		region := c.Zone[:len(c.Zone)-2]
		c.Region = region
	}

	// Authenticating via an account file
	if c.AccountFile != "" {
		if c.VaultGCPOauthEngine != "" && c.ImpersonateServiceAccount != "" {
			errs = packersdk.MultiErrorAppend(errs, fmt.Errorf("You cannot "+
				"specify impersonate_service_account, account_file and vault_gcp_oauth_engine at the same time"))
		}
		cfg, err := ProcessAccountFile(c.AccountFile)
		if err != nil {
			errs = packersdk.MultiErrorAppend(errs, err)
		}
		c.account = cfg
	}

	if c.OmitExternalIP && c.Address != "" {
		errs = packersdk.MultiErrorAppend(fmt.Errorf("you can not specify an external address when 'omit_external_ip' is true"))
	}

	if c.OmitExternalIP && !c.UseInternalIP {
		errs = packersdk.MultiErrorAppend(fmt.Errorf("'use_internal_ip' must be true if 'omit_external_ip' is true"))
	}

	if c.AcceleratorCount > 0 && len(c.AcceleratorType) == 0 {
		errs = packersdk.MultiErrorAppend(fmt.Errorf("'accelerator_type' must be set when 'accelerator_count' is more than 0"))
	}

	if c.AcceleratorCount > 0 && c.OnHostMaintenance != "TERMINATE" {
		errs = packersdk.MultiErrorAppend(fmt.Errorf("'on_host_maintenance' must be set to 'TERMINATE' when 'accelerator_count' is more than 0"))
	}

	// If DisableDefaultServiceAccount is provided, don't allow a value for ServiceAccountEmail
	if c.DisableDefaultServiceAccount && c.ServiceAccountEmail != "" {
		errs = packersdk.MultiErrorAppend(fmt.Errorf("you may not specify a 'service_account_email' when 'disable_default_service_account' is true"))
	}

	if c.StartupScriptFile != "" {
		if _, err := os.Stat(c.StartupScriptFile); err != nil {
			errs = packersdk.MultiErrorAppend(
				errs, fmt.Errorf("startup_script_file: %v", err))
		}

		if c.WrapStartupScriptFile == config.TriUnset {
			c.WrapStartupScriptFile = config.TriTrue
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
	KmsKeyName string `mapstructure:"kmsKeyName" json:"kmsKeyName,omitempty"`

	// RawKey: Specifies a 256-bit customer-supplied encryption key, encoded
	// in RFC 4648 base64 to either encrypt or decrypt this resource.
	RawKey string `mapstructure:"rawKey" json:"rawKey,omitempty"`
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

func SupportsIAPTunnel(c *communicator.Config) bool {
	switch c.Type {
	case "ssh", "winrm":
		return true
	default:
		return false
	}
}

func ApplyIAPTunnel(c *communicator.Config, port int) error {
	switch c.Type {
	case "ssh":
		c.SSHPort = port
		return nil
	case "winrm":
		c.WinRMPort = port
		return nil
	default:
		return fmt.Errorf("IAP tunnel is not implemented for %s communicator", c.Type)
	}
}
