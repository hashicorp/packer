## 1.8.6 (Upcoming)

### IMPROVEMENTS:
* bump github.com/hashicorp/hcp-sdk-go from 0.28.0 to 0.29.0.
     [GH-12163](https://github.com/hashicorp/packer/pull/12163)
* hcp: iteration fingerprints used to be computed from the Git SHA of the repository
       where the template is located when running packer build. This changes with this
       release, and now fingerprints are automatically generated as a ULID. This implies
       that continuing an existing iteration will require users to define the
       fingerprint in the environment manually in order to adopt this behaviour,
       otherwise, by default, a new iteration will be created. This does not impact
       workflows where the fingerprint was defined through the
       HCP_PACKER_ITERATION_FINGERPRINT environment variable, and these builds will work
       exactly as they did before.

### BUG FIXES: 
* core/hcl2: Templates with build blocks referencing an unknown source block
     would display an empty string for the template filename at line 0, which made
     it difficult to identify the broken build block. Packer has been updated to
     display the proper filename and line number where the unknown reference
     resides. [GH-12167](https://github.com/hashicorp/packer/pull/12167)

### NOTES:
* core: Users will see some changes in how names are displayed during a Packer
     build for JSON templates. Previously only the builder type or the builder
     name, if it was set, would be displayed. Now for named builders
     (`"name":"mybuilder"`) the builder'ss type and name will be displayed (i.e
     "<type>.mybuilder". This does not impact the behavior of options such as
     only or except, they will continue to work as they did before.)

## 1.8.5 (December 12, 2022)

### NOTES:
* data/packer-image-iteration has been removed. This was an undocumented and
     unusable data source that was built for experimentation but not released. It
     should not affect users in any way but is being mentioned for visibility
     purposes.
     [GH-12111](https://github.com/hashicorp/packer/pull/12111)

### FEATURES:
* core: Metadata for capturing template types such as JSON or HCL2 has been added to the
     HCP Packer registry metadata. Upon running a `packer build` the type of
     template used during execution will be sent along to the registry as
     additional build metadata.[GH-12132](https://github.com/hashicorp/packer/pull/12132)

### PLUGINS:
The following external plugins have been updated and pinned to address open
    issues. Please see their respective changelogs for details on plugin
    specific bug fixes and improvements.

* alicloud@v1.0.5 - [CHANGELOG](https://github.com/hashicorp/packer-plugin-alicloud/releases/tag/v1.0.5)
* amazon@v1.1.6 - [CHANGELOG](https://github.com/hashicorp/packer-plugin-amazon/releases/tag/v1.1.6)
* proxmox@v1.1.0 - [CHANGELOG](https://github.com/hashicorp/packer-plugin-proxmox/releases/tag/v1.1.0)
* vsphere@v1.1.1 - [CHANGELOG](https://github.com/hashicorp/packer-plugin-vsphere/releases/tag/v1.1.1)
* qemu@v1.0.8 - [CHANGELOG](https://github.com/hashicorp/packer-plugin-qemu/releases/tag/v1.0.8)

### IMPROVEMENTS:
* cmd/hcl2_upgrade: Generate variable block for all referenced user input
     variables. [GH-12136](https://github.com/hashicorp/packer/pull/12136)
* cmd/validate: Add support for the `-evaluate-datasources` flag to evaluate
     the data sources from a template during validation time.
     [GH-12106](https://github.com/hashicorp/packer/pull/12106)
     [GH-12152](https://github.com/hashicorp/packer/pull/12152)
* core/hcl2: Variable definition files containing undeclared variables within
     an HCL2 template will no longer warn during build execution. Warnings will
     be displayed to a user during template validation, which can be disabled by
     passing the `-no-warn-undeclared-var` flag to the validate command.
     [GH-12104](https://github.com/hashicorp/packer/pull/12104)
     [GH-12109](https://github.com/hashicorp/packer/pull/12109)
* core: Docker images have been updated to include the `xorriso` package for
     supporting the creation of ISO files.[GH-12081](https://github.com/hashicorp/packer/pull/12081)
* core: Split HCP Packer publishing components into a separate internal-only
     module. [GH-11](https://github.com/hashicorp/packer/pull/11967)
     [GH-12116](https://github.com/hashicorp/packer/pull/12116)

### BUG FIXES:
* cmd/init: The init command will try to fallback to the next available version
     for a plugin, if the most recent version is not available or has a missing
     checksum file. This should prevent Packer from trying to install versions
     that have a GitHub tag but no actual assets.
     [GH-12103](https://github.com/hashicorp/packer/pull/12103)
* cmd/plugins: The `plugins install` sub-command will try to fallback to the
     next available version for a plugin, if the most recent version is not
     available or has a missing checksum file. This should prevent Packer from
     trying to install versions that have a GitHub tag but no actual assets.
     [GH-12103](https://github.com/hashicorp/packer/pull/12103)
* core: Bump Go version to 1.18.9 to address vulnerability GO-2022-1144, which
     concerns the net/http and golang.org/x/net
     packages.[GH-12153](https://github.com/hashicorp/packer/pull/12153)
     [GH-12158](https://github.com/hashicorp/packer/pull/12158)
* core: Fix the registration of deleted input artifacts in HCP Packer when
     setting `keep_input_artifacts` to
     false.[GH-11462](https://github.com/hashicorp/packer/pull/11967)
* core: Using different template types when building a HCP Packer
     image iteration will now trigger a runtime build error due to the 
     mixing of template types.
     [GH-12132](https://github.com/hashicorp/packer/pull/12132)

## 1.8.4 (October 28 2022)

### NOTES:

* Packer user documentation has been moved to the new HashiCorp developer
     documentation portal. The main Packer site (https://www.packer.io) will
     continue to be the home for Packer but all requests for general
     documentation and binary downloads will be redirected to
     https://developer.hashicorp.com/packer. Users are encouraged to visit the
     developer documentation portal for access to all Packer related
     documentation; including integrations with HCP Packer.

* The Oracle builder and post-processor are no longer vendored with Packer
     core. Users of the Oracle plugin should use `packer init` to install the
     latest version of the plugin. See the [Oracle Plugin
     Documentation](https://github.com/hashicorp/packer-plugin-oracle) for more
     information. [GH-11983](https://github.com/hashicorp/packer/pull/11983)

* HCP Packer environment variables: The behavior of some HCP Packer-specific
     environment variables have changed slightly. Refer to [HCP Packer](https://developer.hashicorp.com/packer/docs/hcp)
     in the Packer documentation for a full list of HCP Packer environment variables. [GH-12059](https://github.com/hashicorp/packer/pull/12059)
    - For JSON templates, the `HCP_PACKER_REGISTRY` environment variable was
     previously required to enable the HCP Packer integration. In this release,
     the environment variable is now optional, and can be used for disabling the
     publishing of metadata for any HCP Packer enabled configuration template.
    - For HCL2 templates, the `HCP_PACKER_REGISTRY` environment variable can be used
     to disable publishing to a HCP Packer registry even if the template defines a
     `hcp_packer_registry` block. This can be useful for testing that a template
     works as intended prior to pushing metadata to HCP Packer.
    - The `HCP_PACKER_BUCKET_NAME` environment variable is now the only
     requirement to push metadata to a HCP Packer registry, in both JSON 
     and HCL2 templates without a `hcp_packer_registry` block.

### FEATURES:

* provisioner/powershell: Add `use_pwsh` configuration argument to support pwsh
     in powershell provisioner. [GH-11950](https://github.com/hashicorp/packer/pull/11950)

### PLUGINS:

* builder/oracle: Remove Oracle plugin from the list of vendored
     plugins. [GH-11983](https://github.com/hashicorp/packer/pull/11983)

### IMPROVEMENTS:

* command/hcl2_upgrade: Has been updated to persist all possible template
     engine options that were supported by the legacy JSON templates. While the
     upgrade command copies the template engine options as is support for the
     template options may not actually work with HCL2 templates; indicated by an
     error similar to `fieldname type <no value> is invalid`. Before executing a
     build with the upgraded template you are encouraged to run `packer validate`
     against the template and fix any invalid `<no value>` references.
     [GH-12068](https://github.com/hashicorp/packer/pull/12068)
* core/hcl2: Packer will now report an error when executing a build with no
     sources selected for execution.
     [GH-12016](https://github.com/hashicorp/packer/pull/12016)
* core/hcp: Configuration errors for HCP Packer enabled builds have been
     consolidated into a single report to help users address all potential
     issues before retrying their build.
     [GH-12031](https://github.com/hashicorp/packer/pull/12031)
* core/hcp: Named builds within a legacy JSON template are now published to a
     HCP Packer registry using its full build name (e.g `happycloud.windows-srv-2019)`, 
     as opposed to just the build name field (e.g `"name"="windows-srv-2019"`). 
     Builders with no defined name will continue to publish build
     metadate using the builder type as the build name (e.g `happycloud`).
     [GH-12059](https://github.com/hashicorp/packer/pull/12059)
* core:hcl2: When a variable is set in a variables definitions file (i.e
     \*.pkrvars.hcl), but isn't defined with the template files (i.e
     \*.pkr.hcl), the outputted error message will now include an example of
     variable block that can be added to the build template to remedy the issue.
     [GH-12020](https://github.com/hashicorp/packer/pull/12020)
* core: Add ppc64le to binary releases for Linux.
     [GH-11966](https://github.com/hashicorp/packer/pull/11966)
* core: Bump github.com/hashicorp/packer-plugin-sdk from 0.3.1 to 0.3.2.
     [GH-11981](https://github.com/hashicorp/packer/pull/11981)
* core: Bump supported Go version to 1.18.
     [GH-11927](https://github.com/hashicorp/packer/pull/11927)

### BUG FIXES:
* command/hcl2_upgrade: special case: Azure `shared_image_gallery` fix.
     [GH-12087](https://github.com/hashicorp/packer/pull/12087)
* core: Bump golang.org/x/sys to address CVE-2022-29526.
     [GH-11953](https://github.com/hashicorp/packer/pull/11953)
* core: Bump golang.org/x/text to v0.3.8.
     [GH-12047](https://github.com/hashicorp/packer/pull/12047)
* core: Update dependency to resolve GO-2022-0969.
     [GH-12009](https://github.com/hashicorp/packer/pull/12009)

## 1.8.3 (August 2, 2022)

### NOTES:
* There's been a change in the way the `ssh_timeout` and the
    `ssh_handshake_attempts` configuration arguments work together. The
    behaviour is unchanged if both or none are specified. However, if only one
    of the two is set the other won't have a default value anymore and will be
    ignored. See [Packer Plugin SDK change](https://github.com/hashicorp/packer-plugin-sdk/pull/116) for details

* packer-plugin-digitalocean: The Digital Ocean Packer plugin has been handed over
    to the Digital Ocean team. New releases for this plugin are available at
    https://github.com/digitalocean/packer-plugin-digitalocean. This plugin is
    still being bundled in the Packer binary but will be removed in a future
    release. Existing references to the plugin will continue to work but
    users are advised to update the required_plugins block to use the new
    plugin source address.
    [GH-11912](https://github.com/hashicorp/packer/pull/11912)
```
required_plugins {
    digitalocean = {
     source =  "github.com/digitalocean/digitalocean"
     version = ">=1.0.8"
    }
}
```
* packer-plugin-outscale:  The Outscale Packer plugin managed by the Outscale
    team, since Packer 1.7.9, has been removed from the Packer binary. Users are
    advised to install the latest version of the plugin by running
    `packer plugins install github.com/outscale/outscale`. [GH-11912](https://github.com/hashicorp/packer/pull/11912)

* packer-plugin-outscale:  The Scaleway Packer plugin managed by the Scaleway
    team, since Packer 1.7.7, has been removed from the Packer binary. Users are
    advised to install the latest version of the plugin by running
    `packer plugins install github.com/scaleway/scaleway`. [GH-11912](https://github.com/hashicorp/packer/pull/11912)

### FEATURES:
* Future Scaffolding: This release contains additional changes that allow
    Packer core to validate that a newly built image is a direct child of a HCP
    Packer registry source image. This feature is only available for HCP Packer
    enabled builds using the `hcp_packer_image` and `hcp_packer_iteration` data
    source for setting a builder's source image.
    [GH-11861](https://github.com/hashicorp/packer/pull/11861)

### PLUGINS:

The following external plugins have been updated and pinned to address open
    issues. Please see their respective changelogs for details on plugin
    specific bug fixes and improvements.

* amazon@v1.1.2 - [CHANGELOG](https://github.com/hashicorp/packer-plugin-amazon/releases/tag/v1.1.2)
* ansible@v1.0.3 - [CHANGELOG](https://github.com/hashicorp/packer-plugin-ansible/releases/tag/v1.0.3)
* azure@v1.3.0 - [CHANGELOG](https://github.com/hashicorp/packer-plugin-azure/releases/tag/v1.3.0)
* docker@v1.0.7 - [CHANGELOG](https://github.com/hashicorp/packer-plugin-docker/releases/tag/v1.0.7)
* googlecompute@v1.0.14 - [CHANGELOG](https://github.com/hashicorp/packer-plugin-googlecompute/releases/tag/v1.0.14)
* lxc@v1.0.2 - [CHANGELOG](https://github.com/hashicorp/packer-plugin-lxc/releases/tag/v1.0.2)
* triton@v1.0.2 - [CHANGELOG](https://github.com/hashicorp/packer-plugin-triton/releases/tag/v1.0.2)
* vsphere@v1.0.7 - [CHANGELOG](https://github.com/hashicorp/packer-plugin-vsphere/releases/tag/v1.0.7)
* yandex@v1.1.2 - [CHANGELOG](https://github.com/hashicorp/packer-plugin-yandex/releases/tag/v1.1.2)


### IMPROVEMENTS:
* Bump packer-plugin-sdk to v0.3.1 to fix inconsistencies between `ssh_timeout`
    and `ssh_handshake_attempts` configuration arguments in the SSH
    communicator. [GH-11909](https://github.com/hashicorp/packer/pull/11909)
* core: During long running builds the HCP Packer registry will mark a build as
    timed out if it has not posted an update after a certain number of minutes.
    For HCP Packer enabled builds a status update will now be sent every 2
    minutes to the registry to prevent long builds from being marked as timed
    out. [GH-11846](https://github.com/hashicorp/packer/pull/11846)
* data/hcp_packer_image: Add `component_type` configuration argument to
    support specifying an exact build image when multiple images exist in the
    same provider and region for a given HCP Packer bucket iteration.
    [GH-11872](https://github.com/hashicorp/packer/pull/11872)
* data/hcp_packer_image: Add support for `channel` as input argument to
    retrieve an image from the associated iteration. If
    using several images from a single iteration, you may prefer sourcing an
    iteration first, and referencing it for subsequent uses, as every
    `hcp_packer_image` with the channel set will generate a potentially
    billable HCP Packer request but if several `hcp_packer_image`s use a
    shared `hcp_packer_iteration` that will only generate one potentially
    billable request.
    [GH-11865](https://github.com/hashicorp/packer/pull/11865)



### BUG FIXES

* core/hcl2: Fix crash when parsing malformed provisioner override blocks.
    [GH-11881](https://github.com/hashicorp/packer/pull/11881)
* core/hcl2: Fix crash when running `packer validate` on templates containing
    one or more  HCP Packer data sources.
    [GH-11883](https://github.com/hashicorp/packer/pull/11883)


## 1.8.2 (June 21, 2022)

### NOTES:
* The Packer plugin SDK includes the latest version of the go-getter library,
    which has been updated to address the vulnerabilities listed in
    [HCSEC-2022-13](https://discuss.hashicorp.com/t/hcsec-2022-13-multiple-vulnerabilities-in-go-getter-library/39930).
    The updated SDK contains changes that can
    be breaking for some plugins as the updated go-getter settings in the SDK
    prevent reading/writing through symlinks and to sub-directories that
    require upward path traversal (e.g /tmp/.../etc/hosts). The updates also
    includes a 30 minute maximum timeout for file downloading, which can be an
    issue for very large or slow downloads if they exceed more than 30 minutes
    to complete.

### SECURITY:
* Bump packer-plugin-sdk to v0.3.0 to address reported vulnerabilities within
    the go-getter library.
    [GH-11843](https://github.com/hashicorp/packer/pull/11843)
* Bump plugins relying on go-getter for downloading remote files to address
    reported vulnerabilities within the go-getter library. See [HCSEC-2022-13](https://discuss.hashicorp.com/t/hcsec-2022-13-multiple-vulnerabilities-in-go-getter-library/39930) for details.
    [GH-11844](https://github.com/hashicorp/packer/pull/11844)

### FEATURES:
* Future Scaffolding: This release contains changes that allow Packer core to
    validate that a newly built image is a direct child of a HCP Packer
    registry source image. This feature is only available for HCP Packer
    enabled builds using the `hcp_packer_image` data source for setting a
    builder's source image.
    [GH-11832](https://github.com/hashicorp/packer/pull/11832)

### PLUGINS:

External plugins have been pinned to the following versions. Please see their
    respective changelogs for details on plugin specific bug fixes and
    improvements.

* azure@v1.1.0 - [CHANGELOG](https://github.com/hashicorp/packer-plugin-azure/releases/tag/v1.1.0)
* hyperv@v1.0.4 - [CHANGELOG](https://github.com/hashicorp/packer-plugin-hyperv/releases/tag/v1.0.4)
* parallels@v1.0.3 - [CHANGELOG](https://github.com/hashicorp/packer-plugin-parallels/releases/tag/v1.0.3)
* proxmox@v1.0.8 - [CHANGELOG](https://github.com/hashicorp/packer-plugin-proxmox/releases/tag/v1.0.8)
* qemu@v1.0.5 - [CHANGELOG](https://github.com/hashicorp/packer-plugin-qemu/releases/tag/v1.0.5)
* vagrant@v1.0.3 - [CHANGELOG](https://github.com/hashicorp/packer-plugin-vagrant/releases/tag/v1.0.3)
* virtualbox@v1.0.4 - [CHANGELOG](https://github.com/hashicorp/packer-plugin-virtualbox/releases/tag/v1.0.4)
* vmware@v1.0.7 - [CHANGELOG](https://github.com/hashicorp/packer-plugin-vmware/releases/tag/v1.0.7)
* vsphere@v1.0.5 - [CHANGELOG](https://github.com/hashicorp/packer-plugin-vsphere/releases/tag/v1.0.5)

### IMPROVEMENTS:
* Add `pause_after` configuration argument to Powershell provisioner.
    [GH-11792](https://github.com/hashicorp/packer/pull/11792)
* HCP Packer data sources will now fail for revoked iterations to prevent building non-compliant images.
    [GH-11854](https://github.com/hashicorp/packer/pull/11854)

### BUG FIXES:

* Add missing support for the `env` configuration argument in remote shell
    provisioners. [GH-11819](https://github.com/hashicorp/packer/pull/11819)
* The preinst and postrm user scripts, including the service configuration
    directives, have been removed from the Packer rpm installations packages,
    as Packer does not require a service user in order to run.
    [GH-11831](https://github.com/hashicorp/packer/pull/11831)

## 1.8.1 (May 27, 2022)

### NOTES:
* All bundled plugins have been updated to their latest release to address any
    open issues pertaining to the legacy SSH key algorithm fix made to the
    Packer plugin SDK.
    [GH-11761](https://github.com/hashicorp/packer/pull/11761)
    [GH-11802](https://github.com/hashicorp/packer/pull/11802)
* This release contains the latest golang.org/x/crypto/ssh module which
    implements client authentication support for signature algorithms based on
    SHA-2 for use with existing RSA keys. Previously, a client would fail to
    authenticate with RSA keys to servers that reject signature algorithms
    based on SHA-1.

### FEATURES:
* **New Datasource** HTTP data source retrieves information from a HTTP
    endpoint to be used during Packer builds.
    [GH-11658](https://github.com/hashicorp/packer/pull/11658)

### PLUGINS:

External plugins have been pinned to the following versions. Please see their
    respective changelogs for details on plugin specific bug fixes and
    improvements.

* alicloud@v1.0.2 - [CHANGELOG](https://github.com/hashicorp/packer-plugin-alicloud/releases/tag/v1.0.2)
* amazon@v1.1.0 - [CHANGELOG](https://github.com/hashicorp/packer-plugin-amazon/releases/tag/v1.1.0)
* ansible@v1.0.2 - [CHANGELOG](https://github.com/hashicorp/packer-plugin-ansible/releases/tag/v1.0.2)
* azure@v1.0.8 - [CHANGELOG](https://github.com/hashicorp/packer-plugin-azure/releases/tag/v1.0.8)
* chef@v1.0.2 - [CHANGELOG](https://github.com/hashicorp/packer-plugin-chef/releases/tag/v1.0.2)
* cloudstack@v1.0.1 - [CHANGELOG](https://github.com/hashicorp/packer-plugin-cloudstack/releases/tag/v1.0.1)
* converge@v1.0.1 - [CHANGELOG](https://github.com/hashicorp/packer-plugin-converge/releases/tag/v1.0.1)
* digitalocean@v1.0.3 - [CHANGELOG](https://github.com/hashicorp/packer-plugin-digitalocean/releases/tag/v1.0.3)
* docker@v1.0.5 - [CHANGELOG](https://github.com/hashicorp/packer-plugin-docker/releases/tag/v1.0.5)
* googlecompute@v1.0.13 - [CHANGELOG](https://github.com/hashicorp/packer-plugin-googlecompute/releases/tag/v1.0.13)
* hcloud@v1.0.4 - [CHANGELOG](https://github.com/hashicorp/packer-plugin-hcloud/releases/tag/v1.0.4)
* hyperone@v1.0.1 - [CHANGELOG](https://github.com/hashicorp/packer-plugin-hyperone/releases/tag/v1.0.1)
* hyperv@v1.0.3 - [CHANGELOG](https://github.com/hashicorp/packer-plugin-hyperv/releases/tag/v1.0.3)
* inspec@v1.0.0 - [CHANGELOG](https://github.com/hashicorp/packer-plugin-inspec/releases/tag/v1.0.0)
* jdcloud@v1.0.1 - [CHANGELOG](https://github.com/hashicorp/packer-plugin-jdcloud/releases/tag/v1.0.1)
* linode@v1.0.3 - [CHANGELOG](https://github.com/hashicorp/packer-plugin-linode/releases/tag/v1.0.3)
* lxc@v1.0.1 - [CHANGELOG](https://github.com/hashicorp/packer-plugin-lxc/releases/tag/v1.0.1)
* lxd@v1.0.1 - [CHANGELOG](https://github.com/hashicorp/packer-plugin-lxd/releases/tag/v1.0.1)
* ncloud@v1.0.3 - [CHANGELOG](https://github.com/hashicorp/packer-plugin-ncloud/releases/tag/v1.0.3)
* oneandone@v1.0.1 - [CHANGELOG](https://github.com/hashicorp/packer-plugin-oneandone/releases/tag/v1.0.1)
* openstack@v1.0.1 - [CHANGELOG](https://github.com/hashicorp/packer-plugin-openstack/releases/tag/v1.0.1)
* oracle@v1.0.2 - [CHANGELOG](https://github.com/hashicorp/packer-plugin-oracle/releases/tag/v1.0.2)
* parallels@v1.0.2 - [CHANGELOG](https://github.com/hashicorp/packer-plugin-parallels/releases/tag/v1.0.2)
* profitbricks@v1.0.2 - [CHANGELOG](https://github.com/hashicorp/packer-plugin-profitbricks/releases/tag/v1.0.2)
* proxmox@v1.0.7 - [CHANGELOG](https://github.com/hashicorp/packer-plugin-proxmox/releases/tag/v1.0.7)
* puppet@v1.0.1 - [CHANGELOG](https://github.com/hashicorp/packer-plugin-puppet/releases/tag/v1.0.1)
* qemu@v1.0.4 - [CHANGELOG](https://github.com/hashicorp/packer-plugin-qemu/releases/tag/v1.0.4)
* salt@v1.0.0 - [CHANGELOG](https://github.com/hashicorp/packer-plugin-salt/releases/tag/v1.0.0)
* tencentcloud@v1.0.5 - [CHANGELOG](https://github.com/hashicorp/packer-plugin-tencentcloud/releases/tag/v1.0.5)
* triton@v1.0.1 - [CHANGELOG](https://github.com/hashicorp/packer-plugin-triton/releases/tag/v1.0.1)
* ucloud@v1.0.1 - [CHANGELOG](https://github.com/hashicorp/packer-plugin-ucloud/releases/tag/v1.0.1)
* vagrant@v1.0.2 - [CHANGELOG](https://github.com/hashicorp/packer-plugin-vagrant/releases/tag/v1.0.2)
* virtualbox@v1.0.3 - [CHANGELOG](https://github.com/hashicorp/packer-plugin-virtualbox/releases/tag/v1.0.3)
* vmware@v1.0.5 - [CHANGELOG](https://github.com/hashicorp/packer-plugin-vmware/releases/tag/v1.0.5)
* vsphere@v1.0.4 - [CHANGELOG](https://github.com/hashicorp/packer-plugin-vsphere/releases/tag/v1.0.4)
* yandex@v1.1.1 - [CHANGELOG](https://github.com/hashicorp/packer-plugin-yandex/releases/tag/v1.1.1)

### IMPROVEMENTS:
* Bump hcp-sdk-go to latest version.
    [GH-11763](https://github.com/hashicorp/packer/pull/11763)
* Plugins installed via the `packer plugins install` command will now
    automatically load for builds using HCL2 templates without a
    required_plugins block.
    [GH-11712](https://github.com/hashicorp/packer/pull/11712)
* Plugins installed via the `packer plugins install` command will now
    automatically load for builds using JSON templates.
    [GH-11712](https://github.com/hashicorp/packer/pull/11712)

### BUG FIXES:
* Bump github.com/hashicorp/go-getter to fix the untarring of OVA files
    regression. [GH-11675](https://github.com/hashicorp/packer/pull/11675)
* Bump github.com/hashicorp/packer-plugin-sdk to address legacy SSH key
    algorithms in SSH communicator.
    [GH-11774](https://github.com/hashicorp/packer/pull/11774)
    [GH-11754](https://github.com/hashicorp/packer/pull/11754)

## 1.8.0 (March 4, 2022)

### NOTES:
* **Breaking Change**: Support for the following  architectures `ppc64le, mips,
     mips64, mipsle, mipsle64, s390x` have been removed from the Packer releases
     page. Packer, along with the HashiCorp maintained Packer plugins, have been
     updated to release binaries for the HashiCorp supported architectures arm,
     arm64, 386, and amd64. A full list of supported platforms can be found
     on the [Packer Downloads](https://www.packer.io/downloads) page.
    [GH-11564](https://github.com/hashicorp/packer/pull/11564)
    [GH-11601](https://github.com/hashicorp/packer/pull/11601)
    [GH-11603](https://github.com/hashicorp/packer/pull/11603)

* **Breaking Change**: The `packer-plugin-check`(github.com/hashicorp/packer/cmd/packer-plugins-check) has been
     replaced by the [`packer-sdc plugin-check` command](https://github.com/hashicorp/packer- plugin sdk/tree/main/cmd/packer-sdc#packer-sdc).
     Plugin maintainers who may be using the packer-plugin-check as part of their
     release pipeline are encouraged to move to the packer-sdc command. As an
     alternative, maintainers can continue to use the packer-plugin-check by
     pinning the command to Packer 1.7.10.
     [GH-11317](https://github.com/hashicorp/packer/pull/11317)

### FEATURES
* **New Command** `packer plugins` command and subcommands to manage external
    plugins. [GH-11553](https://github.com/hashicorp/packer/pull/11553)
    [GH-11625](https://github.com/hashicorp/packer/pull/11625)

### IMPROVEMENTS
* core: Add a `env` argument to provisioner blocks that allow for setting a
    map of key/value pairs to inject prior to the execute_command. The env argument
    is an alternative to using environment_vars for setting environment variables,
    which has the added ability to read from Packer datasources. [GH-11569](https://github.com/hashicorp/packer/pull/11569)
* core: Bump version of go-getter to allow for downloading ISOs with PGP signed
    checksums. [GH-11495](https://github.com/hashicorp/packer/pull/11495)
* core: Docker images are now available for all supported architectures that
    the linux binaries are built for including arm, arm64, 386, and amd64
    [GH-11564](https://github.com/hashicorp/packer/pull/11564)
    [GH-11601](https://github.com/hashicorp/packer/pull/11601)
    [GH-11603](https://github.com/hashicorp/packer/pull/11603)
* core: Packer's Linux package service configs and pre/post install scripts are
    now available under .release/linux.
    [GH-11601](https://github.com/hashicorp/packer/pull/11601)
    [GH-11603](https://github.com/hashicorp/packer/pull/11603)
* core: Packer's Linux packages are now available for all supported Linux
    architectures including arm, arm64, 386, and amd64
    [GH-11564](https://github.com/hashicorp/packer/pull/11564)
    [GH-11601](https://github.com/hashicorp/packer/pull/11601)
    [GH-11603](https://github.com/hashicorp/packer/pull/11603)
* core: The dockerfile that is used to build the packer docker image available
    at hashicorp/packer now lives in the root of this repo. The README that
    describes how to build the docker targets defined in the Dockerfile are
    available under ./release/docker/README.md.
    [GH-11564](https://github.com/hashicorp/packer/pull/11564)
    [GH-11601](https://github.com/hashicorp/packer/pull/11601)
    [GH-11603](https://github.com/hashicorp/packer/pull/11603)
* core: The packer-plugin-check command has been removed. Plugin maintainers
    should update their release configuration to use the `packer-sdc plugin-
    check` command.  [GH-11317](https://github.com/hashicorp/packer/pull/11317)


### BUG FIXES
* core/hcl2: Fixes an issue preventing builds from pausing between provisioners when
    the `--debug` argument has been passed.
    [GH-11537](https://github.com/hashicorp/packer/pull/11537)
* core/hcl2: Fixes a data loss issue when merging an empty-object map to a non-
    empty map variable.
    [GH-11566](https://github.com/hashicorp/packer/pull/11566)
* core/hcl2: Fixes a regression where references to locals via the lookup
    function were failing to find defined keys.
    [GH-11566](https://github.com/hashicorp/packer/pull/11566)
* core/hcl2: Fixes an issue where HCP Packer build labels from the first
    completed build image were being copied to all images within the same
    build. [GH-11574](https://github.com/hashicorp/packer/pull/11574)
    [GH-11584](https://github.com/hashicorp/packer/pull/11584)
* core: HCP Packer datasources will no longer fail for iterations with
    scheduled revocations.
    [GH-11619](https://github.com/hashicorp/packer/pull/11619)
* core: Packer darwin builds now use macOS system DNS resolver for resolving
    hostnames. [GH-9710](https://github.com/hashicorp/packer/issues/9710)
    [GH-11564](https://github.com/hashicorp/packer/pull/11564)

## 1.7.10 (February 02, 2022)

### NOTES:

This patch release fixes a crash in m1 caused by the psutils dep,
    https://github.com/hashicorp/packer/pull/11546, originally reported by the
    HomeBrew team here https://github.com/hashicorp/packer/issues/11512

### IMPROVEMENTS
* Print all locals errors when there is a circular error.
    [GH-11527](https://github.com/hashicorp/packer/pull/11527

### BUG FIXES
* Prevent duplicate local block creation.
    [GH-11534](https://github.com/hashicorp/packer/pull/11534)
* Update psutils dep to avoid crash on Arm64 Macs.
    [GH-11546](https://github.com/hashicorp/packer/pull/11546)

## 1.7.9 (January 19, 2022)
### NOTES:

* HCP Packer compatible plugins will contain a "HCP Packer Ready" badge at the
    top of on their documentation page.
    [GH-11456](https://github.com/hashicorp/packer/pull/11456)
* Packer core [documentation](https://packer.io/docs) now supports versioning
    on Packer.io. Users can select documentation for the version of Packer they
    are using. [GH-11434](https://github.com/hashicorp/packer/pull/11434)
* The `hcp_packer_registry.labels` argument has been deprecated in favor of the
    new `bucket_labels` argument to support custom user generated bucket
    labels. [GH-11401](https://github.com/hashicorp/packer/pull/11401)
* The code of the Outscale builder was handed over to the outscale org
    [GH-11428](https://github.com/hashicorp/packer/pull/11428)
* Version information for external plugins has been added as a badge to the
    plugin documentation.
    [GH-11457](https://github.com/hashicorp/packer/pull/11457)
* The following unmaintained external provisioner plugins have been archived: chef,
    converge, inspec, salt-masterless, puppet. More
    information on what this deprecation means for users can be found on the
    blog post - [Plans to Archive Unmaintained Provisioner
    Plugins](https://www.hashicorp.com/blog/plans-to-archive-unmaintained-packer-provisioner-plugins)


### PLUGINS:

External plugins have been pinned to the following versions. Please see
    their respective changelogs for details on plugin specific bug fixes and
    improvements.

* alicloud@v1.0.1 - [CHANGELOG](https://github.com/hashicorp/packer-plugin-alicloud/releases/tag/v1.0.1)
* amazon@v1.0.6 - [CHANGELOG](https://github.com/hashicorp/packer-plugin-amazon/releases/tag/v1.0.6)
* ansible@v1.0.1 - [CHANGELOG](https://github.com/hashicorp/packer-plugin-ansible/releases/tag/v1.0.1)
* azure@v1.0.4 - [CHANGELOG](https://github.com/hashicorp/packer-plugin-azure/releases/tag/v1.0.4)
* chef@v1.0.2 - [CHANGELOG](https://github.com/hashicorp/packer-plugin-chef/releases/tag/v1.0.2)
* cloudstack@v1.0.0 - [CHANGELOG](https://github.com/hashicorp/packer-plugin-cloudstack/releases/tag/v1.0.0)
* converge@v1.0.1 - [CHANGELOG](https://github.com/hashicorp/packer-plugin-converge/releases/tag/v1.0.1)
* digitalocean@v1.0.1 - [CHANGELOG](https://github.com/hashicorp/packer-plugin-digitalocean/releases/tag/v1.0.1)
* docker@v1.0.3 - [CHANGELOG](https://github.com/hashicorp/packer-plugin-docker/releases/tag/v1.0.3)
* googlecompute@v1.0.9 - [CHANGELOG](https://github.com/hashicorp/packer-plugin-googlecompute/releases/tag/v1.0.9)
* hcloud@v1.0.2 - [CHANGELOG](https://github.com/hashicorp/packer-plugin-hcloud/releases/tag/v1.0.2)
* hyperone@v1.0.0 - [CHANGELOG](https://github.com/hashicorp/packer-plugin-hyperone/releases/tag/v1.0.0)
* hyperv@v1.0.1 - [CHANGELOG](https://github.com/hashicorp/packer-plugin-hyperv/releases/tag/v1.0.1)
* inspec@v1.0.0 - [CHANGELOG](https://github.com/hashicorp/packer-plugin-inspec/releases/tag/v1.0.0)
* jdcloud@v1.0.0 - [CHANGELOG](https://github.com/hashicorp/packer-plugin-jdcloud/releases/tag/v1.0.0)
* linode@v1.0.1 - [CHANGELOG](https://github.com/hashicorp/packer-plugin-linode/releases/tag/v1.0.1)
* lxc@v1.0.0 - [CHANGELOG](https://github.com/hashicorp/packer-plugin-lxc/releases/tag/v1.0.0)
* lxd@v1.0.0 - [CHANGELOG](https://github.com/hashicorp/packer-plugin-lxd/releases/tag/v1.0.0)
* ncloud@v1.0.1 - [CHANGELOG](https://github.com/hashicorp/packer-plugin-ncloud/releases/tag/v1.0.1)
* oneandone@v1.0.0 - [CHANGELOG](https://github.com/hashicorp/packer-plugin-oneandone/releases/tag/v1.0.0)
* openstack@v1.0.0 - [CHANGELOG](https://github.com/hashicorp/packer-plugin-openstack/releases/tag/v1.0.0)
* oracle@v1.0.1 - [CHANGELOG](https://github.com/hashicorp/packer-plugin-oracle/releases/tag/v1.0.1)
* outscale@v1.0.2 - [CHANGELOG](https://github.com/hashicorp/packer-plugin-outscale/releases/tag/v1.0.2)
* parallels@v1.0.0 - [CHANGELOG](https://github.com/hashicorp/packer-plugin-parallels/releases/tag/v1.0.0)
* profitbricks@v1.0.1 - [CHANGELOG](https://github.com/hashicorp/packer-plugin-profitbricks/releases/tag/v1.0.1)
* proxmox@v1.0.4 - [CHANGELOG](https://github.com/hashicorp/packer-plugin-proxmox/releases/tag/v1.0.4)
* puppet@v1.0.1 - [CHANGELOG](https://github.com/hashicorp/packer-plugin-puppet/releases/tag/v1.0.1)
* qemu@v1.0.1 - [CHANGELOG](https://github.com/hashicorp/packer-plugin-qemu/releases/tag/v1.0.1)
* salt@v1.0.0 - [CHANGELOG](https://github.com/hashicorp/packer-plugin-salt/releases/tag/v1.0.0)
* tencentcloud@v1.0.3 - [CHANGELOG](https://github.com/hashicorp/packer-plugin-tencentcloud/releases/tag/v1.0.3)
* triton@v1.0.0 - [CHANGELOG](https://github.com/hashicorp/packer-plugin-triton/releases/tag/v1.0.0)
* ucloud@v1.0.0 - [CHANGELOG](https://github.com/hashicorp/packer-plugin-ucloud/releases/tag/v1.0.0)
* vagrant@v1.0.1 - [CHANGELOG](https://github.com/hashicorp/packer-plugin-vagrant/releases/tag/v1.0.1)
* virtualbox@v1.0.1 - [CHANGELOG](https://github.com/hashicorp/packer-plugin-virtualbox/releases/tag/v1.0.1)
* vmware@v1.0.5 - [CHANGELOG](https://github.com/hashicorp/packer-plugin-vmware/releases/tag/v1.0.5)
* vsphere@v1.0.2 - [CHANGELOG](https://github.com/hashicorp/packer-plugin-vsphere/releases/tag/v1.0.2)
* yandex@v1.0.3 - [CHANGELOG](https://github.com/hashicorp/packer-plugin-yandex/releases/tag/v1.0.3)
* scaleway@v1.0.4 - [CHANGELOG](https://github.com/scaleway/packer-plugin-scaleway/releases/tag/v1.0.4)

### IMPROVEMENTS:

* core/hcl2: Add `bucket_labels` argument to the `hcp_packer_registry` block to
    support custom user generated bucket labels.
    [GH-11401](https://github.com/hashicorp/packer/pull/11401)
* core/hcl2: Add `build_labels` argument to the `hcp_packer_registry` block to
    support custom user generated build labels.
    [GH-11401](https://github.com/hashicorp/packer/pull/11401)
* core/hcl2: Allow for the use of variables and locals within a `build` block.
    [GH-11421](https://github.com/hashicorp/packer/pull/11421)
* core/hcl2: Allow for the use of variables and locals within a
    `hcp_packer_registry` block.
    [GH-11421](https://github.com/hashicorp/packer/pull/11421)
* core/website: Add HCP Packer Ready badge to supported plugins.
    [GH-11456](https://github.com/hashicorp/packer/pull/11456)
* core/website: Add Packer version selection toggle to Packer core
    documentation. [GH-11434](https://github.com/hashicorp/packer/pull/11434)
* core/website: Add version information to external plugin documentation.
    [GH-11456](https://github.com/hashicorp/packer/pull/11456)
* core/website: Extract external plugins documentation from `packer.io/docs/`
    into `packer.io/plugins`.
    [GH-11464](https://github.com/hashicorp/packer/pull/11464)
* core: Add Packer user agent information to HCP Packer client requests.
    [GH-11455](https://github.com/hashicorp/packer/pull/11455)
* core: Bump github.com/hashicorp/packer-plugin-sdk from 0.2.9 to 0.2.11 to
    prevent HCP Packer builds from failing when no SourceImageID is
    provided. [GH-11459](https://github.com/hashicorp/packer/pull/11459)
* core: Bump to latest preview version of hashicorp/hcp-sdk-go to prevent HCP
    Packer builds from trying to update a revoked iteration.
    [GH-11492](https://github.com/hashicorp/packer/pull/11492)
* provisioner/powwershell: Tiny tweaks and fixes for the PowerShell
    provisioner. [GH-11410](https://github.com/hashicorp/packer/pull/11410)

### BUG FIXES:

* core/hcl2: Allow the use of `build.name` for naming provisioners and
    post-processors. [GH-11432](https://github.com/hashicorp/packer/pull/11432)
* core/hcl2: Fix crash when a provisioner `timeout` argument is improperly
    formatted. [GH-11382](https://github.com/hashicorp/packer/pull/11382)
* core/hcl2: HCP Packer builds containing metadata not expected by Packer core
    will no longer fail the build.
    [GH-11458](https://github.com/hashicorp/packer/pull/11458)
* provisioner/file: File provisioner will now perform a noop when no source
    file content is specified; previously missing content resulted in a hard
    fail. [GH-11349](https://github.com/hashicorp/packer/pull/11349)


## 1.7.8 (October 27, 2021)

### BUG FIXES
* builder/amazon: Bump plugin to latest version to address a variable
    interpolation issue for builder `run_tags`. [GH-11360](https://github.com/hashicorp/packer/pull/11360)

## 1.7.7 (October 19, 2021)

### NOTES:

* The code of the Scaleway builder was handed over to the scaleway org
    [GH-11298](https://github.com/hashicorp/packer/pull/11298)
    [GH-11296](https://github.com/hashicorp/packer/pull/11296)

### IMPROVEMENTS:
* Improved support and user experience for HCP Packer registry.
    [GH-11304](https://github.com/hashicorp/packer/pull/11304)
    [GH-11315](https://github.com/hashicorp/packer/pull/11315)
    [GH-11320](https://github.com/hashicorp/packer/pull/11320)
    [GH-11319](https://github.com/hashicorp/packer/pull/11319)

* core: Allow to use build variables in a post processor.
    [GH-11323](https://github.com/hashicorp/packer/pull/11323)
* core: Allow use in top level source variables in a build.source block.
    [GH-11318](https://github.com/hashicorp/packer/pull/11318)
* core: Show successful message upon successful packer validate.
    [GH-11337](https://github.com/hashicorp/packer/pull/11337)
* packer init: better error handling.
    [GH-11330](https://github.com/hashicorp/packer/pull/11330)

## 1.7.6 (September 28, 2021)

### BUG FIXES:
* core: Update Go module dependencies to fix an issue preventing Go 1.16 users
    from installing the packer-plugins-check command via go install. [GH-11282]

## 1.7.5 (September 14, 2021)

### NOTES:
The Exoscale builder and post-processor are no longer vendored with Packer
    core, users of the Exoscale plugin should use `packer init` to install the
    latest version of the plugin. See the [Exoscale Plugin
    Documentation](https://github.com/exoscale/packer-plugin-exoscale#exoscale-packer-plugin) for more information. [GH-11237]

### FEATURES:
* **Future Scaffolding** This release contains no-op refactors in preparation
    for connecting Packer to the HCP Packer Registry.

### IMPROVEMENTS:
* core: Upgrade to Go 1.17. [GH-11237]
* hcl2_upgrade: support strftime function. [GH-11220]
* provisioner/file: add option to set content + tests. [GH-11209]
* provisioner/inspec: Remove inspec provisioner from Packer core. [GH-11230]
* provisioner/salt-masterless: Remove salt-masterless provisioner from Packer
    core to github.com/hashicorp/packer-plugin-salt/provisioner/salt-
    masterless. [GH-11229]

### BUG FIXES:
* builder/azure: Bump plugin to latest version to address a vulnerable
    dependency in azure-sdk-for-go. [GH-11162]
* builder/googlecompute: Bump plugin to latest version to fix an issue when
    building an instance via an IAP tunnel. [GH-11235]
* builder/ncloud: Bump ncloud dependency to fix dependency that got deleted
    from github [GH-11224]
* core: Fix `{{packer_version}}` interpolation regression for HCL and JSON
    templates. [GH-11200]
* hcl2_upgrade: Fix panic when file does not exist. [GH-11206]
* hcl2_upgrade: special case: vsphere fix. [GH-11216]

## 1.7.4 (July 20, 2021)

### BUG FIXES:

* builder/outscale: Update Outscale multi-component plugin to fix `go build`
    failures due to missing Go module dependencies. [GH-11147]

## 1.7.3 (June 15, 2021)

### IMPROVEMENTS:

We've extracted a majority of HashiCorp-maintained and community plugins from the Packer Core repository. They now live in their own multi-component plugin repositories. This is not a breaking change as we are enabling backwards compatibility in this release by vendoring components back into Packer.
However, we encourage users to begin using `packer init` to download and install plugins to get the latest updates to each plugin, and to prepare for Packer v2.0 when we will stop vendoring the above plugins into the main Packer binary.

The following repositories have been created, and their components have been deleted from the "github.com/hashicorp/packer" repository.

* "github.com/hashicorp/packer-plugin-alicloud" [GH-10932]
* "github.com/hashicorp/packer-plugin-amazon" [GH-10800]
* "github.com/hashicorp/packer-plugin-ansible" [GH-10912]
* "github.com/hashicorp/packer-plugin-azure" [GH-10979]
* "github.com/hashicorp/packer-plugin-chef" [GH-10921]
* "github.com/hashicorp/packer-plugin-cloudstack" [GH-10934]
* "github.com/hashicorp/packer-plugin-converge" [GH-10956]
* "github.com/hashicorp/packer-plugin-digitalocean" [GH-10961]
* "github.com/hashicorp/packer-plugin-docker" [GH-10695]
* "github.com/hashicorp/packer-plugin-googlecompute" [GH-10890]
* "github.com/hashicorp/packer-plugin-hcloud" [GH-10966]
* "github.com/hashicorp/packer-plugin-hyperone" [GH-10949]
* "github.com/hashicorp/packer-plugin-hyperv" [GH-10949]
* "github.com/hashicorp/packer-plugin-inspec"
* "github.com/hashicorp/packer-plugin-ionos-cloud"
* "github.com/hashicorp/packer-plugin-jdcloud" [GH-10946]
* "github.com/hashicorp/packer-plugin-linode" [GH-10947]
* "github.com/hashicorp/packer-plugin-lxc" [GH-10965]
* "github.com/hashicorp/packer-plugin-lxd" [GH-10965]
* "github.com/hashicorp/packer-plugin-ncloud" [GH-10937]
* "github.com/hashicorp/packer-plugin-openstack" [GH-10933]
* "github.com/hashicorp/packer-plugin-oracle" [GH-10962]
* "github.com/hashicorp/packer-plugin-outscale" [GH-10941]
* "github.com/hashicorp/packer-plugin-parallels" [GH-10936]
* "github.com/hashicorp/packer-plugin-profitbricks" [GH-11084]
* "github.com/hashicorp/packer-plugin-proxmox" [GH-10930]
* "github.com/hashicorp/packer-plugin-puppet" [GH-10943]
* "github.com/hashicorp/packer-plugin-qemu" [GH-10929]
* "github.com/hashicorp/packer-plugin-salt"
* "github.com/hashicorp/packer-plugin-scaleway" [GH-10939]
* "github.com/hashicorp/packer-plugin-tencentcloud" [GH-10967]
* "github.com/hashicorp/packer-plugin-triton" [GH-10963]
* "github.com/hashicorp/packer-plugin-ucloud" [GH-10953]
* "github.com/hashicorp/packer-plugin-vagrant" [GH-10960]
* "github.com/hashicorp/packer-plugin-virtualbox" [GH-10910]
* "github.com/hashicorp/packer-plugin-vmware" [GH-10920]
* "github.com/hashicorp/packer-plugin-vsphere" [GH-10896]
* "github.com/hashicorp/packer-plugin-yandex" [GH-10970]

The following components will not be removed from the main packer binary:

* `null` builder
* `file` builder
* `breakpoint` provisioner

* `file` provisioner
* `powershell` provisioner
* `shell` provisioner
* `shell-local` provisioner
* `sleep` provisioner
* `windows-restart` provisioner
* `windows-shell` provisioner

* `artifice` post-processor
* `checksum` post-processor
* `compress` post-processor
* `manifest` post-processor
* `shell-local` post-processor

### Bug Fixes:
* builder/azure: Add `keep_os_disk` parameter to control OS disk deletion
    [GH-10045]
* builder/azure: Stop SIG timout from being overridden by PollingDuration
    [GH-10816]
* builder/azure: Support shared image gallery storage account type [GH-10863]
* builder/proxmox: Proxmox builder use ipv4 address instead of always ipv6.
    [GH-10858]
* core/hcl2_upgrade: Allow hcl2_upgrade continue with unknown builders.
    [GH-11049]
* core/hcl2_upgrade: Improve regex to fix escaping on split function.
    [GH-11083]
* core/hcl: Fix Invalid provisioner pause_before panic [GH-10978]
* core: HCL "index" function now actually returns the index of the element
    [GH-11008]
* core: Implemented DEFAULT_NAME handling for datasource plugins [GH-11026]

### Enhancements:

* builder/azure:  Added custom nicname and osdiskname [GH-10938]
* builder/azure: Add support for shared image gallery storage account type
    [GH-10863]
* builder/digitalocean: support ecdsa, ed25519, dsa temporary key types.
    [GH-10856]
* builder/ncloud: Support ncloud vpc version [GH-10870]
* core/fmt: When reading from stdin `packer fmt` will output the contents of
    the formatted file even if the input was already formatted. [GH-11047]
* core/hcl: HCL variables are now supported within the `name`, `only`,
    `except`, and `keep_input_artifact` fields for post-processor blocks.
    [GH-11094]
* core/hcl: Running `packer build` with an `-only` or `-exclude` flag will now
    inform the user if no match was found. [GH-11050]
* post-processor/compress: Add bzip2 support to post-processor [GH-10867]
* post-processor/googlecompute-import: Add Image Storage Locations field
    [GH-10864]
* Removed the golang "vendor" directory in favor of go modules. This should not
    affect end users. [GH-10916]

## 1.7.2 (April 05, 2021)

### IMPROVEMENTS:

* builder/alicloud: Add `ramrole` configuration to ECS instance. [GH-10845]

### BUG FIXES:

* builder/proxmox: Update Proxmox Go API to ensure only the first non-loopback
    IPv4 address gets returned. [GH-10858]
* builder/vsphere: Fix primary disk resize on clone. [GH-10848]
* core: Fix bug where call to "packer version" sent output to stderr instead of
    stdout. [GH-10850]

## 1.7.1 (March 31, 2021)

### NOTES:

* builder/amazon: Has been vendored in this release and will no longer be
    updated with Packer core. In Packer v1.8.0 the plugin will be removed
    entirely. The `amazon` components will continue to work as expected until
    then, but for the latest offerings of the Amazon plugin, users are
    encourage to use the `packer init` command to install the latest release
    version. For more details see [Installing Packer
    Plugins](https://www.packer.io/docs/plugins#installing-plugins)
* builder/docker: Has been vendored in this release and will no longer be
    updated with Packer core. In Packer v1.8.0 the plugin will be removed
    entirely. The `docker` builder will continue to work as expected until
    then, but for the latest offerings of the Docker plugin, users are
    encourage to use the `packer init` command to install the latest release
    version. For more details see [Installing Packer
    Plugins](https://www.packer.io/docs/plugins#installing-plugins)
* darwin/arm64: Packer now includes the darwin/arm64 binary to its releases to
    supports the new OSX M1. [GH-10804]
* post-processor/docker-\*: Have been vendored in this release and will no
    longer be updated with Packer core. In Packer v1.8.0 the plugin will be
    removed entirely. The `docker` builder will continue to work as expected
    until then, but for the latest offerings of the Docker plugin, users are
    encourage to use the `packer init` command to install the latest release
    version. For more details see [Installing Packer
    Plugins](https://www.packer.io/docs/plugins#installing-plugins)
* post-processor/exoscale-import: Has been vendored in this release and will no
    longer be updated with Packer core. In Packer v1.8.0 the plugin will be
    removed entirely. The `exoscale-import` post-processor will continue to
    work as expected until then, but for the latest offerings of the Exoscale
    plugin, users are encourage to use the `packer init` command to install the
    latest release version. For more details see [Exoscale Plugin
    Repostiroy](https://github.com/exoscale/packer-plugin-exoscale). [GH-10709]

### IMPROVEMENTS
* builder/amazon: allow creation of ebs snapshots without volumes. [GH-9591]
* builder/amazon: Fix issue for multi-region AMI build that fail when
    encrypting with KMS and sharing across accounts. [GH-10754]
* builder/azure: Add client_cert_token_timeout option. [GH-10528]
* builder/google: Make Windows password timeout configurable. [GH-10727]
* builder/google: Update public GCP image project as gce-uefi-images are
    deprecated. [GH-10724]
* builder/oracle-oci: Update Oracle Go SDK to add support for OCI flexible
    shapes.  [GH-10833]
* builder/proxmox: Allow using API tokens for Proxmox authentication.
    [GH-10797]
* builder/qemu: Added firmware option. [GH-10683]
* builder/scaleway: add support for timeout in shutdown step. [GH-10503]
* builder/vagrant: Fix logging to be clearer when Vagrant builder overrides
    values retrieved from vagrant's ssh_config call. [GH-10743]
* builder/virtualbox:  Added ISO builder option to create additional disks.
    [GH-10674]
* builder/virtualbox: Add options for nested virtualisation and RTC time base.
    [GH-10736]
* builder/virtualbox: Add template options for chipset, firmware, nic, graphics
    controller, and audio controller. [GH-10671]
* builder/virtualbox: Support for "virtio" storage and ISO drive. [GH-10632]
* builder/vmware: Added "attach_snapshot" parameter to vmware vmx builder.
    [GH-10651]
* command/fmt: Adding recursive flag to formatter to format subdirectories.
    [GH-10457]
* core/hcl2: Add legacy_isotime function. [GH-10780]
* core/hcl2: Add support for generating `dynamic` blocks within a `build`
    block. [GH-10825]
* core/hcl2: Add templatefile function. [GH-10776]
* core/hcl2_upgrade: hcl2_upgrade command can now upgrade json var-files.
    [GH-10676]
* core/init: Add implicit required_plugin blocks feature. [GH-10732]
* core: Add http_content option to serve variables from HTTP at preseed.
    [GH-10801]
* core: Change template parsing error to include warning about file extensions.
    [GH-10652]
* core: Update to gopsutil v3.21.1 to allow builds to work for darwin arm64.
    [GH-10697]
* provisioner/inspec: Allow non-zero exit codes for inspec provisioner.
    [GH-10723]

### BUG FIXES
* buider/azure: Update builder to ensure a proper clean up Azure temporary
    managed Os disks. [GH-10713]
* builder/amazon: Update amazon SDK to fix an SSO login issue. [GH-10668]
* builder/azure: Don't overwrite subscription id if unset. [GH-10659]
* builder/azure: Set default for the parameter client_cert_token_timeout
    [GH-10783]
* builder/google: Add new configuration field `windows_password_timeout` to
    allow user to set configurable timeouts. [GH-10727]
* builder/hyperv: Make Packer respect winrm_host flag in winrm connect func.
    [GH-10748]
* builder/openstack: Make Packer respect winrm_host flag in winrm connect func.
    [GH-10748]
* builder/oracle-oci: Update Oracle Go SDK to fix issue with reading key file.
    [GH-10560] [GH-10774]
* builder/outscale: Fix omi_description that was ignored in Osc builder
    [GH-10792]
* builder/parallels: Make Packer respect winrm_host flag in winrm connect func.
    [GH-10748]
* builder/proxmox: Fixes issue when using `additional_iso_files` in HCL enabled
    templates. [GH-10772]
* builder/qemu: Make Packer respect winrm_host flag in winrm connect func.
    [GH-10748]
* builder/virtualbox: Make Packer respect winrm_host flag in winrm connect
    func. [GH-10748]
* builder/vmware: Added a fallback file check when trying to determine the
    network-mapping configuration. [GH-10543]
* builder/vsphere: Fix invalid device configuration issue when creating a
    vm with multiple disk on the same controller. [GH-10844]
* builder/vsphere: Fix issue where boot command would fail the build do to a
    key typing error. This change will now retry to type the key on error
    before giving up. [GH-10541]
* core/hcl2_upgrade: Check for nil config map when provisioner/post-processor
    doesn't have config. [GH-10730]
* core/hcl2_upgrade: Fix escaped quotes in template functions [GH-10794]
* core/hcl2_upgrade: Make hcl2_upgrade command correctly translate
    pause_before. [GH-10654]
* core/hcl2_upgrade: Make json variables using template engines get stored as
    locals so they can be properly interpolated. [GH-10685]
* core/init: Fixes issue where `packer init` was failing to install valid
    plugins containing a 'v' within its name. [GH-10760]
* core: Packer will now show a proper error message when failing to load the
    contents of PACKER_CONFIG. [GH-10766]
* core: Pin Packer to Golang 1.16 to fix code generation issues. [GH-10702]
* core: Templates previously could not interpolate the environment variable
    PACKER_LOG_PATH. [GH-10660]
* post-processor/vagrant-cloud: Override direct upload based on box size
    [GH-10820]
* provisioner/chef-solo: HCL2 templates can support the json_string option.
    [GH-10655]
* provisioner/inspec: Add new configuration field `valid_exit_codes` to allow
    for non-zero exit codes. [GH-10723]
* provisioner/salt-masterless: Update urls for the bootstrap scripts used by
    salt-masterless provide. [GH-10755]

## 1.7.0 (February 17, 2021)

### FEATURES
* **New Command** (HCL only) `packer init` command will download plugins defined
    in a new `required_plugins` block [GH-10304] [GH-10633].
* **New Plugin Type** Data sources can be implemented (blog post forthcoming).
    [GH-10440]
* **New Plugin** Aws Secrets Manager data source [GH-10505] [GH-10467]

### BACKWARDS INCOMPATIBILITIES
* core: The API that the Packer core uses to communicate with community plugins
    has changed; maintainers of community plugins will need to upgrade their
    plugins in order to make them compatible with v1.7.0. An upgrade guide will
    be available on our guides page https://www.packer.io/guides.

### IMPROVEMENTS
* builder/amazon: Add `skip_create_ami` option for testing and situations where
    artifact is not the ami. [GH-10531]
* builder/amazon: Add IMDSv2 support for AWS EBS builder. [GH-10546]
* builder/amazon: Add resource tags in the launch template used to request spot
    instances. [GH-10456]
* builder/openstack:  Add `skip_create_image` option for testing and situations
    where artifact is not the image. [GH-10496]
* builder/oracle-oci: Add retry strategies to oci calls [GH-10591]
* core/fmt: The `packer fmt` can now read from stdin. [GH-10500]
* core/hcl: Add regex and regexall hcl2 template functions. [GH-10601]
* core/hcl: Templates now support "sensitive" locals. [GH-10509]
* core/hcl: Templates now support error-cleanup-provisioner. [GH-10604]
* hcl2_upgrade: Command now comes with a flag so you can control whether output
    templates are annotated with helpful comments. [GH-10619]
* hcl2_upgrade: Command now gracefully handles options with template engine
    interpolations. [GH-10625]
* hcl2_upgrade: Command will convert amazon filters to use the ami data source.
    [GH-10491]

### BUG FIXES
* amazon/ebssurrogate: Apply snapshot tags at same time as when taking
    snapshot. [GH-10150]
* builder/amazon: Fix bug where validation fails if optional iops value is
    unset. [GH-10518]
* builder/amazon: Wrap API call to get filtered image in a retry. [GH-10610]
* builder/bsusurrogate: override bsu when omi root device is set. [GH-10490]
* builder/google: Fix bug where Packer would fail when run by users who do not
    have permission to access the metadata, even though the metadata is not
    necessary to the run. [GH-10458]
* builder/profitbricks: Profitbricks builder could not connect using SSH
    communicator. [GH-10549]
* builder/proxmox: Ensure ISOs in additional_iso_files are mounted during VM
    creation. [GH-10586]
* builder/proxmox: Improve cloud init error logging for proxmox builder.
    [GH-10499]
* builder/qemu: Fix bug where vnc_min_port set to value greater then 5900 could
    prevent Packer from connecting to QEMU. [GH-10450] [GH-10451]
* builder/qemu: Fix regression with cd indexing when disk_interface is `ide`.
    [GH-10519]
* builder/vmware-esx: Skip credential validation, which requires ovftool to be
    installed, if we are not exporting an image. [GH-10520]
* builder/yandex: Fix cloud-init config for ubuntu 20.04. [GH-10522]
* builder/yandex: Fix incorrect access to `instance_id`. [GH-10522]
* core/hcl: Fix bug where []uint8 types could not be passed to plugins.
* core/hcl: fix bug where HCL core could not handle passing []uint8 to plugins.
    [GH-10516]
* core/hcl: Fix force flag for hcl2 provisioners and post-processors.
    [GH-10571]
* post-processor/vsphere: Fix regression where Packer would not check the exit
    status after streaming UI from the ovftool command. [GH-10468]
* post-processor/yandex-export: Changed dhclient command and supported
    configuring disk for exportupdate-dump-method. Also added support for
    `file` builder. [GH-10488]

## 1.6.6 (December 16, 2020)

### FEATURES
* **New command** `fmt` allows users to format existing HCL2 configuration
    files into a canonical style. Please see [fmt command
    docs](https://packer.io/docs/commands/fmt) for more details. [GH-10225]
    [GH-10377]
* **New function** `env` allows users to set the default value of a variable to
    the value of an environment variable. Please see [env function
    docs](https://www.packer.io/docs/templates/hcl_templates/functions/contextual/env) for
    more details. [GH-10240]
* **Future Scaffolding** This release contains a large number of no-op
    refactoring changes. The Packer team at HashiCorp is preparing to split the
    plugins and core to make it easier for our third party maintainers and
    community members to release and maintain plugins, just like HashiCorp did
    with the Terraform Core-Provider split. The Packer team is committed to
    making sure that this split is seamless for our users and for our community
    maintainers -- if you are a community maintainer, you may want to follow
    along with some of the work by looking at the
    [core-plugin-split github tag.](https://github.com/hashicorp/packer/pulls?q=is%3Apr+label%3Acore-plugin-split)
    No one needs to do anything, yet, but we felt it was worth calling out all
    the work that isn't making it into the changelog. We will be following up
    with lots of documentation and communication in early 2021 with more
    information.

### IMPROVEMENTS
* builder/amazon-ebs: Add tags to launch templates. [GH-10203]
* builder/amazon: Add support for Amazon EBS gp3 volumes. [Gh-10338]
* builder/amazon: Increase default max_retries to lessen throttling issues.
    [GH-10290]
* builder/amazon: Support AWS gp3 volumes [GH-10338]
* builder/amazon: Support root volume encryption for amazon-chroot. [GH-10243]
* builder/amazon: Validate IOPS ratio. [GH-10199]
* builder/azure-arm: Add Azure CLI authentication support to builder.
    [GH-10157]
* builder/azure-arm: Create keyvaults with SoftDelete enabled. [GH-10210]
* builder/digitalocean: New option to provision with private ip. [GH-10093]
* builder/google: Add `wait_to_add_ssh_keys` option to delay the addition of
    SSH configuration that may be disrupted during an instance boot sequence.
    [GH-10320]
* builder/google: Add support for creating shielded VMs. [GH-10172]
* builder/googlecompute-export: Add logging.write to service account scopes.
    [GH-10316]
* builder/oracle-oci: Support image launch mode. [GH-10212]
* builder/outscale: Add outscale.hk endpoint support [GH-10207]
* builder/outscale: Add x509 certificate support. [GH-10161]
* builder/proxmox: New config option for boot-order. [GH-10260]
* builder/scaleway: Use the SDK functions to load profile from file and env.
    [GH-10181]
* builder/virtualbox: Allow attaching guest additions with "none" communicator.
    [GH-10306]
* builder/vmware: Make compatible with MacOS BigSur by using Apple DHCP leases
    instead of VMWare leases [GH-10384]
* builder/vsphere: New option to add additional storage to a cloned vm.
    [GH-10287]
* builder/yandex: More resilient image mounting and initialization. [GH-10335]
* builder/yandex: Update user-data to not use cloud-config fields to prevent
    possible user data collisions. [GH-10385]
* core/hcl: Update to `hcl2_upgrade` command to support complex variable values
    and packer version blocks. [GH-10221]
* hcl2upgrade: Update command to fix `env` call upgrade. [GH-10244]
* post-processor/vagrant-cloud: Add support for uploading directly to storage
    on Vagrant Cloud. [GH-10193]
* post-processor/yandex-export: Add retries and wait after disk attach
    operation. [GH-10303]
* post-processor/yandex-export: Show progress on export. [GH-10368]
* post-processor/yandex-export: Use ssh communicator in export. [GH-10352]
* post-processor/yandex-export: Verify the access to a specific bucket.
    [GH-10188]
* provisioner/salt-masterless: Call winrepo.update_git_repos and
    pkg.refresh_db. [GH-10201]

### BUG FIXES
* builder/amazon: Fix retry logic in AWS spot instance tagging. [GH-10394]
* builder/amazon: Fix single `tag` interpolation to allow for templating engine
    usage. [GH-10224]
* builder/google: Fix crash when using the `-on-error` build flag. [GH-10247]
* builder/google: Fix issue with service account detection when running Packer
    on a compute instance with `use_os_login` enabled. [GH-10360]
* builder/qemu: Fix duplication of main disk when setting "disk_image: true".
    [GH-10337]
* builder/qemu: Fix nil pointer dereference when loading values from state.
    [GH-10249]
* builder/qemu: Fix panic when disk_image=true and source image has no file
    extension. [GH-10226]
* builder/vagrant: Return error if ssh-config command fails. [GH-10213]
* builder/vsphere: WaitForIP should not return an error if an IP is not found
    [GH-10321]
* builder/yandex: Change disk creation method to manual. [GH-10250]
* builder/yandex: Fix issue with UserAgent string. [GH-10361]
* builder/yandex: Fixed using cloud config when using IPv6. [GH-10297]
* core/hcl: Ensure the `reverse` function does not break when given a value of
    type list. [GH-10380]
* post-processor/yandex-export: Check service account id. [GH-10305]

## 1.6.5 (October 30, 2020)

### FEATURES:
* New Builder(s): Proxmox builder has been split into two new builders
    `proxmox-iso` and `promox-clone`. See [Proxmox
    Builder](https://packer.io/docs/builders/proxmox) for more information on
    the builder. For users of the previous `proxmox` builder please use `packer
    fix` to migrate your templates to the new `promox-iso` builder. [GH-9262]

### BUG FIXES:

* builder/amazon: SSM connection now recovers from reboots. [GH-10003]
* builder/azure-arm: Fix build failures due to the deletion of additional
    managed disks defined in "disk_additional_size". [GH-10163]
* builder/azure-chroot: Fix typo in option `exlude_from_latest` to
    `exclude_from_latest`. Old name will still be respected. [GH-10034]
* builder/googlecompute: Fix HCL image_encryption_key fields and use the same
    casing in JSON and HCL2 [GH-10173]
* builder/openstack: Fix source image validation regression when using filters.
    [GH-10065]
* builder/proxmox: Fix unhandled buildvar type for HCL2 enabled build
    templates. [GH-10154]
* builder/qemu: Fix a regression where Packer would not wait properly in
    step_shutdown when a null communicator was used. [GH-10178]
* builder/qemu: Fix crash in step_run of qemu when loading commhostport form
    the statebag in a situation where the communicator is none. [GH-10145]
* builder/vsphere-clone: Packer was not respecting the "destroy" flag set in
    the content library config, and always keeping the source vm. This has been
    fixed. [GH-10165]
* builder/vsphere: Ensure builds are able to continue when no communicator has
    been specified `"communicator": "none"`. [GH-9964]
* builder/vsphere: Fix CD uploads so that Packer does not try to delete a CD
    that was not successfully uploaded. [GH-10155]
* core/hcl: Hide sensitive variables from output. [GH-10031]
* core/hcl: Packer HCL's "Coalesce" function now behaves same way as
    Terraform's. [GH-10016]
* core: Fix artifact handling so that input artifacts are properly preserved in
    postprocessors that don't modify artifacts. [GH-9996]
* core: Fix pathing in cd_files to copy proper directory tree when user
    provided absolute paths. [GH-10022]
* provisioner/ansible: Ansible galaxy no longer forces use of collections in v1
    files. [GH-10010]

### IMPROVEMENTS:

* builder/amazon-ebssurrogate: Apply snapshot tags at snapshot creation time.
    [GH-10150]
* builder/amazon: Add `io2` as a supported volume type. [GH-10102]
* builder/amazon: Add support for source instance tenancy [GH-10085]
* builder/google: Add service account impersonation. [GH-9968] [GH-10054]
* builder/googlecompute: Add `skip_create_image` option. [GH-10115]
* builder/googlecompute: Allow users to select the algorithm to use when
    generating temporary SSH keypair [GH-10111]
* builder/linode: Add `state_timeout` attribute to Linode builder. [GH-10128]
* builder/oracle-oci:  New option to specify image compartment separate from
    build compartment. [GH-10040]
* builder/oracle-oci: New option to specify boot volume size. [GH-10017]
* builder/oracle: Add `base_image_filter` option as alternative to
    `base_image_ocid` [GH-10116]
* builder/outscale: Migrate to new Outscale SDK. [GH-10056]
* builder/proxmox: split Proxmox into proxmox-iso and proxmox-clone. [GH-9626]
    [GH-10166]
* builder/scaleway: Allow the user to use an image label (eg ubuntu_focal)
    instead of a hardcoded UUID on the Scaleway builder. [GH-10061]
* builder/vsphere: Skip iso download if hashed file is already present on
    remote datastore. [GH-10143]
* builder/yandex: Add support for IAM credentials in the token field and
    YC_TOKEN environment variable. [GH-10158]
* core/hcl: Add ability to set version restrictions [GH-10149]
* core/hcl: Add build.name variable so users can access build name in addition
    to source name. [GH-10114]
* core/hcl: Add consul_key function to HCL templates. [GH-10119]
* core/hcl: Add HCL2 aws_secretsmanager function [GH-10124]
* core/hcl: Add packer.version variable to hcl configs so users can access the
    Packer release version. [GH-10117]
* core: Let user provide type of generated ssh key instead of always doing ssh-
    rsa [GH-10101]

## 1.6.4 (September 30, 2020)

### BUG FIXES:
* builder/amazon: Fix authentication issue when using instance profiles or
    assumed roles for loading session-derived credentials. [GH-10007]
* builder/azure: Fix crash when using `azure_tag` or `azure_tags` configuration
    options. [GH-10014]
* builder/qemu: Ensure `qemu_img_args` are honored during the disk convert
    step. [GH-10001]

 ## 1.6.3 (September 25, 2020)

### IMPROVEMENTS:
* builder/amazon: Add `pause_before_ssm` option to pause for some time before
    establishing a Session Manager session; defaults to 10s. [GH-9988]
* builder/amazon: Implement assume_role option that matches Terraform behavior.
    [GH-9981]
* builder/azure: Support publishing to a Shared Image Gallery with a different
    subscription id [GH-9875]
* builder/openstack: Add `external_source_image_url` and
    `external_source_image_format` to support building images from external
    source URLs. [GH-9992]
* builder/openstack: Include API requests and responses as part of the debug
    log output. [GH-9972]
* builder/oracle-oci: Add `create_vnic_details` option for launch details.
    [GH-9856]
* builder/oracle-oci: Allow freeform and defined tags to be added to an instance.
    [GH-9802]
* builder/proxmox: Add `io_thread` option for supporting io threads when using
    a `virtio-scsi-single` controller with a `scsi` or `virtio` disk type.
    [GH-9969]
* builder/proxmox: Add ability to specify interfaces for http_directory and VM.
    [GH-9874]
* builder/proxmox: Allow the mounting of multiple ISOs via the `cd_drive`
    option. [GH-9653]
* builder/proxmox: Fix boot command special keys. [GH-9885]
* builder/qemu: Add `qemu_img_args` option to set special cli flags for calls
    to qemu-img [GH-9956]
* builder/qemu: Add `skip_resize_disk` option to skip the resizing of QCOW2
    images. [GH-9896] [GH-9860]
* builder/qemu: Skip qemu-img convert on MacOS to prevent the creation of
    corrupt images [QEMU
    #1776920](https://bugs.launchpad.net/qemu/+bug/1776920) [GH-9949]
* builder/scaleway: Change default boottype to local. [GH-9853]
* builder/scaleway: Update scaleway to use non-deprecated sdk. [GH-9902]
* builder/vmware: Add `vnc_over_websocket` to allow the sending of a
    `boot_command` to hosts running ESXi 6.7 and above. [GH-9938]
* builder/vmware: Allow user to set vmware tools source path. [GH-9983]
* builder/vsphere-clone: Add ability to set `mac_address` [GH-9930]
* builder/vsphere-clone: Add floppy_files, cd_files, and iso_paths options.
    [GH-9963]
* builder/vsphere-iso: Add NVMe controller support. [GH-9880]
* builder/vsphere: Look for a default resource pool when root resource pool is
    not found. [GH-9809]
* core: Add support for running cygwin/msys2 based cd/iso creation tool
    [GH-9954]
* core: New `cd_files` option to mount iso for modern OSes which don't support
    floppies. [GH-9796] [GH-9919] [GH-9928] [GH-9932] [GH-9941]
* HCL2: When the type of a variable is not known evaluate setting as a literal
    string instead of a variable name. [GH-9863]
* post-processor/vagrant: Support the use of template variables within
    Vagrantfile templates. [GH-9923]
* post-processor/yandex-import: Allow custom API endpoint. [GH-9850]
* provisioner/ansible: Add support for Ansible Galaxy Collections. [GH-9903]

### BUG FIXES:
* builder/amazon-ebs: Fix issue where retrying on invalid IAM instance profile
    error was creating multiple spot instances. [GH-9946]
* builder/amazon-ebssurrogate: Fix issue where builder defaults to AWS managed
    key even when custom `kms_key_id` is set. [GH-9959]
* builder/amazon: Update ssm_driver log polling logic to prevent infinite loops
    when SSM driver is terminated outside of Packer. [GH-9991]
* builder/azure: Fix crash when using HCL2 configs. [GH-9984] [GH-9985]
* builder/qemu: Fix hardcoded lowerbound causing negative ports [GH-9905]
* builder/qemu: Skip compaction when backing file is used. [GH-9918]
* builder/scaleway: Add pre validate step to prevent the creation of multiple
    images with the same name. [GH-9840]
* builder/vmware-iso: Prevent the use of reserved SCSI ID 0:7 when attaching
    multiple disks. [GH-9940]
* builder/vsphere: Fix overly strict iso_path validation regex. [GH-9855]
* command/console: Prevent failure when there are unknown vars. [GH-9864]
* command/inspect: Allow unset variables in HCL2 and JSON. [GH-9832]
* core: Prevent the UI progressbar from hanging and crashing when there is no
    TTY available. [GH-9974]
* core: Use $APPDATA over $HOME on Windows hosts when determining homedir.
    [GH-9830]
* post-processor/digitalocean-import: Fix crash caused by empty artifact.Files
    slice. [GH-9857]
* post-processor/yandex-export: Check for error after runner completes.
    [GH-9925]
* post-processor/yandex-export: Set metadata key to expected value on error.
    [GH-9849]
* post-processor/yandex-import: Fix S3 URL construct process. [GH-9931]

## 1.6.2 (August 28, 2020)

### FEATURES:
* **New command** `hcl2_upgrade` is a JSON to HCL2 transpiler that allows users
    to transform an existing JSON configuration template into its HCL2 template
    equivalent. Please see [hcl2_upgrade command
    docs](https://packer.io/docs/commands/hcl2_upgrade) for more details.
    [GH-9659]

### IMPROVEMENTS:
* builder/amazon:  Add all of the custom AWS template engines to `build`
    template function for use by provisioners. [GH-9751]
* builder/amazon: Add aws_polling config option to override env variables.
    [GH-9777]
* builder/azure: Add FreeBSD support to azure/chroot builder. [GH-9697]
* builder/vmware-esx: Add `network_name` option to vmware so that users can set
    a network without using vmx data. [GH-9718]
* builder/vmware-vmx: Add additional disk configuration option.  Previously
    only implemented for vmware-iso builder [GH-9815]
* builder/vmware: Add a `remote_output_directory option` so users can tell
    Packer where on a datastore to create a vm. [GH-9784]
* builder/vmware: Add option to export to ovf or ova from a local vmware build
    [GH-9825]
* builder/vmware: Add progress tracker to vmware-esx5 iso upload. [GH-9779]
* builder/vsphere-iso:  Add support for building on a single ESXi host
    [GH-9793]
* builder/vsphere: Add new `directory_permission` config export option.
    [GH-9704]
* builder/vsphere: Add option to import OVF templates to the Content Library
    [GH-9755]
* builder/vsphere: Add step and options to customize cloned VMs. [GH-9665]
* builder/vsphere: Update `iso_paths` to support reading ISOs from Content
    Library paths [GH-9801]
* core/hcl: Add provisioner "override" option to HCL2 templates. [GH-9764]
* core/hcl: Add vault integration as an HCL2 function function. [GH-9746]
* core: Add colored prefix to progress bar so it's clearer what build each
    progress bar belongs to. [GH-9780]
* core: Ui now pretty prints build durations. [GH-9749]
* core: When a build is cancelled, Packer will skip postprocessors gracefully
    rather than failing them. [GH-9720]
* integrations/secretsmanager: Add support for plaintext non-key/pair secrets.
    [GH-9773]
* post-processor/vsphere: Improve UI to catch bad credentials and print errors.
    [GH-9649]
* provisioner/ansible-remote: Add `ansible_ssh_extra_args` so users can specify
    extra arguments to pass to ssh [GH-9821]
* provisioner/file: Clean up, bugfix, and document previously-hidden `sources`
    option. [GH-9725] [GH-9735]
* provisioner/salt-masterless: Add option to option to download community
    SaltStack formulas. [GH-9726]

### BUG FIXES:
* build: Fix bug in code generator that caused generation to fail in nested
    packer/packer dirs [GH-9728]
* build: Fix Makefile so that default target doesn't crash and creates dev
    binaries. [GH-9706]
* builder/amazon-ebssurrogate: Make skip_save_build_region option work in the
    ebssurrogate builder, not just the ebs builder. [GH-9666]
* builder/amazon: Add retry logic to the spot instance creation step to handle
    "Invalid IAM Instance Profile name" errors [GH-9810]
* builder/amazon: Update the `aws_secretsmanager` function to read from the AWS
    credentials file for obtaining default region information; fixes the
    'MissingRegion' error when AWS_REGION is not set [GH-9781]
* builder/file: Make sure that UploadDir receives the interpolated destination.
    [GH-9698]
* builder/googlecompute: Fix bug where startup script hang would cause export
    to hang. [GH-9708]
* builder/hyperv: Send boot command in small chunks to make it more stable.
    [GH-9765]
* builder/scaleway: Fix config issue that made scaleway builder fail when used
    with HCL2 config. [GH-9677]
* builder/vmware: Fully destroy vm if it was cancelled or errored. This will
    make orphaned vms easier to destroy through vCenter. [GH-9782]
* builder/vsphere: Fix `alt`, `ctrl`, and `shift` keypresses in the
    boot_command. [GH-9702] [GH-9739]
* builder/vsphere: Fix bug where Packer timed out if two interfaces were
    defined but only one had an available IP. [GH-9748]
* builder/vsphere: Fix the configuration_parameters option so that it is always
    applied, not just when the tool sync policy is set. [GH-9713]
* communicator: Fix `pause_before_connect` option to force a reconnect after
    the pause. [GH-9772]
* core: Make `max_retries` provisioner option a string to allow variable
    interpolation. [GH-9673]
* post-processor/vsphere-template: Fix ReregisterVM to default to true instead
    of false. [GH-9736]
* post-processor/yandex-export: Fix issue when validating region_name [GH-9814]
* provisioner/inspec: Fix the 'Unsupported argument; An argument named
    "command"' error when using the inspec provisioner in an HCL2 configuration
    [GH-9800]

## 1.6.1 (July 30, 2020)

### BACKWARDS INCOMPATABILITIES:

* HCL: builder/vsphere: Add option to add a xhci/usb3 controller; changes
    controller value to an array of strings.  [GH-9574]
* HCL: New HCL-only `post-processors` block to run chained post-processors
    after a build [GH-9638]. Before this, defining multiple `post-processor`
    blocks after provisioning steps would run them sequentially, now doing this
    makes them start from the build artifact. To queue post-processors you now
    have to define them in a `post-processors` block. [GH-9638]
* post-processor/vSphere: We have fixed a bug in the ovftool URL encoding. This
    may mean that you have performed an encoding workaround that is no longer
    necessary. [GH-9589]

### FEATURES:
* **New post-processor** Yandex Import [GH-9553]

### IMPROVEMENTS:
* builder/amazon-ebs: Wrap CreateImage call in a retry to account for eventual
    consistency [GH-9579]
* builder/azure: Disable SSH password authentication unless password is
    explicitly specified. [GH-9603]
* builder/docker: Add options for --cap-add, --cap-drop, --tmpfs, --device
    [GH-9565]
* builder/file: Create parent directories of target file, if they don't exist.
    [GH-9452]
* builder/googlecompute: Add `wrap_startup_script` configuration option to
    allow the disabling of Packer's startup script wrapper [GH-9505]
* builder/googlecompute: Add support for oslogin via the `use_os_login`
    configuration option [GH-9339]
* builder/googlecompute: Make IAP tunnel timeout configurable. [GH-9545]
* builder/googlecompute: Support using WinRM over an IAP tunnel [GH-9610]
* builder/hyper-v: Include secure boot template in box.xml [GH-9552]
* builder/hyperone: Add support for custom username in vm creation. [GH-9497]
* builder/hyperone: Skip chroot device discovery. [GH-9489]
* builder/openstack: Bump gophercloud to latest version [GH-9573]
* builder/proxmox: Add option to disable KVM hardware virtualization in Proxmox
    builder [GH-9587]
* builder/proxmox: Add support for multiple NIC packet queues [GH-9597]
* builder/proxmox: Enable Proxmox builder to toggle the firewall parameter for
    network interfaces. [GH-9487]
* builder/proxmox: Update Proxmox storagePoolTypes [GH-9418]
* builder/qemu: Add 'cdrom_interface' option to QEMU builder [GH-9483]
* builder/tencentcloud: Add `source_image_name` to support getting source image
    by name [GH-9619]
* builder/tencentcloud: Update cvm root disk type to `CLOUD_PREMIUM` [GH-9663]
* builder/ucloud: New access config options and run config options. [GH-9466]
* builder/vsphere-clone: Add `boot_command` support to vsphere-clone builder,
    including support for starting an HTTP server
* builder/vsphere-clone: Add `vapp` configuration option [GH-9507]
* builder/vsphere: Add ability to define multiple disk controllers [GH-9519]
* builder/vsphere: Add `boot_command` support to vsphere-clone builder. [GH-9397]
* builder/vsphere: Add `content_library_destination` to import VM template to a
    Content Library [GH-9551]
* builder/vsphere: Add `force_bios_setup` configuration option [GH-9563]
* builder/vsphere: Add option to add a xhci/usb3 controller [GH-9574]
* builder/vsphere: Create vm output folders if they don't exist [GH-9402]
* builder/vsphere: Fix file size descriptor when exporting OFV [GH-9568]
* builder/vsphere: Look at all available IPs in the waiting for IP loop.
    [GH-9450]
* builder/vsphere: Match network to host when multiple networks are found
    [GH-9556]
* builder/vsphere: Update vsphere `boot_command` to bring it in line with other
    builders' boot_command functionality. [GH-9406]
* builder/vsphere: Use datacenter inventory path for find folder [GH-9390]
* builder/vsphere: Use value from "ip_wait_address" option to determine the
    default for the http server IP [GH-9441]
* builder/yandex: Allow set `min_disk_size` for an image.
* builder/yandex: Support authentication by Service Account on instance
    [GH-9383]
* builder/yandex: yandex: Add new property `min_disk_size` of built image
    [GH-9594]
* communicator/ssh: Add support for OpenSSH certificate signing [GH-9521]
* communicator/ssh: Allow users to provide a list of ciphers that they want
    Packer to support. [GH-9453]
* core/hcl2: Add possibility to name singular build.source blocks to
    differentiate their output and to filter on them [GH-9490]
* core/hcl2: Add the "inspect" command for hcl2 configs. [GH-9468]
* core/hcl2: HCL configs now respect only/except using build names instead of
    types. [GH-9454]
* core/hcl: Allow use of `keep_input_artifact` in post processors. [GH-9477]
* core/hcl: Share build info with Provisioner and Post-Processor via HCL2
    variables [GH-9444] [GH-9534] [GH-9622]
* core: Add on-error flag option to run error-cleanup-provisioner [GH-9429]
* core: communicator/ssh: Add new `ssh_key_exchange_algorithms` option to
    supply custom key exchange algorithms in SSH client [GH-9634]
* core: refactor initialization out from Packer core to allow `validate
    --syntax-only` to no error when a variable is not set [GH-9627]
* hcl2: Handle uint64 buildvars [GH-9585]
* post-processor/yandex-export: Allow users to utilize generated variables in
    templating. [GH-9555]
* post-processor/yandex-export: Support Authentication by Service Account Key
    file [GH-9379]
* post-processor/yandex-import: Support creating an Image based on another one
    [GH-9614]
* post-processor/yandex-import: Support using URL from yandex-export [GH-9601]
* provisioner/ansible: Add template option for templating the inventory file
    lines [GH-9438]

### BUG FIXES:
* builder/amazon: Change "Resource" field in
    `temporary_iam_instance_profile_policy_document` to be an array of strings,
    not just a single string. [GH-9509]
* builder/amazon: HCL2: Add singular `run_volume_tag` block to ebs & ebssurrogate
    builders. [GH-9457]
* builder/amazon: Retry fetching block device mappings if empty. [GH-9480]
* builder/azure: Fix data disks URI. [GH-9467]
* builder/googlecompute: Fix issue with `use_iap` globally changing a user's
    gcloud project configuration, by temporarily setting "project" via project
    flag and not via `gcloud config` [GH-9662]
* builder/googlecompute: Fix the "secure boot" validation for uefi_compatible
    images. [GH-9371]
* builder/qemu: Only set up localhost port forwarding if skipnatmapping is
    false. [GH-9479]
* builder/vagrant: Fix box file validation for remote box files specified using
    `source_path` [GH-9660]
* builder/vagrant: Improve validation and error handling around synced_folder.
    Make sure that synced folder can be defined relative to Packer run
    directory, not the Vagrant output directory. [GH-9577]
* builder/virtualbox-vm: Fix regression where builder would fail if the vm had
    no snapshots. [GH-9435]
* builder/vmware-iso: Try checksum remote file instead of local file. [GH-9584]
* builder/vmware: Fix a race that occurred when parsing the network config.
    [GH-9387]
* builder/vmware: Update vendor library, enabling retries on 502 errors
    [GH-9391]
* builder/vsphere-clone: Fix SSH public key injection for cloned templates
    [GH-9507]
* builder/vsphere: Clean up folder path to remove leading slashes. [GH-9542]
* builder/vsphere: Deduplicate where Firmware is set in vsphere-iso builder
    [GH-9557]
* builder/vsphere: Fix vsphere ToolsSyncTime and ToolsUpgradePolicy [GH-9515]
* builder/vsphere: vSphere driver context is no longer cancelled when Packer
    context is cancelled. [GH-9576]
* communicator/winrm: Add the "no_proxy" environment variable option to have
    winrm bypass the proxy set by the http_proxy or https_proxy environment
    vars, when connecting to the remote instance. [GH-9267]
* core: Do not print download progress bar if a machine-readable UI is chosen.
    [GH-9448]
* post-processor/amazon-import: Add support for retrying RequestLimitExceeded
    errors when importing an image [GH-9537]
* post-processor/docker-import: Fix crash when using docker-import with HCL.
    [GH-9670]
* post-processor/vsphere: Fix password encoding in vsphere post-processor
    ovftool call [GH-9589]
* post-processor/yandex-export: Fix error handling and docs. [GH-9554]
* provisioner/ansible-local: Fix agent auth in SSH communicator  [GH-9639]
* provisioner/ansible: Correct check for whether PackerHttpAddr is implemented
    or not [GH-9498]
* provisioner/ansible: Quote extra-var packer_build_name to handle names with
    spaces [GH-9590]

## 1.6.0 (June 09, 2020)

### BACKWARDS INCOMPATIBILITIES:
* builder/vsphere-iso: The deprecated fields `network`, `network_card`,
    `disk_size`, `disk_thin_provisioned`, and `disk_eagerly_scrub` have been
    removed. Run `packer fix template.json` to migrate an existing template
    [GH-9149]

* For all iso-based builders, the `iso_checksum_type` and `iso_checksum_url`
    fields have been removed in favor of simply setting the `iso_checksum`
    field. Use the [packer fix](https://www.packer.io/docs/commands/fix/)
    command to update a config file. See the [iso_checksum field
    docs](https://www.packer.io/docs/builders/virtualbox/iso/#iso_checksum) to
    read more about this. [GH-8437]

### FEATURES:
* **New post-processor** Yandex Export [GH-9124]
* builder/amazon: Add SSM Session Manager as a SSH interface connection
    [GH-9082]
* builder/google: Implement iap proxy for googlecompute [GH-9105]
* HCL2/core: Add command/console support for HCL2 config files [GH-9359]
* HCL2/core: Add command/validate support for HCL2 config files [GH-9346]
* HCL2/core: it is now possible to set source fields from the `build` block
    [GH-9291]

### IMPROVEMENTS:
* builder/amazon: Added new `build` template function field:
    SourceAMICreationDate [GH-9277]
* builder/azure-arm: Support User Assigned Managed Identity [GH-9293]
* builder/azure-chroot: Copy data disks between shared image galleries
    [GH-9323]
* builder/googlecompute: Changed default disk size. [GH-9071]
* builder/googlecompute: New option to specify storage location for GCP images
    [GH-9326]
* builder/qemu: Add support for using a network bridge [GH-9159]
* builder/qemu: Added `skip_nat_mapping` option to skip the communicator (SSH
    or WinRM) automatic port forward and use the guest port directly. [GH-9307]
* builder/qemu: Replace deprecated `ssh_host_port_min` and `ssh_host_port_max`
    by `host_port_min` and `host_port_max`. [GH-9307]
* builder/virtualbox: Add `output_filename` config option to allow user to set
    a custom filename instead of forcing it to be the same as vm_name.
    [GH-9174]
* builder/vsphere: Add `floppy_label` option. [GH-9187]
* builder/vsphere: Add `ip_wait_address` to allow packer to wait for an ip
    address that fits within the range (cidr address).  [GH-9358]
* builder/vsphere: Add `tools_sync_time` and `tools_upgrade_policy` options to
    set time synchronization and automatic VMware Tools upgrade. [GH-9366]
* core/hcl: Enable "force", "debug", and "on-error" command line flags for hcl2
    builds. [GH-9234]
* core/hcl: Enable only/except for provisioners and post-processors [GH-9357]
* core/hcl: Support named builds in HCL2 templates [GH-9245]
* core/interpolation: Add support for specifying a particular key to fetch from
    the AWS Secrets Manager [GH-9202] [GH-9286]
* core: HCL logs now display source type and source name (`type.name`) in logs
    to differentiate more easily who says what. [GH-9257]
* core: update consul and vault dependencies [GH-9205]
* core: Update vendored "go-getter" library with checksum fixes.
* core: Users can now configure the http bind address of the server launched to
    serve files from http_directory [GH-9313]
* core: When a template contains a deprecated option, Packer will now encourage
    the user to call `packer fix`. [GH-9325]
* post-processor/docker-push: Support pushing multiple tags [GH-9182]
* post-processor/docker-tag: Change field name of docker tag to "tags" instead
    of "tag" since it's a list. Keep "tag" for backwards compatibility.
    [GH-9183]
* provisioner/windows-restart: Display full hostname, instead of just the
    NetBIOS name [GH-9335]

### BUG FIXES:
* buidler/vsphere-iso: Fix bug validating iso_urls [GH-9321]
* builder/azure-arm: Update runtime constants with a valid DataDiskName
    [GH-9251]
* builder/digitalocean: Use correct image type for Droplet creates. [GH-9212]
* builder/openstack: Don't error if metadata can't be set. Old versions of
    openstack don't support that API call. [GH-9198]
* builder/vagrant: Use absolute path for package_include files to prevent them
    from having to be relative to the output vagrant directory. [GH-9260]
* builder/virtualbox: Fix bug using checksum files. [GH-9101]
* builder/vmware: Fix ssh connection when the guest platform's dhcpcd switches
    the IP address in-between a build. [GH-9322]
* builder/vsphere-iso: Use the Datacenter's VmFolder call instead of manually
    generating folder path [GH-9342]
* builder/vsphere: Add option not to set host during datastore upload. [GH-9100
* builder/vsphere: Fix crash in the driver for an interface conversion of
    types.AnyType nil to types.ManagedObjectReference. [GH-9354]
* builder/vsphere: Fix iso config prepare being called incorrectly, which
    caused `iso_url` field to fail. [GH-9197]
* builder/yandex: Do not require 'gpu-standard-v1' platform_id for any GPU-
    based config. [GH-9356]
* core: Ensure HTTP server information `PackerHTTPIP`, `PackerHTTPPort`, and
    `PackerHTTPAddr` are available via the `build` template engine for all
    supported builders [GH-9238]
* core: Fix `iso_url` to accept SMB shared files UNC and windows network paths.
    [GH-8954]
* core: Fix regression that broke use of pwd when retrieving the checksum from
    a file [GH-9129].
* post-processor/vagrant: Add "provider_override" template option to allow
    artifacts from the Artifice post-processor [GH-9239]
* post-processor/vsphere-template: Add VSphere builder's artifact to vsphere-
    template's supported types [GH-9146]
* postprocessor/artifice: Update various core post-processors to accept
    artifacts from the Artifice post-processor [GH-9239]
* provisioner/ansible-remote: Fix the arg order to make sure that the playbook
    file is the last item in the call [GH-9279]
* provisioner/ansible: Ansible provisioner doesn't force ssh key checking if
    user wants to use password to connect instead. [GH-9350]
* provisioner/inspec: Fix build variables interpolation [GH-9262]
* provisioner/powershell: fix interpolation of execute_command in cleanup
    script call. [GH-9275]
* provisioner/powershell: Fix long-wait retry loop caused by cleanup logic
    [GH-9226]
* provisioner/salt-masterless: Ignore the CmdArgs field in hcl2 [GH-9290]

## 1.5.6 (May 1, 2020)

### BACKWARDS INCOMPATIBILITIES:
* core/hcl2: HCL2 singular blocks: use key/value or name/value depending on how
    the service names things [GH-9078]
* core/hcl2: Maps are now treated as settable arguments as opposed to blocks.
    For example `tags = {}` instead of `tags {}` [GH-9035]
* `packer build` command: removed option to set `parallel=false`, use
    `-parallel-builds=1` for this.

### FEATURES:
* **New Builder** azure-dtl allows creation of devtestlabs images in Azure
    [GH-8987]

* **New Core Feature** provisioners now support a `max_retries` option that can
    be used for retrying a provisioner on error [GH-9061]

* **New Post-Processor**: `yandex-export` Upload built image in Yandex Object Storage.

### IMPROVEMENTS:
* builder/azure-arm: Add `boot_diag_storage_account` option for enabling boot
    diagnostics on a virtual machine [GH-9053]
* builder/azure-arm: Add support for setting custom resource names [GH-9028]
* builder/azure-arm: Data disk names are now randomly generated [GH-8986]
* builder/azure: Add shared image destination for azure-chroot [GH-9021]
* builder/azure: Add shared image source for chroot builder [GH-9070]
* builder/google: Add support for custom shielded images [GH-8970]
* builder/google: Allow `source_image_project_id` to be a list of several
    projects to search. [GH-8679]
* builder/hyperv: Add `boot_order` option to override the default boot order
    [GH-9046]
* builder/oracle-oci: Allow Instance Principal Auth for Oracle OCI builder
    [GH-8893]
* builder/oracle-oci: Update Oracle SDK. [GH-9104]
* builder/proxmox: Add ability to add a cloud-init drive [GH-9045]
* builder/vsphere: Add `disable_shutdown` option to support manual vm shutdown
    over the default automatic shutdown step [GH-9095]
* builder/vsphere: Add `vgpu_profile` option for specifying vGPU profiles
    [GH-8946]
* builder/vsphere: Add support for EFI Secure Boot [GH-9018]
* builder/vsphere: Add support for specifying vGPU profiles [GH-8946]
* builder/yandex: Add `target_image_folder_id ` option for changing the folder
    where a built image will be saved to [GH-9080]
* core/hcl2: HCL mode builds now honor -only and -except options [GH-8947]
* core/hcl2: Set `packer_build_name` and `packer_builder_type` variables for
    builder provisioners and post-processors [GH-8956]
* core/HCL: HCL mode now honors -only and -except options. [GH-8947]
* core: New template function: aws_secretsmanager [GH-9099]
* provisioner/ansible: Add `use_proxy` option to not use localhost proxy adapter. Removes
    need for ansible connection_plugin when using WinRM. [GH-8625]
* provisioner/powershell: Add `debug_mode` option to help with debugging
    generated scripts [GH-8996]
* provisioner/powershell: Add cleanup step to remove temporarily created
    scripts; cleanup can be skipped by setting the `skip_clean` option
    [GH-8908]

### BUG FIXES:
* builder/amazon: Fix bug with `launch_block_device_mappings` in spot
    instances. [GH-8945]
* builder/azure-arm: Fix issue where managed image builds were using a
    different `location` then what was specified in the build configuration
    [GH-9068]
* builder/azure: Allow Managed Data Disks to be used with Azure Shared Image
    Gallery  [GH-8912]
* builder/azure: Fix SSH connection for temporary admin users specified in
    `ssh_username` [GH-9103]
* builder/osc: Make compliant with oAPI spec for Outscale osc-bsu builder
    [GH-9093]
* builder/qemu: Remove `net_device` pre-validation [GH-8979]
* builder/tencentcloud: Update builder to handle the
    InstanceOperationInProgress error [GH-9069]
* builder/vsphere-iso: disk_size is no longer required if storage is defined
    [GH-8975]
* builder/vsphere: Add exported files to VSphere artifact [GH-9020]
* builder/vsphere: Fix issue where -force is not working with vsphere builders
    [GH-9039]
* core: Fix crash in wrapperreadline helper when calling `os.NewFile` on
    unknown file descriptor [GH-9037]
* core: Make sure CLI variables supersede variables from var files [GH-8964]
* provisioner/powershell: Fix integer decoding issue in the execution policy
    parser [GH-8997]

## 1.5.5 (March 25,2020)

### BACKWARDS INCOMPATIBILITIES:
* core: Interpolation of undefined variables will now error as expected, in
    previous versions variables were allowed to be set as a command line
    argument but that was because the validation was being ignored

### IMPROVEMENTS:
* builder/azure: Add support for configurable KeyVault SKU [GH-8879]
* builder/hyperv: Add `first_boot_device` setting to allow the selection of the
    initial device or device class used for booting the VM. [GH-8714]
* builder/hyperv: Fix Hyper-V compacted disk size comparison [GH-8811]
* builder/openstack: Add new `image_auto_accept_members` option [GH-8931]
* builder/proxmox: Add ability to specify vga adapter [GH-8892]
* builder/proxmox: Add onboot directive support [GH-8935]
* builder/tencentcloud: Show tencentcloud image id after copy to destination
    region. [GH-8763]
* builder/vmware-iso: Add `cleanup_remote_cache` config option to [GH-8917]
* builder/vmware-iso: Do not perform dial test of NIC when ssh bastion is
    required [GH-8877]
* builder/vsphere-clone: Add ability to export VM to OVF file [GH-8764]
* builder/vsphere-iso: Add ability to define multiple disks. [GH-8787]
* builder/vsphere-iso: Add ability to export VM to OVF file [GH-8764]
* builder/vsphere-iso: Add support for eagerly zeroed / scrubbed disks.
    [GH-8756]
* builder/vsphere-iso: Add the remote iso first so that it is first in boot
    order, and clarify boot behavior. [GH-8732]
* communicator/ssh: Add flag to enable support for keyboard-interactive auth to
    connect bastion [GH-8847]
* core/hcl2: Add support for singular blocks [GH-8889]
* core/hcl2: Add support in HCL2 configs for dynamic blocks, document for loops
    and splat expressions [GH-8720]
* core/hcl2: Fix HCL2 local variables decoding to allow local usage within
    another local in the same locals block [GH-8755]
* core/hcl2: Import new replace and regex_replace funcs from go-cty +
    documentation [GH-8863]
* core: Enable hcl files as var files in HCL mode [GH-8882]
* core: Make "build" engine template variables SSHPublicKey and SSHPrivateKey
    strings [GH-8829]

### Bug Fixes:
* builder/azure: Fix `winrm_password` attribution and allow users to set
    `winrm_username` [GH-8928]
* builder/azure: Fix azure key vault cleanup failure [GH-8905]
* builder/azure: Fix HCL2 bug that prevented Azure and other builders from
    loading properly. [GH-8785]
* builder/googlecompute: Fix WinRMPassword template engine. [GH-8890]
* builder/googlecompute: Replace deprecated "sshKeys" metadata with "ssh-keys"
    to fix SSH authentication issue [GH-8942]
* builder/proxmox: Add new validation to catch that template_name cannot
    contain spaces. [GH-8799]
* builder/proxmox: Bump proxmox-api-go to fix upstream bug where users hit open
    file limit. [GH-8800]
* builder/vagrant: Fix path validation in ssh config step. [GH-8826]
* builder/virtualbox-vm: Fix crash when VM has no snapshots. [GH-8906]
* builder/virtualbox: Remove all floppy controllers before adding a new one.
    [GH-8828]
* builder/vsphere-clone: Fix issue preventing the cloning of VMs with the same
    name in different folders [GH-8938]
* builder/vsphere-iso: Fix issue preventing the creation of VMs with the same
    name in different folders [GH-8938]
* builder/vsphere: Fix network object interface panic. [GH-8753]
* core/hcl2: Fix crash when an unset variable is used [GH-8837]
* core/hcl2: Fix logic for parsing literal value variables [GH-8834]
* core/hcl2: Make sure locals are evaluated only after variables are. [GH-8918]
* core: Fix "build" template engine interpolation for certain fields in certain
    provisioners. [GH-8771]
* core: Fix bug where user var recursion could fail intermittently when used
    with env vars [GH-8875]
* core: Interpolation of undefined variables will now error as expected
* plugins: Make plugin discovery stricter with respect to periods so that users
    can disable plugins by renaming the extension [GH-8735]
* provisioner/salt: Fix `no_exit_on_failure` config to work correctly as
    expected. [GH-9119]
* provisioner/shell: "inline" config option is now a template engine. [GH-8883]

## 1.5.4 (February 14, 2020)
no-change release to fix code-signing on OSX binaries. Since checksums for these
binaries has changed, we are releasing a second time to prevent confusion.

## 1.5.3 (February 14, 2020)

### IMPROVEMENTS:
* builder/vsphere: Add ability to define multiple NICs for vsphere-iso
    [GH-8739]
* builder/vsphere: Add option to remove CD-ROM drives. [GH-8690]
* core: Add validation to catch when users accidentally add duplicate fields to
    template [GH-8725]

### Bug Fixes:
* core/hcl2: Fix template prepare/validation for HCL2 templates [GH-8742]
* core: Fix `build` template function interpolation [GH-8727]

## 1.5.2 (February 12, 2020)
**New Builder** The vsphere-iso builder, previously maintained by JetBrains,
has been merged with the Packer core. It will be officially supported by the
Packer team at HashiCorp moving forward. [GH-8480]

**HCL2 variables & functions** HCL2 configurations can now use `variable`,
`variables`, `locals`, and functions [GH-8588].

### IMPROVEMENTS:
* builder/alicloud: Add AlicloudProfile option. [GH-8560]
* builder/amazon: Add max_retries option to aws builders [GH-8709]
* builder/amazon: Add source AMI owner ID/name to template engines [GH-8550]
* builder/amazon: Update instance waiters to use global waiter settings set by
    `AWS_POLL_DELAY_SECONDS` and `AWS_TIMEOUT_SECONDS` [GH-8699]
* builder/azure: Allow users to use custom key vault for storing Windows
    certificates [GH-8704]
* builder/azure: Set expiry for image versions in SIG [GH-8561]
* builder/proxmox: Add option to upload the boot ISO rather than pointing out a
    previously manually uploaded one. [GH-8624]
* builder/vagrant: Fix a crash in the Vagrant driver [GH-8607]
* builder/yandex: Add service account ID to config [GH-8717]
* communicator/winrm: Users can now override winrm_host with a static IP even
    when using cloud builders. [GH-8675]
* core/hcl2: Fix bug preventing reading slices within other slices [GH-8669]
* core:  Interpolation within post-processors can now access build-specific
    values like Host IP, communicator password, and more [GH-8632]
* core: Add `PACKER_PLUGIN_PATH` to list of supported paths for plugin
    discovery [GH-8616]
* core: clean up messy log line in plugin execution. [GH-8542]
* core: Ensure `PACKER_HTTP_ADDR` is always set for any builder that provides a
    HTTP server for file transfer [GH-8654]
* core: Fix loading external plugins defined in PACKER_CONFIG [GH-8582]
* core: Log name of postprocessor running to disambiguate long chains of post-
    processors. [GH-8613]
* core: Packer can use isos in-place on Windows again, instead of copying them
    into its cache. [GH-7627]
* core: step_download: return without error if Urls is empty [GH-8579]
* post-processor/vsphere-template] Simplify method to use vm.MarkAsTemplate
    (optionally) [GH-8511]
* scripts: Fix some issues with mapstructure-to-hcl2 code generator. [GH-8574]
* scripts: Update Vagrant bootstrapping scripts [GH-8604]

### Bug Fixes:
* builder/alicloud: Fix "security group doesn't exist" error when there are >10
    security groups. [GH-8535]
* builder/amazon: Allow AWS builder pre-validation to pass when subnet filters
    are present [GH-8622]
* builder/azure: Fix bug where deployments were not being cleaned up: [GH-8496]
* builder/azure: Fix issue where WinRMPassword was being left unset [GH-8670]
* builder/lxd: Fix file uploading issue when using the file provisioner
    [GH-8636]
* builder/null: Fix crash when configuring builder using HCL2. [GH-8612]
* builder/osc: Fix ssh host detection in Public Cloud and Nets [GH-8414]
* builder/vagrant: Fix bug with reading key from a path with spaces [GH-8605]
* builder/virtualbox-ovf: Remove config dependency from StepImport [GH-8509]
* builder/virtualbox-vm: use config as a non pointer to avoid a panic [GH-8576]
* core: Fix crash when build.sources is set to an invalid name [GH-8569]
* core: Fix error loading .packerconfig [GH-8623]
* core: Fix loading local ISO files when using `iso_target_path` [GH-8689]
* core: Fix loading of external plugins. GH-8543]
* post-processor/docker-tag: Fix regression if no tags were specified.
    [GH-8593]
* post-processor/vagrant-cloud: Update error handling for Vagrant Cloud API
    [GH-8594]
* post-processor/vagrant: correctly handle the diskSize property as a qemu size
    string [GH-8567]
* provisioner/ansible: Fix password sanitization to account for empty string
    values. [GH-8570]
* provisioner/shell: Fix bug with shell provisioner failing to clean up the
    environment var file when env_var_file is true. [GH-8639]

## 1.5.1 (December 20, 2019)
This was a fast-follow release to fix a number of panics that we introduced when
making changes for HCL2.

### IMPROVEMENTS:
* builder/alicloud: Add show_expired option for describing images [GH-8425]

### Bug Fixes:
* builder/cloudstack: Fix panics associated with loading config [GH-8513]
* builder/hyperv/iso: Fix panics associated with loading config [GH-8513]
* builder/hyperv/vmcx: Fix panics associated with loading config [GH-8513]
* builder/jdcloud: Update jdcloud statebag to use pointers for config [GH-8518]
* builder/linode: Fix panics associated with loading config [GH-8513]
* builder/lxc: Fix panics associated with loading config [GH-8513]
* builder/lxd: Fix panics associated with loading config [GH-8513]
* builder/oneandone: Fix panics associated with loading config [GH-8513]
* builder/oracle/classic: Fix panics associated with loading config [GH-8513]
* builder/oracle/oci: Fix panics associated with loading config [GH-8513]
* builder/osc/bsuvolume: Fix panics associated with loading config [GH-8513]
* builder/parallels/pvm: Fix panics associated with loading config [GH-8513]
* builder/profitbricks: Fix panics associated with loading config [GH-8513]
* builder/scaleway: Fix panics associated with loading config [GH-8513]
* builder/vagrant: Fix panics associated with loading config [GH-8513]
* builder/virtualbox/ovf: Fix panics associated with loading config [GH-8513]
* builder/virtualbox: Configure NAT interface before forwarded port mapping
    #8514
* post-processor/vagrant-cloud: Configure NAT interface before forwarded port
    mapping [GH-8514]

## 1.5.0 (December 18, 2019)

### IMPROVEMENTS:
* builder/amazon: Add no_ephemeral template option to remove ephemeral drives
    from launch mappings. [GH-8393]
* builder/amazon: Add validation for "subnet_id" when specifying "vpc_id"
    [GH-8360] [GH-8387] [GH-8391]
* builder/amazon: allow enabling ena/sr-iov on ebssurrogate spot instances
    [GH-8397]
* builder/amazon: Retry runinstances aws api call to mitigate throttling
    [GH-8342]
* builder/hyperone: Update builder schema and tags [GH-8444]
* builder/qemu: Add display template option for qemu. [GH-7676]
* builder/qemu: Disk Size is now read as a string to support units. [GH-8320]
    [GH-7546]
* builder/qemu: Add fixer to convert disk size from int to string [GH-8390]
* builder/qemu: Disk Size is now read as a string to support units. [GH-8320]
    [GH-7546]
* builder/qemu: When a user adds a new drive in qemuargs, process it to make
    sure that necessary settings are applied to that drive. [GH-8380]
* builder/vmware: Fix error message when ovftool is missing [GH-8371]
* core: Cleanup logging for external plugins [GH-8471]
* core: HCL2 template support is now in beta. [GH-8423]
* core: Interpolation within provisioners can now access build-specific values
    like Host IP, communicator password, and more. [GH-7866]
* core: Various fixes to error handling. [GH-8343] [GH-8333] [GH-8316]
    [GH-8354] [GH-8361] [GH-8363] [GH-8370]
* post-processor/docker-tag: Add support for multiple tags. [GH-8392]
* post-processor/shell-local: Add "valid_exit_codes" option to shell-local.
    [GH-8401]
* provisioner/chef-client: Add version selection option. [GH-8468]
* provisioner/shell-local: Add "valid_exit_codes" option to shell-local.
    [GH-8401]
* provisioner/shell: Add support for the "env_var_format" parameter [GH-8319]

### BUG FIXES:
* builder/amazon: Fix request retry mechanism to launch aws instance [GH-8430]
* builder/azure: Fix PollDuration option which was overridden in some clients.
    [GH-8490]
* builder/hyperv: Fix bug in checking VM name that could cause flakiness if
    many VMs are defined. [GH-8357]
* builder/vagrant: Use absolute path for Vagrantfile [GH-8321]
* builder/virtualbox: Fix panic in snapshot builder. [GH-8336] [GH-8329]
* communicator/winrm: Resolve ntlm nil pointer bug by bumping go-ntlmssp
    dependency [GH-8369]
* communicator: Fix proxy connection settings to use "SSHProxyUsername" and
    "SSHProxyPassword" where relevant instead of bastion username and password.
    [GH-8375]
* core: Fix bug where Packer froze if asked to log an extremely long line
    [GH-8356]
* core: Fix iso_target_path option; don't cache when target path is non-nil
    [GH-8394]
* core: Return exit code 1 when builder type is not found [GH-8474]
* core: Return exit code 1 when builder type is not found [GH-8475]
* core: Update to newest version of go-tty to re-enable CTRL-S and CTRL-Q usage
    [GH-8364]

### BACKWARDS INCOMPATIBILITIES:
* builder/amazon: Complete deprecation of clean_ami_name template func
    [GH-8320] [GH-8193]
* core: Changes have been made to both the Prepare() method signature on the
    builder interface and on the Provision() method signature on the
    provisioner interface. [GH-7866]
* provisioner/ansible-local: The "galaxycommand" option has been renamed to
    "galaxy_command". A fixer has been written for this, which can be invoked
    with `packer fix`. [GH-8411]

## 1.4.5 (November 4, 2019)

### IMPROVEMENTS:
* added ucloud-import post-processsor to import custom image for UCloud UHost
    instance [GH-8261]
* builder/amazon: New option to specify IAM policy for a temporary instance
    profile [GH-8247]
* builder/amazon: improved validation around encrypt_boot and kms_key_id for a
    better experience [GH-8288]
* builder/azure-arm: Allow specification of polling duration [GH-8226]
* builder/azure-chroot: Add Azure chroot builder [GH-8185] & refactored some
    common code together after it [GH-8269]
* builder/azure: Deploy NSG if list of IP addresses is provided in config
    [GH-8203]
* builder/azure: Set correct user agent for Azure client set [GH-8259]
* builder/cloudstack: Add instance_display_name for cloudstack builder
    [GH-8280]
* builder/hyperv: Add the additional_disk_size option tho the hyperv vmcx
    builder. [GH-8246]
* builder/openstack: Add option to discover provisioning network [GH-8279]
* builder/oracle-oci: Support defined tags for oci builder [GH-8172]
* builder/proxmox: Add ability to select CPU type [GH-8201]
* builder/proxmox: Add support for SCSI controller selection [GH-8199]
* builder/proxmoz: Bump Proxmox dependency: [GH-8241]
* builder/tencent: Add retry on remote api call [GH-8250]
* builder/vagrant: Pass through logs from vagrant in real time rather than
    buffering until command is complete [GH-8274]
* builder/vagrant: add insert_key option for toggling whether to add Vagrant's
    insecure key [GH-8274]
* builder/virtualbox: enabled pcie disks usage, but this feature is in beta and
  won't work out of the box yet [GH-8305]
* communicator/winrm: Prevent busy loop while waiting for WinRM connection
    [GH-8213]
* core: Add strftime function in templates [GH-8208]
* core: Improve error message when comment is bad [GH-8267]
* post-processor/amazon-import: delete intermediary snapshots [GH-8307]
* Fix various dropped errors an removed unused code: [GH-8230] [GH-8265]
    [GH-8276] [GH-8281] [GH-8309] [GH-8311] [GH-8304] [GH-8303] [GH-8293]

### BUG FIXES:
* builder/amazon: Fix region copy for non-ebs amazon builders [GH-8212]
* builder/amazon: Fix spot instance bug where builder would fail if one
    availability zone could not support the requested spot instance type, even
    if another AZ could do so. [GH-8184]
* builder/azure: Fix build failure after a retry config generation error.
    [GH-8209]
* builder/docker: Use a unique temp dir for each build to prevent concurrent
    builds from stomping on each other [GH-8192]
* builder/hyperv: Improve filter for determining which files to compact
    [GH-8248]
* builder/hyperv: Use first adapter, rather than failing, when multiple
    adapters are attached to host OS's VM switch [GH-8234]
* builder/openstack: Fix setting openstack metadata for use_blockstorage_volume
    [GH-8186]
* builder/openstack: Warn instead of failing on terminate if instance is
    already shut down [GH-8176]
* post-processor/digitalocean-import: Fix panic when 'image_regions' not set
    [GH-8179]
* provisioner/powershell: Fix powershell syntax error causing failed builds
    [GH-8195]

## 1.4.4 (October 1, 2019)

### IMPROVEMENTS:
**New Core Feature** Error cleanup provisioner [GH-8155]
* builder/amazon: Add ability to set `run_volume_tags` [GH-8051]
* builder/amazon: Add AWS API call reties on AMI prevalidation [GH-8034]
* builder/azure: Refactor client config [GH-8121]
* builder/cloudstack: New step to detach iso. [GH-8106]
* builder/googlecompute: Fail fast when image name is invalid. [GH-8112]
* builder/googlecompute: Users can now query Vault for an Oauth token rather
    than setting an account file [GH-8143]
* builder/hcloud: Allow selecting image based on filters [GH-7945]
* builder/hyper-v: Decrease the delay between Hyper-V VM startup and hyper-v
    builder's ability to send keystrokes to the target VM. [GH-7970]
* builder/openstack: Store WinRM password for provisioners to use [GH-7940]
* builder/proxmox: Shorten default boot_key_interval to 5ms from 100ms
    [GH-8088]
* builder/proxmox: Allow running the template VM in a Proxmox resource pool
    [GH-7862]
* builder/ucloud: Make ucloud builder's base url configurable [GH-8095]
* builder/virtualbox-vm: Make target snapshot optional [GH-8011] [GH-8004]
* builder/vmware: Allow user to attach floppy files to remote vmx builds
    [GH-8132]
* builder/yandex: Add ability to retry API requests [GH-8142]
* builder/yandex: Support GPU instances and set source image by name [GH-8091]
* communicator/ssh: Support for SSH port tunneling [GH-7918]
* core: Add a new `floppy_label` option [GH-8099]
* core: Added version compatibility to console command [GH-8080]
* post-processor/vagrant-cloud: Allow blank access_token for private vagrant
    box hosting [GH-8097]
* post-processor/vagrant-cloud: Allow use of the Artifice post-processor with
    the Vagrant Cloud post-processor [GH-8018] [GH-8027]
* post-processor/vsphere: Removed redundant whitelist check for builders,
    allowing users to use post-processor without the VMWare builder [GH-8064]

### BUG FIXES:
* builder/amazon: Fix FleetID crash. [GH-8013]
* builder/amazon: Gracefully handle rate limiting when retrieving winrm
    password. [GH-8087]
* builder/amazon: Fix race condition in spot instance launching [GH-8165]
* builder/amazon: Amazon builders now respect ssh_host option [GH-8162]
* builder/amazon: Update the AWS sdk to resolve some credential handling issues
    [GH-8131]
* builder/azure: Avoid a panic in getObjectIdFromToken [GH-8047]
* builder/googlecompute: Fix crash caused by nil account file. [GH-8102]
* builder/hyper-v: Fix when management interface is not part of virtual switch
    [GH-8017]
* builder/openstack: Fix dropped error when creating image client. [GH-8110]
* builder/openstack: Fix race condition created when adding metadata [GH-8016]
* builder/outscale: Get SSH Host from VM.Nics instead of VM Root [GH-8077]
* builder/proxmox: Bump proxmox api dep, fixing bug with checking http status
    during boot command [GH-8083]
* builder/proxmox: Check that disk format is set when pool type requires it
    [GH-8084]
* builder/proxmox: Fix panic caused by cancelling build [GH-8067] [GH-8072]
* builder/qemu: Fix dropped error when retrieving version [GH-8050]
* builder/vagrant: Fix dropped errors in code and tests. [GH-8118]
* builder/vagrant: Fix provisioning boxes, define source and output boxes
    [GH-7957]
* builder/vagrant: Fix ssh and package steps to use source syntax. [GH-8125]
* builder/vagrant: Use GlobalID when provided [GH-8092]
* builder/virtualbox: Fix windows pathing problem for guest additions checksum
    download. [GH-7996]
* builder/virtualbox: LoadSnapshots succeeds even if machine has no snapshots
    [GH-8096]
* builder/vmware: fix dropped test errors [GH-8170]
* core: Fix bug where sensitive variables containing commas were not being
    properly sanitized in UI calls. [GH-7997]
* core: Fix handling of booleans where "unset" is a value distinct from
    "false". [GH-8021]
* core: Fix tests that swallowed errors in goroutines [GH-8094]
* core: Fix bug where Packer could no longer run as background process [GH-8101]
* core: Fix zsh auto-completion [GH-8160]
* communicator/ssh: Friendlier message warning user that their creds may be
    wrong [GH-8167]
* post-processor/amazon-import: Fix non-default encryption. [GH-8113]
* post-processor/vagrant-cloud: Fix dropped errors [GH-8156]
* provisioner/ansible: Fix provisioner dropped errors [GH-8045]

### BACKWARDS INCOMPATIBILITIES:
* core: "sed" template function has been deprecated in favor of "replace" and
    "replace_all" functions [GH-8119]

## 1.4.3 (August 14, 2019)

### IMPROVEMENTS:
* **new builder** UCloud builder [GH-7775]
* **new builder** Outscale [GH-7459]
* **new builder** VirtualBox Snapshot [GH-7780]
* **new builder** JDCloud [GH-7962]
* **new post-processor** Exoscale Import post-processor [GH-7822] [GH-7946]
* build: Change Makefile to behave differently inside and outside the gopath
    when generating code. [GH-7827]
* builder/amazon: Don't calculate spot bids; Amazon has changed spot pricing to
    no longer require this. [GH-7813]
* builder/google: Add suse-byos-cloud to list of public GCP cloud image
    projects [GH-7935]
* builder/openstack: New `image_min_disk` option [GH-7290]
* builder/openstack: New option `use_blockstorage_volume` to set openstack
    image metadata [GH-7792]
* builder/openstack: Select instance network on which to assign floating ip
    [GH-7884]
* builder/qemu: Implement VNC password functionality [GH-7836]
* builder/scaleway: Allow removing volume after image creation for Scaleway
    builder [GH-7887]
* builder/tencent: Add `run_tags` to option to tag instance. [GH-7810]
* builder/tencent: Remove unnecessary image name validation check. [GH-7786]
* builder/tencent: Support data disks for tencentcloud builder [GH-7815]
* builder/vmware: Fix intense CPU usage because of poorly handled errors.
    [GH-7877]
* communicator: Use context for timeouts, interruption in ssh and winrm
    communicators [GH-7868]
* core: Change how on-error=abort is handled to prevent EOF errors that mask
    real issues [GH-7913]
* core: Clean up logging vs ui call in step download [GH-7936]
* core: New environment var option to allow user to set location of config
    directory [GH-7912]
* core: Remove obsolete Cancel functions from builtin provisioners [GH-7917]
* post-processor/vagrant:  Add option to allow box Vagrantfiles to be generated
    during the build [GH-7951]
* provisioner/ansible: Add support for installing roles with ansible-galaxy
    [GH-7916
* provisioner/salt-masterless: Modify file upload to handle non-root case.
    [GH-7833]

### BUG FIXES:
* builder/amazon: Add error to warn users of spot_tags regression. [GH-7989]
* builder/amazon: Allow EC2 Spot Fleet packer instances to run in parallel
    [GH-7818]
* builder/amazon: Fix failures and duplication in Amazon region copy and
    encryption step. [GH-7870] [GH-7923]
* builder/amazon: No longer store names of volumes which get deleted on
    termination inside ebssurrogate artifact. [GH-7829]
* builder/amazon: Update aws-sdk-go to v1.22.2, resolving some AssumeRole
    issues [GH-7967]
* builder/azure: Create configurable polling duration and set higher default
    for image copies to prevent timeouts on successful copies [GH-7920]
* builder/digitalocean: increase timeout for Digital Ocean snapshot creation.
    [GH-7841]
* builder/docker: Check container os, not host os, when creating container dir
    default [GH-7939]
* builder/docker: Fix bug where PACKER_TMP_DIR was created with root perms on
    linux [GH-7905]
* builder/docker: Fix file download hang caused by blocking ReadAll call
    [GH-7814]
* builder/google: Fix outdated oauth URL. [GH-7835] [GH-7927]
* builder/hyperv: Improve code for detecting IP address [GH-7880]
* builder/ucloud: Update the api about stop instance to fix the read-only image
    build by ucloud-uhost [GH-7914]
* builder/vagrant: Fix bug where source_path was being used instead of box_name
    when generating the Vagrantfile. [GH-7859]
* builder/virtualbox: Honor value of 'Comment' field in ssh keypair generation.
    [GH-7922]
* builder/vmware: Fix validation regression that occurred when user provided a
    checksum file [GH-7804]
* buildere/azure: Fix crash with managed images not published to shared image
    gallery. [GH-7837]
* communicator/ssh: Move ssh_interface back into individual builders from ssh
    communicator to prevent validation issues where it isn't implemented.
    [GH-7831]
* console: Fix console help text [GH-7960]
* core: Fix bug in template parsing where function errors were getting
    swallowed. [GH-7854]
* core: Fix regression where a local filepath containing `//` was no longer
    properly resolving to `/`. [GH-7888]
* core: Fix regression where we could no longer access isos on SMB shares.
    [GH-7800]
* core: Make ssh_host template option always override all builders' IP
    discovery. [GH-7832]
* core: Regenerate boot_command PEG code [GH-7977]
* fix: clean up help text and fixer order to make sure all fixers are called
    [GH-7903]
* provisioner/inspec: Use --input-file instead of --attrs to avoid deprecation
    warning [GH-7893]
* provisioner/salt-masterless: Make salt-masterless provisioner respect
    disable_sudo directive for all commands [GH-7774]

## 1.4.2 (June 26, 2019)

### IMPROVEMENTS:
* **new feature:** Packer console [GH-7726]
* builder/alicloud: cleanup image and snapshot if target image is still not
    available after timeout [GH-7744]
* builder/alicloud: let product API determine the default value of io_optimized
    [GH-7747]
* builder/amazon: Add new `skip_save_build_region` option to fix naming
    conflicts when building in a region you don't want the final image saved
    in. [GH-7759]
* builder/amazon: Add retry for temp key-pair generation in amazon-ebs
    [GH-7731]
* builder/amazon: Enable encrypted AMI sharing across accounts [GH-7707]
* builder/amazon: New SpotInstanceTypes feature for spot instance users.
    [GH-7682]
* builder/azure: Allow users to publish Managed Images to Azure Shared Image
    Gallery (same Subscription) [GH-7778]
* builder/azure: Update Azure SDK for Go to v30.0.0 [GH-7706]
* builder/cloudstack: Add tags to instance upon creation [GH-7526]
* builder/docker: Better windows defaults [GH-7678]
* builder/google: Add feature to import user-data from a file [GH-7720]
* builder/hyperv: Abort build if there's a name collision [GH-7746]
* builder/hyperv: Clarify pathing requirements for hyperv-vmcx [GH-7790]
* builder/hyperv: Increase MaxRamSize to match modern Windows [GH-7785]
* builder/openstack: Add image filtering on properties. [GH-7597]
* builder/qemu: Add additional disk support [GH-7791]
* builder/vagrant: Allow user to override vagrant ssh-config details [GH-7782]
* builder/yandex: Gracefully shutdown instance, allow metadata from file, and
    create preemptible instance type [GH-7734]
* core: scrub out sensitive variables in scrub out sensitive variables logs
    [GH-7743]

### BUG FIXES:
* builder/alicloud: Fix describing snapshots issue when image_ignore_data_disks
    is provided [GH-7736]
* builder/amazon: Fix bug in region copy which produced badly-named AMIs in the
    build region. [GH-7691]
* builder/amazon: Fix failure that happened when spot_tags was set but ami_tags
    wasn't [GH-7712]
* builder/cloudstack: Update go-cloudstack sdk, fixing compatibility with
    CloudStack v 4.12 [GH-7694]
* builder/proxmox: Update proxmox-api-go dependency, fixing issue calculating
    VMIDs. [GH-7755]
* builder/tencent: Correctly remove tencentcloud temporary keypair. [GH-7787]
* core: Allow timestamped AND colorless ui messages [GH-7769]
* core: Apply logSecretFilter to output from ui.Say [GH-7739]
* core: Fix "make bin" command to use reasonable defaults. [GH-7752]
* core: Fix user var interpolation for variables set via -var-file and from
    command line [GH-7733]
* core: machine-readable UI now writes UI calls to logs. [GH-7745]
* core: Switch makefile to use "GO111MODULE=auto" to allow for modern gomodule
    usage. [GH-7753]
* provisioner/ansible: prevent nil pointer dereference after a language change
    [GH-7738]
* provisioner/chef: Accept chef license by default to prevent hangs in latest
    Chef [GH-7653]
* provisioner/powershell: Fix crash caused by error in retry logic check in
    powershell provisioner [GH-7657]
* provisioner/powershell: Fix null file descriptor error that occurred when
    remote_path provided is a directory and not a file. [GH-7705]

## 1.4.1 (May 15, 2019)

### IMPROVEMENTS:
* **new builder:** new proxmox builder implemented [GH-7490]
* **new builder:** new yandex cloud builder implemented [GH-7484]
* **new builder:** new linode builder implemented [GH-7508]
* build: Circle CI now generates test binaries for all pull requests [GH-7624]
    [GH-7625] [GH-7630]
* builder/alicloud: Support encryption with default service key [GH-7574]
* builder/amazon: Users of chroot and ebssurrogate builders may now choose
    between "x86_64" and "arm64" architectures when registering their AMIs.
    [GH-7620]
* builder/amazon: Users of the ebssurrogage builder may now choose to omit
    certain launch_block_devices from the final AMI mapping by using the
    omit_from_artifact feature. [GH-7612]
* builder/azure: Update Azure SDK [GH-7563]
* builder/docker: Better error messaging with container downloads. [GH-7513]
* builder/google: add image encryption support [GH-7551]
* builder/hyperv: Add keep_registered option to hyperv [GH-7498]
* builder/qemu: Replace dot-based parsing with hashicorp/go-version [GH-7614]
* builder/vmware: Add 30 minute timeout for destroying a VM [GH-7553]
* core: Cleanup cache of used port after closing [GH-7613]
* core: New option to set number of builds running in parallel & test
    BuildCommand more [GH-7501]
* packer compiles on s390x [GH-7567]
* provisioner/file: Added warnings about writeable locations [GH-7494]


### BUG FIXES:
* builder/amazon: Fix bug that always encrypted build region with default key.
    [GH-7507]
* builder/amazon: Fix bug that wasn't deleting unencrypted temporary snapshots
    [GH-7521]
* builder/amazon: Fix EBSsurrogate copy, encryption, and deletion of temporary
    unencrypted amis. [GH-7598]
* builder/hyperv: Fixes IP detection error if more than one VMNetworkAdapter is
    found [GH-7480]
* builder/qemu: Fix mistake switching ssh port mix/max for vnc port min/max
    [GH-7615]
* builder/vagrant: Fix bug with builder and vagrant-libvirt plugin [GH-7633]
* builder/virtualbox: Don't fail download when checksum is not set. [GH-7512]
* builder/virtualbox: Fix ovf download failures by using local ovf files in
    place instead of symlinking [GH-7497]
* builder/vmware: Fix panic configuring VNC for remote builds [GH-7509]
* core/build: Allow building Packer on solaris by removing progress bar and tty
    imports on solaris [GH-7618]
* core: Fix race condition causing hang [GH-7579]
* core: Fix tty related panics [GH-7517]
* core: Step download: Always copy local files on windows rather than
    symlinking them [GH-7575]
* packer compiles on Solaris again [GH-7589] [GH-7618]
* post-processor/vagrant: Fix bug in retry logic that caused failed upload to
    report success. [GH-7554]

## 1.4.0 (April 11, 2019)

### IMPROVEMENTS:
* builder/alicloud: Improve error message for conflicting images name [GH-7415]
* builder/amazon-chroot: Allow users to specify custom block device mapping
    [GH-7370]
* builder/ansible: Documentation fix explaining how to use ansible 2.7 + winrm
    [GH-7461]
* builder/azure-arm: specify zone resilient image from config [GH-7211]
* builder/docker: Add support for windows containers [GH-7444]
* builder/openstack: Allow both ports and networks in openstack builder
    [GH-7451]
* builder/openstack: Expose force_delete for openstack builder [GH-7395]
* builder/OpenStack: Support Application Credential Authentication [GH-7300]
* builder/virtualbox: Add validation for 'none' communicator. [GH-7419]
* builder/virtualbox: create ephemeral SSH key pair for build process [GH-7287]
* core: Add functionality to marshal a Template to valid Packer JSON [GH-7339]
* core: Allow user variables to be interpreted within the variables section
    [GH-7390]
* core: Incorporate the go-getter to handle downloads [GH-6999]
* core: Lock Packer VNC ports using a lock file to prevent collisions [GH-7422]
* core: Print VerifyChecksum log for the download as ui.Message output
    [GH-7387]
* core: Users can now set provisioner timeouts [GH-7466]
* core: Switch to using go mod for managing dependencies [GH-7270]
* core: Select a new VNC port if initial port is busy [GH-7423]
* post-processor/googlecompute-export: Set network project id to builder
    [GH-7359]
* post-processor/vagrant-cloud: support for the vagrant builder [GH-7397]
* post-processor/Vagrant: Option to ignore SSL verification when using on-
    premise vagrant cloud [GH-7377]
* postprocessor/amazon-import: Support S3 and AMI encryption. [GH-7396]
* provisioner/shell provisioner/windows-shell: allow to specify valid exit
    codes [GH-7385]
* core: Filter sensitive variables out of the ui as well as the logs
    [GH-7462]

### BUG FIXES:
* builder/alibaba: Update to latest Alibaba Cloud official image to fix
    acceptance tests [GH-7375]
* builder/amazon-chroot: Fix building PV images and where mount_partition is
    set [GH-7337]
* builder/amazon: Fix http_proxy env var regression [GH-7361]
* builder/azure: Fix: Power off before taking snapshot (windows) [GH-7464]
* builder/hcloud: Fix usage of freebsd64 rescue image [GH-7381]
* builder/vagrant: windows : fix docs and usage [GH-7416] [GH-7417]
* builder/vmware-esxi: properly copy .vmxf files in remote vmx builds [GH-7357]
* core: fix bug where Packer didn't pause in debug on certain linux platforms.
    [GH-7352]
* builder/amazon: Fix bug copying encrypted images between regions [GH-7342]

### BACKWARDS INCOMPATIBILITIES:
* builder/amazon: Change `temporary_security_group_source_cidr` to
    `temporary_security_group_source_cidrs` and allow it to accept a list of
    strings. [GH-7450]
* builder/amazon: If users do not pass any encrypt setting, retain any initial
    encryption setting of the AMI. [GH-6787]
* builder/docker: Update docker's default config to use /bin/sh instead of
    /bin/bash [GH-7106]
* builder/hyperv: Change option names cpu->cpus and ram_size->memory to bring
    naming in line with vmware and virtualbox builders [GH-7447]
* builder/oracle-classic: Remove default ssh_username from oracle classic
    builder, but add note to docs with oracle's default user. [GH-7446]
* builder/scaleway: Renamed attribute api_access_key to organization_id.
    [GH-6983]
* Change clean_image name and clean_ami_name to a more general clean_resource
    name for Googlecompute, Azure, and AWS builders. [GH-7456]
* core/post-processors: Change interface for post-processors to allow an
    overridable default for keeping input artifacts. [GH-7463]

## 1.3.5 (February 28, 2019)

### IMPROVEMENTS:
* builder/alicloud: Update aliyun sdk to support eu-west-1 region [GH-7338]
* builder/amazon: AWS users can now use the Vault AWS engine to generate
    temporary credentials. [GH-7282]
* builder/azure: IMDS to get subscription for Azure MSI [GH-7332]
* builder/openstack: Replaced deprecated compute/ api with imageservice/
    [GH-7038]
* builder/virtualbox: New "guest_additions_interface" option to enable
    attaching via a SATA interface. [GH-7298]
* builder/vmware: Add `cores` option for specifying the number of cores per
    socket. [GH-7191]
* bulder/openstac: Deprecated compute/v2/images API [GH-7268]
* core: Add validation check to help folks who swap their iso_path and
    checksum_path [GH-7311]
* fixer/amazon: Make the amazon-private-ip fixer errors more visible [GH-7336]
* post-processor/googlecompute-export: Extend auth for the GCE-post-processors
    to act like the GCE builder. [GH-7222]
* post-processor/googlecompute-import: Extend auth for the GCE-post-processors
    to act like the GCE builder. [GH-7222]
* post-processor/manifest: Add "custom_data" key to packer manifest post-
    processor [GH-7248]

### BUG FIXES:
* builder/amazon: Fix support for aws-us-gov [GH-7347]
* builder/amazon: Move snapshot deletion to cleanup phase. [GH-7343]
* builder/azure: Fixed Azure interactive authentication [GH-7276]
* builder/cloudstack: Updated sdk version; can now use ostype name in
    template_os option.  [GH-7264]
* builder/google: Change metadata url to use a FQDN fixing bug stemming from
    differing DNS/search domains. [GH-7260]
* builder/hyper-v: Fix integer overflows in 32-bit builds [GH-7251]
* builder/hyper-v: Fix regression where we improperly handled spaces in switch
    names [GH-7266]
* builder/openstack: Pass context So we know to cancel during WaitForImage
    [GH-7341]
* builder/vmware-esxi: Strip \r\n whitespace from end of names of
    files stored on esxi. [GH-7310]
* builder/vmware: Add "--noSSLVerify" to args in ovftool Validation [GH-7314]
* core: clean up Makefile [GH-7254][GH-7265]
* core: Fixes mismatches in checksums for dependencies for Go 1.11.4+ [GH-7261]
* core: make sure 'only' option is completely ignored by post-processors
    [GH-7262]
* core: name a post-processor to its type when it is not named [GH-7330]
* provisioner/salt: Force powershell to overwrite duplicate files [GH-7281]

### Features:
* **new builder** `vagrant` allows users to call vagrant to provision starting
    from vagrant boxes and save them as new vagrant boxes. [GH-7221]
* **new builder:** `hyperone` for building new images on HyperOne Platform on
    top of existing image or from the scratch with the use of chroot. [GH-7294]
* **new post-processor** `digitalocean-import`Add digitalocean-import post-
    processor. [GH-7060]
* **new provisioner**`inspec` Added inspec.io provisioner [GH-7180]
* communicator: Add configurable pause after communicator can connect but
    before it performs provisioning tasks [GH-7317] [GH-7351]

## 1.3.4 (January 30, 2019)
### IMPROVEMENTS:
* builder/alicloud: delete copied image and snapshots if corresponding options
    are specified [GH-7050]
* builder/amazon: allow to interpolate more variables [GH-7059]
* builder/amazon: Check that the KMS key ID is valid [GH-7090]
* builder/amazon: Clean up logging for aws waiters so that it only runs once
    per builder [GH-7080]
* builder/amazon: don't Cleanup Temp Keys when there is no communicator to
    avoid a panic [GH-7100] [GH-7095]
* builder/amazon: Don't try to guess region from metadata if not set & update
    aws-sdk-go [GH-7230]
* builder/amazon: Import error messages should now contain reason for failure
    [GH-7207]
* builder/azure: add certificate authentication [GH-7189]
* builder/azure: allow to configure disk caching [GH-7061]
* builder/azure: use deallocate instead of just power-off [GH-7203]
* builder/hyperv: Add support for legacy network adapters on Hyper-V. [GH-7128]
* builder/hyperv: Allow user to set `version` option in the New-VM command.
    [GH-7136]
* builder/openstack: Add `volume_size` option [GH-7130]
* builder/openstack: Don't require network v2 [GH-6933]
* builder/openstack: Support for tagging new images [GH-7037]
* builder/qemu: Add configuration options to specify cpu count and memory size
    [GH-7156]
* builder/qemu: Add support for whpx accelerator to qemu builder [GH-7151]
* builder/vmware: Escape query as suggested in issue #7200 [GH-7223]
* core/shell: Add env vars "PACKER_HTTP_IP" and "PACKER_HTTP_PORT" to shell
    provisioners [GH-7075]
* core: allow to use `-except` on post-processors [GH-7183]
* core: Clean up internal handling and creation of temporary directories
    [GH-7102]
* core: Deprecate mitchellh/go-homedir package in favor of os/user [GH-7062]
* core: Download checksum match failures will now log the received checksum.
    [GH-7210]
* core: Explicitly set ProxyFromEnvironment in httpclients when creating an aws
    session [GH-7226]
* core: make packer inspect not print sensitive variables [GH-7084]
* post-processor/google: Add new `guest-os-features` option. [GH-7218]
* postprocessor/docker-import: Added `change` support [GH-7127]
* provisioner/ansible-remote: add `-o IdentitiesOnly=yes` as a default flag
    [GH-7115]
* provisioner/chef-client: Elevated support for chef-client provisioner
    [GH-7078]
* provisioner/puppet: Elevated support for puppet-* provisioner [GH-7078]
* provisioner/windows-restart: wait for already-scheduled reboot [GH-7056] and
    ignore reboot specific errors [GH-7071]


### BUG FIXES:
* builder/azure: Ensure the Windows Guest Agent is fully functional before
    Sysprep is executed. [GH-7176]
* builder/azure: Fix snapshot regression [GH-7111]
* builder/docker: Ensure that entrypoint and arguments get passed to docker,
    not the image. [GH-7091]
* builder/hcloud: fix go mod dependency [GH-7099]
* builder/hcloud: prevent panic when ssh key was not passed [GH-7118]
* builder/hyperv: Fix the Hyper-V gen 1 guest boot order. [GH-7147]
* builder/hyperv: hyper-v builder no longer ignores `ssh_host` option.
    [GH-7154]
* builder/oracle-oci: Fix crash that occurs when image is nil [GH-7126]
* builder/parallels: Fix attaching prl tools [GH-7158]
* builder/virtualbox: Fix handling of portcount argument for version 6 beta
    [GH-7174] [GH-7094]
* builder/vmware: Fix bug caused by 'nil' dir field in artifact struct when
    building locally [GH-7116]
* communicator/docker: Fix docker file provisioner on Windows [GH-7163]
* core: prioritize AppData over default user directory ( UserProfile )
    [GH-7166]
* core: removed a flaky race condition in tests [GH-7119]
* postprocessor/vsphere: Stop setting HDDOrder, since it was breaking uploads
    [GH-7108]


## 1.3.3 (December 5, 2018)
### IMPROVEMENTS:
* builder/alicloud: Add options for system disk properties [GH-6939]
* builder/alicloud: Apply tags to relevant snapshots [GH-7040]
* builder/alicloud: Support creating image without data disks [GH-7022]
* builder/amazon: Add option for skipping TLS verification [GH-6842]
* builder/azure: Add options for Managed Image OS Disk and Data Disk snapshots
    [GH-6980]
* builder/hcloud: Add `snapshot_labels` option to hcloud builder [GH-7046]
* builder/hcloud: Add ssh_keys config to hcloud builder [GH-7028]
* builder/hcloud: Update hcloud-go version and support builds using rescue mode
    [GH-7034]
* builder/oracle: Parameterized volume size support for Oracle classic builder
    [GH-6918]
* builder/parallels: Add configuration options to parallels builder to specify
    cpu count and memory size [GH-7018]
* builder/virtualbox: Add configuration options to virtualbox builder to
    specify cpu count and memory size [GH-7017]
* builder/virtualbox: expose the VBoxManage export --iso option [GH-5950]
* builder/vmware: Add configuration options to vmware builder to specify cpu
    count and memory size [GH-7019]
* builder/vmware: Add new display_name template option [GH-6984]
* builder/vmware: Extend vmware-vmx builder to allow esxi builds. [GH-4591]
    [GH-6927]
* builder/vmware: Validate username/password for ovftool during prepare.
    [GH-6977]
* builder/vmware: Warn users if their vmx_data overrides data that Packer uses
    the template engine to set in its default vmx template. [GH-6987]
* communicator/ssh: Expand user path for SSH private key [GH-6946]
* core: Add a sed template engine [GH-6580]
* core: More explicit error message in rpc/ui.go [GH-6981]
* core: Replaced unsafe method of determining homedir with os/user
    implementation [GH-7036]
* core: Update vagrantfile's go version. [GH-6841]
* post-processor/amazon-import: Support ova, raw, vmdk, and vhdx formats in the
    amazon-import post-processor. [GH-6938]
* post-processor/vsphere-template: Add option to snapshot vm before marking as
    template [GH-6969]
* provisioner/breakpoint: Add a new breakpoint provisioner. [GH-7058]
* provisioner/powershell: Allow Powershell provisioner to use service accounts
    [GH-6972]
* provisioner/shell: Add PauseAfter option to shell provisioner [GH-6913]

### BUG FIXES:
* builder/amazon: Better error handling of region/credential guessing from
    metadata [GH-6931]
* builder/amazon: move region validation to run so that we don't break
    validation when no credentials are set [GH-7032]
* builder/hyperv: Remove -Copy:$false when calling Hyper-V\Compare-VM
    compatability report [GH-7030]
* builder/qemu: Do not set detect-zeroes option when we want it "off" [GH-7064]
* builder/vmware-esxi: Create export directories for vmx and ovf file types
    [GH-6985]
* builder/vmware: Correctly parse version for VMware Fusion Tech Preview
    [GH-7016]
* builder/vmware: Escape vSphere username when putting it into the export call
    [GH-6962]
* post-processor/vagrant: Add "hvf" as a libvirt driver [GH-6955]
* provisioner/ansible: inventory is no longer set to inventory_directory
    [GH-7065]

## 1.3.2 (October 29, 2018)
### IMPROVEMENTS:
* builder/alicloud: Add new `disable_stop_instance` option. [GH-6764]
* builder/alicloud: Support adding tags to image. [GH-6719]
* builder/alicloud: Support ssh with private ip address. [GH-6688]
* builder/amazon: Add support to explicitly control ENA support [GH-6872]
* builder/amazon: Add suppport for `vpc_filter`, `subnet_filter`, and
    `security_group_filter`. [GH-6374]
* builder/amazon: Add validation for required `device_name` parameter in
    `block_device_mappings`. [GH-6845]
* builder/amazon: Clean up security group wait code. [GH-6843]
* builder/amazon: Update aws-sdk-go to v1.15.54, adding support for
    `credential_source`. [GH-6849]
* builder/amazon: Use DescribeRegions for aws region validation. [GH-6512],
    [GH-6904]
* builder/azure: Add new `shared_image_gallery` option. [GH-6798]
* builder/googlecompute: Return an error if `startup_script_file` is specified,
    but file does not exist. [GH-6848]
* builder/hcloud: Add Hetzner Cloud builder. [GH-6871]
* builder/openstack: Add new `disk_format` option. [GH-6702]
* builder/openstack: Fix bug where `source_image_name` wasn't being used to
    properly find a UUID. [GH-6751]
* builder/openstack: Wait for volume availability when cleaning up [GH-6703]
* builder/qemu: Add `disk_detect_zeroes` option. [GH-6827]
* builder/scaleway: Add `boottype` parameter to config. [GH-6772]
* builder/scaleway: Update scaleway-cli vendor. [GH-6771]
* core: New option to add timestamps to UI output. [GH-6784]
* post-processor/vagrant-cloud: Validate vagrant cloud auth token doing an auth
    request [GH-6914]
* provisioner/file: Improve error messaging when file destination is a
    directory with no trailing slash. [GH-6756]
* provisioner/powershell: Provide better error when Packer can't find
    Powershell executable. [GH-6817]
* provisioner/shell-local: Add ability to specify OSs where shell-local can run
    [GH-6878]

### BUG FIXES:
* builder/alicloud: Fix ssh configuration pointer issues that could cause a bug
    [GH-6720]
* builder/alicloud: Fix type error in step_create_tags [GH-6763]
* builder/amazon: Error validating credentials is no longer obscured by a
    region validation error. and some region validation refactors and
    improvements [GH-6865]
* builder/amazon: Fix error calculating defaults in AWS waiters. [GH-6727]
* builder/amazon: Increase default wait for image import to one hour. [GH-6818]
* builder/amazon: Waiter now fails rather than hanging for extra time when an
    image import fails. [GH-6747]
* builder/azure: Updated Azure/go-ntlmssp dependency to resolve an issue with
    the winrm communicator not connecting to Windows machines requiring NTLMv2
    session security
* builder/digitalocean: Fix ssh configuration pointer issues that could cause a
    panic [GH-6729]
* builder/hyperv/vmcx: Allow to set generation from buildfile [GH-6909]
* builder/scaleway: Fix issues with ssh keys. [GH-6768]
* core: Fix error where logging was always enabled when Packer was run from
    inside Terraform. [GH-6758]
* core: Fix issue with with names containing spaces in ESX5Driver and in ssh
    communicator [GH-6891], [GH-6823]
* core: Fix logger so it doesn't accidentally try to format unescaped strings.
    [GH-6824]
* core: Fix race conditions in progress bar code [GH-6858], [GH-6788],
    [GH-6851]
* core: Fix various places in multiple builders where config was not being
    passed as a pointer. [GH-6739]
* post-processor/manifest: No longer provides an empty ID string for Azure's
    managed image artifact [GH-6822]
* provisioner/powershell: Fix a bug in the way we set the ProgressPreference
    variable in the default `execute_command` [GH-6838]
* provisioner/windows-restart: Fix extraneous break which forced early exit
    from our wait loop. [GH-6792]

## 1.3.1 (September 13, 2018)

### IMPROVEMENTS:
* builder/amazon: automatically decode encoded authorization messages if
    possible [GH-5415]
* builder:amazon: Optional cleanup of the authorized keys file [GH-6713]
* builder/qemu: Fixed bug where a -device in the qemuargs would override the default network settings, resulting in no network [GH-6807]

### BUG FIXES:
* builder/amazon: fix bugs relating to spot instances provisioning [GH-6697]
    [GH-6693]
* builder/openstack: fix ssh keypair not attached [GH-6701]
* core: progressbar: fix deadlock locking builds after first display [GH-6698]

## 1.3.0 (September 11, 2018)

### IMPROVEMENTS:
* azure/arm: Retry cleanup of individual resources on error [GH-6644]
* builder/alicloud: Support source image coming from marketplace [GH-6588]
* builder/amazon-chroot: Add new `root_volume_type` option. [GH-6669]
* builder/amazon-chroot: If you have a PV source AMI, with the Amazon Chroot
    builder, and the destination AMI is type HVM, you can now enable
    ena_support, example: [GH-6670]
* builder/amazon-chroot: New feature `root_volume_tags` to tag the created
    volumes. [GH-6504]
* builder/amazon: Create a random interim AMI name when encrypt_boot is true so
    that ami name is not searchable. [GH-6657]
* builder/azure: Implement clean_image_name template engine. [GH-6558]
* builder/cloudstack: Add option to use a fixed port via public_port. [GH-6532]
* builder/digitalocean: Add support for tagging to instances [GH-6546]
* builder/googlecompute: Add new `min_cpu_platform` feature [GH-6607]
* builder/googlecompute: Update the list of public image projects that we
    search, based on GCE documentation. [GH-6648]
* builder/lxc: Allow unplivileged LXC containers. [GH-6279]
* builder/oci: Add `metadata` feature to Packer config. [GH-6498]
* builder/openstack: Add support for getting config from clouds-public.yaml.
    [GH-6595]
* builder/openstack: Add support for ports. [GH-6570]
* builder/openstack: Add support for source_image_filter. [GH-6490]
* builder/openstack: Migrate floating IP usage to Network v2 API from Compute
    API. [GH-6373]
* builder/openstack: Support Block Storage volumes as boot volume. [GH-6596]
* builder/oracle-oci: Add support for freeform tagging of OCI images [GH-6338]
* builder/qemu: add ssh agent support. [GH-6541]
* builder/qemu: New `use_backing_file` feature [GH-6249]
* builder/vmware-iso: Add support for disk compaction [GH-6411]
* builder/vmware-iso: Try to use ISO files uploaded to the datastore when
    building remotely instead of uploading them freshly every time [GH-5165]
* command/validate: Warn users if config needs fixing. [GH-6423]
* core: Add a 'split' function to parse template variables. [GH-6357]
* core: Add a template function allowing users to read keys from consul
    [GH-6577]
* core: Add a template function allowing users to read keys from vault
    [GH-6533]
* core: Add progress-bar to download step. [GH-5851]
* core: Create a new root-level Packer template option, "sensitive-variables"
    which allows users to list which variables they would like to have scrubbed
    from the Packer logs. [GH-6610]
* core: Create new config options, "boot_keygroup_interval" and
    "boot_key_interval" that can be set at the builder-level to supercede
    PACKER_KEY_INTERVAL for the bootcommand. [GH-6616]
* core: Deduplicate ui and log lines that stream to terminal [GH-6611]
* core: Refactor and deduplicate ssh code across builders. This should be a no-
    op but is a big win for maintainability. [GH-6621] [GH-6613]
* post-processor/compress: Add support for xz compression [GH-6534]
* post-processor/vagrant: Support for Docker images. [GH-6494]
* post-processor/vsphere: Add new `esxi_host` option. [GH-5366]
* postprocessor/vagrant: Add support for Azure. [GH-6576]
* provisioner/ansible: Add new "extra var", packer_http_addr. [GH-6501]
* provisioner/ansible: Enable {{.WinRMPassword}} template engine. [GH-6450]
* provisioner/shell-local: Create PACKER_HTTP_ADDR environment variable
    [GH-6503]


### BUG FIXES:
* builder/amazon-ebssurrogate: Clean up volumes at end of build. [GH-6514]
* builder/amazon: Increase default waiter timeout for AWS
    WaitUntilImageAvailable command [GH-6601]
* builder/amazon: Increase the MaxRetries in the Amazon client from the default
    to 20, to work around users who regularly reach their requestlimit and are
    being throttled. [GH-6641]
* builder/amazon: Properly apply environment overrides to our custom-written
    waiters. [GH-6649]
* builder/azure: Generated password satisfies Azure password requirements
    [GH-6480]
* builder/hyper-v: Builder no longer errors if skip_compaction isn't true when
    skip_export is true, and compaction efficiency is improved [GH-6393]
* builder/lxc: Correctly pass "config" option to "lxc launch". [GH-6563]
* builder/lxc: Determine lxc root according to the running user [GH-6543]
* builder/lxc: Fix file copying for unprivileged LXC containers [GH-6544]
* builder/oracle-oci: Update OCI sdk, fixing validation bug that occurred when
    RSA key was encrypted. [GH-6492]
* builder/vmware-iso: Fix crash caused by invalid datacenter url. [GH-6529]
* builder/vmware: Maintain original boot order during CreateVMX step for
    vmware-iso builder [GH-6204]
* communicator/chroot: Fix quote escaping so that ansible provisioner works
    properly. [GH-6635]
* core: Better error handling in downloader when connection error occurs.
    [GH-6557]
* core: Fix broken pathing checks in checksum files. [GH-6525]
* provisioner/shell Create new template option allowing users to choose to
    source env vars from a file rather than declaring them inline. This
    resolves a bug that occurred when users had complex quoting in their
    `execute_command`s [GH-6636]
* provisioner/shell-local: Windows inline scripts now default to being appended
    with ".cmd", fixing a backwards incompatibility in v1.2.5 [GH-6626]
* provisioner/windows-restart: Provisioner now works when used in conjunction
    with SSH communicator [GH-6606]

### BACKWARDS INCOMPATIBILITIES:
* builder/amazon: "owners" field on source_ami_filter is now required for
    secuirty reasons. [GH-6585]
* builder/vmware-iso: validation will fail for templates using esxi that have the "disk_type_id" set to something other than "thin" or "" and that do not have "skip_compaction": true also set. Use `packer fix` to fix this. [GH-6411]

## 1.2.5 (July 16, 2018)

### BUG FIXES:
* builder/alickoud: Fix issue where internet_max_bandwidth_out template option
    was not being passed to builder. [GH-6416]
* builder/alicloud: Fix an issue with VPC cleanup. [GH-6418]
* builder/amazon-chroot: Fix communicator bug that broke chroot builds.
    [GH-6363]
* builder/amazon: Replace packer's waiters with those from the AWS sdk, solving
    several timeout bugs. [GH-6332]
* builder/azure: update azure-sdk-for-go, fixing 32-bit build errors. [GH-6479]
* builder/azure: update the max length of managed_image_resource_group to match
    new increased length of 90 characters. [GH-6477]
* builder/hyper-v: Fix secure boot template feature so that it properly passes
    the temolate for MicrosoftUEFICertificateAuthority. [GH-6415]
* builder/hyperv: Fix bug in HyperV IP lookups that was causing breakages in
    FreeBSD/OpenBSD builds. [GH-6416]
* builder/qemu: Fix error race condition in qemu builder that caused convert to
    fail on ubuntu 18.x [GH-6437]
* builder/qemu: vnc_bind_address was not being passed to qemu. [GH-6467]
* builder/virtualbox: Allow iso_url to be a symlink. [GH-6370]
* builder/vmware: Don't fail on DHCP lease files that cannot be read, fixing
    bug where builder failed on NAT networks that don't serve DHCP. [GH-6415]
* builder/vmware: Fix bug where we couldn't discover IP if vm_name differed
    from the vmx displayName. [GH-6448]
* builder/vmware: Fix validation to prevent hang when remopte_password is not
    sent but vmware is building on esxi. [GH-6424]
* builder/vmware:Correctly default the vm export format to ovf; this is what
    the docs claimed we already did, but we didn't. [GH-4538]
* communicator/winrm: Revert an attempt to determine whether remote upload
    destinations were files or directories, as this broke uploads on systems
    without Powershell installed. [GH-6481]
* core: Fix bug in parsing of iso checksum files that arose when setting
    iso_url to a relative filepath. [GH-6488]
* core: Fix Packer crash caused by improper error handling in the downloader.
    [GH-6381]
* fix: Fix bug where fixer for ssh_private_ip that failed when boolean values
    are passed as strings. [GH-6458]
* provisioner/powershell: Make upload of powershell variables retryable, in
    case of system restarts. [GH-6388]

### IMPROVEMENTS:
* builder/amazon: Add the ap-northeast-3 region. [GH-6385]
* builder/amazon: Spot requests may now have tags applied using the `spot_tags`
    option [GH-5452]
* builder/cloudstack: Add support for Projectid and new config option
    prevent_firewall_changes. [GH-6487]
* builder/openstack: Add support for token authorization and cloud.yaml.
    [GH-6362]
* builder/oracle-oci: Add new "instance_name" template option. [GH-6408]
* builder/scaleway: Add new "bootscript" parameter, allowing the user to not
    use the default local bootscript [GH-6439]
* builder/vmware: Add support for linked clones to vmware-vmx. [GH-6394]
* debug: The -debug flag will now cause Packer to pause between provisioner
    scripts in addition to Packer steps. [GH-4663]
* post-processor/googlecompute-import: Added new googlecompute-import post-
    processor [GH-6451]
* provisioner/ansible: Add new "playbook_files" option to execute multiple
    playbooks within one provisioner call. [GH-5086]

## 1.2.4 (May 29, 2018)

### BUG FIXES:

* builder/amazon: Can now force the chroot builder to mount an entire block
    device instead of a partition [GH-6194]
* builder/azure: windows-sql-cloud is now in the default list of projects to
    check for provided images. [GH-6210]
* builder/chroot: A new template option, `nvme_device_path` has been added to
    provide a workaround for users who need the amazon-chroot builder to mount
    a NVMe volume on their instances. [GH-6295]
* builder/hyper-v: Fix command for mounting multiple disks [GH-6267]
* builder/hyperv: Enable IP retrieval for Server 2008 R2 hosts. [GH-6219]
* builder/hyperv: Fix bug in MAC address specification on Hyper-V. [GH-6187]
* builder/parallels-pvm: Add missing disk compaction step. [GH-6202]
* builder/vmware-esxi: Remove floppy files from the remote server on cleanup.
    [GH-6206]
* communicator/winrm: Updated dependencies to fix a race condition [GH-6261]
* core: When using `-on-error=[abort|ask]`, output the error to the user.
    [GH-6252]
* provisioner/puppet: Extra-Arguments are no longer prematurely
    interpolated. [GH-6215]
* provisioner/shell: Remove file stat that was causing problems uploading files
    [GH-6239]

### IMPROVEMENTS:

* builder/amazon: Amazon builders other than `chroot` now support T2 unlimited
    instances [GH-6265]
* builder/azure: Allow device login for US government cloud. [GH-6105]
* builder/azure: Devicelogin Support for Windows [GH-6285]
* builder/azure: Enable simultaneous builds within one resource group.
    [GH-6231]
* builder/azure: Faster deletion of Azure Resource Groups. [GH-6269]
* builder/azure: Updated Azure SDK to v15.0.0 [GH-6224]
* builder/hyper-v: Hyper-V builds now connect to vnc display by default when
    building [GH-6243]
* builder/hyper-v: New `use_fixed_vhd_format` allows vm export in an Azure-
    compatible format [GH-6101]
* builder/hyperv: New config option for specifying what secure boot template to
    use, allowing secure boot of linux vms. [GH-5883]
* builder/qemu: Add support for hvf accelerator. [GH-6193]
* builder/scaleway: Fix SSH communicator connection issue. [GH-6238]
* core: Add opt-in Packer top-level command autocomplete [GH-5454]
* post-processor/shell-local: New options have been added to create feature
    parity with the shell-local provisioner. This feature now works on Windows
    hosts. [GH-5956]
* provisioner/chef: New config option allows user to skip cleanup of chef
    client staging directory. [GH-4300]
* provisioner/shell-local: Can now access automatically-generated WinRM
    password as variable [GH-6251]
* provisoner/shell-local: New options have been added to create feature parity
    with the shell-local post-processor. This feature now works on Windows
    hosts. [GH-5956]
* builder/virtualbox: Use HTTPS to download guest editions, now that it's
    available. [GH-6406]

## 1.2.3 (April 25, 2018)

### BUG FIXES:

* builder/azure: Azure CLI may now be logged into several accounts. [GH-6087]
* builder/ebssurrogate: Snapshot all launch devices. [GH-6056]
* builder/hyper-v: Fix CopyExportedVirtualMachine script so it works with
    links. [GH-6082]
* builder/oracle-classic: Fix panics when cleaning up resources that haven't
    been created. [GH-6095]
* builder/parallels: Allow user to cancel build while the OS is starting up.
    [GH-6166]
* builder/qemu: Avoid warning when using raw format. [GH-6080]
* builder/scaleway: Fix compilation issues on solaris/amd64. [GH-6069]
* builder/virtualbox: Fix broken scancodes in boot_command. [GH-6067]
* builder/vmware-iso: Fail in validation if user gives wrong remote_type value.
    [GH-4563]
* builder/vmware: Fixed a case-sensitivity issue when determining the network
    type during the cloning step in the vmware-vmx builder. [GH-6057]
* builder/vmware: Fixes the DHCP lease and configuration pathfinders for VMware
    Player. [GH-6096]
* builder/vmware: Multi-disk VM's can be properly handled by the compacting
    stage. [GH-6074]
* common/bootcommand: Fix numerous bugs in the boot command code, and make
    supported features consistent across builders. [GH-6129]
* communicator/ssh: Stop trying to discover whether destination is a directory
    from uploader. [GH-6124]
* post-processor/vagrant: Large VMDKs should no longer show a 0-byte size on OS
    X. [GH-6084]
* post-processor/vsphere: Fix encoding of spaces in passwords for upload.
    [GH-6110]
* provisioner/ansible: Pass the inventory_directory configuration option to
    ansible -i when it is set. [GH-6065]
* provisioner/powershell: fix bug with SSH communicator + cygwin. [GH-6160]
* provisioner/powershell: The {{.WinRMPassword}} template variable now works
    with parallel builders. [GH-6144]

### IMPROVEMENTS:

* builder/alicloud: Update aliyungo common package. [GH-6157]
* builder/amazon: Expose more source ami data as template variables. [GH-6088]
* builder/amazon: Setting `force_delete` will only delete AMIs owned by the
    user. This should prevent failures where we try to delete an AMI with a
    matching name, but owned by someone else. [GH-6111]
* builder/azure: Users of Powershell provisioner may access the randomly-
    generated winrm password using the template variable {{.WinRMPassword}}.
    [GH-6113]
* builder/google: Users of Powershell provisioner may access the randomly-
    generated winrm password using the template variable {{.WinRMPassword}}.
    [GH-6141]
* builder/hyper-v: User can now configure hyper-v disk block size. [GH-5941]
* builder/openstack: Add configuration option for `instance_name`. [GH-6041]
* builder/oracle-classic: Better validation of destination image name.
    [GH-6089]
* builder/oracle-oci: New config options for user data and user data file.
    [GH-6079]
* builder/oracle-oci: use the official OCI sdk instead of handcrafted client.
    [GH-6142]
* builder/triton: Add support to Skip TLS Verification of Triton Certificate.
    [GH-6039]
* provisioner/ansible: Ansible users may provide a custom inventory file.
    [GH-6107]
* provisioner/file: New `generated` tag allows users to upload files created
    during Packer run. [GH-3891]

## 1.2.2 (March 26, 2018)

### BUG FIXES:

* builder/amazon: Fix AWS credential defaulting [GH-6019]
* builder/LXC: make sleep timeout easily configurable [GH-6038]
* builder/virtualbox: Correctly send multi-byte scancodes when typing boot
    command. [GH-5987]
* builder/virtualbox: Special boot-commands no longer overwrite previous
    commands [GH-6002]
* builder/vmware: Default to disabling XHCI bus for USB on the vmware-iso
    builder. [GH-5975]
* builder/vmware: Handle multiple devices per VMware network type [GH-5985]
* communicator/ssh: Handle errors uploading files more gracefully [GH-6033]
* provisioner/powershell: Fix environment variable file escaping. [GH-5973]


### IMPROVEMENTS:

* builder/amazon: Added new region `cn-northwest-1`. [GH-5960]
* builder/amazon: Users may now access the amazon-generated administrator
    password [GH-5998]
* builder/azure: Add support concurrent deployments in the same resource group.
    [GH-6005]
* builder/azure: Add support for building with additional disks. [GH-5944]
* builder/azure: Add support for marketplace plan information. [GH-5970]
* builder/azure: Make all command output human readable. [GH-5967]
* builder/azure: Respect `-force` for managed image deletion. [GH-6003]
* builder/google: Add option to specify a service account, or to run without
    one. [GH-5991] [GH-5928]
* builder/oracle-oci: Add new "use_private_ip" option. [GH-5893]
* post-processor/vagrant: Add LXC support. [GH-5980]
* provisioner/salt-masterless: Added Windows support. [GH-5702]
* provisioner/salt: Add windows support to salt provisioner [GH-6012] [GH-6012]


## 1.2.1 (February 23, 2018)

### BUG FIXES:

* builder/amazon: Fix authorization using assume role. [GH-5914]
* builder/hyper-v: Fix command collisions with VMWare PowerCLI. [GH-5861]
* builder/vmware-iso: Fix panic when building on esx5 remotes. [GH-5931]
* builder/vmware: Fix issue detecting host IP. [GH-5898] [GH-5900]
* provisioner/ansible-local: Fix conflicting escaping schemes for vars provided
    via `--extra-vars`. [GH-5888]

### IMPROVEMENTS:

* builder/oracle-classic: Add `snapshot_timeout` option to control how long we
    wait for the snapshot to be created. [GH-5932]
* builder/oracle-classic: Add support for WinRM connections. [GH-5929]


## 1.2.0 (February 9, 2018)

### BACKWARDS INCOMPATIBILITIES:

* 3rd party plugins: We have moved internal dependencies, meaning your 3rd
    party plugins will no longer compile (however existing builds will still
    work fine); the work to fix them is minimal and documented in GH-5810.
    [GH-5810]
* builder/amazon: The `ssh_private_ip` option has been removed. Instead, please
    use `"ssh_interface": "private"`. A fixer has been written for this, which
    can be invoked with `packer fix`. [GH-5876]
* builder/openstack: Extension support has been removed. To use OpenStack
    builder with the OpenStack Newton (Oct 2016) or earlier, we recommend you
    use Packer v1.1.2 or earlier version.
* core: Affects Windows guests: User variables containing Powershell special
    characters no longer need to be escaped. [GH-5376]
* provisioner/file: We've made destination semantics more consistent across the
    various communicators. In general, if the destination is a directory, files
    will be uploaded into the directory instead of failing. This mirrors the
    behavior of `rsync`. There's a chance some users might be depending on the
    previous buggy behavior, so it's worth ensuring your configuration is
    correct. [GH-5426]
* provisioner/powershell: Regression from v1.1.1 forcing extra escaping of
    environment variables in the non-elevated provisioner has been fixed.
    [GH-5515] [GH-5872]

### IMPROVEMENTS:

* **New builder:** `ncloud` for building server images using the NAVER Cloud
    Platform. [GH-5791]
* **New builder:** `oci-classic` for building new custom images for use with
    Oracle Cloud Infrastructure Classic Compute. [GH-5819]
* **New builder:** `scaleway` - The Scaleway Packer builder is able to create
    new images for use with Scaleway BareMetal and Virtual cloud server.
    [GH-4770]
* builder/amazon: Add `kms_key_id` option to block device mappings. [GH-5774]
* builder/amazon: Add `skip_metadata_api_check` option to skip consulting the
    amazon metadata service. [GH-5764]
* builder/amazon: Add Paris region (eu-west-3) [GH-5718]
* builder/amazon: Give better error messages if we have trouble during
    authentication. [GH-5764]
* builder/amazon: Remove Session Token (STS) from being shown in the log.
    [GH-5665]
* builder/amazon: Replace `InstanceStatusOK` check with `InstanceReady`. This
    reduces build times universally while still working for all instance types.
    [GH-5678]
* builder/amazon: Report which authentication provider we're using. [GH-5764]
* builder/amazon: Timeout early if metadata service can't be reached. [GH-5764]
* builder/amazon: Warn during prepare if we didn't get both an access key and a
    secret key when we were expecting one. [GH-5762]
* builder/azure: Add validation for incorrect VHD URLs [GH-5695]
* builder/docker: Remove credentials from being shown in the log. [GH-5666]
* builder/google: Support specifying licenses for images. [GH-5842]
* builder/hyper-v: Allow MAC address specification. [GH-5709]
* builder/hyper-v: New option to use differential disks and Inline disk
    creation to improve build time and reduce disk usage [GH-5631]
* builder/qemu: Add Intel HAXM support to QEMU builder [GH-5738]
* builder/triton: Triton RBAC is now supported. [GH-5741]
* builder/triton: Updated triton-go dependencies, allowing better error
    handling. [GH-5795]
* builder/vmware-iso: Add support for cdrom and disk adapter types. [GH-3417]
* builder/vmware-iso: Add support for setting network type and network adapter
    type. [GH-3417]
* builder/vmware-iso: Add support for usb/serial/parallel ports. [GH-3417]
* builder/vmware-iso: Add support for virtual soundcards. [GH-3417]
* builder/vmware-iso: More reliably retrieve the guest networking
    configuration. [GH-3417]
* builder/vmware: Add support for "super" key in `boot_command`. [GH-5681]
* communicator/ssh: Add session-level keep-alives [GH-5830]
* communicator/ssh: Detect dead connections. [GH-4709]
* core: Gracefully clean up resources on SIGTERM. [GH-5318]
* core: Improved error logging in floppy file handling. [GH-5802]
* core: Improved support for downloading and validating a uri containing a
    Windows UNC path or a relative file:// scheme. [GH-2906]
* post-processor/amazon-import: Allow user to specify role name in amazon-
    import [GH-5817]
* post-processor/docker: Remove credentials from being shown in the log.
    [GH-5666]
* post-processor/google-export: Synchronize credential semantics with the
    Google builder. [GH-4148]
* post-processor/vagrant: Add vagrant post-processor support for Google
    [GH-5732]
* post-processor/vsphere-template: Now accepts artifacts from the vSphere post-
    processor. [GH-5380]
* provisioner/amazon: Use Amazon SDK's InstanceRunning waiter instead of
    InstanceStatusOK waiter [GH-5773]
* provisioner/ansible: Improve user retrieval. [GH-5758]
* provisioner/chef: Add support for 'trusted_certs_dir' chef-client
    configuration option [GH-5790]
* provisioner/chef: Added Policyfile support to chef-client provisioner.
    [GH-5831]

### BUG FIXES:

* builder/alicloud-ecs: Attach keypair before starting instance in alicloud
    builder [GH-5739]
* builder/amazon: Fix tagging support when building in us-gov/china. [GH-5841]
* builder/amazon: NewSession now inherits MaxRetries and other settings.
    [GH-5719]
* builder/virtualbox: Fix interpolation ordering so that edge cases around
    guest_additions_url are handled correctly [GH-5757]
* builder/virtualbox: Fix regression affecting users running Packer on a
    Windows host that kept Packer from finding Virtualbox guest additions if
    Packer ran on a different drive from the one where the guest additions were
    stored. [GH-5761]
* builder/vmware: Fix case where artifacts might not be cleaned up correctly.
    [GH-5835]
* builder/vmware: Fixed file handle leak that may have caused race conditions
    in vmware builder [GH-5767]
* communicator/ssh: Add deadline to SSH connection to prevent Packer hangs
    after script provisioner reboots vm [GH-4684]
* communicator/winrm: Fix issue copying empty directories. [GH-5763]
* provisioner/ansible-local: Fix support for `--extra-vars` in
    `extra_arguments`. [GH-5703]
* provisioner/ansible-remote: Fixes an error where Packer's private key can be
    overridden by inherited `ansible_ssh_private_key` options. [GH-5869]
* provisioner/ansible: The "default extra variables" feature added in Packer
    v1.0.1 caused the ansible-local provisioner to fail when an --extra-vars
    argument was specified in the extra_arguments configuration option; this
    has been fixed. [GH-5335]
* provisioner/powershell: Regression from v1.1.1 forcing extra escaping of
    environment variables in the non-elevated provisioner has been fixed.
    [GH-5515] [GH-5872]


## 1.1.3 (December 8, 2017)

### IMPROVEMENTS:

* builder/alicloud-ecs: Add security token support and set TLS handshake
    timeout through environment variable. [GH-5641]
* builder/amazon: Add a new parameter `ssh_interface`. Valid values include
    `public_ip`, `private_ip`, `public_dns` or `private_dns`. [GH-5630]
* builder/azure: Add sanity checks for resource group names [GH-5599]
* builder/azure: Allow users to specify an existing resource group to use,
    instead of creating a new one for every run. [GH-5548]
* builder/hyper-v: Add support for differencing disk. [GH-5458]
* builder/vmware-iso: Improve logging of network errors. [GH-5456]
* core: Add new `packer_version` template engine. [GH-5619]
* core: Improve logic checking for downloaded ISOs in case where user has
    provided more than one URL in `iso_urls` [GH-5632]
* provisioner/ansible-local: Add ability to clean staging directory. [GH-5618]

### BUG FIXES:

* builder/amazon: Allow `region` to appear in `ami_regions`. [GH-5660]
* builder/amazon: `C5` instance types now build more reliably. [GH-5678]
* builder/amazon: Correctly set AWS region if given in template along with a
    profile. [GH-5676]
* builder/amazon: Prevent `sriov_support` and `ena_support` from being used
    with spot instances, which would cause a build failure. [GH-5679]
* builder/hyper-v: Fix interpolation context for user variables in
    `boot_command` [GH-5547]
* builder/qemu: Set default disk size to 40960 MB to prevent boot failures.
    [GH-5588]
* builder/vmware: Correctly detect Windows boot on vmware workstation.
    [GH-5672]
* core: Fix windows path regression when downloading ISOs. [GH-5591]
* provisioner/chef: Fix chef installs on Windows. [GH-5649]

## 1.1.2 (November 15, 2017)

### IMPROVEMENTS:

* builder/amazon: Correctly deregister AMIs when `force_deregister` is set.
    [GH-5525]
* builder/digitalocean: Add `ipv6` option to enable on droplet. [GH-5534]
* builder/docker: Add `aws_profile` option to control the aws profile for ECR.
    [GH-5470]
* builder/google: Add `clean_image_name` template engine. [GH-5463]
* builder/google: Allow selecting container optimized images. [GH-5576]
* builder/google: Interpolate network and subnetwork values, rather than
    relying on an API call that packer may not have permission for. [GH-5343]
* builder/hyper-v: Add `disk_additional_size` option to allow for up to 64
    additional disks. [GH-5491]
* builder/hyper-v: Also disable automatic checkpoints for gen 2 VMs. [GH-5517]
* builder/lxc: Add new `publish_properties` field to set image properties.
    [GH-5475]
* builder/lxc: Add three new configuration option categories to LXC builder:
    `create_options`, `start_options`, and `attach_options`. [GH-5530]
* builder/triton: Add `source_machine_image_filter` option to select an image
    ID based on a variety of parameters. [GH-5538]
* builder/virtualbox-ovf: Error during prepare if source path doesn't exist.
    [GH-5573]
* builder/virtualbox-ovf: Retry while removing VM to solve for transient
    errors. [GH-5512]
* communicator/ssh: Add socks 5 proxy support. [GH-5439]
* core/iso_config: Support relative paths in checksum file. [GH-5578]
* core: Rewrite vagrantfile code to make cross-platform development easier.
    [GH-5539]
* post-processor/docker-push: Add `aws_profile` option to control the aws
    profile for ECR. [GH-5470]
* post-processor/vsphere: Properly capture `ovftool` output. [GH-5499]

### BUG FIXES:

* builder/amazon: Add a delay option to security group waiter. [GH-5536]
* builder/amazon: Fix regressions relating to spot instances and EBS volumes.
    [GH-5495]
* builder/amazon: Set region from profile, if profile is set, rather than being
    overridden by metadata. [GH-5562]
* builder/docker: Remove `login_email`, which no longer exists in the docker
    client. [GH-5511]
* builder/hyperv: Fix admin check that was causing powershell failures.
    [GH-5510]
* builder/oracle: Defaulting of OCI builder region will first check the packer
    template and the OCI config file. [GH-5407]
* builder/triton: Fix a bug where partially created images can be reported as
    complete. [GH-5566]
* post-processor/vsphere: Use the vm disk path information to re-create the vmx
    datastore path. [GH-5567]
* provisioner/windows-restart: Wait for restart no longer endlessly loops if
    user specifies a custom restart check command. [GH-5563]

## 1.1.1 (October 13, 2017)

### IMPROVEMENTS:

* **New builder:** `hyperv-vmcx` for building images from existing VMs.
    [GH-4944] [GH-5444]
* builder/amazon-instance: Add `.Token` as a variable in the
    `BundleUploadCommand` template. [GH-5288]
* builder/amazon: Add `temporary_security_group_source_cidr` option to control
    ingress to source instances. [GH-5384]
* builder/amazon: Output AMI Name during prevalidation. [GH-5389]
* builder/amazon: Support template functions in tag keys. [GH-5381]
* builder/amazon: Tag volumes on creation instead of as a separate step.
    [GH-5417]
* builder/docker: Add option to set `--user` flag when running `exec`.
    [GH-5406]
* builder/docker: Set file owner to container user when uploading. Can be
    disabled by setting `fix_upload_owner` to `false`. [GH-5422]
* builder/googlecompute: Support setting labels on the resulting image.
    [GH-5356]
* builder/hyper-v: Add `vhd_temp_path` option to control where the VHD resides
    while it's being provisioned. [GH-5206]
* builder/hyper-v: Allow vhd or vhdx source images instead of just ISO.
    [GH-4944] [GH-5444]
* builder/hyper-v: Disable automatic checkpoints. [GH-5374]
* builder/virtualbox-ovf: Add `keep_registered` option. [GH-5336]
* builder/vmware: Add `disable_vnc` option to prevent VNC connections from
    being made. [GH-5436]
* core: Releases will now be built for ppc64le.
* post-processor/vagrant: When building from a builder/hyper-v artifact, link
    instead of copy when available. [GH-5207]


### BUG FIXES:

* builder/cloudstack: Fix panic if build is aborted. [GH-5388]
* builder/hyper-v: Respect `enable_dynamic_memory` flag. [GH-5363]
* builder/puppet-masterless: Make sure directories created with sudo are
    writable by the packer user. [GH-5351]
* provisioner/chef-solo: Fix issue installing chef-solo on Windows. [GH-5357]
* provisioner/powershell: Fix issue setting environment variables by writing
    them to a file, instead of the command line. [GH-5345]
* provisioner/powershell: Fix issue where powershell scripts could hang.
    [GH-5082]
* provisioner/powershell: Fix Powershell progress stream leak to stderr for
    normal and elevated commands. [GH-5365]
* provisioner/puppet-masterless: Fix bug where `puppet_bin_dir` wasn't being
    respected. [GH-5340]
* provisioner/puppet: Fix setting facter vars on Windows. [GH-5341]


## 1.1.0 (September 12, 2017)

### IMPROVEMENTS:

* builder/alicloud: Update alicloud go sdk and enable multi sites support for
    alicloud [GH-5219]
* builder/amazon: Upgrade aws-sdk-go to 1.10.14, add tags at instance run time.
    [GH-5196]
* builder/azure: Add object_id to windows_custom_image.json. [GH-5285]
* builder/azure: Add support for storage account for managed images. [GH-5244]
* builder/azure: Update pkcs12 package. [GH-5301]
* builder/cloudstack: Add support for Security Groups. [GH-5175]
* builder/docker: Uploading files and directories now use docker cp. [GH-5273]
    [GH-5333]
* builder/googlecompute: Add `labels` option for labeling launched instances.
    [GH-5308]
* builder/googlecompute: Add support for accelerator api. [GH-5137]
* builder/profitbricks: added support for Cloud API v4. [GH-5233]
* builder/vmware-esxi: Remote builds now respect `output_directory` [GH-4592]
* builder/vmware: Set artifact ID to `VMName`. [GH-5187]
* core: Build solaris binary by default. [GH-5268] [GH-5248]
* core: Remove LGPL dependencies. [GH-5262]
* provisioner/puppet: Add `guest_os_type` option to add support for Windows.
    [GH-5252]
* provisioner/salt-masterless: Also use sudo to clean up if we used sudo to
    install. [GH-5240]

### BACKWARDS INCOMPATIBILITIES:

* builder/amazon: Changes way that AMI artifacts are printed out after build,
    aligning them to builder. Could affect output parsing. [GH-5281]
* builder/amazon: Split `enhanced_networking` into `sriov_support` and
    `ena_support` to support finer grained control. Use `packer fix
    <template.json>` to automatically update your template to use `ena_support`
    where previously there was only `enhanced_networking`. Make sure to also
    add `sriov_support` if you need that feature, and to ensure `ena_support`
    is what you intended to be in your template. [GH-5284]
* builder/cloudstack: Setup temporary SSH keypair; backwards incompatible in
    the uncommon case that the source image allowed SSH auth with password but
    not with keypair. [GH-5174]
* communicator/ssh: Renamed `ssh_disable_agent` to
    `ssh_disable_agent_forwarding`. Need to run fixer on packer configs that
    use `ssh_disable_agent`. [GH-5024]
* communicator: Preserve left-sided white-space in remote command output. Make
    sure any scripts that parse this output can handle the new whitespace
    before upgrading. [GH-5167]
* provisioner/shell: Set default for `ExpectDisconnect` to `false`. If your
    script causes the connection to be reset, you should set this to `true` to
    prevent errors. [GH-5283]

### BUG FIXES:

* builder/amazon: `force_deregister` works in all regions, not just original
    region. [GH-5250]
* builder/docker: Directory uploads now use the same semantics as the rest of
    the communicators. [GH-5333]
* builder/vmware: Fix timestamp in default VMName. [GH-5274]
* builder/winrm: WinRM now waits to make sure commands can run successfully
    before considering itself connected. [GH-5300]
* core: Fix issue where some builders wouldn't respect `-on-error` behavior.
    [GH-5297]
* provisioner/windows-restart: The first powershell provisioner after a restart
    now works. [GH-5272]

### FEATURES:

* **New builder**: Oracle Cloud Infrastructure (OCI) builder for creating
    custom images. [GH-4554]
* **New builder:** `lxc` for building lxc images. [GH-3523]
* **New builder:** `lxd` for building lxd images. [GH-3625]
* **New post-processor**: vSphere Template post-processor to be used with
    vmware-iso builder enabling user to mark a VM as a template. [GH-5114]

## 1.0.4 (August 11, 2017)

### IMPROVEMENTS:

* builder/alicloud: Increase polling timeout. [GH-5148]
* builder/azure: Add `private_virtual_network_with_public_ip` option to
    optionally obtain a public IP. [GH-5222]
* builder/googlecompute: use a more portable method of obtaining zone.
    [GH-5192]
* builder/hyperv: Properly interpolate user variables in template. [GH-5184]
* builder/parallels: Remove soon to be removed --vmtype flag in createvm.
    [GH-5172]
* contrib: add json files to zsh completion. [GH-5195]

### BUG FIXES:

* builder/amazon: Don't delete snapshots we didn't create. [GH-5211]
* builder/amazon: fix builds when using the null communicator. [GH-5217]
* builder/docker: Correctly handle case when uploading an empty directory.
    [GH-5234]
* command/push: Don't push variables if they are unspecified. Reverts to
    behavior in 1.0.1. [GH-5235]
* command/push: fix handling of symlinks. [GH-5226]
* core: Strip query parameters from ISO URLs when checking against a checksum
    file. [GH-5181]
* provisioner/ansible-remote: Fix issue where packer could hang communicating
    with ansible-remote. [GH-5146]

## 1.0.3 (July 17, 2017)

### IMPROVEMENTS:
* builder/azure: Update to latest Azure SDK, enabling support for managed
    disks. [GH-4511]
* builder/cloudstack: Add default cidr_list [ 0.0.0.0/0 ]. [GH-5125]
* builder/cloudstack: Add support for ssh_agent_auth. [GH-5130]
* builder/cloudstack: Add support for using a HTTP server. [GH-5017]
* builder/cloudstack: Allow reading api_url, api_key, and secret_key from env
    vars. [GH-5124]
* builder/cloudstack: Make expunge optional and improve logging output.
    [GH-5099]
* builder/googlecompute: Allow using URL's for network and subnetwork.
    [GH-5035]
* builder/hyperv: Add support for floppy_dirs with hyperv-iso builder.
* builder/hyperv: Add support for override of system %temp% path.
* core: Experimental Android ARM support. [GH-5111]
* post-processor/atlas: Disallow packer push of vagrant.box artifacts to atlas.
    [GH-4780]
* postprocessor/atlas: Disallow pushing vagrant.box artifacts now that Vagrant
    cloud is live. [GH-4780]

### BUG FIXES:
* builder/amazon: Fix panic that happens if ami_block_device_mappings is empty.
    [GH-5059]
* builder/azure: Write private SSH to file in debug mode. [GH-5070] [GH-5074]
* builder/cloudstack: Properly report back errors. [GH-5103] [GH-5123]
* builder/docker: Fix windows filepath in docker-toolbox call [GH-4887]
* builder/docker: Fix windows filepath in docker-toolbox call. [GH-4887]
* builder/hyperv: Use SID to verify membership in Admin group, fixing for non-
    english users. [GH-5022]
* builder/hyperv: Verify membership in the group Hyper-V Administrators by SID
    not name. [GH-5022]
* builder/openstack: Update gophercloud version, fixing builds  > 1 hr long.
    [GH-5046]
* builder/parallels: Skip missing paths when looking for unnecessary files.
    [GH-5058]
* builder/vmware-esxi: Fix VNC port discovery default timeout. [GH-5051]
* communicator/ssh: Add ProvisionerTypes to communicator tests, resolving panic
    [GH-5116]
* communicator/ssh: Resolve race condition that sometimes truncates ssh
    provisioner stdout [GH-4719]
* post-processor/checksum: Fix interpolation of "output". [GH-5112]
* push: Push vars in packer config, not just those set from command line and in
    var-file. [GH-5101]

## 1.0.2 (June 21, 2017)

### BUG FIXES:
* communicator/ssh: Fix truncated stdout from remote ssh provisioner. [GH-5050]
* builder/amazon: Fix bugs related to stop instance command. [GH-4719]
* communicator/ssh: Fix ssh connection errors. [GH-5038]
* core: Remove logging that shouldn't be there when running commands. [GH-5042]
* provisioner/shell: Fix bug where scripts were being run under `sh`. [GH-5043]

### IMPROVEMENTS:

* provisioner/windows-restart: make it clear that timeouts come from the
    provisioner, not winrm. [GH-5040]

## 1.0.1 (June 19, 2017)

### IMPROVEMENTS:

* builder/amazon: Allow amis to be copied to other regions, encrypted with
    custom KMS keys. [GH-4948]
* builder/amazon: Allow configuration of api endpoint to support api-compatible
    cloud providers. [GH-4896]
* builder/amazon: Fix regex used for ami name validation [GH-4902]
* builder/amazon: Look up vpc from subnet id if no vpc was specified. [GH-4879]
* builder/amazon: Print temporary security group name to the UI. [GH-4997]
* builder/amazon: Support Assume Role with MFA and ECS Task Roles. Also updates
    to a newer version of aws-sdk-go. [GH-4996]
* builder/amazon: Use retry logic when creating instance tags. [GH-4876]
* builder/amazon: Validate ami name. [GH-4762]
* builder/azure: Add build output to artifact. [GH-4953]
* builder/azure: Use disk URI as artifact ID. [GH-4981]
* builder/digitalocean: Added support for monitoring. [GH-4782]
* builder/digitalocean: Support for copying snapshot to other regions.
    [GH-4893]
* builder/hyper-v: Remove the check for administrator rights when sending key
    strokes to Hyper-V. [GH-4687] # builder/openstack: Fix private key error
    message to match documentation [GH-4898]
* builder/null: Support SSH agent auth [GH-4956]
* builder/openstack: Add ssh agent support. [GH-4655]
* builder/openstack: Support client x509 certificates. [GH-4921]
* builder/parallels-iso: Configuration of disk type, plain or expanding.
    [GH-4621]
* builder/triton: An SSH agent can be used to authenticate requests, making
    `triton_key_material` optional. [GH-4838]
* builder/triton: If no source machine networks are specified, instances are
    started on the default public and internal networks. [GH-4838]
* builder/virtualbox: Add sata port count configuration option. [GH-4699]
* builder/virtualbox: Don't add port forwarding when using "none" communicator.
    [GH-4960]
* builder/vmware: Add option to remove interfaces from the vmx. [GH-4927]
* builder/vmware: Properly remove mounted CDs on OS X. [GH-4810]
* builder/vmware: VNC probe timeout is configurable. [GH-4919]
* command/push: add `-sensitive` flag to mark pushed vars are sensitive.
    [GH-4970]
* command/push: Vagrant support in Terraform Enterprise is deprecated.
    [GH-4950]
* communicator/ssh: Add ssh agent support for bastion connections. [GH-4940]
* communicator/winrm: Add NTLM authentication support. [GH-4979]
* communicator/winrm: Add support for file downloads. [GH-4748]
* core: add telemetry for better product support. [GH-5015]
* core: Build binaries for arm64 [GH-4892]
* post-processor/amazon-import: Add support for `license_type`. [GH-4634]
* post-processor/vagrant-cloud: Get vagrant cloud token from environment.
    [GH-4982]
* provisioner/ansible-local: Add extra-vars `packer_build_name`,
    `packer_builder_type`, and `packer_http_addr`. [GH-4821]
* provisioner/ansible: Add `inventory_directory` option to control where to
    place the generated inventory file. [GH-4760]
* provisioner/ansible: Add `skip_version_check` flag for when ansible will be
    installed from a prior provisioner. [GH-4983]
* provisioner/ansible: Add extra-vars `packer_build_name` and
    `packer_builder_type`. [GH-4821]
* provisioner/chef-solo: Add option to select Chef version. [GH-4791]
* provisioner/salt: Add salt bin directory configuration. [GH-5009]
* provisioner/salt: Add support for grains. [GH-4961]
* provisioner/shell: Use `env` to set environment variables to support freebsd
    out of the box. [GH-4909]
* website/docs: Clarify language, improve formatting. [GH-4866]
* website/docs: Update docker metadata fields that can be changed. [GH-4867]


### BUG FIXES:

* builder/amazon-ebssurrogate: Use ami device settings when creating the AMI.
    [GH-4972]
* builder/amazon: don't try to delete extra volumes during clean up. [GH-4930]
* builder/amazon: fix `force_delete_snapshot` when the launch instance has
    extra volumes. [GH-4931]
* builder/amazon: Only delete temporary key if we created one. [GH-4850]
* builder/azure: Replace calls to panic with error returns. [GH-4846]
* communicator/winrm: Use KeepAlive to keep long-running connections open.
    [GH-4952]
* core: Correctly reject config files which have junk after valid json.
    [GH-4906]
* post-processor/checksum: fix crash when invalid checksum is used. [GH-4812]
* post-processor/vagrant-cloud: don't read files to upload in to memory first.
    [GH-5005]
* post-processor/vagrant-cloud: only upload once under normal conditions.
    [GH-5008]
* provisioner/ansible-local: Correctly set the default staging directory under
    Windows. [GH-4792]

### FEATURES:

* **New builder:** `alicloud-ecs` for building Alicloud ECS images. [GH-4619]


## 1.0.0 (April 4, 2017)

### BUG FIXES:

* builder/amazon: Fix b/c issue by reporting again the tags we create.
    [GH-4704]
* builder/amazon: Fix crash in `step_region_copy`. [GH-4642]
* builder/googlecompute: Correct values for `on_host_maintenance`. [GH-4643]
* builder/googlecompute: Use "default" service account. [GH-4749]
* builder/hyper-v: Don't wait for shutdown_command to return. [GH-4691]
* builder/virtualbox: fix `none` communicator by allowing skipping upload of
    version file. [GH-4678]
* builder/virtualbox: retry removing floppy controller. [GH-4705]
* communicator/ssh: don't return error if we can't close connection. [GH-4741]
* communicator/ssh: fix nil pointer error. [GH-4690]
* core: fix version number
* core: Invoking packer `--help` or `--version` now exits with status 0.
    [GH-4723]
* core: show correct step name when debugging. [GH-4672]
* communicator/winrm: Directory uploads behave more like scp. [GH-4438]

### IMPROVEMENTS:

* builder/amazon-chroot: Ability to give an empty list in `copy_files` to
    prevent the default `/etc/resolv.conf` file from being copied. If
    `copy_files` isn't given at all, the default behavior remains. [GH-4708]
* builder/amazon: set force_deregister to true on -force. [GH-4649]
* builder/amazon: validate ssh key name/file. [GH-4665]
* builder/ansible: Clearer error message when we have problems getting the
    ansible version. [GH-4694]
* builder/hyper-v: validate output dir in step, not in config. [GH-4645]
* More diligently try to complete azure-setup.sh. [GH-4752]
* website: fix display on ios devices. [GH-4618]

## 0.12.3 (March 1, 2017)

### BACKWARDS INCOMPATIBILITIES:

* provisioner/ansible: by default, the staging dir will be randomized. [GH-4472]

### FEATURES:

* **New builder:** `ebs-surrogate` for building AMIs from EBS volumes. [GH-4351]

### IMPROVEMENTS:

* builder/amazon-chroot: support encrypted boot volume. [GH-4584]
* builder/amazon: Add BuildRegion and SourceAMI template variables. [GH-4399]
* builder/amazon: Change EC2 Windows password timeout to 20 minutes. [GH-4590]
* builder/amazon: enable ena when `enhanced_networking` is set. [GH-4578]
* builder/azure:: add two new config variables for temp_compute_name and
    temp_resource_group_name. [GH-4468]
* builder/docker: create export dir if needed. [GH-4439]
* builder/googlecompute: Add `on_host_maintenance` option. [GH-4544]
* builder/openstack: add reuse_ips option to try to re-use existing IPs.
    [GH-4564]
* builder/vmware-esxi: try for longer to connect to vnc port. [GH-4480]
    [GH-4610]
* builder/vmware: allow extra options for ovftool. [GH-4536]
* builder/vmware: don't cache ip address so we know if it changes. [GH-4532]
* communicator/docker: preserve file mode. [GH-4443]
* communicator/ssh: Use SSH agent when enabled for bastion step. [GH-4598]
* communicator/winrm: support ProxyFromEnvironment. [GH-4463]
* core: don't show ui color if we're not colorized. [GH-4525]
* core: make VNC links clickable in terminal. [GH-4497] [GH-4498]
* docs: add community page. [GH-4550]
* post-processor/amazon-import: support AMI attributes on import [GH-4216]
* post-processor/docker-import: print stderr on docker import failure.
    [GH-4529]

### BUG FIXES:

* builder/amazon-ebsvolume: Fix interpolation of block_device. [GH-4464]
* builder/amazon: Fix ssh agent authentication. [GH-4597]
* builder/docker: Don't force tag if using a docker version that doesn't
    support it. [GH-4560]
* builder/googlecompute: fix bug when creating image from custom image_family.
    [GH-4518]
* builder/virtualbox: remove guest additions before saving image. [GH-4496]
* core: always check for an error first when walking a path. [GH-4467]
* core: update crypto/ssh lib to fix large file uploads. [GH-4546]
* provisioner/chef-client: only upload knife config if we're cleaning.
    [GH-4534]

## 0.12.2 (January 20, 2017)

### FEATURES:

* **New builder:** `triton` for building images for Joyent Triton. [GH-4325]
* **New provisioner:** `converge` for provisioning with converge.sh. [GH-4326]

### IMPROVEMENTS:

* builder/hyperv-iso: add `iso_target_extension` option. [GH-4294]
* builder/openstack: Add support for instance metadata. [GH-4361]
* builder/openstack: Attempt to use existing floating IPs before allocating a
    new one. [GH-4357]
* builder/parallels-iso: add `iso_target_extension` option. [GH-4294]
* builder/qemu: add `iso_target_extension` option. [GH-4294]
* builder/qemu: add `use_default_display` option for osx compatibility.
    [GH-4293]
* builder/qemu: Detect input disk image format during copy/convert. [GH-4343]
* builder/virtualbox-iso: add `iso_target_extension` option. [GH-4294]
* builder/virtualbox: add `skip_export` option to skip exporting the VM after
    build completes. [GH-4339]
* builder/vmware & builder/qemu: Allow configurable delay between keystrokes
    when typing boot command. [GH-4403]
* builder/vmware-iso: add `iso_target_extension` option. [GH-4294]
* builder/vmware-iso: add `skip_export` option to skip exporting the VM after
    build completes. [GH-4378]
* builder/vmware: Try to use `ip address` to find host IP. [GH-4411]
* common/step_http\_server: set `PACKER_HTTP_ADDR` env var for accessing http
    server from inside builder. [GH-4409]
* provisioner/powershell: Allow equals sign in value of environment variables.
    [GH-4328]
* provisioner/puppet-server: Add default facts.  [GH-4286]

### BUG FIXES:

* builder/amazon-chroot: Panic in AMI region copy step. [GH-4341]
* builder/amazon: Crashes when new EBS vols are used. [GH-4308]
* builder/amazon: Fix crash in amazon-instance. [GH-4372]
* builder/amazon: fix run volume tagging [GH-4420]
* builder/amazon: fix when using non-existent security\_group\_id. [GH-4425]
* builder/amazon: Properly error if we don't have the
    ec2:DescribeSecurityGroups permission. [GH-4304]
* builder/amazon: Properly wait for security group to exist. [GH-4369]
* builder/docker: Fix crash when performing log in to ECR with an invalid URL.
    [GH-4385]
* builder/openstack: fix for finding resource by ID. [GH-4301]
* builder/qemu: Explicitly set WinRMPort for StepConnect. [GH-4321]
* builder/virtualbox: Explicitly set WinRMPort for StepConnect. [GH-4321]
* builder/virtualbox: Pause between each boot command element in -debug.
    [GH-4346]
* builder/vmware builder/parallels: Fix hang when shutting down windows in
    certain cases. [GH-4436]
* command/push: Don't interpolate variables when pushing. [GH-4389]
* common/step_http_server: make port range inclusive. [GH-4398]
* communicator/winrm: update winrm client, resolving `MaxMemoryPerShellMB`
    errors and properly error logging instead of panicking. [GH-4412] [GH-4424]
* provider/windows-shell: Allows equals sign in env var value. [GH-4423]

## 0.12.1 (December 15, 2016)

### BACKWARDS INCOMPATIBILITIES:

* `ssh_username` is now required if using communicator ssh. [GH-4172]
* builder/amazon: Change `shutdown_behaviour` to `shutdown_behavior`.  Run
    "packer fix template.json" to migrate a template. [GH-4285]
* builder/openstack: No long supports the `api_key` option for rackspace.
    [GH-4283]
* post-processor/manifest: Changed `filename` field to be `output`, to be more
    consistent with other post-processors. `packer fix` will fix this for you.
    [GH-4192]
* post-processor/shell-local: Now runs per-builder instead of per-file. The
    filename is no longer passed in as an argument to the script, but instead
    needs to be gleaned from the manifest post-processor. [GH-4189]

### FEATURES:

* **New builder:** "Hyper-V" Added new builder for Hyper-V on Windows.
    [GH-2576]
* **New builder:** "1&1" Added new builder for [1&1](https://www.1and1.com/).
    [GH-4163]

### IMPROVEMENTS:

* builder/amazon-ebs: Support specifying KMS key for encryption. [GH-4023]
* builder/amazon-ebsvolume: Add artifact output. [GH-4141]
* builder/amazon: Add `snapshot_tag` overrides. [GH-4015]
* builder/amazon: Added new region London - eu-west-2. [GH-4284]
* builder/amazon: Added ca-central-1 to list of known aws regions. [GH-4274]
* builder/amazon: Adds `force_delete_snapshot` flag to also cleanup snapshots
    if we're removing a preexisting image, as with `force_deregister_image`.
    [GH-4223]
* builder/amazon: Support `snapshot_users` and `snapshot_groups` for sharing
    ebs snapshots. [GH-4243]
* builder/cloudstack: Support reusing an already associated public IP.
    [GH-4149]
* builder/docker: Introduce docker commit changes, author, and message.
    [GH-4202]
* builder/googlecompute: Support `source_image_family`. [GH-4162]
* builder/googlecompute: enable support for Google Compute XPN. [GH-4288]
* builder/openstack: Added `image_members` to add new members to image after
    it's created. [GH-4283]
* builder/openstack: Added `image_visibility` field to specify visibility of
    created image. [GH-4283]
* builder/openstack: Automatically reauth as needed. [GH-4262]
* builder/virtualbox-ovf: Can now give a URL to an ova file. [GH-3982]
* communicator/ssh: adds ability to download download directories and
    wildcards, fix destination file mode (not hardcoded anymore). [GH-4210]
* post-processor/shell-local: support spaces in script path. [GH-4144]
* provisioner/ansible: Allow `winrm` communicator. [GH-4209]
* provisioner/salt: Bootstrap fallback on wget if curl failed. [GH-4244]

### BUG FIXES:

* builder/amazon: Correctly assign key from `ssh_keypair_name` to source
    instance. [GH-4222]
* builder/amazon: Fix `source_ami_filter` ignores `owners`. [GH-4235]
* builder/amazon: Fix launching spot instances in EC2 Classic [GH-4204]
* builder/qemu: Fix issue where multiple <waitXX> commands on a single line
    in boot_command wouldn't be parsed correctly. [GH-4269]
* core: Unbreak glob patterns in `floppy_files`. [GH-3890]
* post-processor/checksum: cleanup, and fix output to specified file with
    more than one artifacts. [GH-4210]
* post-processor/checksum: reset hash after each artifact file. [GH-4210]
* provisioner/file: fix for directory download. [GH-4210]
* provisioner/file: fix issue uploading multiple files to a directory,
    mentioned in [GH-4049]. [GH-4210]
* provisioner/shell: Treat disconnects as retryable when running cleanup. If
    you have a reboot in your script, we'll now wait until the host is
    available before attempting to cleanup the script. [GH-4197]

## 0.12.0 (November 15, 2016)

### FEATURES:

* **New builder:** "cloudstack" Can create new templates for use with
    CloudStack taking either an ISO or existing template as input. [GH-3909]
* **New builder:** "profitbricks" Builder for creating images in the
    ProfitBricks cloud. [GH-3660]
* **New builder:** "amazon-ebsvolume" Can create Amazon EBS volumes which are
    preinitialized with a filesystem and data. [GH-4088]


### IMPROVEMENTS:

* builder/amazon: Allow polling delay override with `AWS_POLL_DELAY_SECONDS`.
    [GH-4083]
* builder/amazon: Allow use of local SSH Agent. [GH-4050]
* builder/amazon: Dynamic source AMI [GH-3817]
* builder/amazon: Show AMI ID found when using `source_ami_filter`. [GH-4096]
* builder/googlecompute: Support `ssh_private_key_file` in communicator.
    [GH-4101]
* builder/googlecompute: Support custom scopes. [GH-4043]
* command/push: Fix variable pushes to Atlas. Still needs Atlas server to be
    updated before the issue will be fixed completely. [GH-4089]
* communicator/ssh: Improved SSH upload performance. [GH-3940]
* contrib/azure-setup.sh: Support for azure-cli 0.10.7. [GH-4133]
* docs: Fix command line variable docs. [GH-4143]
* post-processor/vagrant: Fixed inconsistency between vagrant-libvirt driver
    and packer QEMU accelerator. [GH-4104]
* provisioner/ansible: Move info messages to log [GH-4123]
* provisioner/puppet: Add `puppet_bin_dir` option. [GH-4014]
* provisioner/salt: Add `salt_call_args` option. [GH-4158]

### BUG FIXES:

* builder/amazon: Fixed an error where we wouldn't fail the build even if we
    timed out waiting for the temporary security group to become available.
    [GH-4099]
* builder/amazon: Properly cleanup temporary key pairs. [GH-4080]
* builder/google: Fix issue where we'd hang waiting for a startup script
    which doesn't exist. [GH-4102]
* builder/qemu: Fix keycodes for ctrl, shift and alt keys. [GH-4115]
* builder/vmware: Fix keycodes for ctrl, shift and alt keys. [GH-4115]
* builder/vmware: Fixed build error when shutting down. [GH-4041]
* common/step_create_floppy: Fixed support for 1.44MB floppies on Windows.
    [GH-4135]
* post-processor/googlecompute-export: Fixes scopes. [GH-4147]
* provisioner/powershell: Reverted [GH-3371] fixes quoting issue. [GH-4069]
* scripts: Fix build under Windows for go 1.5. [GH-4142]

## 0.11.0 (October 21, 2016)

### BACKWARDS INCOMPATIBILITIES:

* VNC and VRDP-like features in VirtualBox, VMware, and QEMU now configurable
    but bind to 127.0.0.1 by default to improve security. See the relevant
    builder docs for more info.
* Docker builder requires Docker > 1.3
* provisioner/chef-solo: default staging directory renamed to
    `packer-chef-solo`. [GH-3971]

### FEATURES:

* **New Checksum post-processor**: Create a checksum file from your build
    artifacts as part of your build. [GH-3492] [GH-3790]
* **New build flag** `-on-error` to allow inspection and keeping artifacts on
    builder errors. [GH-3885]
* **New Google Compute Export post-processor**: exports an image from a Packer
    googlecompute builder run and uploads it to Google Cloud Storage.
    [GH-3760]
* **New Manifest post-processor**: writes metadata about packer's output
    artifacts data to a JSON file. [GH-3651]


### IMPROVEMENTS:

* builder/amazon: Added `disable_stop_instance` option to prevent automatic
    shutdown when the build is complete. [GH-3352]
* builder/amazon: Added `shutdown_behavior` option to support `stop` or
    `terminate` at the end of the build. [GH-3556]
* builder/amazon: Added `skip_region_validation` option to allow newer or
    custom AWS regions. [GH-3598]
* builder/amazon: Added `us-east-2` and `ap-south-1` regions. [GH-4021]
    [GH-3663]
* builder/amazon: Support building from scratch with amazon-chroot builder.
    [GH-3855] [GH-3895]
* builder/amazon: Support create an AMI with an `encrypt_boot` volume.
    [GH-3382]
* builder/azure: Add `os_disk_size_gb`. [GH-3995]
* builder/azure: Add location to setup script. [GH-3803]
* builder/azure: Allow user to set custom data. [GH-3996]
* builder/azure: Made `tenant_id` optional. [GH-3643]
* builder/azure: Now pre-validates `capture_container_name` and
    `capture_name_prefix` [GH-3537]
* builder/azure: Removed superfluous polling code for deployments. [GH-3638]
* builder/azure: Support for a user defined VNET. [GH-3683]
* builder/azure: Support for custom images. [GH-3575]
* builder/azure: tag all resources. [GH-3764]
* builder/digitalocean: Added `user_data_file` support. [GH-3933]
* builder/digitalocean: Fixes timeout waiting for snapshot. [GH-3868]
* builder/digitalocean: Use `state_timeout` for unlock and off transitions.
    [GH-3444]
* builder/docker: Improved support for Docker pull from Amazon ECR. [GH-3856]
* builder/google: Add `-force` option to delete old image before creating new
    one. [GH-3918]
* builder/google: Add image license metadata. [GH-3873]
* builder/google: Added support for `image_family` [GH-3531]
* builder/google: Added support for startup scripts. [GH-3639]
* builder/google: Create passwords for Windows instances. [GH-3932]
* builder/google: Enable to select NVMe images. [GH-3338]
* builder/google: Signal that startup script fished via metadata. [GH-3873]
* builder/google: Use gcloud application default credentials. [GH-3655]
* builder/google: provision VM without external IP address. [GH-3774]
* builder/null: Can now be used with WinRM. [GH-2525]
* builder/openstack: Added support for `ssh_password` instead of generating
    ssh keys. [GH-3976]
* builder/parallels: Add support for ctrl, shift and alt keys in
    `boot_command`.  [GH-3767]
* builder/parallels: Copy directories recursively with `floppy_dirs`.
    [GH-2919]
* builder/parallels: Now pauses between `boot_command` entries when running
    with `-debug` [GH-3547]
* builder/parallels: Support future versions of Parallels by using the latest
    driver. [GH-3673]
* builder/qemu: Add support for ctrl, shift and alt keys in `boot_command`.
    [GH-3767]
* builder/qemu: Added `vnc_bind_address` option. [GH-3574]
* builder/qemu: Copy directories recursively with `floppy_dirs`. [GH-2919]
* builder/qemu: Now pauses between `boot_command` entries when running with
    `-debug` [GH-3547]
* builder/qemu: Specify disk format when starting qemu. [GH-3888]
* builder/virtualbox-iso: Added `hard_drive_nonrotational` and
    `hard_drive_discard` options to enable trim/discard. [GH-4013]
* builder/virtualbox-iso: Added `keep_registered` option to skip cleaning up
    the image. [GH-3954]
* builder/virtualbox: Add support for ctrl, shift and alt keys in
    `boot_command`.  [GH-3767]
* builder/virtualbox: Added `post_shutdown_delay` option to wait after
    shutting down to prevent issues removing floppy drive. [GH-3952]
* builder/virtualbox: Added `vrdp_bind_address` option. [GH-3566]
* builder/virtualbox: Copy directories recursively with `floppy_dirs`.
    [GH-2919]
* builder/virtualbox: Now pauses between `boot_command` entries when running
    with `-debug` [GH-3542]
* builder/vmware-vmx: Added `tools_upload_flavor` and `tools_upload_path` to
    docs.
* builder/vmware: Add support for ctrl, shift and alt keys in `boot_command`.
    [GH-3767]
* builder/vmware: Added `vnc_bind_address` option. [GH-3565]
* builder/vmware: Adds passwords for VNC. [GH-2325]
* builder/vmware: Copy directories recursively with `floppy_dirs`. [GH-2919]
* builder/vmware: Handle connection to VM with more than one NIC on ESXi
    [GH-3347]
* builder/vmware: Now paused between `boot_command` entries when running with
    `-debug` [GH-3542]
* core: Supress plugin discovery from plugins. [GH-4002]
* core: Test floppy disk files actually exist. [GH-3756]
* core: setting `PACKER_LOG=0` now disables logging. [GH-3964]
* post-processor/amazon-import: Support `ami_name` for naming imported AMI.
    [GH-3941]
* post-processor/compress: Added support for bgzf compression. [GH-3501]
* post-processor/docker: Improved support for Docker push to Amazon ECR.
    [GH-3856]
* post-processor/docker: Preserve tags when running docker push. [GH-3631]
* post-processor/vagrant: Added vsphere-esx hosts to supported machine types.
    [GH-3967]
* provisioner/ansible-local: Support for ansible-galaxy. [GH-3350] [GH-3836]
* provisioner/ansible: Improved logging and error handling. [GH-3477]
* provisioner/ansible: Support scp. [GH-3861]
* provisioner/chef: Added `knife_command` option and added a correct default
    value for Windows. [GH-3622]
* provisioner/chef: Installs 64bit chef on Windows if available. [GH-3848]
* provisioner/file: Now makes destination directory. [GH-3692]
* provisioner/puppet: Added `execute_command` option. [GH-3614]
* provisioner/salt: Added `custom_state` to specify state to run instead of
    `highstate`. [GH-3776]
* provisioner/shell: Added `expect_disconnect` flag to fail if remote
    unexpectedly disconnects. [GH-4034]
* scripts: Added `help` target to Makefile. [GH-3290]
* vendor: Moving from Godep to govendor. See `CONTRIBUTING.md` for details.
    [GH-3956]
* website: code examples now use inconsolata. Improve code font rendering on
    linux.

### BUG FIXES:

* builder/amazon: Add 0.5 cents to discovered spot price. [GH-3662]
* builder/amazon: Allow using `ssh_private_key_file` and `ssh_password`.
    [GH-3953]
* builder/amazon: Fix packer crash when waiting for SSH. [GH-3865]
* builder/amazon: Honor ssh_private_ip flag in EC2-Classic. [GH-3752]
* builder/amazon: Properly clean up EBS volumes on failure. [GH-3789]
* builder/amazon: Use `temporary_key_pair_name` when specified. [GH-3739]
* builder/amazon: retry creating tags on images since the images might take
    some time to become available. [GH-3938]
* builder/azure: Fix authorization setup script failing to creating service
    principal. [GH-3812]
* builder/azure: check for empty resource group. [GH-3606]
* builder/azure: fix token validity test. [GH-3609]
* builder/docker: Fix file provisioner dotfile matching. [GH-3800]
* builder/docker: fix docker builder with ansible provisioner. [GH-3476]
* builder/qemu: Don't fail on communicator set to `none`. [GH-3681]
* builder/qemu: Make `ssh_host_port_max` an inclusive bound. [GH-2784]
* builder/virtualbox: Make `ssh_host_port_max` an inclusive bound. [GH-2784]
* builder/virtualbox: Respect `ssh_host` [GH-3617]
* builder/vmware: Do not add remotedisplay.vnc.ip to VMX data on ESXi
    [GH-3740]
* builder/vmware: Don't check for poweron errors on ESXi. [GH-3195]
* builder/vmware: Re-introduce case sensitive VMX keys. [GH-2707]
* builder/vmware: Respect `ssh_host`/`winrm_host` on ESXi. [GH-3738]
* command/push: Allows dot (`.`) in image names. [GH-3937]
* common/iso_config: fix potential panic when iso checksum url was given but
    not the iso url. [GH-4004]
* communicator/ssh: fixed possible panic when reconnecting fails. [GH-4008]
* communicator/ssh: handle error case where server closes the connection but
    doesn't give us an error code. [GH-3966]
* post-processor/shell-local: Do not set execute bit on artifact file.
    [GH-3505]
* post-processor/vsphere: Fix upload failures with vsphere. [GH-3321]
* provisioner/ansible: Properly set host key checking even when a custom ENV
    is specified. [GH-3568]
* provisioner/file: Fix directory download. [GH-3899]
* provisioner/powershell: fixed issue with setting environment variables.
    [GH-2785]
* website: improved rendering on iPad. [GH-3780]

## 0.10.2 (September 20, 2016)

### BUG FIXES:

* Rebuilding with OS X Sierra and go 1.7.1 to fix bug  in Sierra

## 0.10.1 (May 7, 2016)

### FEATURES:

* `azure-arm` builder: Can now build Windows images, and supports additional
    configuration. Please refer to the documentation for details.

### IMPROVEMENTS:

* core: Added support for `ATLAS_CAFILE` and `ATLAS_CAPATH` [GH-3494]
* builder/azure: Improved build cancellation and cleanup of partially-
    provisioned resources. [GH-3461]
* builder/azure: Improved logging. [GH-3461]
* builder/azure: Added support for US Government and China clouds. [GH-3461]
* builder/azure: Users may now specify an image version. [GH-3461]
* builder/azure: Added device login. [GH-3461]
* builder/docker: Added `privileged` build option. [GH-3475]
* builder/google: Packer now identifies its version to the service. [GH-3465]
* provisioner/shell: Added `remote_folder` and `remote_file` options
    [GH-3462]
* post-processor/compress: Added support for `bgzf` format and added
    `format` option. [GH-3501]

### BUG FIXES:

* core: Fix hang after pressing enter key in `-debug` mode. [GH-3346]
* provisioner/chef: Use custom values for remote validation key path
    [GH-3468]

## 0.10.0 (March 14, 2016)

### BACKWARDS INCOMPATIBILITIES:

* Building Packer now requires go >= 1.5 (>= 1.6 is recommended). If you want
    to continue building with go 1.4 you can remove the `azurearmbuilder` line
    from `command/plugin.go`.

### FEATURES:

* **New `azure-arm` builder**: Build virtual machines in Azure Resource
    Manager

### IMPROVEMENTS:

* builder/google: Added support for `disk_type` [GH-2830]
* builder/openstack: Added support for retrieving the Administrator password
    when using WinRM if no `winrm_password` is set. [GH-3209]
* provisioner/ansible: Added the `empty_groups` parameter. [GH-3232]
* provisioner/ansible: Added the `user` parameter. [GH-3276]
* provisioner/ansible: Don't use deprecated ssh option with Ansible 2.0
    [GH-3291]
* provisioner/puppet-masterless: Add `ignore_exit_codes` parameter. [GH-3349]

### BUG FIXES:

* builders/parallels: Handle `output_directory` containing `.` and `..`
    [GH-3239]
* provisioner/ansible: os.Environ() should always be passed to the ansible
    command. [GH-3274]

## 0.9.0 (February 19, 2016)

### BACKWARDS INCOMPATIBILITIES:

* Packer now ships as a single binary, including plugins. If you install packer
    0.9.0 over a previous packer installation, **you must delete all of the
    packer-* plugin files** or packer will load out-of-date plugins from disk.
* Release binaries are now provided via <https://releases.hashicorp.com>.
* Packer 0.9.0 is now built with Go 1.6.
* core: Plugins that implement the Communicator interface must now implement
    a DownloadDir method. [GH-2618]
* builder/amazon: Inline `user_data` for EC2 is now base64 encoded
    automatically. [GH-2539]
* builder/parallels: `parallels_tools_host_path` and `guest_os_distribution`
    have been replaced by `guest_os_type`; use `packer fix` to update your
    templates. [GH-2751]

### FEATURES:

* **Chef on Windows**: The chef provisioner now has native support for
    Windows using Powershell and WinRM. [GH-1215]
* **New `vmware-esxi` feature**: Packer can now export images from vCloud or
    vSphere during the build. [GH-1921]
* **New Ansible Provisioner**: `ansible` provisioner supports remote
    provisioning to keep your build image cleaner. [GH-1969]
* **New Amazon Import post-processor**: `amazon-import` allows you to upload an
    OVA-based VM to Amazon EC2. [GH-2962]
* **Shell Local post-processor**: `shell-local` allows you to run shell
    commands on the host after a build has completed for custom packaging or
    publishing of your artifacts. [GH-2706]
* **Artifice post-processor**: Override packer artifacts during post-
    processing. This allows you to extract artifacts from a packer builder and
    use them with other post-processors like compress, docker, and Atlas.

### IMPROVEMENTS:

* core: Packer plugins are now compiled into the main binary, reducing file
    size and build times, and making packer easier to install. The overall
    plugin architecture has not changed and third-party plugins can still be
    loaded from disk. Please make sure your plugins are up-to-date! [GH-2854]
* core: Packer now indicates line numbers for template parse errors. [GH-2742]
* core: Scripts are executed via `/usr/bin/env bash` instead of `/bin/bash`
    for broader compatibility. [GH-2913]
* core: `target_path` for builder downloads can now be specified. [GH-2600]
* core: WinRM communicator now supports HTTPS protocol. [GH-3061]
* core: Template syntax errors now show line, column, offset. [GH-3180]
* core: SSH communicator now supports downloading directories. [GH-2618]
* builder/amazon: Add support for `ebs_optimized` [GH-2806]
* builder/amazon: You can now specify `0` for `spot_price` to switch to on
    demand instances. [GH-2845]
* builder/amazon: Added `ap-northeast-2` (Seoul) [GH-3056]
* builder/amazon: packer will try to derive the AZ if only a subnet is
    specified. [GH-3037]
* builder/digitalocean: doubled instance wait timeouts to power off or
    shutdown (now 4 minutes) and to complete a snapshot (now 20 minutes)
    [GH-2939]
* builder/google: `account_file` can now be provided as a JSON string
    [GH-2811]
* builder/google: added support for `preemptible` instances. [GH-2982]
* builder/google: added support for static external IPs via `address` option
    [GH-3030]
* builder/openstack: added retry on WaitForImage 404. [GH-3009]
* builder/openstack: Can specify `source_image_name` instead of the ID
    [GH-2577]
* builder/openstack: added support for SSH over IPv6. [GH-3197]
* builder/parallels: Improve support for Parallels 11. [GH-2662]
* builder/parallels: Parallels disks are now compacted by default. [GH-2731]
* builder/parallels: Packer will look for Parallels in
    `/Applications/Parallels Desktop.app` if it is not detected automatically
    [GH-2839]
* builder/qemu: qcow2 images are now compacted by default. [GH-2748]
* builder/qemu: qcow2 images can now be compressed. [GH-2748]
* builder/qemu: Now specifies `virtio-scsi` by default. [GH-2422]
* builder/qemu: Now checks for version-specific options. [GH-2376]
* builder/qemu: Can now bypass disk cache using `iso_skip_cache` [GH-3105]
* builder/qemu: `<wait>` in `boot_command` now accepts an arbitrary duration
    like <wait1m30s> [GH-3129]
* builder/qemu: Expose `{{ .SSHHostPort }}` in templates. [GH-2884]
* builder/virtualbox: Added VRDP for debugging. [GH-3188]
* builder/vmware-esxi: Added private key auth for remote builds via
    `remote_private_key_file` [GH-2912]
* post-processor/atlas: Added support for compile ID. [GH-2775]
* post-processor/docker-import: Can now import Artifice artifacts. [GH-2718]
* provisioner/chef: Added `encrypted_data_bag_secret_path` option. [GH-2653]
* provisioner/puppet: Added the `extra_arguments` parameter. [GH-2635]
* provisioner/salt: Added `no_exit_on_failure`, `log_level`, and improvements
    to salt command invocation. [GH-2660]

### BUG FIXES:

* core: Random number generator is now seeded. [GH-2640]
* core: Packer should now have a lot less race conditions. [GH-2824]
* builder/amazon: The `no_device` option for block device mappings is now handled correctly. [GH-2398]
* builder/amazon: AMI name validation now matches Amazon's spec. [GH-2774]
* builder/amazon: Use snapshot size when volume size is unspecified. [GH-2480]
* builder/amazon: Pass AccessKey and SecretKey when uploading bundles for
    instance-backed AMIs. [GH-2596]
* builder/parallels: Added interpolation in `prlctl_post` [GH-2828]
* builder/vmware: `format` option is now read correctly. [GH-2892]
* builder/vmware-esxi: Correct endless loop in destroy validation logic
    [GH-2911]
* provisioner/shell: No longer leaves temp scripts behind. [GH-1536]
* provisioner/winrm: Now waits for reboot to complete before continuing with provisioning. [GH-2568]
* post-processor/artifice: Fix truncation of files downloaded from Docker. [GH-2793]


## 0.8.6 (Aug 22, 2015)

### IMPROVEMENTS:

* builder/docker: Now supports Download so it can be used with the file
    provisioner to download a file from a container. [GH-2585]
* builder/docker: Now verifies that the artifact will be used before the build
    starts, unless the `discard` option is specified. This prevent failures
    after the build completes. [GH-2626]
* post-processor/artifice: Now supports glob-like syntax for filenames. [GH-2619]
* post-processor/vagrant: Like the compress post-processor, vagrant now uses a
    parallel gzip algorithm to compress vagrant boxes. [GH-2590]

### BUG FIXES:

* core: When `iso_url` is a local file and the checksum is invalid, the local
    file will no longer be deleted. [GH-2603]
* builder/parallels: Fix interpolation in `parallels_tools_guest_path` [GH-2543]

## 0.8.5 (Aug 10, 2015)

### FEATURES:

* **[Beta]** Artifice post-processor: Override packer artifacts during post-
    processing. This allows you to extract artifacts from a packer builder
    and use them with other post-processors like compress, docker, and Atlas.

### IMPROVEMENTS:

* Many docs have been updated and corrected; big thanks to our contributors!
* builder/openstack: Add debug logging for IP addresses used for SSH. [GH-2513]
* builder/openstack: Add option to use existing SSH keypair. [GH-2512]
* builder/openstack: Add support for Glance metadata. [GH-2434]
* builder/qemu and builder/vmware: Packer's VNC connection no longer asks for
    an exclusive connection. [GH-2522]
* provisioner/salt-masterless: Can now customize salt remote directories. [GH-2519]

### BUG FIXES:

* builder/amazon: Improve instance cleanup by storing id sooner. [GH-2404]
* builder/amazon: Only fetch windows password when using WinRM communicator. [GH-2538]
* builder/openstack: Support IPv6 SSH address. [GH-2450]
* builder/openstack: Track new IP address discovered during RackConnect. [GH-2514]
* builder/qemu: Add 100ms delay between VNC key events. [GH-2415]
* post-processor/atlas: atlas_url configuration option works now. [GH-2478]
* post-processor/compress: Now supports interpolation in output config. [GH-2414]
* provisioner/powershell: Elevated runs now receive environment variables. [GH-2378]
* provisioner/salt-masterless: Clarify error messages when we can't create or
    write to the temp directory. [GH-2518]
* provisioner/salt-masterless: Copy state even if /srv/salt exists already. [GH-1699]
* provisioner/salt-masterless: Make sure /etc/salt exists before writing to it. [GH-2520]
* provisioner/winrm: Connect to the correct port when using NAT with
    VirtualBox / VMware. [GH-2399]

## Note: 0.8.3 was pulled and 0.8.4 was skipped.

## 0.8.2 (July 17, 2015)

### IMPROVEMENTS:

* builder/docker: Add option to use a Pty. [GH-2425]

### BUG FIXES:

* core: Fix crash when `min_packer_version` is specified in a template. [GH-2385]
* builder/amazon: Fix EC2 devices being included in EBS mappings. [GH-2459]
* builder/googlecompute: Fix default name for GCE images. [GH-2400]
* builder/null: Fix error message with missing ssh_host. [GH-2407]
* builder/virtualbox: Use --portcount on VirtualBox 5.x. [GH-2438]
* provisioner/puppet: Packer now correctly handles a directory for manifest_file. [GH-2463]
* provisioner/winrm: Fix potential crash with WinRM. [GH-2416]

## 0.8.1 (July 2, 2015)

### IMPROVEMENTS:

* builder/amazon: When debug mode is enabled, the Windows administrator
    password for Windows instances will be shown. [GH-2351]

### BUG FIXES:

* core: `min_packer_version`  field in configs work. [GH-2356]
* core: The `build_name` and `build_type` functions work in provisioners. [GH-2367]
* core: Handle timeout in SSH handshake. [GH-2333]
* command/build: Fix reading configuration from stdin. [GH-2366]
* builder/amazon: Fix issue with sharing AMIs when using `ami_users` [GH-2308]
* builder/amazon: Fix issue when using multiple Security Groups. [GH-2381]
* builder/amazon: Fix for tag creation when creating new ec2 instance. [GH-2317]
* builder/amazon: Fix issue with creating AMIs with multiple device mappings. [GH-2320]
* builder/amazon: Fix failing AMI snapshot tagging when copying to other
    regions. [GH-2316]
* builder/amazon: Fix setting AMI launch permissions. [GH-2348]
* builder/amazon: Fix spot instance cleanup to remove the correct request. [GH-2327]
* builder/amazon: Fix `bundle_prefix` not interpolating `timestamp` [GH-2352]
* builder/amazon-instance: Fix issue with creating AMIs without specifying a
    virtualization type. [GH-2330]
* builder/digitalocean: Fix builder using private IP instead of public IP. [GH-2339]
* builder/google: Set default communicator settings properly. [GH-2353]
* builder/vmware-iso: Setting `checksum_type` to `none` for ESX builds
    now works. [GH-2323]
* provisioner/chef: Use knife config file vs command-line params to
    clean up nodes so full set of features can be used. [GH-2306]
* post-processor/compress: Fixed crash in compress post-processor plugin. [GH-2311]

## 0.8.0 (June 23, 2015)

### BACKWARDS INCOMPATIBILITIES:

* core: SSH connection will no longer request a PTY by default. This
    can be enabled per builder.
* builder/digitalocean: no longer supports the v1 API which has been
    deprecated for some time. Most configurations should continue to
    work as long as you use the `api_token` field for auth.
* builder/digitalocean: `image`, `region`, and `size` are now required.
* builder/openstack: auth parameters have been changed to better
    reflect OS terminology. Existing environment variables still work.

### FEATURES:

* **WinRM:** You can now connect via WinRM with almost every builder.
    See the docs for more info. [GH-2239]
* **Windows AWS Support:** Windows AMIs can now be built without any
    external plugins: Packer will start a Windows instance, get the
    admin password, and can use WinRM (above) to connect through. [GH-2240]
* **Disable SSH:** Set `communicator` to "none" in any builder to disable SSH
    connections. Note that provisioners won't work if this is done. [GH-1591]
* **SSH Agent Forwarding:** SSH Agent Forwarding will now be enabled
    to allow access to remote servers such as private git repos. [GH-1066]
* **SSH Bastion Hosts:** You can now specify a bastion host for
    SSH access (works with all builders). [GH-387]
* **OpenStack v3 Identity:** The OpenStack builder now supports the
    v3 identity API.
* **Docker builder supports SSH**: The Docker builder now supports containers
    with SSH, just set `communicator` to "ssh" [GH-2244]
* **File provisioner can download**: The file provisioner can now download
    files out of the build process. [GH-1909]
* **New config function: `build_name`**: The name of the currently running
    build. [GH-2232]
* **New config function: `build_type`**: The type of the currently running
    builder. This is useful for provisioners. [GH-2232]
* **New config function: `template_dir`**: The directory to the template
    being built. This should be used for template-relative paths. [GH-54]
* **New provisioner: shell-local**: Runs a local shell script. [GH-770]
* **New provisioner: powershell**: Provision Windows machines
    with PowerShell scripts. [GH-2243]
* **New provisioner: windows-shell**: Provision Windows machines with
    batch files. [GH-2243]
* **New provisioner: windows-restart**: Restart a Windows machines and
    wait for it to come back online. [GH-2243]
* **Compress post-processor supports multiple algorithms:** The compress
    post-processor now supports lz4 compression and compresses gzip in
    parallel for much faster throughput.

### IMPROVEMENTS:

* core: Interrupt handling for SIGTERM signal as well. [GH-1858]
* core: HTTP downloads support resuming. [GH-2106]
    * builder/*: Add `ssh_handshake_attempts` to configure the number of
    handshake attempts done before failure. [GH-2237]
* builder/amazon: Add `force_deregister` option for automatic AMI
    deregistration. [GH-2221]
* builder/amazon: Now applies tags to EBS snapshots. [GH-2212]
* builder/amazon: Clean up orphaned volumes from Source AMIs. [GH-1783]
* builder/amazon: Support custom keypairs. [GH-1837]
* builder/amazon-chroot: Can now resize the root volume of the resulting
    AMI with the `root_volume_size` option. [GH-2289]
* builder/amazon-chroot: Add `mount_options` configuration option for providing
    options to the `mount` command. [GH-2296]
* builder/digitalocean: Save SSH key to pwd if debug mode is on. [GH-1829]
* builder/digitalocean: User data support. [GH-2113]
* builder/googlecompute: Option to use internal IP for connections. [GH-2152]
* builder/parallels: Support Parallels Desktop 11. [GH-2199]
* builder/openstack: Add `rackconnect_wait` for Rackspace customers to wait for
    RackConnect data to appear
* builder/openstack: Add `ssh_interface` option for rackconnect for users that
    have prohibitive firewalls
* builder/openstack: Flavor names can be used as well as refs
* builder/openstack: Add `availability_zone` [GH-2016]
* builder/openstack: Machine will be stopped prior to imaging if the
    cluster supports the `startstop` extension. [GH-2223]
* builder/openstack: Support for user data. [GH-2224]
* builder/qemu: Default accelerator to "tcg" on Windows. [GH-2291]
* builder/virtualbox: Added option: `ssh_skip_nat_mapping` to skip the
    automatic port forward for SSH and to use the guest port directly. [GH-1078]
* builder/virtualbox: Added SCSI support
* builder/vmware: Support for additional disks. [GH-1382]
* builder/vmware: Can now customize the template used for adding disks. [GH-2254]
* command/fix: After fixing, the template is validated. [GH-2228]
* command/push: Add `-name` flag for specifying name from CLI. [GH-2042]
* command/push: Push configuration in templates supports variables. [GH-1861]
* post-processor/docker-save: Can be chained. [GH-2179]
* post-processor/docker-tag: Support `force` option. [GH-2055]
* post-processor/docker-tag: Can be chained. [GH-2179]
* post-processor/vsphere: Make more fields optional, support empty
    resource pools. [GH-1868]
* provisioner/puppet-masterless: `working_directory` option. [GH-1831]
* provisioner/puppet-masterless: `packer_build_name` and
    `packer_build_type` are default facts. [GH-1878]
* provisioner/puppet-server: `ignore_exit_codes` option added. [GH-2280]

### BUG FIXES:

* core: Fix potential panic for post-processor plugin exits. [GH-2098]
* core: `PACKER_CONFIG` may point to a non-existent file. [GH-2226]
* builder/amazon: Allow spaces in AMI names when using `clean_ami_name` [GH-2182]
* builder/amazon: Remove deprecated ec2-upload-bundle parameter. [GH-1931]
* builder/amazon: Use IAM Profile to upload bundle if provided. [GH-1985]
* builder/amazon: Use correct exit code after SSH authentication failed. [GH-2004]
* builder/amazon: Retry finding created instance for eventual
    consistency. [GH-2129]
* builder/amazon: If no AZ is specified, use AZ chosen automatically by
    AWS for spot instance. [GH-2017]
* builder/amazon: Private key file (only available in debug mode)
    is deleted on cleanup. [GH-1801]
* builder/amazon: AMI copy won't copy to the source region. [GH-2123]
* builder/amazon: Validate AMI doesn't exist with name prior to build. [GH-1774]
* builder/amazon: Improved retry logic around waiting for instances. [GH-1764]
* builder/amazon: Fix issues with creating Block Devices. [GH-2195]
* builder/amazon/chroot: Retry waiting for disk attachments. [GH-2046]
* builder/amazon/chroot: Only unmount path if it is mounted. [GH-2054]
* builder/amazon/instance: Use `-i` in sudo commands so PATH is inherited. [GH-1930]
* builder/amazon/instance: Use `--region` flag for bundle upload command. [GH-1931]
* builder/digitalocean: Wait for droplet to unlock before changing state,
    should lower the "pending event" errors.
* builder/digitalocean: Ignore invalid fields from the ever-changing v2 API
* builder/digitalocean: Private images can be used as a source. [GH-1792]
* builder/docker: Fixed hang on prompt while copying script
* builder/docker: Use `docker exec` for newer versions of Docker for
    running scripts. [GH-1993]
* builder/docker: Fix crash that could occur at certain timed ctrl-c. [GH-1838]
* builder/docker: validate that `export_path` is not a directory. [GH-2105]
* builder/google: `ssh_timeout` is respected. [GH-1781]
* builder/openstack: `ssh_interface` can be used to specify the interface
    to retrieve the SSH IP from. [GH-2220]
* builder/qemu: Add `disk_discard` option. [GH-2120]
* builder/qemu: Use proper SSH port, not hardcoded to 22. [GH-2236]
* builder/qemu: Find unused SSH port if SSH port is taken. [GH-2032]
* builder/virtualbox: Bind HTTP server to IPv4, which is more compatible with
    OS installers. [GH-1709]
* builder/virtualbox: Remove the floppy controller in addition to the
    floppy disk. [GH-1879]
* builder/virtualbox: Fixed regression where downloading ISO without a
    ".iso" extension didn't work. [GH-1839]
* builder/virtualbox: Output dir is verified at runtime, not template
    validation time. [GH-2233]
* builder/virtualbox: Find unused SSH port if SSH port is taken. [GH-2032]
* builder/vmware: Add 100ms delay between keystrokes to avoid subtle
    timing issues in most cases. [GH-1663]
* builder/vmware: Bind HTTP server to IPv4, which is more compatible with
    OS installers. [GH-1709]
* builder/vmware: Case-insensitive match of MAC address to find IP. [GH-1989]
* builder/vmware: More robust IP parsing from ifconfig output. [GH-1999]
* builder/vmware: Nested output directories for ESXi work. [GH-2174]
* builder/vmware: Output dir is verified at runtime, not template
    validation time. [GH-2233]
* command/fix: For the `virtualbox` to `virtualbox-iso` builder rename,
    provisioner overrides are now also fixed. [GH-2231]
* command/validate: don't crash for invalid builds. [GH-2139]
* post-processor/atlas: Find common archive prefix for Windows. [GH-1874]
* post-processor/atlas: Fix index out of range panic. [GH-1959]
* post-processor/vagrant-cloud: Fixed failing on response
* post-processor/vagrant-cloud: Don't delete version on error. [GH-2014]
* post-processor/vagrant-cloud: Retry failed uploads a few times
* provisioner/chef-client: Fix permissions issues on default dir. [GH-2255]
* provisioner/chef-client: Node cleanup works now. [GH-2257]
* provisioner/puppet-masterless: Allow manifest_file to be a directory
* provisioner/salt-masterless: Add `--retcode-passthrough` to salt-call
* provisioner/shell: chmod executable script to 0755, not 0777. [GH-1708]
* provisioner/shell: inline commands failing will fail the provisioner. [GH-2069]
* provisioner/shell: single quotes in env vars are escaped. [GH-2229]
* provisioner/shell: Temporary file is deleted after run. [GH-2259]
* provisioner/shell: Randomize default script name to avoid strange
    race issues from Windows. [GH-2270]

## 0.7.5 (December 9, 2014)

### FEATURES:

* **New command: `packer push`**: Push template and files to HashiCorp's
    Atlas for building your templates automatically.
* **New post-processor: `atlas`**: Send artifact to HashiCorp's Atlas for
    versioning and storing artifacts. These artifacts can then be queried
    using the API, Terraform, etc.

### IMPROVEMENTS:

* builder/googlecompute: Support for ubuntu-os-cloud project
* builder/googlecompute: Support for OAuth2 to avoid client secrets file
* builder/googlecompute: GCE image from persistent disk instead of tarball
* builder/qemu: Checksum type "none" can be used
* provisioner/chef: Generate a node name if none available
* provisioner/chef: Added ssl_verify_mode configuration

### BUG FIXES:

* builder/parallels: Fixed attachment of ISO to cdrom device
* builder/parallels: Fixed boot load ordering
* builder/digitalocean: Fixed decoding of size
* builder/digitalocean: Fixed missing content-type header in request
* builder/digitalocean: Fixed use of private IP
* builder/digitalocean: Fixed the artifact ID generation
* builder/vsphere: Fixed credential escaping
* builder/qemu: Fixed use of CDROM with disk_image
* builder/aws: Fixed IP address for SSH in VPC
* builder/aws: Fixed issue with multiple block devices
* builder/vmware: Upload VMX to ESX5 after editing
* communicator/docker: Fix handling of symlinks during upload
* provisioner/chef: Fixed use of sudo in some cases
* core: Fixed build name interpolation
* postprocessor/vagrant: Fixed check for Vagrantfile template

## 0.7.2 (October 28, 2014)

### FEATURES:

* builder/digitalocean: API V2 support. [GH-1463]
* builder/parallels: Don't depend on _prl-utils_. [GH-1499]

### IMPROVEMENTS:

* builder/amazon/all: Support new AWS Frankfurt region.
* builder/docker: Allow remote `DOCKER_HOST`, which works as long as
    volumes work. [GH-1594]
* builder/qemu: Can set cache mode for main disk. [GH-1558]
* builder/qemu: Can build from pre-existing disk. [GH-1342]
* builder/vmware: Can specify path to Fusion installation with environmental
    variable `FUSION_APP_PATH`. [GH-1552]
* builder/vmware: Can specify the HW version for the VMX. [GH-1530]
* builder/vmware/esxi: Will now cache ISOs/floppies remotely. [GH-1479]
* builder/vmware/vmx: Source VMX can have a disk connected via SATA. [GH-1604]
* post-processors/vagrant: Support Qemu (libvirt) boxes. [GH-1330]
* post-processors/vagrantcloud: Support self-hosted box URLs.

### BUG FIXES:

* core: Fix loading plugins from pwd. [GH-1521]
* builder/amazon: Prefer token in config if given. [GH-1544]
* builder/amazon/all: Extended timeout for waiting for AMI. [GH-1533]
* builder/virtualbox: Can read VirtualBox version on FreeBSD. [GH-1570]
* builder/virtualbox: More robust reading of guest additions URL. [GH-1509]
* builder/vmware: Always remove floppies/drives. [GH-1504]
* builder/vmware: Wait some time so that post-VMX update aren't
    overwritten. [GH-1504]
* builder/vmware/esxi: Retry power on if it fails. [GH-1334]
* builder/vmware-vmx: Fix issue with order of boot command support. [GH-1492]
* builder/amazon: Extend timeout and allow user override. [GH-1533]
* builder/parallels: Ignore 'The fdd0 device does not exist' [GH-1501]
* builder/parallels: Rely on Cleanup functions to detach devices. [GH-1502]
* builder/parallels: Create VM without hdd and then add it later. [GH-1548]
* builder/parallels: Disconnect cdrom0. [GH-1605]
* builder/qemu: Don't use `-redir` flag anymore, replace with
    `hostfwd` options. [GH-1561]
* builder/qemu: Use `pc` as default machine type instead of `pc-1.0`.
* providers/aws: Ignore transient network errors. [GH-1579]
* provisioner/ansible: Don't buffer output so output streams in. [GH-1585]
* provisioner/ansible: Use inventory file always to avoid potentially
    deprecated feature. [GH-1562]
* provisioner/shell: Quote environmental variables. [GH-1568]
* provisioner/salt: Bootstrap over SSL. [GH-1608]
* post-processors/docker-push: Work with docker-tag artifacts. [GH-1526]
* post-processors/vsphere: Append "/" to object address. [GH-1615]

## 0.7.1 (September 10, 2014)

### FEATURES:

* builder/vmware: VMware Fusion Pro 7 is now supported. [GH-1478]

### BUG FIXES:

* core: SSH will connect slightly faster if it is ready immediately.
* provisioner/file: directory uploads no longer hang. [GH-1484]
* provisioner/file: fixed crash on large files. [GH-1473]
* scripts: Windows executable renamed to packer.exe. [GH-1483]

## 0.7.0 (September 8, 2014)

### BACKWARDS INCOMPATIBILITIES:

* The authentication configuration for Google Compute Engine has changed.
    The new method is much simpler, but is not backwards compatible.
    `packer fix` will _not_ fix this. Please read the updated GCE docs.

### FEATURES:

* **New Post-Processor: `compress`** - Gzip compresses artifacts with files.
* **New Post-Processor: `docker-save`** - Save an image. This is similar to
    export, but preserves the image hierarchy.
* **New Post-Processor: `docker-tag`** - Tag a created image.
* **New Template Functions: `upper`, `lower`** - See documentation for
    more details.
* core: Plugins are automatically discovered if they're named properly.
    Packer will look in the PWD and the directory with `packer` for
    binaries named `packer-TYPE-NAME`.
* core: Plugins placed in `~/.packer.d/plugins` are now automatically
    discovered.
* builder/amazon: Spot instances can now be used to build EBS backed and
    instance store images. [GH-1139]
* builder/docker: Images can now be committed instead of exported. [GH-1198]
* builder/virtualbox-ovf: New `import_flags` setting can be used to add
    new command line flags to `VBoxManage import` to allow things such
    as EULAs to be accepted. [GH-1383]
* builder/virtualbox-ovf: Boot commands and the HTTP server are supported.
    [GH-1169]
* builder/vmware: VMware Player 6 is now supported. [GH-1168]
* builder/vmware-vmx: Boot commands and the HTTP server are supported.
    [GH-1169]

### IMPROVEMENTS:

* core: `isotime` function can take a format. [GH-1126]
* builder/amazon/all: `AWS_SECURITY_TOKEN` is read and can also be
    set with the `token` configuration. [GH-1236]
* builder/amazon/all: Can force SSH on the private IP address with
    `ssh_private_ip`. [GH-1229]
* builder/amazon/all: String fields in device mappings can use variables. [GH-1090]
* builder/amazon-instance: EBS AMIs can be used as a source. [GH-1453]
* builder/digitalocean: Can set API URL endpoint. [GH-1448]
* builder/digitalocean: Region supports variables. [GH-1452]
* builder/docker: Can now specify login credentials to pull images.
* builder/docker: Support mounting additional volumes. [GH-1430]
* builder/parallels/all: Path to tools ISO is calculated automatically. [GH-1455]
* builder/parallels-pvm: `reassign_mac` option to choose whether or not
    to generate a new MAC address. [GH-1461]
* builder/qemu: Can specify "none" acceleration type. [GH-1395]
* builder/qemu: Can specify "tcg" acceleration type. [GH-1395]
* builder/virtualbox/all: `iso_interface` option to mount ISO with SATA. [GH-1200]
* builder/vmware-vmx: Proper `floppy_files` support. [GH-1057]
* command/build: Add `-color=false` flag to disable color. [GH-1433]
* post-processor/docker-push: Can now specify login credentials. [GH-1243]
* provisioner/chef-client: Support `chef_environment`. [GH-1190]

### BUG FIXES:

* core: nicer error message if an encrypted private key is used for
    SSH. [GH-1445]
* core: Fix crash that could happen with a well timed double Ctrl-C.
    [GH-1328] [GH-1314]
* core: SSH TCP keepalive period is now 5 seconds (shorter). [GH-1232]
* builder/amazon-chroot: Can properly build HVM images now. [GH-1360]
* builder/amazon-chroot: Fix crash in root device check. [GH-1360]
* builder/amazon-chroot: Add description that Packer made the snapshot
    with a time. [GH-1388]
* builder/amazon-ebs: AMI is deregistered if an error. [GH-1186]
* builder/amazon-instance: Fix deprecation warning for `ec2-bundle-vol`
    [GH-1424]
* builder/amazon-instance: Add `--no-filter` to the `ec2-bundle-vol`
    command by default to avoid corrupting data by removing package
    manager certs. [GH-1137]
* builder/amazon/all: `delete_on_termination` set to false will work.
* builder/amazon/all: Fix race condition on setting tags. [GH-1367]
* builder/amazon/all: More descriptive error messages if Amazon only
    sends an error code. [GH-1189]
* builder/docker: Error if `DOCKER_HOST` is set.
* builder/docker: Remove the container during cleanup. [GH-1206]
* builder/docker: Fix case where not all output would show up from
    provisioners.
* builder/googlecompute: add `disk_size` option. [GH-1397]
* builder/googlecompute: Auth works with latest formats on Google Cloud
    Console. [GH-1344]
* builder/openstack: Region is not required. [GH-1418]
* builder/parallels-iso: ISO not removed from VM after install. [GH-1338]
* builder/parallels/all: Add support for Parallels Desktop 10. [GH-1438]
* builder/parallels/all: Added some navigation keys. [GH-1442]
* builder/qemu: If headless, sdl display won't be used. [GH-1395]
* builder/qemu: Use `512M` as `-m` default. [GH-1444]
* builder/virtualbox/all: Search `VBOX_MSI_INSTALL_PATH` for path to
    `VBoxManage` on Windows. [GH-1337]
* builder/virtualbox/all: Seed RNG to avoid same ports. [GH-1386]
* builder/virtualbox/all: Better error if guest additions URL couldn't be
    detected. [GH-1439]
* builder/virtualbox/all: Detect errors even when `VBoxManage` exits
    with a zero exit code. [GH-1119]
* builder/virtualbox/iso: Append timestamp to default name for parallel
    builds. [GH-1365]
* builder/vmware/all: No more error when Packer stops an already-stopped
    VM. [GH-1300]
* builder/vmware/all: `ssh_host` accepts templates. [GH-1396]
* builder/vmware/all: Don't remount floppy in VMX post step. [GH-1239]
* builder/vmware/vmx: Do not re-add floppy disk files to VMX. [GH-1361]
* builder/vmware-iso: Fix crash when `vnc_port_min` and max were the
    same value. [GH-1288]
* builder/vmware-iso: Finding an available VNC port on Windows works. [GH-1372]
* builder/vmware-vmx: Nice error if Clone is not supported (not VMware
    Fusion Pro). [GH-787]
* post-processor/vagrant: Can supply your own metadata.json. [GH-1143]
* provisioner/ansible-local: Use proper path on Windows. [GH-1375]
* provisioner/file: Mode will now be preserved. [GH-1064]

## 0.6.1 (July 20, 2014)

### FEATURES:

* **New post processor:** `vagrant-cloud` - Push box files generated by
    vagrant post processor to Vagrant Cloud. [GH-1289]
* Vagrant post-processor can now packer Hyper-V boxes.

### IMPROVEMENTS:

* builder/amazon: Support for enhanced networking on HVM images. [GH-1228]
* builder/amazon-ebs: Support encrypted EBS volumes. [GH-1194]
* builder/ansible: Add `playbook_dir` option. [GH-1000]
* builder/openstack: Add ability to configure networks. [GH-1261]
* builder/openstack: Skip certificate verification. [GH-1121]
* builder/parallels/all: Add ability to select interface to connect to.
* builder/parallels/pvm: Support `boot_command`. [GH-1082]
* builder/virtualbox/all: Attempt to use local guest additions ISO
    before downloading from internet. [GH-1123]
* builder/virtualbox/ovf: Supports `guest_additions_mode` [GH-1035]
* builder/vmware/all: Increase cleanup timeout to 120 seconds. [GH-1167]
* builder/vmware/all: Add `vmx_data_post` for modifying VMX data
    after shutdown. [GH-1149]
* builder/vmware/vmx: Supports tools uploading. [GH-1154]

### BUG FIXES:

* core: `isotime` is the same time during the entire build. [GH-1153]
* builder/amazon-common: Sort AMI strings before outputting. [GH-1305]
* builder/amazon: User data can use templates/variables. [GH-1343]
* builder/amazon: Can now build AMIs in GovCloud.
* builder/null: SSH info can use templates/variables. [GH-1343]
* builder/openstack: Workaround for gophercloud.ServerById crashing. [GH-1257]
* builder/openstack: Force IPv4 addresses from address pools. [GH-1258]
* builder/parallels: Do not delete entire CDROM device. [GH-1115]
* builder/parallels: Errors while creating floppy disk. [GH-1225]
* builder/parallels: Errors while removing floppy drive. [GH-1226]
* builder/virtualbox-ovf: Supports guest additions options. [GH-1120]
* builder/vmware-iso: Fix esx5 path separator in windows. [GH-1316]
* builder/vmware: Remote ESXi builder now uploads floppy. [GH-1106]
* builder/vmware: Remote ESXi builder no longer re-uploads ISO every
    time. [GH-1244]
* post-processor/vsphere: Accept DOMAIN\account usernames. [GH-1178]
    * provisioner/chef-*: Fix remotePaths for Windows. [GH-394]

## 0.6.0 (May 2, 2014)

### FEATURES:

* **New builder:** `null` - The null builder does not produce any
    artifacts, but is useful for debugging provisioning scripts. [GH-970]
* **New builder:** `parallels-iso` and `parallels-pvm` - These can be
    used to build Parallels virtual machines. [GH-1101]
* **New provisioner:** `chef-client` - Provision using a the `chef-client`
    command, which talks to a Chef Server. [GH-855]
* **New provisioner:** `puppet-server` - Provision using Puppet by
    communicating to a Puppet master. [GH-796]
* `min_packer_version` can be specified in a Packer template to force
    a minimum version. [GH-487]

### IMPROVEMENTS:

* core: RPC transport between plugins switched to MessagePack
* core: Templates array values can now be comma separated strings.
    Most importantly, this allows for user variables to fill
    array configurations. [GH-950]
* builder/amazon: Added `ssh_private_key_file` option. [GH-971]
* builder/amazon: Added `ami_virtualization_type` option. [GH-1021]
* builder/digitalocean: Regions, image names, and sizes can be
    names that are looked up for their valid ID. [GH-960]
* builder/googlecompute: Configurable instance name. [GH-1065]
* builder/openstack: Support for conventional OpenStack environmental
    variables such as `OS_USERNAME`, `OS_PASSWORD`, etc. [GH-768]
* builder/openstack: Support `openstack_provider` option to automatically
    fill defaults for different OpenStack variants. [GH-912]
* builder/openstack: Support security groups. [GH-848]
* builder/qemu: User variable expansion in `ssh_key_path` [GH-918]
* builder/qemu: Floppy disk files list can also include globs
    and directories. [GH-1086]
* builder/virtualbox: Support an `export_opts` option which allows
    specifying arbitrary arguments when exporting the VM. [GH-945]
* builder/virtualbox: Added `vboxmanage_post` option to run vboxmanage
    commands just before exporting. [GH-664]
* builder/virtualbox: Floppy disk files list can also include globs
    and directories. [GH-1086]
* builder/vmware: Workstation 10 support for Linux. [GH-900]
* builder/vmware: add cloning support on Windows. [GH-824]
* builder/vmware: Floppy disk files list can also include globs
    and directories. [GH-1086]
* command/build: Added `-parallel` flag so you can disable parallelization
    with `-no-parallel`. [GH-924]
* post-processors/vsphere: `disk_mode` option. [GH-778]
* provisioner/ansible: Add `inventory_file` option. [GH-1006]
* provisioner/chef-client: Add `validation_client_name` option. [GH-1056]

### BUG FIXES:

* core: Errors are properly shown when adding bad floppy files. [GH-1043]
* core: Fix some URL parsing issues on Windows.
* core: Create Cache directory only when it is needed. [GH-367]
* builder/amazon-instance: Use S3Endpoint for ec2-upload-bundle arg,
    which works for every region. [GH-904]
* builder/digitalocean: updated default image_id. [GH-1032]
* builder/googlecompute: Create persistent disk as boot disk via
    API v1. [GH-1001]
* builder/openstack: Return proper error on invalid instance states. [GH-1018]
* builder/virtualbox-iso: Retry unregister a few times to deal with
    VBoxManage randomness. [GH-915]
* provisioner/ansible: Fix paths when provisioning Linux from
    Windows. [GH-963]
* provisioner/ansible: set cwd to staging directory. [GH-1016]
* provisioners/chef-client: Don't chown directory with Ubuntu. [GH-939]
* provisioners/chef-solo: Deeply nested JSON works properly. [GH-1076]
* provisioners/shell: Env var values can have equal signs. [GH-1045]
* provisioners/shell: chmod the uploaded script file to 0777. [GH-994]
* post-processor/docker-push: Allow repositories with ports. [GH-923]
* post-processor/vagrant: Create parent directories for `output` path. [GH-1059]
* post-processor/vsphere: datastore, network, and folder are no longer
    required. [GH-1091]

## 0.5.2 (02/21/2014)

### FEATURES:

*  **New post-processor:** `docker-import` - Import a Docker image and give it
    a specific repository/tag.
*  **New post-processor:** `docker-push` - Push an imported image to
    a registry.

### IMPROVEMENTS:

* core: Most downloads made by Packer now use a custom user agent. [GH-803]
* builder/googlecompute: SSH private key will be saved to disk if `-debug` is
    specified. [GH-867]
* builder/qemu: Can specify the name of the qemu binary. [GH-854]
* builder/virtualbox-ovf: Can specify import options such as "keepallmacs".
    [GH-883]

### BUG FIXES:

* core: Fix crash case if blank parameters are given to Packer. [GH-832]
* core: Fix crash if big file uploads are done. [GH-897]
* core: Fix crash if machine-readable output is going to a closed
    pipe. [GH-875]
* builder/docker: user variables work properly. [GH-777]
* builder/qemu: reboots are now possible in provisioners. [GH-864]
* builder/virtualbox,vmware: iso\_checksum is not required if the
    checksum type is "none"
* builder/virtualbox,vmware/qemu: Support for additional scancodes for
    `boot_command` such as `<up>`, `<left>`, `<insert>`, etc. [GH-808]
* communicator/ssh: Send TCP keep-alives on connections. [GH-872]
* post-processor/vagrant: AWS/DigitalOcean keep input artifacts by
    default. [GH-55]
* provisioners/ansible-local: Properly upload custom playbooks. [GH-829]
* provisioners/ansible-local: Better error if ansible isn't installed.
    [GH-836]

## 0.5.1 (01/02/2014)

### BUG FIXES:

* core: If a stream ID loops around, don't let it use stream ID 0. [GH-767]
* core: Fix issue where large writes to plugins would result in stream
    corruption. [GH-727]
* builders/virtualbox-ovf: `shutdown_timeout` config works. [GH-772]
* builders/vmware-iso: Remote driver works properly again. [GH-773]

## 0.5.0 (12/30/2013)

### BACKWARDS INCOMPATIBILITIES:

* "virtualbox" builder has been renamed to "virtualbox-iso". Running your
    template through `packer fix` will resolve this.
* "vmware" builder has been renamed to "vmware-iso". Running your template
    through `packer fix` will resolve this.
* post-processor/vagrant: Syntax for overriding by provider has changed.
    See the documentation for more information. Running your template
    through `packer fix` should resolve this.
* post-processor/vsphere: Some available configuration options were
    changed. Running your template through `packer fix` should resolve
    this.
* provisioner/puppet-masterless: The `execute_command` no longer has
    the `Has*` variables, since the templating language now supports
    comparison operations. See the Go documentation for more info:
    http://golang.org/pkg/text/template/

### FEATURES:

* **New builder:** Google Compute Engine. You can now build images for
    use in Google Compute Engine. See the documentation for more information.
    [GH-715]
* **New builder:** "virtualbox-ovf" can build VirtualBox images from
    an existing OVF or OVA. [GH-201]
* **New builder:** "vmware-vmx" can build VMware images from an existing
    VMX. [GH-201]
* Environmental variables can now be accessed as default values for
    user variables using the "env" function. See the documentation for more
    information.
* "description" field in templates: write a human-readable description
    of what a template does. This will be shown in `packer inspect`.
* Vagrant post-processor now accepts a list of files to include in the
    box.
* All provisioners can now have a "pause\_before" parameter to wait
    some period of time before running that provisioner. This is useful
    for reboots. [GH-737]

### IMPROVEMENTS:

* core: Plugins communicate over a single TCP connection per plugin now,
    instead of sometimes dozens. Performance around plugin communication
    dramatically increased.
* core: Build names are now template processed so you can use things
    like user variables in them. [GH-744]
* core: New "pwd" function available globally that returns the working
    directory. [GH-762]
* builder/amazon/all: Launched EC2 instances now have a name of
    "Packer Builder" so that they are easily recognizable. [GH-642]
* builder/amazon/all: Copying AMIs to multiple regions now happens
    in parallel. [GH-495]
* builder/amazon/all: Ability to specify "run\_tags" to tag the instance
    while running. [GH-722]
* builder/digitalocean: Private networking support. [GH-698]
* builder/docker: A "run\_command" can be specified, configuring how
    the container is started. [GH-648]
* builder/openstack: In debug mode, the generated SSH keypair is saved
    so you can SSH into the machine. [GH-746]
* builder/qemu: Floppy files are supported. [GH-686]
* builder/qemu: Next `run_once` option tells Qemu to run only once,
    which is useful for Windows installs that handle reboots for you.
    [GH-687]
* builder/virtualbox: Nice errors if Packer can't write to
    the output directory.
* builder/virtualbox: ISO is ejected prior to export.
* builder/virtualbox: Checksum type can be "none" [GH-471]
* builder/vmware: Can now specify path to the Fusion application. [GH-677]
* builder/vmware: Checksum type can be "none" [GH-471]
* provisioner/puppet-masterless: Can now specify a `manifest_dir` to
    upload manifests to the remote machine for imports. [GH-655]

### BUG FIXES:

* core: No colored output in machine-readable output. [GH-684]
* core: User variables can now be used for non-string fields. [GH-598]
* core: Fix bad download paths if the download URL contained a "."
    before a "/" [GH-716]
* core: "{{timestamp}}" values will always be the same for the entire
    duration of a build. [GH-744]
* builder/amazon: Handle cases where security group isn't instantly
    available. [GH-494]
* builder/virtualbox: don't download guest additions if disabled. [GH-731]
* post-processor/vsphere: Uploads VM properly. [GH-694]
* post-processor/vsphere: Process user variables.
* provisioner/ansible-local: all configurations are processed as templates
    [GH-749]
* provisioner/ansible-local: playbook paths are properly validated
    as directories, not files. [GH-710]
* provisioner/chef-solo: Environments are recognized. [GH-726]

## 0.4.1 (December 7, 2013)

### IMPROVEMENTS:

* builder/amazon/ebs: New option allows associating a public IP with
    non-default VPC instances. [GH-660]
* builder/openstack: A "proxy\_url" setting was added to define an HTTP
    proxy to use when building with this builder. [GH-637]

### BUG FIXES:

* core: Don't change background color on CLI anymore, making things look
    a tad nicer in some terminals.
* core: multiple ISO URLs works properly in all builders. [GH-683]
* builder/amazon/chroot: Block when obtaining file lock to allow
    parallel builds. [GH-689]
* builder/amazon/instance: Add location flag to upload bundle command
    so that building AMIs works out of us-east-1. [GH-679]
* builder/qemu: Qemu arguments are templated. [GH-688]
* builder/vmware: Cleanup of VMX keys works properly so cd-rom won't
    get stuck with ISO. [GH-685]
* builder/vmware: File cleanup is more resilient to file delete races
    with the operating system. [GH-675]
* provisioner/puppet-masterless: Check for hiera config path existence
    properly. [GH-656]

## 0.4.0 (November 19, 2013)

### FEATURES:

* Docker builder: build and export Docker containers, easily provisioned
    with any of the Packer built-in provisioners.
* QEMU builder: builds a new VM compatible with KVM or Xen using QEMU.
* Remote ESXi builder: builds a VMware VM using ESXi remotely using only
    SSH to an ESXi machine directly.
* vSphere post-processor: Can upload VMware artifacts to vSphere
* Vagrant post-processor can now make DigitalOcean provider boxes. [GH-504]

### IMPROVEMENTS:

* builder/amazon/all: Can now specify a list of multiple security group
    IDs to apply. [GH-499]
* builder/amazon/all: AWS API requests are now retried when a temporary
    network error occurs as well as 500 errors. [GH-559]
* builder/virtualbox: Use VBOX\_INSTALL\_PATH env var on Windows to find
    VBoxManage. [GH-628]
* post-processor/vagrant: skips gzip compression when compression_level=0
* provisioner/chef-solo: Encrypted data bag support. [GH-625]

### BUG FIXES:

* builder/amazon/chroot: Copying empty directories works. [GH-588]
* builder/amazon/chroot: Chroot commands work with shell provisioners. [GH-581]
* builder/amazon/chroot: Don't choose a mount point that is a partition of
    an already mounted device. [GH-635]
* builder/virtualbox: Ctrl-C interrupts during waiting for boot. [GH-618]
* builder/vmware: VMX modifications are now case-insensitive. [GH-608]
* builder/vmware: VMware Fusion won't ask for VM upgrade.
* builder/vmware: Ctrl-C interrupts during waiting for boot. [GH-618]
* provisioner/chef-solo: Output is slightly prettier and more informative.

## 0.3.11 (November 4, 2013)

### FEATURES:

* builder/amazon/ebs: Ability to specify which availability zone to create
    instance in. [GH-536]

### IMPROVEMENTS:

* core: builders can now give warnings during validation. warnings won't
    fail the build but may hint at potential future problems.
* builder/digitalocean: Can now specify a droplet name
* builder/virtualbox: Can now disable guest addition download entirely
    by setting "guest_additions_mode" to "disable" [GH-580]
* builder/virtualbox,vmware: ISO urls can now be https. [GH-587]
* builder/virtualbox,vmware: Warning if shutdown command is not specified,
    since it is a common case of data loss.

### BUG FIXES:

* core: Won't panic when writing to a bad pipe. [GH-560]
* builder/amazon/all: Properly scrub access key and secret key from logs.
    [GH-554]
* builder/openstack: Properly scrub password from logs. [GH-554]
* builder/virtualbox: No panic if SSH host port min/max is the same. [GH-594]
* builder/vmware: checks if `ifconfig` is in `/sbin` [GH-591]
* builder/vmware: Host IP lookup works for non-C locales. [GH-592]
* common/uuid: Use cryptographically secure PRNG when generating
    UUIDs. [GH-552]
* communicator/ssh: File uploads that exceed the size of memory no longer
    cause crashes. [GH-561]

## 0.3.10 (October 20, 2013)

### FEATURES:

* Ansible provisioner

### IMPROVEMENTS:

* post-processor/vagrant: support instance-store AMIs built by Packer. [GH-502]
* post-processor/vagrant: can now specify compression level to use
    when creating the box. [GH-506]

### BUG FIXES:

* builder/all: timeout waiting for SSH connection is a failure. [GH-491]
* builder/amazon: Scrub sensitive data from the logs. [GH-521]
* builder/amazon: Handle the situation where an EC2 instance might not
    be immediately available. [GH-522]
* builder/amazon/chroot: Files copied into the chroot remove destination
    before copy, fixing issues with dangling symlinks. [GH-500]
* builder/digitalocean: don't panic if erroneous API response doesn't
    contain error message. [GH-492]
* builder/digitalocean: scrub API keys from config debug output. [GH-516]
* builder/virtualbox: error if VirtualBox version cant be detected. [GH-488]
* builder/virtualbox: detect if vboxdrv isn't properly setup. [GH-488]
* builder/virtualbox: sleep a bit before export to ensure the session
    is unlocked. [GH-512]
* builder/virtualbox: create SATA drives properly on VirtualBox 4.3. [GH-547]
* builder/virtualbox: support user templates in SSH key path. [GH-539]
* builder/vmware: support user templates in SSH key path. [GH-539]
* communicator/ssh: Fix issue where a panic could arise from a nil
    dereference. [GH-525]
* post-processor/vagrant: Fix issue with VirtualBox OVA. [GH-548]
* provisioner/salt: Move salt states to correct remote directory. [GH-513]
* provisioner/shell: Won't block on certain scripts on Windows anymore.
    [GH-507]

## 0.3.9 (October 2, 2013)

### FEATURES:

* The Amazon chroot builder is now able to run without any `sudo` privileges
    by using the "command_wrapper" configuration. [GH-430]
* Chef provisioner supports environments. [GH-483]

### BUG FIXES:

* core: default user variable values don't need to be strings. [GH-456]
* builder/amazon-chroot: Fix errors with waiting for state change. [GH-459]
* builder/digitalocean: Use proper error message JSON key (DO API change).
* communicator/ssh: SCP uploads now work properly when directories
    contain symlinks. [GH-449]
* provisioner/chef-solo: Data bags and roles path are now properly
    populated when set. [GH-470]
* provisioner/shell: Windows line endings are actually properly changed
    to Unix line endings. [GH-477]

## 0.3.8 (September 22, 2013)

### FEATURES:

* core: You can now specify `only` and `except` configurations on any
    provisioner or post-processor to specify a list of builds that they
    are valid for. [GH-438]
* builders/virtualbox: Guest additions can be attached rather than uploaded,
    easier to handle for Windows guests. [GH-405]
* provisioner/chef-solo: Ability to specify a custom Chef configuration
    template.
* provisioner/chef-solo: Roles and data bags support. [GH-348]

### IMPROVEMENTS:

* core: User variables can now be used for integer, boolean, etc.
    values. [GH-418]
* core: Plugins made with incompatible versions will no longer load.
* builder/amazon/all: Interrupts work while waiting for AMI to be ready.
* provisioner/shell: Script line-endings are automatically converted to
    Unix-style line-endings. Can be disabled by setting "binary" to "true".
    [GH-277]

### BUG FIXES:

* core: Set TCP KeepAlives on internally created RPC connections so that
    they don't die. [GH-416]
* builder/amazon/all: While waiting for AMI, will detect "failed" state.
* builder/amazon/all: Waiting for state will detect if the resource (AMI,
    instance, etc.) disappears from under it.
* builder/amazon/instance: Exclude only contents of /tmp, not /tmp
    itself. [GH-437]
* builder/amazon/instance: Make AccessKey/SecretKey available to bundle
    command even when they come from the environment. [GH-434]
* builder/virtualbox: F1-F12 and delete scancodes now work. [GH-425]
* post-processor/vagrant: Override configurations properly work. [GH-426]
* provisioner/puppet-masterless: Fix failure case when both facter vars
    are used and prevent_sudo. [GH-415]
* provisioner/puppet-masterless: User variables now work properly in
    manifest file and hiera path. [GH-448]

## 0.3.7 (September 9, 2013)

### BACKWARDS INCOMPATIBILITIES:

* The "event_delay" option for the DigitalOcean builder is now gone.
    The builder automatically waits for events to go away. Run your templates
    through `packer fix` to get rid of these.

### FEATURES:

* **NEW PROVISIONER:** `puppet-masterless`. You can now provision with
    a masterless Puppet setup. [GH-234]
* New globally available template function: `uuid`. Generates a new random
    UUID.
* New globally available template function: `isotime`. Generates the
    current time in ISO standard format.
* New Amazon template function: `clean_ami_name`. Substitutes '-' for
    characters that are illegal to use in an AMI name.

### IMPROVEMENTS:

* builder/amazon/all: Ability to specify the format of the temporary
    keypair created. [GH-389]
* builder/amazon/all: Support the NoDevice flag for block mappings. [GH-396]
* builder/digitalocean: Retry on any pending event errors.
* builder/openstack: Can now specify a project. [GH-382]
* builder/virtualbox: Can now attach hard drive over SATA. [GH-391]
* provisioner/file: Can now upload directories. [GH-251]

### BUG FIXES:

* core: Detect if SCP is not enabled on the other side. [GH-386]
* builder/amazon/all: When copying AMI to multiple regions, copy
    the metadata (tags and attributes) as well. [GH-388]
* builder/amazon/all: Fix panic case where eventually consistent
    instance state caused an index out of bounds.
* builder/virtualbox: The `vm_name` setting now properly sets the OVF
    name of the output. [GH-401]
* builder/vmware: Autoanswer VMware dialogs. [GH-393]
* command/inspect: Fix weird output for default values for optional vars.

## 0.3.6 (September 2, 2013)

### FEATURES:

* User variables can now be specified as "required", meaning the user
    MUST specify a value. Just set the default value to "null". [GH-374]

### IMPROVEMENTS:

* core: Much improved interrupt handling. For example, interrupts now
    cancel much more quickly within provisioners.
* builder/amazon: In `-debug` mode, the keypair used will be saved to
    the current directory so you can access the machine. [GH-373]
* builder/amazon: In `-debug` mode, the DNS is outputted.
* builder/openstack: IPv6 addresses supported for SSH. [GH-379]
* communicator/ssh: Support for private keys encrypted using PKCS8. [GH-376]
* provisioner/chef-solo: You can now use user variables in the `json`
    configuration for Chef. [GH-362]

### BUG FIXES:

* core: Concurrent map access is completely gone, fixing rare issues
    with runtime memory corruption. [GH-307]
* core: Fix possible panic when ctrl-C during provisioner run.
* builder/digitalocean: Retry destroy a few times because DO sometimes
    gives false errors.
* builder/openstack: Properly handle the case no image is made. [GH-375]
* builder/openstack: Specifying a region is now required in a template.
* provisioners/salt-masterless: Use filepath join to properly join paths.

## 0.3.5 (August 28, 2013)

### FEATURES:

* **NEW BUILDER:** `openstack`. You can now build on OpenStack. [GH-155]
* **NEW PROVISIONER:** `chef-solo`. You can now provision with Chef
    using `chef-solo` from local cookbooks.
* builder/amazon: Copy AMI to multiple regions with `ami_regions`. [GH-322]
* builder/virtualbox,vmware: Can now use SSH keys as an auth mechanism for
    SSH using `ssh_key_path`. [GH-70]
* builder/virtualbox,vmware: Support SHA512 as a checksum type. [GH-356]
* builder/vmware: The root hard drive type can now be specified with
    "disk_type_id" for advanced users. [GH-328]
* provisioner/salt-masterless: Ability to specify a minion config. [GH-264]
* provisioner/salt-masterless: Ability to upload pillars. [GH-353]

### IMPROVEMENTS:

* core: Output message when Ctrl-C received that we're cleaning up. [GH-338]
* builder/amazon: Tagging now works with all amazon builder types.
* builder/vmware: Option `ssh_skip_request_pty` for not requesting a PTY
    for the SSH connection. [GH-270]
* builder/vmware: Specify a `vmx_template_path` in order to customize
    the generated VMX. [GH-270]
* command/build: Machine-readable output now contains build errors, if any.
* command/build: An "end" sentinel is outputted in machine-readable output
    for artifact listing so it is easier to know when it is over.

### BUG FIXES:

* core: Fixed a couple cases where a double ctrl-C could panic.
* core: Template validation fails if an override is specified for a
    non-existent builder. [GH-336]
* core: The SSH connection is heartbeated so that drops can be
    detected. [GH-200]
* builder/amazon/instance: Remove check for ec2-ami-tools because it
    didn't allow absolute paths to work properly. [GH-330]
* builder/digitalocean: Send a soft shutdown request so that files
    are properly synced before shutdown. [GH-332]
* command/build,command/validate: If a non-existent build is specified to
    '-only' or '-except', it is now an error. [GH-326]
* post-processor/vagrant: Setting OutputPath with a timestamp now
    always works properly. [GH-324]
* post-processor/vagrant: VirtualBox OVA formats now turn into
    Vagrant boxes properly. [GH-331]
* provisioner/shell: Retry upload if start command fails, making reboot
    handling much more robust.

## 0.3.4 (August 21, 2013)

### IMPROVEMENTS:

* post-processor/vagrant: the file being compressed will be shown
    in the UI. [GH-314]

### BUG FIXES:

* core: Avoid panics when double-interrupting Packer.
* provisioner/shell: Retry shell script uploads, making reboots more
    robust if they happen to fail in this stage. [GH-282]

## 0.3.3 (August 19, 2013)

### FEATURES:

* builder/virtualbox: support exporting in OVA format. [GH-309]

### IMPROVEMENTS:

* core: All HTTP downloads across Packer now support the standard
    proxy environmental variables (`HTTP_PROXY`, `NO_PROXY`, etc.) [GH-252]
* builder/amazon: API requests will use HTTP proxy if specified by
    environmental variables.
* builder/digitalocean: API requests will use HTTP proxy if specified
    by environmental variables.

### BUG FIXES:

* core: TCP connection between plugin processes will keep-alive. [GH-312]
* core: No more "unused key keep_input_artifact" for post processors. [GH-310]
* post-processor/vagrant: `output_path` templates now work again.

## 0.3.2 (August 18, 2013)

### FEATURES:

* New command: `packer inspect`. This command tells you the components of
    a template. It respects the `-machine-readable` flag as well so you can
    parse out components of a template.
* Packer will detect its own crashes (always a bug) and save a "crash.log"
    file.
* builder/virtualbox: You may now specify multiple URLs for an ISO
    using "iso_url" in a template. The URLs will be tried in order.
* builder/vmware: You may now specify multiple URLs for an ISO
    using "iso_url" in a template. The URLs will be tried in order.

### IMPROVEMENTS:

* core: built with Go 1.1.2
* core: packer help output now loads much faster.
* builder/virtualbox: guest_additions_url can now use the `Version`
    variable to get the VirtualBox version. [GH-272]
* builder/virtualbox: Do not check for VirtualBox as part of template
    validation; only check at execution.
* builder/vmware: Do not check for VMware as part of template validation;
    only check at execution.
* command/build: A path of "-" will read the template from stdin.
* builder/amazon: add block device mappings. [GH-90]

### BUG FIXES:

* windows: file URLs are easier to get right as Packer
    has better parsing and error handling for Windows file paths. [GH-284]
* builder/amazon/all: Modifying more than one AMI attribute type no longer
    crashes.
* builder/amazon-instance: send IAM instance profile data. [GH-294]
* builder/digitalocean: API request parameters are properly URL
    encoded. [GH-281]
* builder/virtualbox: download progress won't be shown until download
    actually starts. [GH-288]
* builder/virtualbox: floppy files names of 13 characters are now properly
    written to the FAT12 filesystem. [GH-285]
* builder/vmware: download progress won't be shown until download
    actually starts. [GH-288]
* builder/vmware: interrupt works while typing commands over VNC.
* builder/virtualbox: floppy files names of 13 characters are now properly
    written to the FAT12 filesystem. [GH-285]
* post-processor/vagrant: Process user variables. [GH-295]

## 0.3.1 (August 12, 2013)

### IMPROVEMENTS:

* provisioner/shell: New setting `start_retry_timeout` which is the timeout
    for the provisioner to attempt to _start_ the remote process. This allows
    the shell provisioner to work properly with reboots. [GH-260]

### BUG FIXES:

* core: Remote command output containing '\r' now looks much better
    within the Packer output.
* builder/vmware: Fix issue with finding driver files. [GH-279]
* provisioner/salt-masterless: Uploads work properly from Windows. [GH-276]

## 0.3.0 (August 12, 2013)

### BACKWARDS INCOMPATIBILITIES:

* All `{{.CreateTime}}` variables within templates (such as for AMI names)
    are now replaced with `{{timestamp}}`. Run `packer fix` to fix your
    templates.

### FEATURES:

* **User Variables** allow you to specify variables within your templates
    that can be replaced using the command-line, files, or environmental
    variables. This dramatically improves the portability of packer templates.
    See the documentation for more information.
* **Machine-readable output** can be enabled by passing the
    `-machine-readable` flag to _any_ Packer command.
* All strings in a template are now processed for variables/functions,
    so things like `{{timestamp}}` can be used everywhere. More features will
    be added in the future.
* The `amazon` builders (all of them) can now have attributes of their
    resulting AMIs modified, such as access permissions and product codes.

### IMPROVEMENTS:

* builder/amazon/all: User data can be passed to start the instances. [GH-253]
* provisioner/salt-masterless: `local_state_tree` is no longer required,
    allowing you to use shell provisioner (or others) to bring this down.
    [GH-269]

### BUG FIXES:

* builder/amazon/ebs,instance: Retry deleing security group a few times.
    [GH-278]
* builder/vmware: Workstation works on Windows XP now. [GH-238]
* builder/vmware: Look for files on Windows in multiple locations
    using multiple environmental variables. [GH-263]
* provisioner/salt-masterless: states aren't deleted after the run
    anymore. [GH-265]
* provisioner/salt-masterless: error if any commands exit with a non-zero
    exit status. [GH-266]

## 0.2.3 (August 7, 2013)

### IMPROVEMENTS:

* builder/amazon/all: Added Amazon AMI tag support. [GH-233]

### BUG FIXES:

* core: Absolute/relative filepaths on Windows now work for iso_url
    and other settings. [GH-240]
* builder/amazon/all: instance info is refreshed while waiting for SSH,
    allowing Packer to see updated IP/DNS info. [GH-243]

## 0.2.2 (August 1, 2013)

### FEATURES:

* New builder: `amazon-chroot` can create EBS-backed AMIs without launching
    a new EC2 instance. This can shave minutes off of the AMI creation process.
    See the docs for more info.
* New provisioner: `salt-masterless` will provision the node using Salt
    without a master.
* The `vmware` builder now works with Workstation 9 on Windows. [GH-222]
* The `vmware` builder now works with Player 5 on Linux. [GH-190]

### IMPROVEMENTS:

* core: Colors won't be outputted on Windows unless in Cygwin.
* builder/amazon/all: Added `iam_instance_profile` to launch the source
    image with a given IAM profile. [GH-226]

### BUG FIXES:

* builder/virtualbox,vmware: relative paths work properly as URL
    configurations. [GH-215]
* builder/virtualbox,vmware: fix race condition in deleting the output
    directory on Windows by retrying.

## 0.2.1 (July 26, 2013)

### FEATURES:

* New builder: `amazon-instance` can create instance-storage backed
    AMIs.
* VMware builder now works with Workstation 9 on Linux.

### IMPROVEMENTS:

* builder/amazon/all: Ctrl-C while waiting for state change works
* builder/amazon/ebs: Can now launch instances into a VPC for added protection. [GH-210]
* builder/virtualbox,vmware: Add backspace, delete, and F1-F12 keys to the boot
    command.
* builder/virtualbox: massive performance improvements with big ISO files because
    an expensive copy is avoided. [GH-202]
* builder/vmware: CD is removed prior to exporting final machine. [GH-198]

### BUG FIXES:

* builder/amazon/all: Gracefully handle when AMI appears to not exist
    while AWS state is propagating. [GH-207]
* builder/virtualbox: Trim carriage returns for Windows to properly
    detect VM state on Windows. [GH-218]
* core: build names no longer cause invalid config errors. [GH-197]
* command/build: If any builds fail, exit with non-zero exit status.
* communicator/ssh: SCP exit codes are tested and errors are reported. [GH-195]
* communicator/ssh: Properly change slash direction for Windows hosts. [GH-218]

## 0.2.0 (July 16, 2013)

### BACKWARDS INCOMPATIBILITIES:

* "iso_md5" in the virtualbox and vmware builders is replaced with
    "iso_checksum" and "iso_checksum_type" (with the latter set to "md5").
    See the announce below on `packer fix` to automatically fix your templates.

### FEATURES:

* **NEW COMMAND:** `packer fix` will attempt to fix templates from older
    versions of Packer that are now broken due to backwards incompatibilities.
    This command will fix the backwards incompatibilities introduced in this
    version.
* Amazon EBS builder can now optionally use a pre-made security group
    instead of randomly generating one.
* DigitalOcean API key and client IDs can now be passed in as
    environmental variables. See the documentation for more details.
* VirtualBox and VMware can now have `floppy_files` specified to attach
    floppy disks when booting. This allows for unattended Windows installs.
* `packer build` has a new `-force` flag that forces the removal of
    existing artifacts if they exist. [GH-173]
* You can now log to a file (instead of just stderr) by setting the
    `PACKER_LOG_FILE` environmental variable. [GH-168]
* Checksums other than MD5 can now be used. SHA1 and SHA256 can also
    be used. See the documentation on `iso_checksum_type` for more info. [GH-175]

### IMPROVEMENTS:

* core: invalid keys in configuration are now considered validation
    errors. [GH-104]
* core: all builders now share a common SSH connection core, improving
    SSH reliability over all the builders.
* amazon-ebs: Credentials will come from IAM role if available. [GH-160]
* amazon-ebs: Verify the source AMI is EBS-backed before launching. [GH-169]
* shell provisioner: the build name and builder type are available in
    the `PACKER_BUILD_NAME` and `PACKER_BUILDER_TYPE` env vars by default,
    respectively. [GH-154]
* vmware: error if shutdown command has non-zero exit status.

### BUG FIXES:

* core: UI messages are now properly prefixed with spaces again.
* core: If SSH connection ends, re-connection attempts will take
    place. [GH-152]
* virtualbox: "paused" doesn't mean the VM is stopped, improving
    shutdown detection.
* vmware: error if guest IP could not be detected. [GH-189]

## 0.1.5 (July 7, 2013)

### FEATURES:

* "file" uploader will upload files from the machine running Packer to the
    remote machine.
* VirtualBox guest additions URL and checksum can now be specified, allowing
    the VirtualBox builder to have the ability to be used completely offline.

### IMPROVEMENTS:

* core: If SCP is not available, a more descriptive error message
    is shown telling the user. [GH-127]
* shell: Scripts are now executed by default according to their shebang,
    not with `/bin/sh`. [GH-105]
* shell: You can specify what interpreter you want inline scripts to
    run with `inline_shebang`.
* virtualbox: Delete the packer-made SSH port forwarding prior to
    exporting the VM.

### BUG FIXES:

* core: Non-200 response codes on downloads now show proper errors.
    [GH-141]
* amazon-ebs: SSH handshake is retried. [GH-130]
* vagrant: The `BuildName` template property works properly in
    the output path.
* vagrant: Properly configure the provider-specific post-processors so
    things like `vagrantfile_template` work. [GH-129]
* vagrant: Close filehandles when copying files so Windows can
    rename files. [GH-100]

## 0.1.4 (July 2, 2013)

### FEATURES:

* virtualbox: Can now be built headless with the "Headless" option. [GH-99]
* virtualbox: <wait5> and <wait10> codes for waiting 5 and 10 seconds
    during the boot sequence, respectively. [GH-97]
* vmware: Can now be built headless with the "Headless" option. [GH-99]
* vmware: <wait5> and <wait10> codes for waiting 5 and 10 seconds
    during the boot sequence, respectively. [GH-97]
* vmware: Disks are defragmented and compacted at the end of the build.
    This can be disabled using "skip_compaction"

### IMPROVEMENTS:

* core: Template syntax errors now show line and character number. [GH-56]
* amazon-ebs: Access key and secret access key default to
    environmental variables. [GH-40]
* virtualbox: Send password for keyboard-interactive auth. [GH-121]
* vmware: Send password for keyboard-interactive auth. [GH-121]

### BUG FIXES:

* vmware: Wait until shut down cleans up properly to avoid corrupt
    disk files. [GH-111]

## 0.1.3 (July 1, 2013)

### FEATURES:

* The VMware builder can now upload the VMware tools for you into
    the VM. This is opt-in, you must specify the `tools_upload_flavor`
    option. See the website for more documentation.

### IMPROVEMENTS:

* digitalocean: Errors contain human-friendly error messages. [GH-85]

### BUG FIXES:

* core: More plugin server fixes that avoid hangs on OS X 10.7. [GH-87]
* vagrant: AWS boxes will keep the AMI artifact around. [GH-55]
* virtualbox: More robust version parsing for uploading guest additions. [GH-69]
* virtualbox: Output dir and VM name defaults depend on build name,
    avoiding collisions. [GH-91]
* vmware: Output dir and VM name defaults depend on build name,
    avoiding collisions. [GH-91]

## 0.1.2 (June 29, 2013)

### IMPROVEMENTS:

* core: Template doesn't validate if there are no builders.
* vmware: Delete any VMware files in the VM that aren't necessary for
    it to function.

### BUG FIXES:

* core: Plugin servers consider a port in use if there is any
    error listening to it. This fixes I18n issues and Windows. [GH-58]
* amazon-ebs: Sleep between checking instance state to avoid
    RequestLimitExceeded. [GH-50]
* vagrant: Rename VirtualBox ovf to "box.ovf" [GH-64]
* vagrant: VMware boxes have the correct provider type.
* vmware: Properly populate files in artifact so that the Vagrant
    post-processor works. [GH-63]

## 0.1.1 (June 28, 2013)

### BUG FIXES:

* core: plugins listen explicitly on 127.0.0.1, fixing odd hangs. [GH-37]
* core: fix race condition on verifying checksum of large ISOs which
    could cause panics. [GH-52]
* virtualbox: `boot_wait` defaults to "10s" rather than 0. [GH-44]
* virtualbox: if `http_port_min` and max are the same, it will no longer
    panic. [GH-53]
* vmware: `boot_wait` defaults to "10s" rather than 0. [GH-44]
* vmware: if `http_port_min` and max are the same, it will no longer
    panic. [GH-53]

## 0.1.0 (June 28, 2013)

* Initial release

