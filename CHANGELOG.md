## 0.6.1 (unreleased)

IMPROVEMENTS:

  * builder/ansible: Add `playbook_dir` option. [GH-1000]
  * builder/openstack: Skip certificate verification. [GH-1121]
  * builder/virtualbox/all: Attempt to use local guest additions ISO
      before downloading from internet. [GH-1123]
  * builder/vmware/all: Add `vmx_data_post` for modifying VMX data
      after shutdown. [GH-1149]
  * builder/vmware/vmx: Supports tools uploading. [GH-1154]

BUG FIXES:

  * core: `isotime` is the same time during the entire build. [GH-1153]
  * builder/parallels: Do not delete entire CDROM device. [GH-1115]
  * builder/virtualbox-ovf: Supports guest additions options. [GH-1120]
  * builder/vmware: Remote ESXi builder now uploads floppy. [GH-1106]

## 0.6.0 (May 2, 2014)

FEATURES:

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

IMPROVEMENTS:

  * core: RPC transport between plugins switched to MessagePack
  * core: Templates array values can now be comma separated strings.
      Most importantly, this allows for user variables to fill
      array configurations. [GH-950]
  * builder/amazon: Added `ssh_private_key_file` option [GH-971]
  * builder/amazon: Added `ami_virtualization_type` option [GH-1021]
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
      commands just before exporting [GH-664]
  * builder/virtualbox: Floppy disk files list can also include globs
      and directories. [GH-1086]
  * builder/vmware: Workstation 10 support for Linux. [GH-900]
  * builder/vmware: add cloning support on Windows [GH-824]
  * builder/vmware: Floppy disk files list can also include globs
      and directories. [GH-1086]
  * command/build: Added `-parallel` flag so you can disable parallelization
    with `-no-parallel`. [GH-924]
  * post-processors/vsphere: `disk_mode` option. [GH-778]
  * provisioner/ansible: Add `inventory_file` option [GH-1006]
  * provisioner/chef-client: Add `validation_client_name` option. [GH-1056]

BUG FIXES:

  * core: Errors are properly shown when adding bad floppy files. [GH-1043]
  * core: Fix some URL parsing issues on Windows.
  * core: Create Cache directory only when it is needed. [GH-367]
  * builder/amazon-instance: Use S3Endpoint for ec2-upload-bundle arg,
      which works for every region. [GH-904]
  * builder/digitalocean: updated default image_id [GH-1032]
  * builder/googlecompute: Create persistent disk as boot disk via
      API v1 [GH-1001]
  * builder/openstack: Return proper error on invalid instance states [GH-1018]
  * builder/virtualbox-iso: Retry unregister a few times to deal with
      VBoxManage randomness. [GH-915]
  * provisioner/ansible: Fix paths when provisioning Linux from
      Windows [GH-963]
  * provisioner/ansible: set cwd to staging directory [GH-1016]
  * provisioners/chef-client: Don't chown directory with Ubuntu. [GH-939]
  * provisioners/chef-solo: Deeply nested JSON works properly. [GH-1076]
  * provisioners/shell: Env var values can have equal signs. [GH-1045]
  * provisioners/shell: chmod the uploaded script file to 0777. [GH-994]
  * post-processor/docker-push: Allow repositories with ports. [GH-923]
  * post-processor/vagrant: Create parent directories for `output` path [GH-1059]
  * post-processor/vsphere: datastore, network, and folder are no longer
      required. [GH-1091]

## 0.5.2 (02/21/2014)

FEATURES:

* **New post-processor:** `docker-import` - Import a Docker image
  and give it a specific repository/tag.
* **New post-processor:** `docker-push` - Push an imported image to
  a registry.

IMPROVEMENTS:

* core: Most downloads made by Packer now use a custom user agent. [GH-803]
* builder/googlecompute: SSH private key will be saved to disk if `-debug`
  is specified. [GH-867]
* builder/qemu: Can specify the name of the qemu binary. [GH-854]
* builder/virtualbox-ovf: Can specify import options such as "keepallmacs".
  [GH-883]

BUG FIXES:

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

BUG FIXES:

* core: If a stream ID loops around, don't let it use stream ID 0 [GH-767]
* core: Fix issue where large writes to plugins would result in stream
  corruption. [GH-727]
* builders/virtualbox-ovf: `shutdown_timeout` config works. [GH-772]
* builders/vmware-iso: Remote driver works properly again. [GH-773]

## 0.5.0 (12/30/2013)

BACKWARDS INCOMPATIBILITIES:

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

FEATURES:

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

IMPROVEMENTS:

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

BUG FIXES:

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

IMPROVEMENTS:

* builder/amazon/ebs: New option allows associating a public IP with
  non-default VPC instances. [GH-660]
* builder/openstack: A "proxy\_url" setting was added to define an HTTP
  proxy to use when building with this builder. [GH-637]

BUG FIXES:

* core: Don't change background color on CLI anymore, making things look
  a tad nicer in some terminals.
* core: multiple ISO URLs works properly in all builders. [GH-683]
* builder/amazon/chroot: Block when obtaining file lock to allow
  parallel builds. [GH-689]
* builder/amazon/instance: Add location flag to upload bundle command
  so that building AMIs works out of us-east-1 [GH-679]
* builder/qemu: Qemu arguments are templated. [GH-688]
* builder/vmware: Cleanup of VMX keys works properly so cd-rom won't
  get stuck with ISO. [GH-685]
* builder/vmware: File cleanup is more resilient to file delete races
  with the operating system. [GH-675]
* provisioner/puppet-masterless: Check for hiera config path existence
  properly. [GH-656]

## 0.4.0 (November 19, 2013)

FEATURES:

* Docker builder: build and export Docker containers, easily provisioned
  with any of the Packer built-in provisioners.
* QEMU builder: builds a new VM compatible with KVM or Xen using QEMU.
* Remote ESXi builder: builds a VMware VM using ESXi remotely using only
  SSH to an ESXi machine directly.
* vSphere post-processor: Can upload VMware artifacts to vSphere
* Vagrant post-processor can now make DigitalOcean provider boxes. [GH-504]

IMPROVEMENTS:

* builder/amazon/all: Can now specify a list of multiple security group
  IDs to apply. [GH-499]
* builder/amazon/all: AWS API requests are now retried when a temporary
  network error occurs as well as 500 errors. [GH-559]
* builder/virtualbox: Use VBOX\_INSTALL\_PATH env var on Windows to find
  VBoxManage. [GH-628]
* post-processor/vagrant: skips gzip compression when compression_level=0
* provisioner/chef-solo: Encrypted data bag support [GH-625]

BUG FIXES:

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

FEATURES:

* builder/amazon/ebs: Ability to specify which availability zone to create
  instance in. [GH-536]

IMPROVEMENTS:

* core: builders can now give warnings during validation. warnings won't
  fail the build but may hint at potential future problems.
* builder/digitalocean: Can now specify a droplet name
* builder/virtualbox: Can now disable guest addition download entirely
  by setting "guest_additions_mode" to "disable" [GH-580]
* builder/virtualbox,vmware: ISO urls can now be https [GH-587]
* builder/virtualbox,vmware: Warning if shutdown command is not specified,
  since it is a common case of data loss.

BUG FIXES:

* core: Won't panic when writing to a bad pipe. [GH-560]
* builder/amazon/all: Properly scrub access key and secret key from logs.
  [GH-554]
* builder/openstack: Properly scrub password from logs [GH-554]
* builder/virtualbox: No panic if SSH host port min/max is the same. [GH-594]
* builder/vmware: checks if `ifconfig` is in `/sbin` [GH-591]
* builder/vmware: Host IP lookup works for non-C locales. [GH-592]
* common/uuid: Use cryptographically secure PRNG when generating
  UUIDs. [GH-552]
* communicator/ssh: File uploads that exceed the size of memory no longer
  cause crashes. [GH-561]

## 0.3.10 (October 20, 2013)

FEATURES:

* Ansible provisioner

IMPROVEMENTS:

* post-processor/vagrant: support instance-store AMIs built by Packer. [GH-502]
* post-processor/vagrant: can now specify compression level to use
  when creating the box. [GH-506]

BUG FIXES:

* builder/all: timeout waiting for SSH connection is a failure. [GH-491]
* builder/amazon: Scrub sensitive data from the logs. [GH-521]
* builder/amazon: Handle the situation where an EC2 instance might not
  be immediately available. [GH-522]
* builder/amazon/chroot: Files copied into the chroot remove destination
  before copy, fixing issues with dangling symlinks. [GH-500]
* builder/digitalocean: don't panic if erroneous API response doesn't
  contain error message. [GH-492]
* builder/digitalocean: scrub API keys from config debug output [GH-516]
* builder/virtualbox: error if VirtualBox version cant be detected. [GH-488]
* builder/virtualbox: detect if vboxdrv isn't properly setup. [GH-488]
* builder/virtualbox: sleep a bit before export to ensure the sesssion
  is unlocked. [GH-512]
* builder/virtualbox: create SATA drives properly on VirtualBox 4.3 [GH-547]
* builder/virtualbox: support user templates in SSH key path. [GH-539]
* builder/vmware: support user templates in SSH key path. [GH-539]
* communicator/ssh: Fix issue where a panic could arise from a nil
  dereference. [GH-525]
* post-processor/vagrant: Fix issue with VirtualBox OVA. [GH-548]
* provisioner/salt: Move salt states to correct remote directory. [GH-513]
* provisioner/shell: Won't block on certain scripts on Windows anymore.
  [GH-507]

## 0.3.9 (October 2, 2013)

FEATURES:

* The Amazon chroot builder is now able to run without any `sudo` privileges
  by using the "command_wrapper" configuration. [GH-430]
* Chef provisioner supports environments. [GH-483]

BUG FIXES:

* core: default user variable values don't need to be strings. [GH-456]
* builder/amazon-chroot: Fix errors with waitin for state change. [GH-459]
* builder/digitalocean: Use proper error message JSON key (DO API change).
* communicator/ssh: SCP uploads now work properly when directories
  contain symlinks. [GH-449]
* provisioner/chef-solo: Data bags and roles path are now properly
  populated when set. [GH-470]
* provisioner/shell: Windows line endings are actually properly changed
  to Unix line endings. [GH-477]

## 0.3.8 (September 22, 2013)

FEATURES:

* core: You can now specify `only` and `except` configurations on any
  provisioner or post-processor to specify a list of builds that they
  are valid for. [GH-438]
* builders/virtualbox: Guest additions can be attached rather than uploaded,
  easier to handle for Windows guests. [GH-405]
* provisioner/chef-solo: Ability to specify a custom Chef configuration
  template.
* provisioner/chef-solo: Roles and data bags support. [GH-348]

IMPROVEMENTS:

* core: User variables can now be used for integer, boolean, etc.
  values. [GH-418]
* core: Plugins made with incompatible versions will no longer load.
* builder/amazon/all: Interrupts work while waiting for AMI to be ready.
* provisioner/shell: Script line-endings are automatically converted to
  Unix-style line-endings. Can be disabled by setting "binary" to "true".
  [GH-277]

BUG FIXES:

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

BACKWARDS INCOMPATIBILITIES:

* The "event_delay" option for the DigitalOcean builder is now gone.
  The builder automatically waits for events to go away. Run your templates
  through `packer fix` to get rid of these.

FEATURES:

* **NEW PROVISIONER:** `puppet-masterless`. You can now provision with
  a masterless Puppet setup. [GH-234]
* New globally available template function: `uuid`. Generates a new random
  UUID.
* New globally available template function: `isotime`. Generates the
  current time in ISO standard format.
* New Amazon template function: `clean_ami_name`. Substitutes '-' for
  characters that are illegal to use in an AMI name.

IMPROVEMENTS:

* builder/amazon/all: Ability to specify the format of the temporary
  keypair created. [GH-389]
* builder/amazon/all: Support the NoDevice flag for block mappings. [GH-396]
* builder/digitalocean: Retry on any pending event errors.
* builder/openstack: Can now specify a project. [GH-382]
* builder/virtualbox: Can now attach hard drive over SATA. [GH-391]
* provisioner/file: Can now upload directories. [GH-251]

BUG FIXES:

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

FEATURES:

* User variables can now be specified as "required", meaning the user
  MUST specify a value. Just set the default value to "null". [GH-374]

IMPROVEMENTS:

* core: Much improved interrupt handling. For example, interrupts now
  cancel much more quickly within provisioners.
* builder/amazon: In `-debug` mode, the keypair used will be saved to
  the current directory so you can access the machine. [GH-373]
* builder/amazon: In `-debug` mode, the DNS is outputted.
* builder/openstack: IPv6 addresses supported for SSH. [GH-379]
* communicator/ssh: Support for private keys encrypted using PKCS8. [GH-376]
* provisioner/chef-solo: You can now use user variables in the `json`
  configuration for Chef. [GH-362]

BUG FIXES:

* core: Concurrent map access is completely gone, fixing rare issues
  with runtime memory corruption. [GH-307]
* core: Fix possible panic when ctrl-C during provisioner run.
* builder/digitalocean: Retry destroy a few times because DO sometimes
  gives false errors.
* builder/openstack: Properly handle the case no image is made. [GH-375]
* builder/openstack: Specifying a region is now required in a template.
* provisioners/salt-masterless: Use filepath join to properly join paths.

## 0.3.5 (August 28, 2013)

FEATURES:

* **NEW BUILDER:** `openstack`. You can now build on OpenStack. [GH-155]
* **NEW PROVISIONER:** `chef-solo`. You can now provision with Chef
  using `chef-solo` from local cookbooks.
* builder/amazon: Copy AMI to multiple regions with `ami_regions`. [GH-322]
* builder/virtualbox,vmware: Can now use SSH keys as an auth mechanism for
  SSH using `ssh_key_path`. [GH-70]
* builder/virtualbox,vmware: Support SHA512 as a checksum type. [GH-356]
* builder/vmware: The root hard drive type can now be specified with
  "disk_type_id" for advanced users. [GH-328]
* provisioner/salt-masterless: Ability to specfy a minion config. [GH-264]
* provisioner/salt-masterless: Ability to upload pillars. [GH-353]

IMPROVEMENTS:

* core: Output message when Ctrl-C received that we're cleaning up. [GH-338]
* builder/amazon: Tagging now works with all amazon builder types.
* builder/vmware: Option `ssh_skip_request_pty` for not requesting a PTY
  for the SSH connection. [GH-270]
* builder/vmware: Specify a `vmx_template_path` in order to customize
  the generated VMX. [GH-270]
* command/build: Machine-readable output now contains build errors, if any.
* command/build: An "end" sentinel is outputted in machine-readable output
  for artifact listing so it is easier to know when it is over.

BUG FIXES:

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

IMPROVEMENTS:

* post-processor/vagrant: the file being compressed will be shown
  in the UI [GH-314]

BUG FIXES:

* core: Avoid panics when double-interrupting Packer.
* provisioner/shell: Retry shell script uploads, making reboots more
  robust if they happen to fail in this stage. [GH-282]

## 0.3.3 (August 19, 2013)

FEATURES:

* builder/virtualbox: support exporting in OVA format. [GH-309]

IMPROVEMENTS:

* core: All HTTP downloads across Packer now support the standard
  proxy environmental variables (`HTTP_PROXY`, `NO_PROXY`, etc.) [GH-252]
* builder/amazon: API requests will use HTTP proxy if specified by
  enviromental variables.
* builder/digitalocean: API requests will use HTTP proxy if specified
  by environmental variables.

BUG FIXES:

* core: TCP connection between plugin processes will keep-alive. [GH-312]
* core: No more "unused key keep_input_artifact" for post processors [GH-310]
* post-processor/vagrant: `output_path` templates now work again.

## 0.3.2 (August 18, 2013)

FEATURES:

* New command: `packer inspect`. This command tells you the components of
  a template. It respects the `-machine-readable` flag as well so you can
  parse out components of a template.
* Packer will detect its own crashes (always a bug) and save a "crash.log"
  file.
* builder/virtualbox: You may now specify multiple URLs for an ISO
  using "iso_url" in a template. The URLs will be tried in order.
* builder/vmware: You may now specify multiple URLs for an ISO
  using "iso_url" in a template. The URLs will be tried in order.

IMPROVEMENTS:

* core: built with Go 1.1.2
* core: packer help output now loads much faster.
* builder/virtualbox: guest_additions_url can now use the `Version`
  variable to get the VirtualBox version. [GH-272]
* builder/virtualbox: Do not check for VirtualBox as part of template
  validation; only check at execution.
* builder/vmware: Do not check for VMware as part of template validation;
  only check at execution.
* command/build: A path of "-" will read the template from stdin.
* builder/amazon: add block device mappings [GH-90]

BUG FIXES:

* windows: file URLs are easier to get right as Packer
  has better parsing and error handling for Windows file paths. [GH-284]
* builder/amazon/all: Modifying more than one AMI attribute type no longer
  crashes.
* builder/amazon-instance: send IAM instance profile data. [GH-294]
* builder/digitalocean: API request parameters are properly URL
  encoded. [GH-281]
* builder/virtualbox: dowload progress won't be shown until download
  actually starts. [GH-288]
* builder/virtualbox: floppy files names of 13 characters are now properly
  written to the FAT12 filesystem. [GH-285]
* builder/vmware: dowload progress won't be shown until download
  actually starts. [GH-288]
* builder/vmware: interrupt works while typing commands over VNC.
* builder/virtualbox: floppy files names of 13 characters are now properly
  written to the FAT12 filesystem. [GH-285]
* post-processor/vagrant: Process user variables. [GH-295]

## 0.3.1 (August 12, 2013)

IMPROVEMENTS:

* provisioner/shell: New setting `start_retry_timeout` which is the timeout
  for the provisioner to attempt to _start_ the remote process. This allows
  the shell provisioner to work properly with reboots. [GH-260]

BUG FIXES:

* core: Remote command output containing '\r' now looks much better
  within the Packer output.
* builder/vmware: Fix issue with finding driver files. [GH-279]
* provisioner/salt-masterless: Uploads work properly from Windows. [GH-276]

## 0.3.0 (August 12, 2013)

BACKWARDS INCOMPATIBILITIES:

* All `{{.CreateTime}}` variables within templates (such as for AMI names)
  are now replaced with `{{timestamp}}`. Run `packer fix` to fix your
  templates.

FEATURES:

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

IMPROVEMENTS:

* builder/amazon/all: User data can be passed to start the instances. [GH-253]
* provisioner/salt-masterless: `local_state_tree` is no longer required,
  allowing you to use shell provisioner (or others) to bring this down.
  [GH-269]

BUG FIXES:

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

IMPROVEMENTS:

* builder/amazon/all: Added Amazon AMI tag support [GH-233]

BUG FIXES:

* core: Absolute/relative filepaths on Windows now work for iso_url
  and other settings. [GH-240]
* builder/amazon/all: instance info is refreshed while waiting for SSH,
  allowing Packer to see updated IP/DNS info. [GH-243]

## 0.2.2 (August 1, 2013)

FEATURES:

* New builder: `amazon-chroot` can create EBS-backed AMIs without launching
  a new EC2 instance. This can shave minutes off of the AMI creation process.
  See the docs for more info.
* New provisioner: `salt-masterless` will provision the node using Salt
  without a master.
* The `vmware` builder now works with Workstation 9 on Windows. [GH-222]
* The `vmware` builder now works with Player 5 on Linux. [GH-190]

IMPROVEMENTS:

* core: Colors won't be outputted on Windows unless in Cygwin.
* builder/amazon/all: Added `iam_instance_profile` to launch the source
  image with a given IAM profile. [GH-226]

BUG FIXES:

* builder/virtualbox,vmware: relative paths work properly as URL
  configurations. [GH-215]
* builder/virtualbox,vmware: fix race condition in deleting the output
  directory on Windows by retrying.

## 0.2.1 (July 26, 2013)

FEATURES:

* New builder: `amazon-instance` can create instance-storage backed
  AMIs.
* VMware builder now works with Workstation 9 on Linux.

IMPROVEMENTS:

* builder/amazon/all: Ctrl-C while waiting for state change works
* builder/amazon/ebs: Can now launch instances into a VPC for added protection [GH-210]
* builder/virtualbox,vmware: Add backspace, delete, and F1-F12 keys to the boot
  command.
* builder/virtualbox: massive performance improvements with big ISO files because
  an expensive copy is avoided. [GH-202]
* builder/vmware: CD is removed prior to exporting final machine. [GH-198]

BUG FIXES:

* builder/amazon/all: Gracefully handle when AMI appears to not exist
  while AWS state is propogating. [GH-207]
* builder/virtualbox: Trim carriage returns for Windows to properly
  detect VM state on Windows. [GH-218]
* core: build names no longer cause invalid config errors. [GH-197]
* command/build: If any builds fail, exit with non-zero exit status.
* communicator/ssh: SCP exit codes are tested and errors are reported. [GH-195]
* communicator/ssh: Properly change slash direction for Windows hosts. [GH-218]

## 0.2.0 (July 16, 2013)

BACKWARDS INCOMPATIBILITIES:

* "iso_md5" in the virtualbox and vmware builders is replaced with
  "iso_checksum" and "iso_checksum_type" (with the latter set to "md5").
  See the announce below on `packer fix` to automatically fix your templates.

FEATURES:

* **NEW COMMAND:** `packer fix` will attempt to fix templates from older
  versions of Packer that are now broken due to backwards incompatibilities.
  This command will fix the backwards incompatibilities introduced in this
  version.
* Amazon EBS builder can now optionally use a pre-made security group
  instead of randomly generating one.
* DigitalOcean API key and client IDs can now be passed in as
  environmental variables. See the documentatin for more details.
* VirtualBox and VMware can now have `floppy_files` specified to attach
  floppy disks when booting. This allows for unattended Windows installs.
* `packer build` has a new `-force` flag that forces the removal of
  existing artifacts if they exist. [GH-173]
* You can now log to a file (instead of just stderr) by setting the
  `PACKER_LOG_FILE` environmental variable. [GH-168]
* Checksums other than MD5 can now be used. SHA1 and SHA256 can also
  be used. See the documentation on `iso_checksum_type` for more info. [GH-175]

IMPROVEMENTS:

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

BUG FIXES:

* core: UI messages are now properly prefixed with spaces again.
* core: If SSH connection ends, re-connection attempts will take
  place. [GH-152]
* virtualbox: "paused" doesn't mean the VM is stopped, improving
  shutdown detection.
* vmware: error if guest IP could not be detected. [GH-189]

## 0.1.5 (July 7, 2013)

FEATURES:

* "file" uploader will upload files from the machine running Packer to the
  remote machine.
* VirtualBox guest additions URL and checksum can now be specified, allowing
  the VirtualBox builder to have the ability to be used completely offline.

IMPROVEMENTS:

* core: If SCP is not available, a more descriptive error message
  is shown telling the user. [GH-127]
* shell: Scripts are now executed by default according to their shebang,
  not with `/bin/sh`. [GH-105]
* shell: You can specify what interpreter you want inline scripts to
  run with `inline_shebang`.
* virtualbox: Delete the packer-made SSH port forwarding prior to
  exporting the VM.

BUG FIXES:

* core: Non-200 response codes on downloads now show proper errors.
  [GH-141]
* amazon-ebs: SSH handshake is retried. [GH-130]
* vagrant: The `BuildName` template propery works properly in
  the output path.
* vagrant: Properly configure the provider-specific post-processors so
  things like `vagrantfile_template` work. [GH-129]
* vagrant: Close filehandles when copying files so Windows can
  rename files. [GH-100]

## 0.1.4 (July 2, 2013)

FEATURES:

* virtualbox: Can now be built headless with the "Headless" option. [GH-99]
* virtualbox: <wait5> and <wait10> codes for waiting 5 and 10 seconds
  during the boot sequence, respectively. [GH-97]
* vmware: Can now be built headless with the "Headless" option. [GH-99]
* vmware: <wait5> and <wait10> codes for waiting 5 and 10 seconds
  during the boot sequence, respectively. [GH-97]
* vmware: Disks are defragmented and compacted at the end of the build.
  This can be disabled using "skip_compaction"

IMPROVEMENTS:

* core: Template syntax errors now show line and character number. [GH-56]
* amazon-ebs: Access key and secret access key default to
  environmental variables. [GH-40]
* virtualbox: Send password for keyboard-interactive auth [GH-121]
* vmware: Send password for keyboard-interactive auth [GH-121]

BUG FIXES:

* vmware: Wait until shut down cleans up properly to avoid corrupt
  disk files [GH-111]

## 0.1.3 (July 1, 2013)

FEATURES:

* The VMware builder can now upload the VMware tools for you into
  the VM. This is opt-in, you must specify the `tools_upload_flavor`
  option. See the website for more documentation.

IMPROVEMENTS:

* digitalocean: Errors contain human-friendly error messages. [GH-85]

BUG FIXES:

* core: More plugin server fixes that avoid hangs on OS X 10.7 [GH-87]
* vagrant: AWS boxes will keep the AMI artifact around [GH-55]
* virtualbox: More robust version parsing for uploading guest additions. [GH-69]
* virtualbox: Output dir and VM name defaults depend on build name,
  avoiding collisions. [GH-91]
* vmware: Output dir and VM name defaults depend on build name,
  avoiding collisions. [GH-91]

## 0.1.2 (June 29, 2013)

IMPROVEMENTS:

* core: Template doesn't validate if there are no builders.
* vmware: Delete any VMware files in the VM that aren't necessary for
  it to function.

BUG FIXES:

* core: Plugin servers consider a port in use if there is any
  error listening to it. This fixes I18n issues and Windows. [GH-58]
* amazon-ebs: Sleep between checking instance state to avoid
  RequestLimitExceeded [GH-50]
* vagrant: Rename VirtualBox ovf to "box.ovf" [GH-64]
* vagrant: VMware boxes have the correct provider type.
* vmware: Properly populate files in artifact so that the Vagrant
  post-processor works. [GH-63]

## 0.1.1 (June 28, 2013)

BUG FIXES:

* core: plugins listen explicitly on 127.0.0.1, fixing odd hangs. [GH-37]
* core: fix race condition on verifying checksum of large ISOs which
  could cause panics [GH-52]
* virtualbox: `boot_wait` defaults to "10s" rather than 0. [GH-44]
* virtualbox: if `http_port_min` and max are the same, it will no longer
  panic [GH-53]
* vmware: `boot_wait` defaults to "10s" rather than 0. [GH-44]
* vmware: if `http_port_min` and max are the same, it will no longer
  panic [GH-53]

## 0.1.0 (June 28, 2013)

* Initial release
