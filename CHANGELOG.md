## 0.11.0 (October 21, 2016)

BACKWARDS INCOMPATIBILITIES:

  * VNC and VRDP-like features in VirtualBox, VMware, and QEMU now configurable
    but bind to 127.0.0.1 by default to improve security. See the relevant
    builder docs for more info.
  * Docker builder requires Docker > 1.3
  * provisioner/chef-solo: default staging directory renamed to
      `packer-chef-solo`. [#3971]

FEATURES:

  * **New Checksum post-processor**: Create a checksum file from your build
      artifacts as part of your build. [#3492] [#3790]
  * **New build flag** `-on-error` to allow inspection and keeping artifacts on
      builder errors. [#3885]
  * **New Google Compute Export post-processor**: exports an image from
      a Packer googlecompute builder run and uploads it to Google Cloud
      Storage.  [#3760]
  * **New Manifest post-processor**: writes metadata about packer's output
      artifacts data to a JSON file. [#3651]


IMPROVEMENTS:

  * builder/amazon: Added `disable_stop_instance` option to prevent automatic
      shutdown when the build is complete. [#3352]
  * builder/amazon: Added `shutdown_behavior` option to support `stop` or
      `terminate` at the end of the build. [#3556]
  * builder/amazon: Added `skip_region_validation` option to allow newer or
      custom AWS regions. [#3598]
  * builder/amazon: Added `us-east-2` and `ap-south-1` regions. [#4021]
      [#3663]
  * builder/amazon: Support building from scratch with amazon-chroot builder.
      [#3855] [#3895]
  * builder/amazon: Support create an AMI with an `encrypted_boot` volume.
      [#3382]
  * builder/azure: Add `os_disk_size_gb`. [#3995]
  * builder/azure: Add location to setup script. [#3803]
  * builder/azure: Allow user to set custom data. [#3996]
  * builder/azure: Made `tenant_id` optional. [#3643]
  * builder/azure: Now pre-validates `capture_container_name` and
      `capture_name_prefix` [#3537]
  * builder/azure: Removed superfluous polling code for deployments. [#3638]
  * builder/azure: Support for a user defined VNET. [#3683]
  * builder/azure: Support for custom images. [#3575]
  * builder/azure: tag all resources. [#3764]
  * builder/digitalocean: Added `user_data_file` support. [#3933]
  * builder/digitalocean: Fixes timeout waiting for snapshot. [#3868]
  * builder/digitalocean: Use `state_timeout` for unlock and off transitions.
      [#3444]
  * builder/docker: Improved support for Docker pull from Amazon ECR. [#3856]
  * builder/google: Add `-force` option to delete old image before creating new
      one. [#3918]
  * builder/google: Add image license metadata. [#3873]
  * builder/google: Added support for `image_family` [#3531]
  * builder/google: Added support for startup scripts. [#3639]
  * builder/google: Create passwords for Windows instances. [#3932]
  * builder/google: Enable to select NVMe images. [#3338]
  * builder/google: Signal that startup script fished via metadata. [#3873]
  * builder/google: Use gcloud application default credentials. [#3655]
  * builder/google: provision VM without external IP address. [#3774]
  * builder/null: Can now be used with WinRM. [#2525]
  * builder/openstack: Added support for `ssh_password` instead of generating
      ssh keys. [#3976]
  * builder/parallels: Add support for ctrl, shift and alt keys in
      `boot_command`.  [#3767]
  * builder/parallels: Copy directories recursively with `floppy_dirs`.
      [#2919]
  * builder/parallels: Now pauses between `boot_command` entries when running
      with `-debug` [#3547]
  * builder/parallels: Support future versions of Parallels by using the latest
      driver. [#3673]
  * builder/qemu: Add support for ctrl, shift and alt keys in `boot_command`.
      [#3767]
  * builder/qemu: Added `vnc_bind_address` option. [#3574]
  * builder/qemu: Copy directories recursively with `floppy_dirs`. [#2919]
  * builder/qemu: Now pauses between `boot_command` entries when running with
      `-debug` [#3547]
  * builder/qemu: Specify disk format when starting qemu. [#3888]
  * builder/virtualbox-iso: Added `hard_drive_nonrotational` and
      `hard_drive_discard` options to enable trim/discard. [#4013]
  * builder/virtualbox-iso: Added `keep_registed` option to skip cleaning up
      the image. [#3954]
  * builder/virtualbox: Add support for ctrl, shift and alt keys in
      `boot_command`.  [#3767]
  * builder/virtualbox: Added `post_shutdown_delay` option to wait after
      shutting down to prevent issues removing floppy drive. [#3952]
  * builder/virtualbox: Added `vrdp_bind_address` option. [#3566]
  * builder/virtualbox: Copy directories recursively with `floppy_dirs`.
      [#2919]
  * builder/virtualbox: Now pauses between `boot_command` entries when running
      with `-debug` [#3542]
  * builder/vmware-vmx: Added `tools_upload_flavor` and `tools_upload_path` to
      docs.
  * builder/vmware: Add support for ctrl, shift and alt keys in `boot_command`.
      [#3767]
  * builder/vmware: Added `vnc_bind_address` option. [#3565]
  * builder/vmware: Adds passwords for VNC. [#2325]
  * builder/vmware: Copy directories recursively with `floppy_dirs`. [#2919]
  * builder/vmware: Handle connection to VM with more than one NIC on ESXi
      [#3347]
  * builder/vmware: Now paused between `boot_command` entries when running with
      `-debug` [#3542]
  * core: Supress plugin discovery from plugins. [#4002]
  * core: Test floppy disk files actually exist. [#3756]
  * core: setting `PACKER_LOG=0` now disables logging. [#3964]
  * post-processor/amazon-import: Support `ami_name` for naming imported AMI.
      [#3941]
  * post-processor/compress: Added support for bgzf compression. [#3501]
  * post-processor/docker: Improved support for Docker push to Amazon ECR.
      [#3856]
  * post-processor/docker: Preserve tags when running docker push. [#3631]
  * post-processor/vagrant: Added vsphere-esx hosts to supported machine types.
      [#3967]
  * provisioner/ansible-local: Support for ansible-galaxy. [#3350] [#3836]
  * provisioner/ansible: Improved logging and error handling. [#3477]
  * provisioner/ansible: Support scp. [#3861]
  * provisioner/chef: Added `knife_command` option and added a correct default
      value for Windows. [#3622]
  * provisioner/chef: Installs 64bit chef on Windows if available. [#3848]
  * provisioner/file: Now makes destination directory. [#3692]
  * provisioner/puppet: Added `execute_command` option. [#3614]
  * provisioner/salt: Added `custom_state` to specify state to run instead of
      `highstate`. [#3776]
  * provisioner/shell: Added `expect_disconnect` flag to fail if remote
      unexpectedly disconnects. [#4034]
  * scripts: Added `help` target to Makefile. [#3290]
  * vendor: Moving from Godep to govendor. See `CONTRIBUTING.md` for details.
      [#3956]
  * website: code examples now use inconsolata. Improve code font rendering on
      linux.

BUG FIXES:

  * builder/amazon: Add 0.5 cents to discovered spot price. [#3662]
  * builder/amazon: Allow using `ssh_private_key_file` and `ssh_password`.
      [#3953]
  * builder/amazon: Fix packer crash when waiting for SSH. [#3865]
  * builder/amazon: Honor ssh_private_ip flag in EC2-Classic. [#3752]
  * builder/amazon: Properly clean up EBS volumes on failure. [#3789]
  * builder/amazon: Use `temporary_key_pair_name` when specified. [#3739]
  * builder/amazon: add retry logic when creating tags.
  * builder/amazon: retry creating tags on images since the images might take
      some time to become available. [#3938]
  * builder/azure: Fix authorization setup script failing to creating service
      principal. [#3812]
  * builder/azure: check for empty resource group. [#3606]
  * builder/azure: fix token validity test. [#3609]
  * builder/docker: Fix file provisioner dotfile matching. [#3800]
  * builder/docker: fix docker builder with ansible provisioner. [#3476]
  * builder/qemu: Don't fail on communicator set to `none`. [#3681]
  * builder/qemu: Make `ssh_host_port_max` an inclusive bound. [#2784]
  * builder/virtualbox: Make `ssh_host_port_max` an inclusive bound. [#2784]
  * builder/virtualbox: Respect `ssh_host` [#3617]
  * builder/vmware: Do not add remotedisplay.vnc.ip to VMX data on ESXi
      [#3740]
  * builder/vmware: Don't check for poweron errors on ESXi. [#3195]
  * builder/vmware: Re-introduce case sensitive VMX keys. [#2707]
  * builder/vmware: Respect `ssh_host`/`winrm_host` on ESXi. [#3738]
  * command/push: Allows dot (`.`) in image names. [#3937]
  * common/iso_config: fix potential panic when iso checksum url was given but
      not the iso url. [#4004]
  * communicator/ssh: fixed possible panic when reconnecting fails. [#4008]
  * communicator/ssh: handle error case where server closes the connection but
      doesn't give us an error code. [#3966]
  * post-processor/shell-local: Do not set execute bit on artifact file.
      [#3505]
  * post-processor/vsphere: Fix upload failures with vsphere. [#3321]
  * provisioner/ansible: Properly set host key checking even when a custom ENV
      is specified. [#3568]
  * provisioner/file: Fix directory download. [#3899]
  * provisioner/powershell: fixed issue with setting environment variables.
      [#2785]
  * website: improved rendering on iPad. [#3780]

## 0.10.2 (September 20, 2016)

BUG FIXES:

  * Rebuilding with OS X Sierra and go 1.7.1 to fix bug  in Sierra

## 0.10.1 (May 7, 2016)

FEATURES:

  * `azure-arm` builder: Can now build Windows images, and supports additional
    configuration. Please refer to the documentation for details.

IMPROVEMENTS:

  * core: Added support for `ATLAS_CAFILE` and `ATLAS_CAPATH` [#3494]
  * builder/azure: Improved build cancellation and cleanup of partially-
    provisioned resources. [#3461]
  * builder/azure: Improved logging. [#3461]
  * builder/azure: Added support for US Government and China clouds. [#3461]
  * builder/azure: Users may now specify an image version. [#3461]
  * builder/azure: Added device login. [#3461]
  * builder/docker: Added `privileged` build option. [#3475]
  * builder/google: Packer now identifies its version to the service. [#3465]
  * provisioner/shell: Added `remote_folder` and `remote_file` options
    [#3462]
  * post-processor/compress: Added support for `bgzf` format and added
    `format` option. [#3501]

BUG FIXES:

  * core: Fix hang after pressing enter key in `-debug` mode. [#3346]
  * provisioner/chef: Use custom values for remote validation key path
    [#3468]

## 0.10.0 (March 14, 2016)

BACKWARDS INCOMPATIBILITIES:

  * Building Packer now requires go >= 1.5 (>= 1.6 is recommended). If you want
    to continue building with go 1.4 you can remove the `azurearmbuilder` line
    from `command/plugin.go`.

FEATURES:

  * **New `azure-arm` builder**: Build virtual machines in Azure Resource
    Manager

IMPROVEMENTS:

  * builder/google: Added support for `disk_type` [#2830]
  * builder/openstack: Added support for retrieving the Administrator password
    when using WinRM if no `winrm_password` is set. [#3209]
  * provisioner/ansible: Added the `empty_groups` parameter. [#3232]
  * provisioner/ansible: Added the `user` parameter. [#3276]
  * provisioner/ansible: Don't use deprecated ssh option with Ansible 2.0
    [#3291]
  * provisioner/puppet-masterless: Add `ignore_exit_codes` parameter. [#3349]

BUG FIXES:

  * builders/parallels: Handle `output_directory` containing `.` and `..`
    [#3239]
  * provisioner/ansible: os.Environ() should always be passed to the ansible
    command. [#3274]

## 0.9.0 (February 19, 2016)

BACKWARDS INCOMPATIBILITIES:

  * Packer now ships as a single binary, including plugins. If you install
    packer 0.9.0 over a previous packer installation, **you must delete all of
    the packer-* plugin files** or packer will load out-of-date plugins from
    disk.
  * Release binaries are now provided via <https://releases.hashicorp.com>.
  * Packer 0.9.0 is now built with Go 1.6.
  * core: Plugins that implement the Communicator interface must now implement
    a DownloadDir method. [#2618]
  * builder/amazon: Inline `user_data` for EC2 is now base64 encoded
    automatically. [#2539]
  * builder/parallels: `parallels_tools_host_path` and `guest_os_distribution`
    have been replaced by `guest_os_type`; use `packer fix` to update your
    templates. [#2751]

FEATURES:

  * **Chef on Windows**: The chef provisioner now has native support for
    Windows using Powershell and WinRM. [#1215]
  * **New `vmware-esxi` feature**: Packer can now export images from vCloud or
    vSphere during the build. [#1921]
  * **New Ansible Provisioner**: `ansible` provisioner supports remote
    provisioning to keep your build image cleaner. [#1969]
  * **New Amazon Import post-processor**: `amazon-import` allows you to upload an OVA-based VM to Amazon EC2. [#2962]
  * **Shell Local post-processor**: `shell-local` allows you to run shell
    commands on the host after a build has completed for custom packaging or
    publishing of your artifacts. [#2706]
  * **Artifice post-processor**: Override packer artifacts during post-
    processing. This allows you to extract artifacts from a packer builder and
    use them with other post-processors like compress, docker, and Atlas.

IMPROVEMENTS:

  * core: Packer plugins are now compiled into the main binary, reducing file
    size and build times, and making packer easier to install. The overall
    plugin architecture has not changed and third-party plugins can still be
    loaded from disk. Please make sure your plugins are up-to-date! [#2854]
  * core: Packer now indicates line numbers for template parse errors. [#2742]
  * core: Scripts are executed via `/usr/bin/env bash` instead of `/bin/bash`
    for broader compatibility. [#2913]
  * core: `target_path` for builder downloads can now be specified. [#2600]
  * core: WinRM communicator now supports HTTPS protocol. [#3061]
  * core: Template syntax errors now show line, column, offset. [#3180]
  * core: SSH communicator now supports downloading directories. [#2618]
  * builder/amazon: Add support for `ebs_optimized` [#2806]
  * builder/amazon: You can now specify `0` for `spot_price` to switch to on
    demand instances. [#2845]
  * builder/amazon: Added `ap-northeast-2` (Seoul) [#3056]
  * builder/amazon: packer will try to derive the AZ if only a subnet is
    specified. [#3037]
  * builder/digitalocean: doubled instance wait timeouts to power off or
    shutdown (now 4 minutes) and to complete a snapshot (now 20 minutes)
    [#2939]
  * builder/google: `account_file` can now be provided as a JSON string
    [#2811]
  * builder/google: added support for `preemptible` instances. [#2982]
  * builder/google: added support for static external IPs via `address` option
    [#3030]
  * builder/openstack: added retry on WaitForImage 404. [#3009]
  * builder/openstack: Can specify `source_image_name` instead of the ID
    [#2577]
  * builder/openstack: added support for SSH over IPv6. [#3197]
  * builder/parallels: Improve support for Parallels 11. [#2662]
  * builder/parallels: Parallels disks are now compacted by default. [#2731]
  * builder/parallels: Packer will look for Parallels in
    `/Applications/Parallels Desktop.app` if it is not detected automatically
    [#2839]
  * builder/qemu: qcow2 images are now compacted by default. [#2748]
  * builder/qemu: qcow2 images can now be compressed. [#2748]
  * builder/qemu: Now specifies `virtio-scsi` by default. [#2422]
  * builder/qemu: Now checks for version-specific options. [#2376]
  * builder/qemu: Can now bypass disk cache using `iso_skip_cache` [#3105]
  * builder/qemu: `<wait>` in `boot_command` now accepts an arbitrary duration
    like <wait1m30s> [#3129]
  * builder/qemu: Expose `{{ .SSHHostPort }}` in templates. [#2884]
  * builder/virtualbox: Added VRDP for debugging. [#3188]
  * builder/vmware-esxi: Added private key auth for remote builds via
    `remote_private_key_file` [#2912]
  * post-processor/atlas: Added support for compile ID. [#2775]
  * post-processor/docker-import: Can now import Artifice artifacts. [#2718]
  * provisioner/chef: Added `encrypted_data_bag_secret_path` option. [#2653]
  * provisioner/puppet: Added the `extra_arguments` parameter. [#2635]
  * provisioner/salt: Added `no_exit_on_failure`, `log_level`, and improvements
    to salt command invocation. [#2660]

BUG FIXES:

  * core: Random number generator is now seeded. [#2640]
  * core: Packer should now have a lot less race conditions. [#2824]
  * builder/amazon: The `no_device` option for block device mappings is now handled correctly. [#2398]
  * builder/amazon: AMI name validation now matches Amazon's spec. [#2774]
  * builder/amazon: Use snapshot size when volume size is unspecified. [#2480]
  * builder/amazon: Pass AccessKey and SecretKey when uploading bundles for
    instance-backed AMIs. [#2596]
  * builder/parallels: Added interpolation in `prlctl_post` [#2828]
  * builder/vmware: `format` option is now read correctly. [#2892]
  * builder/vmware-esxi: Correct endless loop in destroy validation logic
    [#2911]
  * provisioner/shell: No longer leaves temp scripts behind. [#1536]
  * provisioner/winrm: Now waits for reboot to complete before continuing with provisioning. [#2568]
  * post-processor/artifice: Fix truncation of files downloaded from Docker. [#2793]


## 0.8.6 (Aug 22, 2015)

IMPROVEMENTS:

  * builder/docker: Now supports Download so it can be used with the file
      provisioner to download a file from a container. [#2585]
  * builder/docker: Now verifies that the artifact will be used before the build
      starts, unless the `discard` option is specified. This prevent failures
      after the build completes. [#2626]
  * post-processor/artifice: Now supports glob-like syntax for filenames. [#2619]
  * post-processor/vagrant: Like the compress post-processor, vagrant now uses a
      parallel gzip algorithm to compress vagrant boxes. [#2590]

BUG FIXES:

  * core: When `iso_url` is a local file and the checksum is invalid, the local
      file will no longer be deleted. [#2603]
  * builder/parallels: Fix interpolation in `parallels_tools_guest_path` [#2543]

## 0.8.5 (Aug 10, 2015)

FEATURES:

  * **[Beta]** Artifice post-processor: Override packer artifacts during post-
      processing. This allows you to extract artifacts from a packer builder
      and use them with other post-processors like compress, docker, and Atlas.

IMPROVEMENTS:

  * Many docs have been updated and corrected; big thanks to our contributors!
  * builder/openstack: Add debug logging for IP addresses used for SSH. [#2513]
  * builder/openstack: Add option to use existing SSH keypair. [#2512]
  * builder/openstack: Add support for Glance metadata. [#2434]
  * builder/qemu and builder/vmware: Packer's VNC connection no longer asks for
      an exclusive connection. [#2522]
  * provisioner/salt-masterless: Can now customize salt remote directories. [#2519]

BUG FIXES:

  * builder/amazon: Improve instance cleanup by storing id sooner. [#2404]
  * builder/amazon: Only fetch windows password when using WinRM communicator. [#2538]
  * builder/openstack: Support IPv6 SSH address. [#2450]
  * builder/openstack: Track new IP address discovered during RackConnect. [#2514]
  * builder/qemu: Add 100ms delay between VNC key events. [#2415]
  * post-processor/atlas: atlas_url configuration option works now. [#2478]
  * post-processor/compress: Now supports interpolation in output config. [#2414]
  * provisioner/powershell: Elevated runs now receive environment variables. [#2378]
  * provisioner/salt-masterless: Clarify error messages when we can't create or
      write to the temp directory. [#2518]
  * provisioner/salt-masterless: Copy state even if /srv/salt exists already. [#1699]
  * provisioner/salt-masterless: Make sure /etc/salt exists before writing to it. [#2520]
  * provisioner/winrm: Connect to the correct port when using NAT with
      VirtualBox / VMware. [#2399]

Note: 0.8.3 was pulled and 0.8.4 was skipped.

## 0.8.2 (July 17, 2015)

IMPROVEMENTS:

  * builder/docker: Add option to use a Pty. [#2425]

BUG FIXES:

  * core: Fix crash when `min_packer_version` is specified in a template. [#2385]
  * builder/amazon: Fix EC2 devices being included in EBS mappings. [#2459]
  * builder/googlecompute: Fix default name for GCE images. [#2400]
  * builder/null: Fix error message with missing ssh_host. [#2407]
  * builder/virtualbox: Use --portcount on VirtualBox 5.x. [#2438]
  * provisioner/puppet: Packer now correctly handles a directory for manifest_file. [#2463]
  * provisioner/winrm: Fix potential crash with WinRM. [#2416]

## 0.8.1 (July 2, 2015)

IMPROVEMENTS:

  * builder/amazon: When debug mode is enabled, the Windows administrator
      password for Windows instances will be shown. [#2351]

BUG FIXES:

  * core: `min_packer_version`  field in configs work. [#2356]
  * core: The `build_name` and `build_type` functions work in provisioners. [#2367]
  * core: Handle timeout in SSH handshake. [#2333]
  * command/build: Fix reading configuration from stdin. [#2366]
  * builder/amazon: Fix issue with sharing AMIs when using `ami_users` [#2308]
  * builder/amazon: Fix issue when using multiple Security Groups. [#2381]
  * builder/amazon: Fix for tag creation when creating new ec2 instance. [#2317]
  * builder/amazon: Fix issue with creating AMIs with multiple device mappings. [#2320]
  * builder/amazon: Fix failing AMI snapshot tagging when copying to other
      regions. [#2316]
  * builder/amazon: Fix setting AMI launch permissions. [#2348]
  * builder/amazon: Fix spot instance cleanup to remove the correct request. [#2327]
  * builder/amazon: Fix `bundle_prefix` not interpolating `timestamp` [#2352]
  * builder/amazon-instance: Fix issue with creating AMIs without specifying a
      virtualization type. [#2330]
  * builder/digitalocean: Fix builder using private IP instead of public IP. [#2339]
  * builder/google: Set default communicator settings properly. [#2353]
  * builder/vmware-iso: Setting `checksum_type` to `none` for ESX builds
      now works. [#2323]
  * provisioner/chef: Use knife config file vs command-line params to
      clean up nodes so full set of features can be used. [#2306]
  * post-processor/compress: Fixed crash in compress post-processor plugin. [#2311]

## 0.8.0 (June 23, 2015)

BACKWARDS INCOMPATIBILITIES:

  * core: SSH connection will no longer request a PTY by default. This
      can be enabled per builder.
  * builder/digitalocean: no longer supports the v1 API which has been
      deprecated for some time. Most configurations should continue to
      work as long as you use the `api_token` field for auth.
  * builder/digitalocean: `image`, `region`, and `size` are now required.
  * builder/openstack: auth parameters have been changed to better
      reflect OS terminology. Existing environment variables still work.

FEATURES:

  * **WinRM:** You can now connect via WinRM with almost every builder.
      See the docs for more info. [#2239]
  * **Windows AWS Support:** Windows AMIs can now be built without any
      external plugins: Packer will start a Windows instance, get the
      admin password, and can use WinRM (above) to connect through. [#2240]
  * **Disable SSH:** Set `communicator` to "none" in any builder to disable SSH
      connections. Note that provisioners won't work if this is done. [#1591]
  * **SSH Agent Forwarding:** SSH Agent Forwarding will now be enabled
      to allow access to remote servers such as private git repos. [#1066]
  * **SSH Bastion Hosts:** You can now specify a bastion host for
      SSH access (works with all builders). [#387]
  * **OpenStack v3 Identity:** The OpenStack builder now supports the
      v3 identity API.
  * **Docker builder supports SSH**: The Docker builder now supports containers
      with SSH, just set `communicator` to "ssh" [#2244]
  * **File provisioner can download**: The file provisioner can now download
      files out of the build process. [#1909]
  * **New config function: `build_name`**: The name of the currently running
      build. [#2232]
  * **New config function: `build_type`**: The type of the currently running
      builder. This is useful for provisioners. [#2232]
  * **New config function: `template_dir`**: The directory to the template
      being built. This should be used for template-relative paths. [#54]
  * **New provisioner: shell-local**: Runs a local shell script. [#770]
  * **New provisioner: powershell**: Provision Windows machines
      with PowerShell scripts. [#2243]
  * **New provisioner: windows-shell**: Provision Windows machines with
      batch files. [#2243]
  * **New provisioner: windows-restart**: Restart a Windows machines and
      wait for it to come back online. [#2243]
  * **Compress post-processor supports multiple algorithms:** The compress
      post-processor now supports lz4 compression and compresses gzip in
      parallel for much faster throughput.

IMPROVEMENTS:

  * core: Interrupt handling for SIGTERM signal as well. [#1858]
  * core: HTTP downloads support resuming. [#2106]
  * builder/*: Add `ssh_handshake_attempts` to configure the number of
      handshake attempts done before failure. [#2237]
  * builder/amazon: Add `force_deregister` option for automatic AMI
      deregistration. [#2221]
  * builder/amazon: Now applies tags to EBS snapshots. [#2212]
  * builder/amazon: Clean up orphaned volumes from Source AMIs. [#1783]
  * builder/amazon: Support custom keypairs. [#1837]
  * builder/amazon-chroot: Can now resize the root volume of the resulting
      AMI with the `root_volume_size` option. [#2289]
  * builder/amazon-chroot: Add `mount_options` configuration option for providing
      options to the `mount` command. [#2296]
  * builder/digitalocean: Save SSH key to pwd if debug mode is on. [#1829]
  * builder/digitalocean: User data support. [#2113]
  * builder/googlecompute: Option to use internal IP for connections. [#2152]
  * builder/parallels: Support Parallels Desktop 11. [#2199]
  * builder/openstack: Add `rackconnect_wait` for Rackspace customers to wait for
      RackConnect data to appear
  * buidler/openstack: Add `ssh_interface` option for rackconnect for users that
      have prohibitive firewalls
  * builder/openstack: Flavor names can be used as well as refs
  * builder/openstack: Add `availability_zone` [#2016]
  * builder/openstack: Machine will be stopped prior to imaging if the
      cluster supports the `startstop` extension. [#2223]
  * builder/openstack: Support for user data. [#2224]
  * builder/qemu: Default accelerator to "tcg" on Windows. [#2291]
  * builder/virtualbox: Added option: `ssh_skip_nat_mapping` to skip the
      automatic port forward for SSH and to use the guest port directly. [#1078]
  * builder/virtualbox: Added SCSI support
  * builder/vmware: Support for additional disks. [#1382]
  * builder/vmware: Can now customize the template used for adding disks. [#2254]
  * command/fix: After fixing, the template is validated. [#2228]
  * command/push: Add `-name` flag for specifying name from CLI. [#2042]
  * command/push: Push configuration in templates supports variables. [#1861]
  * post-processor/docker-save: Can be chained. [#2179]
  * post-processor/docker-tag: Support `force` option. [#2055]
  * post-processor/docker-tag: Can be chained. [#2179]
  * post-processor/vsphere: Make more fields optional, support empty
      resource pools. [#1868]
  * provisioner/puppet-masterless: `working_directory` option. [#1831]
  * provisioner/puppet-masterless: `packer_build_name` and
      `packer_build_type` are default facts. [#1878]
  * provisioner/puppet-server: `ignore_exit_codes` option added. [#2280]

BUG FIXES:

  * core: Fix potential panic for post-processor plugin exits. [#2098]
  * core: `PACKER_CONFIG` may point to a non-existent file. [#2226]
  * builder/amazon: Allow spaces in AMI names when using `clean_ami_name` [#2182]
  * builder/amazon: Remove deprecated ec2-upload-bundle paramger. [#1931]
  * builder/amazon: Use IAM Profile to upload bundle if provided. [#1985]
  * builder/amazon: Use correct exit code after SSH authentication failed. [#2004]
  * builder/amazon: Retry finding created instance for eventual
      consistency. [#2129]
  * builder/amazon: If no AZ is specified, use AZ chosen automatically by
      AWS for spot instance. [#2017]
  * builder/amazon: Private key file (only available in debug mode)
      is deleted on cleanup. [#1801]
  * builder/amazon: AMI copy won't copy to the source region. [#2123]
  * builder/amazon: Validate AMI doesn't exist with name prior to build. [#1774]
  * builder/amazon: Improved retry logic around waiting for instances. [#1764]
  * builder/amazon: Fix issues with creating Block Devices. [#2195]
  * builder/amazon/chroot: Retry waiting for disk attachments. [#2046]
  * builder/amazon/chroot: Only unmount path if it is mounted. [#2054]
  * builder/amazon/instance: Use `-i` in sudo commands so PATH is inherited. [#1930]
  * builder/amazon/instance: Use `--region` flag for bundle upload command. [#1931]
  * builder/digitalocean: Wait for droplet to unlock before changing state,
      should lower the "pending event" errors.
  * builder/digitalocean: Ignore invalid fields from the ever-changing v2 API
  * builder/digitalocean: Private images can be used as a source. [#1792]
  * builder/docker: Fixed hang on prompt while copying script
  * builder/docker: Use `docker exec` for newer versions of Docker for
      running scripts. [#1993]
  * builder/docker: Fix crash that could occur at certain timed ctrl-c. [#1838]
  * builder/docker: validate that `export_path` is not a directory. [#2105]
  * builder/google: `ssh_timeout` is respected. [#1781]
  * builder/openstack: `ssh_interface` can be used to specify the interface
      to retrieve the SSH IP from. [#2220]
  * builder/qemu: Add `disk_discard` option. [#2120]
  * builder/qemu: Use proper SSH port, not hardcoded to 22. [#2236]
  * builder/qemu: Find unused SSH port if SSH port is taken. [#2032]
  * builder/virtualbox: Bind HTTP server to IPv4, which is more compatible with
      OS installers. [#1709]
  * builder/virtualbox: Remove the floppy controller in addition to the
      floppy disk. [#1879]
  * builder/virtualbox: Fixed regression where downloading ISO without a
      ".iso" extension didn't work. [#1839]
  * builder/virtualbox: Output dir is verified at runtime, not template
      validation time. [#2233]
  * builder/virtualbox: Find unused SSH port if SSH port is taken. [#2032]
  * builder/vmware: Add 100ms delay between keystrokes to avoid subtle
      timing issues in most cases. [#1663]
  * builder/vmware: Bind HTTP server to IPv4, which is more compatible with
      OS installers. [#1709]
  * builder/vmware: Case-insensitive match of MAC address to find IP. [#1989]
  * builder/vmware: More robust IP parsing from ifconfig output. [#1999]
  * builder/vmware: Nested output directories for ESXi work. [#2174]
  * builder/vmware: Output dir is verified at runtime, not template
      validation time. [#2233]
  * command/fix: For the `virtualbox` to `virtualbox-iso` builder rename,
      provisioner overrides are now also fixed. [#2231]
  * command/validate: don't crash for invalid builds. [#2139]
  * post-processor/atlas: Find common archive prefix for Windows. [#1874]
  * post-processor/atlas: Fix index out of range panic. [#1959]
  * post-processor/vagrant-cloud: Fixed failing on response
  * post-processor/vagrant-cloud: Don't delete version on error. [#2014]
  * post-processor/vagrant-cloud: Retry failed uploads a few times
  * provisioner/chef-client: Fix permissions issues on default dir. [#2255]
  * provisioner/chef-client: Node cleanup works now. [#2257]
  * provisioner/puppet-masterless: Allow manifest_file to be a directory
  * provisioner/salt-masterless: Add `--retcode-passthrough` to salt-call
  * provisioner/shell: chmod executable script to 0755, not 0777. [#1708]
  * provisioner/shell: inline commands failing will fail the provisioner. [#2069]
  * provisioner/shell: single quotes in env vars are escaped. [#2229]
  * provisioner/shell: Temporary file is deleted after run. [#2259]
  * provisioner/shell: Randomize default script name to avoid strange
      race issues from Windows. [#2270]

## 0.7.5 (December 9, 2014)

FEATURES:

  * **New command: `packer push`**: Push template and files to HashiCorp's
      Atlas for building your templates automatically.
  * **New post-processor: `atlas`**: Send artifact to HashiCorp's Atlas for
      versioning and storing artifacts. These artifacts can then be queried
      using the API, Terraform, etc.

IMPROVEMENTS:

  * builder/googlecompute: Support for ubuntu-os-cloud project
  * builder/googlecompute: Support for OAuth2 to avoid client secrets file
  * builder/googlecompute: GCE image from persistant disk instead of tarball
  * builder/qemu: Checksum type "none" can be used
  * provisioner/chef: Generate a node name if none available
  * provisioner/chef: Added ssl_verify_mode configuration

BUG FIXES:

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

FEATURES:

  * builder/digitalocean: API V2 support. [#1463]
  * builder/parallels: Don't depend on _prl-utils_. [#1499]

IMPROVEMENTS:

  * builder/amazon/all: Support new AWS Frankfurt region.
  * builder/docker: Allow remote `DOCKER_HOST`, which works as long as
      volumes work. [#1594]
  * builder/qemu: Can set cache mode for main disk. [#1558]
  * builder/qemu: Can build from pre-existing disk. [#1342]
  * builder/vmware: Can specify path to Fusion installation with environmental
      variable `FUSION_APP_PATH`. [#1552]
  * builder/vmware: Can specify the HW version for the VMX. [#1530]
  * builder/vmware/esxi: Will now cache ISOs/floppies remotely. [#1479]
  * builder/vmware/vmx: Source VMX can have a disk connected via SATA. [#1604]
  * post-processors/vagrant: Support Qemu (libvirt) boxes. [#1330]
  * post-processors/vagrantcloud: Support self-hosted box URLs.

BUG FIXES:

  * core: Fix loading plugins from pwd. [#1521]
  * builder/amazon: Prefer token in config if given. [#1544]
  * builder/amazon/all: Extended timeout for waiting for AMI. [#1533]
  * builder/virtualbox: Can read VirtualBox version on FreeBSD. [#1570]
  * builder/virtualbox: More robust reading of guest additions URL. [#1509]
  * builder/vmware: Always remove floppies/drives. [#1504]
  * builder/vmware: Wait some time so that post-VMX update aren't
      overwritten. [#1504]
  * builder/vmware/esxi: Retry power on if it fails. [#1334]
  * builder/vmware-vmx: Fix issue with order of boot command support. [#1492]
  * builder/amazon: Extend timeout and allow user override. [#1533]
  * builder/parallels: Ignore 'The fdd0 device does not exist' [#1501]
  * builder/parallels: Rely on Cleanup functions to detach devices. [#1502]
  * builder/parallels: Create VM without hdd and then add it later. [#1548]
  * builder/parallels: Disconnect cdrom0. [#1605]
  * builder/qemu: Don't use `-redir` flag anymore, replace with
      `hostfwd` options. [#1561]
  * builder/qemu: Use `pc` as default machine type instead of `pc-1.0`.
  * providers/aws: Ignore transient network errors. [#1579]
  * provisioner/ansible: Don't buffer output so output streams in. [#1585]
  * provisioner/ansible: Use inventory file always to avoid potentially
      deprecated feature. [#1562]
  * provisioner/shell: Quote environmental variables. [#1568]
  * provisioner/salt: Bootstrap over SSL. [#1608]
  * post-processors/docker-push: Work with docker-tag artifacts. [#1526]
  * post-processors/vsphere: Append "/" to object address. [#1615]

## 0.7.1 (September 10, 2014)

FEATURES:

  * builder/vmware: VMware Fusion Pro 7 is now supported. [#1478]

BUG FIXES:

  * core: SSH will connect slightly faster if it is ready immediately.
  * provisioner/file: directory uploads no longer hang. [#1484]
  * provisioner/file: fixed crash on large files. [#1473]
  * scripts: Windows executable renamed to packer.exe. [#1483]

## 0.7.0 (September 8, 2014)

BACKWARDS INCOMPATIBILITIES:

  * The authentication configuration for Google Compute Engine has changed.
      The new method is much simpler, but is not backwards compatible.
      `packer fix` will _not_ fix this. Please read the updated GCE docs.

FEATURES:

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
      instance store images. [#1139]
  * builder/docker: Images can now be committed instead of exported. [#1198]
  * builder/virtualbox-ovf: New `import_flags` setting can be used to add
      new command line flags to `VBoxManage import` to allow things such
      as EULAs to be accepted. [#1383]
  * builder/virtualbox-ovf: Boot commands and the HTTP server are supported.
      [#1169]
  * builder/vmware: VMware Player 6 is now supported. [#1168]
  * builder/vmware-vmx: Boot commands and the HTTP server are supported.
      [#1169]

IMPROVEMENTS:

  * core: `isotime` function can take a format. [#1126]
  * builder/amazon/all: `AWS_SECURITY_TOKEN` is read and can also be
      set with the `token` configuration. [#1236]
  * builder/amazon/all: Can force SSH on the private IP address with
      `ssh_private_ip`. [#1229]
  * builder/amazon/all: String fields in device mappings can use variables. [#1090]
  * builder/amazon-instance: EBS AMIs can be used as a source. [#1453]
  * builder/digitalocean: Can set API URL endpoint. [#1448]
  * builder/digitalocean: Region supports variables. [#1452]
  * builder/docker: Can now specify login credentials to pull images.
  * builder/docker: Support mounting additional volumes. [#1430]
  * builder/parallels/all: Path to tools ISO is calculated automatically. [#1455]
  * builder/parallels-pvm: `reassign_mac` option to choose wehther or not
      to generate a new MAC address. [#1461]
  * builder/qemu: Can specify "none" acceleration type. [#1395]
  * builder/qemu: Can specify "tcg" acceleration type. [#1395]
  * builder/virtualbox/all: `iso_interface` option to mount ISO with SATA. [#1200]
  * builder/vmware-vmx: Proper `floppy_files` support. [#1057]
  * command/build: Add `-color=false` flag to disable color. [#1433]
  * post-processor/docker-push: Can now specify login credentials. [#1243]
  * provisioner/chef-client: Support `chef_environment`. [#1190]

BUG FIXES:

  * core: nicer error message if an encrypted private key is used for
      SSH. [#1445]
  * core: Fix crash that could happen with a well timed double Ctrl-C.
      [#1328] [#1314]
  * core: SSH TCP keepalive period is now 5 seconds (shorter). [#1232]
  * builder/amazon-chroot: Can properly build HVM images now. [#1360]
  * builder/amazon-chroot: Fix crash in root device check. [#1360]
  * builder/amazon-chroot: Add description that Packer made the snapshot
      with a time. [#1388]
  * builder/amazon-ebs: AMI is deregistered if an error. [#1186]
  * builder/amazon-instance: Fix deprecation warning for `ec2-bundle-vol`
      [#1424]
  * builder/amazon-instance: Add `--no-filter` to the `ec2-bundle-vol`
      command by default to avoid corrupting data by removing package
      manager certs. [#1137]
  * builder/amazon/all: `delete_on_termination` set to false will work.
  * builder/amazon/all: Fix race condition on setting tags. [#1367]
  * builder/amazon/all: More desctriptive error messages if Amazon only
      sends an error code. [#1189]
  * builder/docker: Error if `DOCKER_HOST` is set.
  * builder/docker: Remove the container during cleanup. [#1206]
  * builder/docker: Fix case where not all output would show up from
      provisioners.
  * builder/googlecompute: add `disk_size` option. [#1397]
  * builder/googlecompute: Auth works with latest formats on Google Cloud
      Console. [#1344]
  * builder/openstack: Region is not required. [#1418]
  * builder/parallels-iso: ISO not removed from VM after install. [#1338]
  * builder/parallels/all: Add support for Parallels Desktop 10. [#1438]
  * builder/parallels/all: Added some navigation keys. [#1442]
  * builder/qemu: If headless, sdl display won't be used. [#1395]
  * builder/qemu: Use `512M` as `-m` default. [#1444]
  * builder/virtualbox/all: Search `VBOX_MSI_INSTALL_PATH` for path to
      `VBoxManage` on Windows. [#1337]
  * builder/virtualbox/all: Seed RNG to avoid same ports. [#1386]
  * builder/virtualbox/all: Better error if guest additions URL couldn't be
      detected. [#1439]
  * builder/virtualbox/all: Detect errors even when `VBoxManage` exits
      with a zero exit code. [#1119]
  * builder/virtualbox/iso: Append timestamp to default name for parallel
      builds. [#1365]
  * builder/vmware/all: No more error when Packer stops an already-stopped
      VM. [#1300]
  * builder/vmware/all: `ssh_host` accepts templates. [#1396]
  * builder/vmware/all: Don't remount floppy in VMX post step. [#1239]
  * builder/vmware/vmx: Do not re-add floppy disk files to VMX. [#1361]
  * builder/vmware-iso: Fix crash when `vnc_port_min` and max were the
      same value. [#1288]
  * builder/vmware-iso: Finding an available VNC port on Windows works. [#1372]
  * builder/vmware-vmx: Nice error if Clone is not supported (not VMware
      Fusion Pro). [#787]
  * post-processor/vagrant: Can supply your own metadata.json. [#1143]
  * provisioner/ansible-local: Use proper path on Windows. [#1375]
  * provisioner/file: Mode will now be preserved. [#1064]

## 0.6.1 (July 20, 2014)

FEATURES:

  * **New post processor:** `vagrant-cloud` - Push box files generated by
    vagrant post processor to Vagrant Cloud. [#1289]
  * Vagrant post-processor can now packer Hyper-V boxes.

IMPROVEMENTS:

  * builder/amazon: Support for enhanced networking on HVM images. [#1228]
  * builder/amazon-ebs: Support encrypted EBS volumes. [#1194]
  * builder/ansible: Add `playbook_dir` option. [#1000]
  * builder/openstack: Add ability to configure networks. [#1261]
  * builder/openstack: Skip certificate verification. [#1121]
  * builder/parallels/all: Add ability to select interface to connect to.
  * builder/parallels/pvm: Support `boot_command`. [#1082]
  * builder/virtualbox/all: Attempt to use local guest additions ISO
      before downloading from internet. [#1123]
  * builder/virtualbox/ovf: Supports `guest_additions_mode` [#1035]
  * builder/vmware/all: Increase cleanup timeout to 120 seconds. [#1167]
  * builder/vmware/all: Add `vmx_data_post` for modifying VMX data
      after shutdown. [#1149]
  * builder/vmware/vmx: Supports tools uploading. [#1154]

BUG FIXES:

  * core: `isotime` is the same time during the entire build. [#1153]
  * builder/amazon-common: Sort AMI strings before outputting. [#1305]
  * builder/amazon: User data can use templates/variables. [#1343]
  * builder/amazon: Can now build AMIs in GovCloud.
  * builder/null: SSH info can use templates/variables. [#1343]
  * builder/openstack: Workaround for gophercloud.ServerById crashing. [#1257]
  * builder/openstack: Force IPv4 addresses from address pools. [#1258]
  * builder/parallels: Do not delete entire CDROM device. [#1115]
  * builder/parallels: Errors while creating floppy disk. [#1225]
  * builder/parallels: Errors while removing floppy drive. [#1226]
  * builder/virtualbox-ovf: Supports guest additions options. [#1120]
  * builder/vmware-iso: Fix esx5 path separator in windows. [#1316]
  * builder/vmware: Remote ESXi builder now uploads floppy. [#1106]
  * builder/vmware: Remote ESXi builder no longer re-uploads ISO every
      time. [#1244]
  * post-processor/vsphere: Accept DOMAIN\account usernames. [#1178]
  * provisioner/chef-*: Fix remotePaths for Windows. [#394]

## 0.6.0 (May 2, 2014)

FEATURES:

  * **New builder:** `null` - The null builder does not produce any
    artifacts, but is useful for debugging provisioning scripts. [#970]
  * **New builder:** `parallels-iso` and `parallels-pvm` - These can be
    used to build Parallels virtual machines. [#1101]
  * **New provisioner:** `chef-client` - Provision using a the `chef-client`
    command, which talks to a Chef Server. [#855]
  * **New provisioner:** `puppet-server` - Provision using Puppet by
    communicating to a Puppet master. [#796]
  * `min_packer_version` can be specified in a Packer template to force
    a minimum version. [#487]

IMPROVEMENTS:

  * core: RPC transport between plugins switched to MessagePack
  * core: Templates array values can now be comma separated strings.
      Most importantly, this allows for user variables to fill
      array configurations. [#950]
  * builder/amazon: Added `ssh_private_key_file` option. [#971]
  * builder/amazon: Added `ami_virtualization_type` option. [#1021]
  * builder/digitalocean: Regions, image names, and sizes can be
      names that are looked up for their valid ID. [#960]
  * builder/googlecompute: Configurable instance name. [#1065]
  * builder/openstack: Support for conventional OpenStack environmental
      variables such as `OS_USERNAME`, `OS_PASSWORD`, etc. [#768]
  * builder/openstack: Support `openstack_provider` option to automatically
      fill defaults for different OpenStack variants. [#912]
  * builder/openstack: Support security groups. [#848]
  * builder/qemu: User variable expansion in `ssh_key_path` [#918]
  * builder/qemu: Floppy disk files list can also include globs
      and directories. [#1086]
  * builder/virtualbox: Support an `export_opts` option which allows
      specifying arbitrary arguments when exporting the VM. [#945]
  * builder/virtualbox: Added `vboxmanage_post` option to run vboxmanage
      commands just before exporting. [#664]
  * builder/virtualbox: Floppy disk files list can also include globs
      and directories. [#1086]
  * builder/vmware: Workstation 10 support for Linux. [#900]
  * builder/vmware: add cloning support on Windows. [#824]
  * builder/vmware: Floppy disk files list can also include globs
      and directories. [#1086]
  * command/build: Added `-parallel` flag so you can disable parallelization
    with `-no-parallel`. [#924]
  * post-processors/vsphere: `disk_mode` option. [#778]
  * provisioner/ansible: Add `inventory_file` option. [#1006]
  * provisioner/chef-client: Add `validation_client_name` option. [#1056]

BUG FIXES:

  * core: Errors are properly shown when adding bad floppy files. [#1043]
  * core: Fix some URL parsing issues on Windows.
  * core: Create Cache directory only when it is needed. [#367]
  * builder/amazon-instance: Use S3Endpoint for ec2-upload-bundle arg,
      which works for every region. [#904]
  * builder/digitalocean: updated default image_id. [#1032]
  * builder/googlecompute: Create persistent disk as boot disk via
      API v1. [#1001]
  * builder/openstack: Return proper error on invalid instance states. [#1018]
  * builder/virtualbox-iso: Retry unregister a few times to deal with
      VBoxManage randomness. [#915]
  * provisioner/ansible: Fix paths when provisioning Linux from
      Windows. [#963]
  * provisioner/ansible: set cwd to staging directory. [#1016]
  * provisioners/chef-client: Don't chown directory with Ubuntu. [#939]
  * provisioners/chef-solo: Deeply nested JSON works properly. [#1076]
  * provisioners/shell: Env var values can have equal signs. [#1045]
  * provisioners/shell: chmod the uploaded script file to 0777. [#994]
  * post-processor/docker-push: Allow repositories with ports. [#923]
  * post-processor/vagrant: Create parent directories for `output` path. [#1059]
  * post-processor/vsphere: datastore, network, and folder are no longer
      required. [#1091]

## 0.5.2 (02/21/2014)

FEATURES:

* **New post-processor:** `docker-import` - Import a Docker image
  and give it a specific repository/tag.
* **New post-processor:** `docker-push` - Push an imported image to
  a registry.

IMPROVEMENTS:

* core: Most downloads made by Packer now use a custom user agent. [#803]
* builder/googlecompute: SSH private key will be saved to disk if `-debug`
  is specified. [#867]
* builder/qemu: Can specify the name of the qemu binary. [#854]
* builder/virtualbox-ovf: Can specify import options such as "keepallmacs".
  [#883]

BUG FIXES:

* core: Fix crash case if blank parameters are given to Packer. [#832]
* core: Fix crash if big file uploads are done. [#897]
* core: Fix crash if machine-readable output is going to a closed
  pipe. [#875]
* builder/docker: user variables work properly. [#777]
* builder/qemu: reboots are now possible in provisioners. [#864]
* builder/virtualbox,vmware: iso\_checksum is not required if the
  checksum type is "none"
* builder/virtualbox,vmware/qemu: Support for additional scancodes for
  `boot_command` such as `<up>`, `<left>`, `<insert>`, etc. [#808]
* communicator/ssh: Send TCP keep-alives on connections. [#872]
* post-processor/vagrant: AWS/DigitalOcean keep input artifacts by
  default. [#55]
* provisioners/ansible-local: Properly upload custom playbooks. [#829]
* provisioners/ansible-local: Better error if ansible isn't installed.
  [#836]

## 0.5.1 (01/02/2014)

BUG FIXES:

* core: If a stream ID loops around, don't let it use stream ID 0. [#767]
* core: Fix issue where large writes to plugins would result in stream
  corruption. [#727]
* builders/virtualbox-ovf: `shutdown_timeout` config works. [#772]
* builders/vmware-iso: Remote driver works properly again. [#773]

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
  [#715]
* **New builder:** "virtualbox-ovf" can build VirtualBox images from
  an existing OVF or OVA. [#201]
* **New builder:** "vmware-vmx" can build VMware images from an existing
  VMX. [#201]
* Environmental variables can now be accessed as default values for
  user variables using the "env" function. See the documentation for more
  information.
* "description" field in templates: write a human-readable description
  of what a template does. This will be shown in `packer inspect`.
* Vagrant post-processor now accepts a list of files to include in the
  box.
* All provisioners can now have a "pause\_before" parameter to wait
  some period of time before running that provisioner. This is useful
  for reboots. [#737]

IMPROVEMENTS:

* core: Plugins communicate over a single TCP connection per plugin now,
  instead of sometimes dozens. Performance around plugin communication
  dramatically increased.
* core: Build names are now template processed so you can use things
  like user variables in them. [#744]
* core: New "pwd" function available globally that returns the working
  directory. [#762]
* builder/amazon/all: Launched EC2 instances now have a name of
  "Packer Builder" so that they are easily recognizable. [#642]
* builder/amazon/all: Copying AMIs to multiple regions now happens
  in parallel. [#495]
* builder/amazon/all: Ability to specify "run\_tags" to tag the instance
  while running. [#722]
* builder/digitalocean: Private networking support. [#698]
* builder/docker: A "run\_command" can be specified, configuring how
  the container is started. [#648]
* builder/openstack: In debug mode, the generated SSH keypair is saved
  so you can SSH into the machine. [#746]
* builder/qemu: Floppy files are supported. [#686]
* builder/qemu: Next `run_once` option tells Qemu to run only once,
  which is useful for Windows installs that handle reboots for you.
  [#687]
* builder/virtualbox: Nice errors if Packer can't write to
  the output directory.
* builder/virtualbox: ISO is ejected prior to export.
* builder/virtualbox: Checksum type can be "none" [#471]
* builder/vmware: Can now specify path to the Fusion application. [#677]
* builder/vmware: Checksum type can be "none" [#471]
* provisioner/puppet-masterless: Can now specify a `manifest_dir` to
  upload manifests to the remote machine for imports. [#655]

BUG FIXES:

* core: No colored output in machine-readable output. [#684]
* core: User variables can now be used for non-string fields. [#598]
* core: Fix bad download paths if the download URL contained a "."
  before a "/" [#716]
* core: "{{timestamp}}" values will always be the same for the entire
  duration of a build. [#744]
* builder/amazon: Handle cases where security group isn't instantly
  available. [#494]
* builder/virtualbox: don't download guest additions if disabled. [#731]
* post-processor/vsphere: Uploads VM properly. [#694]
* post-processor/vsphere: Process user variables.
* provisioner/ansible-local: all configurations are processed as templates
  [#749]
* provisioner/ansible-local: playbook paths are properly validated
  as directories, not files. [#710]
* provisioner/chef-solo: Environments are recognized. [#726]

## 0.4.1 (December 7, 2013)

IMPROVEMENTS:

* builder/amazon/ebs: New option allows associating a public IP with
  non-default VPC instances. [#660]
* builder/openstack: A "proxy\_url" setting was added to define an HTTP
  proxy to use when building with this builder. [#637]

BUG FIXES:

* core: Don't change background color on CLI anymore, making things look
  a tad nicer in some terminals.
* core: multiple ISO URLs works properly in all builders. [#683]
* builder/amazon/chroot: Block when obtaining file lock to allow
  parallel builds. [#689]
* builder/amazon/instance: Add location flag to upload bundle command
  so that building AMIs works out of us-east-1. [#679]
* builder/qemu: Qemu arguments are templated. [#688]
* builder/vmware: Cleanup of VMX keys works properly so cd-rom won't
  get stuck with ISO. [#685]
* builder/vmware: File cleanup is more resilient to file delete races
  with the operating system. [#675]
* provisioner/puppet-masterless: Check for hiera config path existence
  properly. [#656]

## 0.4.0 (November 19, 2013)

FEATURES:

* Docker builder: build and export Docker containers, easily provisioned
  with any of the Packer built-in provisioners.
* QEMU builder: builds a new VM compatible with KVM or Xen using QEMU.
* Remote ESXi builder: builds a VMware VM using ESXi remotely using only
  SSH to an ESXi machine directly.
* vSphere post-processor: Can upload VMware artifacts to vSphere
* Vagrant post-processor can now make DigitalOcean provider boxes. [#504]

IMPROVEMENTS:

* builder/amazon/all: Can now specify a list of multiple security group
  IDs to apply. [#499]
* builder/amazon/all: AWS API requests are now retried when a temporary
  network error occurs as well as 500 errors. [#559]
* builder/virtualbox: Use VBOX\_INSTALL\_PATH env var on Windows to find
  VBoxManage. [#628]
* post-processor/vagrant: skips gzip compression when compression_level=0
* provisioner/chef-solo: Encrypted data bag support. [#625]

BUG FIXES:

* builder/amazon/chroot: Copying empty directories works. [#588]
* builder/amazon/chroot: Chroot commands work with shell provisioners. [#581]
* builder/amazon/chroot: Don't choose a mount point that is a partition of
  an already mounted device. [#635]
* builder/virtualbox: Ctrl-C interrupts during waiting for boot. [#618]
* builder/vmware: VMX modifications are now case-insensitive. [#608]
* builder/vmware: VMware Fusion won't ask for VM upgrade.
* builder/vmware: Ctrl-C interrupts during waiting for boot. [#618]
* provisioner/chef-solo: Output is slightly prettier and more informative.

## 0.3.11 (November 4, 2013)

FEATURES:

* builder/amazon/ebs: Ability to specify which availability zone to create
  instance in. [#536]

IMPROVEMENTS:

* core: builders can now give warnings during validation. warnings won't
  fail the build but may hint at potential future problems.
* builder/digitalocean: Can now specify a droplet name
* builder/virtualbox: Can now disable guest addition download entirely
  by setting "guest_additions_mode" to "disable" [#580]
* builder/virtualbox,vmware: ISO urls can now be https. [#587]
* builder/virtualbox,vmware: Warning if shutdown command is not specified,
  since it is a common case of data loss.

BUG FIXES:

* core: Won't panic when writing to a bad pipe. [#560]
* builder/amazon/all: Properly scrub access key and secret key from logs.
  [#554]
* builder/openstack: Properly scrub password from logs. [#554]
* builder/virtualbox: No panic if SSH host port min/max is the same. [#594]
* builder/vmware: checks if `ifconfig` is in `/sbin` [#591]
* builder/vmware: Host IP lookup works for non-C locales. [#592]
* common/uuid: Use cryptographically secure PRNG when generating
  UUIDs. [#552]
* communicator/ssh: File uploads that exceed the size of memory no longer
  cause crashes. [#561]

## 0.3.10 (October 20, 2013)

FEATURES:

* Ansible provisioner

IMPROVEMENTS:

* post-processor/vagrant: support instance-store AMIs built by Packer. [#502]
* post-processor/vagrant: can now specify compression level to use
  when creating the box. [#506]

BUG FIXES:

* builder/all: timeout waiting for SSH connection is a failure. [#491]
* builder/amazon: Scrub sensitive data from the logs. [#521]
* builder/amazon: Handle the situation where an EC2 instance might not
  be immediately available. [#522]
* builder/amazon/chroot: Files copied into the chroot remove destination
  before copy, fixing issues with dangling symlinks. [#500]
* builder/digitalocean: don't panic if erroneous API response doesn't
  contain error message. [#492]
* builder/digitalocean: scrub API keys from config debug output. [#516]
* builder/virtualbox: error if VirtualBox version cant be detected. [#488]
* builder/virtualbox: detect if vboxdrv isn't properly setup. [#488]
* builder/virtualbox: sleep a bit before export to ensure the sesssion
  is unlocked. [#512]
* builder/virtualbox: create SATA drives properly on VirtualBox 4.3. [#547]
* builder/virtualbox: support user templates in SSH key path. [#539]
* builder/vmware: support user templates in SSH key path. [#539]
* communicator/ssh: Fix issue where a panic could arise from a nil
  dereference. [#525]
* post-processor/vagrant: Fix issue with VirtualBox OVA. [#548]
* provisioner/salt: Move salt states to correct remote directory. [#513]
* provisioner/shell: Won't block on certain scripts on Windows anymore.
  [#507]

## 0.3.9 (October 2, 2013)

FEATURES:

* The Amazon chroot builder is now able to run without any `sudo` privileges
  by using the "command_wrapper" configuration. [#430]
* Chef provisioner supports environments. [#483]

BUG FIXES:

* core: default user variable values don't need to be strings. [#456]
* builder/amazon-chroot: Fix errors with waitin for state change. [#459]
* builder/digitalocean: Use proper error message JSON key (DO API change).
* communicator/ssh: SCP uploads now work properly when directories
  contain symlinks. [#449]
* provisioner/chef-solo: Data bags and roles path are now properly
  populated when set. [#470]
* provisioner/shell: Windows line endings are actually properly changed
  to Unix line endings. [#477]

## 0.3.8 (September 22, 2013)

FEATURES:

* core: You can now specify `only` and `except` configurations on any
  provisioner or post-processor to specify a list of builds that they
  are valid for. [#438]
* builders/virtualbox: Guest additions can be attached rather than uploaded,
  easier to handle for Windows guests. [#405]
* provisioner/chef-solo: Ability to specify a custom Chef configuration
  template.
* provisioner/chef-solo: Roles and data bags support. [#348]

IMPROVEMENTS:

* core: User variables can now be used for integer, boolean, etc.
  values. [#418]
* core: Plugins made with incompatible versions will no longer load.
* builder/amazon/all: Interrupts work while waiting for AMI to be ready.
* provisioner/shell: Script line-endings are automatically converted to
  Unix-style line-endings. Can be disabled by setting "binary" to "true".
  [#277]

BUG FIXES:

* core: Set TCP KeepAlives on internally created RPC connections so that
  they don't die. [#416]
* builder/amazon/all: While waiting for AMI, will detect "failed" state.
* builder/amazon/all: Waiting for state will detect if the resource (AMI,
  instance, etc.) disappears from under it.
* builder/amazon/instance: Exclude only contents of /tmp, not /tmp
  itself. [#437]
* builder/amazon/instance: Make AccessKey/SecretKey available to bundle
  command even when they come from the environment. [#434]
* builder/virtualbox: F1-F12 and delete scancodes now work. [#425]
* post-processor/vagrant: Override configurations properly work. [#426]
* provisioner/puppet-masterless: Fix failure case when both facter vars
  are used and prevent_sudo. [#415]
* provisioner/puppet-masterless: User variables now work properly in
  manifest file and hiera path. [#448]

## 0.3.7 (September 9, 2013)

BACKWARDS INCOMPATIBILITIES:

* The "event_delay" option for the DigitalOcean builder is now gone.
  The builder automatically waits for events to go away. Run your templates
  through `packer fix` to get rid of these.

FEATURES:

* **NEW PROVISIONER:** `puppet-masterless`. You can now provision with
  a masterless Puppet setup. [#234]
* New globally available template function: `uuid`. Generates a new random
  UUID.
* New globally available template function: `isotime`. Generates the
  current time in ISO standard format.
* New Amazon template function: `clean_ami_name`. Substitutes '-' for
  characters that are illegal to use in an AMI name.

IMPROVEMENTS:

* builder/amazon/all: Ability to specify the format of the temporary
  keypair created. [#389]
* builder/amazon/all: Support the NoDevice flag for block mappings. [#396]
* builder/digitalocean: Retry on any pending event errors.
* builder/openstack: Can now specify a project. [#382]
* builder/virtualbox: Can now attach hard drive over SATA. [#391]
* provisioner/file: Can now upload directories. [#251]

BUG FIXES:

* core: Detect if SCP is not enabled on the other side. [#386]
* builder/amazon/all: When copying AMI to multiple regions, copy
  the metadata (tags and attributes) as well. [#388]
* builder/amazon/all: Fix panic case where eventually consistent
  instance state caused an index out of bounds.
* builder/virtualbox: The `vm_name` setting now properly sets the OVF
  name of the output. [#401]
* builder/vmware: Autoanswer VMware dialogs. [#393]
* command/inspect: Fix weird output for default values for optional vars.

## 0.3.6 (September 2, 2013)

FEATURES:

* User variables can now be specified as "required", meaning the user
  MUST specify a value. Just set the default value to "null". [#374]

IMPROVEMENTS:

* core: Much improved interrupt handling. For example, interrupts now
  cancel much more quickly within provisioners.
* builder/amazon: In `-debug` mode, the keypair used will be saved to
  the current directory so you can access the machine. [#373]
* builder/amazon: In `-debug` mode, the DNS is outputted.
* builder/openstack: IPv6 addresses supported for SSH. [#379]
* communicator/ssh: Support for private keys encrypted using PKCS8. [#376]
* provisioner/chef-solo: You can now use user variables in the `json`
  configuration for Chef. [#362]

BUG FIXES:

* core: Concurrent map access is completely gone, fixing rare issues
  with runtime memory corruption. [#307]
* core: Fix possible panic when ctrl-C during provisioner run.
* builder/digitalocean: Retry destroy a few times because DO sometimes
  gives false errors.
* builder/openstack: Properly handle the case no image is made. [#375]
* builder/openstack: Specifying a region is now required in a template.
* provisioners/salt-masterless: Use filepath join to properly join paths.

## 0.3.5 (August 28, 2013)

FEATURES:

* **NEW BUILDER:** `openstack`. You can now build on OpenStack. [#155]
* **NEW PROVISIONER:** `chef-solo`. You can now provision with Chef
  using `chef-solo` from local cookbooks.
* builder/amazon: Copy AMI to multiple regions with `ami_regions`. [#322]
* builder/virtualbox,vmware: Can now use SSH keys as an auth mechanism for
  SSH using `ssh_key_path`. [#70]
* builder/virtualbox,vmware: Support SHA512 as a checksum type. [#356]
* builder/vmware: The root hard drive type can now be specified with
  "disk_type_id" for advanced users. [#328]
* provisioner/salt-masterless: Ability to specfy a minion config. [#264]
* provisioner/salt-masterless: Ability to upload pillars. [#353]

IMPROVEMENTS:

* core: Output message when Ctrl-C received that we're cleaning up. [#338]
* builder/amazon: Tagging now works with all amazon builder types.
* builder/vmware: Option `ssh_skip_request_pty` for not requesting a PTY
  for the SSH connection. [#270]
* builder/vmware: Specify a `vmx_template_path` in order to customize
  the generated VMX. [#270]
* command/build: Machine-readable output now contains build errors, if any.
* command/build: An "end" sentinel is outputted in machine-readable output
  for artifact listing so it is easier to know when it is over.

BUG FIXES:

* core: Fixed a couple cases where a double ctrl-C could panic.
* core: Template validation fails if an override is specified for a
  non-existent builder. [#336]
* core: The SSH connection is heartbeated so that drops can be
  detected. [#200]
* builder/amazon/instance: Remove check for ec2-ami-tools because it
  didn't allow absolute paths to work properly. [#330]
* builder/digitalocean: Send a soft shutdown request so that files
  are properly synced before shutdown. [#332]
* command/build,command/validate: If a non-existent build is specified to
  '-only' or '-except', it is now an error. [#326]
* post-processor/vagrant: Setting OutputPath with a timestamp now
  always works properly. [#324]
* post-processor/vagrant: VirtualBox OVA formats now turn into
  Vagrant boxes properly. [#331]
* provisioner/shell: Retry upload if start command fails, making reboot
  handling much more robust.

## 0.3.4 (August 21, 2013)

IMPROVEMENTS:

* post-processor/vagrant: the file being compressed will be shown
  in the UI. [#314]

BUG FIXES:

* core: Avoid panics when double-interrupting Packer.
* provisioner/shell: Retry shell script uploads, making reboots more
  robust if they happen to fail in this stage. [#282]

## 0.3.3 (August 19, 2013)

FEATURES:

* builder/virtualbox: support exporting in OVA format. [#309]

IMPROVEMENTS:

* core: All HTTP downloads across Packer now support the standard
  proxy environmental variables (`HTTP_PROXY`, `NO_PROXY`, etc.) [#252]
* builder/amazon: API requests will use HTTP proxy if specified by
  enviromental variables.
* builder/digitalocean: API requests will use HTTP proxy if specified
  by environmental variables.

BUG FIXES:

* core: TCP connection between plugin processes will keep-alive. [#312]
* core: No more "unused key keep_input_artifact" for post processors. [#310]
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
  variable to get the VirtualBox version. [#272]
* builder/virtualbox: Do not check for VirtualBox as part of template
  validation; only check at execution.
* builder/vmware: Do not check for VMware as part of template validation;
  only check at execution.
* command/build: A path of "-" will read the template from stdin.
* builder/amazon: add block device mappings. [#90]

BUG FIXES:

* windows: file URLs are easier to get right as Packer
  has better parsing and error handling for Windows file paths. [#284]
* builder/amazon/all: Modifying more than one AMI attribute type no longer
  crashes.
* builder/amazon-instance: send IAM instance profile data. [#294]
* builder/digitalocean: API request parameters are properly URL
  encoded. [#281]
* builder/virtualbox: dowload progress won't be shown until download
  actually starts. [#288]
* builder/virtualbox: floppy files names of 13 characters are now properly
  written to the FAT12 filesystem. [#285]
* builder/vmware: dowload progress won't be shown until download
  actually starts. [#288]
* builder/vmware: interrupt works while typing commands over VNC.
* builder/virtualbox: floppy files names of 13 characters are now properly
  written to the FAT12 filesystem. [#285]
* post-processor/vagrant: Process user variables. [#295]

## 0.3.1 (August 12, 2013)

IMPROVEMENTS:

* provisioner/shell: New setting `start_retry_timeout` which is the timeout
  for the provisioner to attempt to _start_ the remote process. This allows
  the shell provisioner to work properly with reboots. [#260]

BUG FIXES:

* core: Remote command output containing '\r' now looks much better
  within the Packer output.
* builder/vmware: Fix issue with finding driver files. [#279]
* provisioner/salt-masterless: Uploads work properly from Windows. [#276]

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

* builder/amazon/all: User data can be passed to start the instances. [#253]
* provisioner/salt-masterless: `local_state_tree` is no longer required,
  allowing you to use shell provisioner (or others) to bring this down.
  [#269]

BUG FIXES:

* builder/amazon/ebs,instance: Retry deleing security group a few times.
  [#278]
* builder/vmware: Workstation works on Windows XP now. [#238]
* builder/vmware: Look for files on Windows in multiple locations
  using multiple environmental variables. [#263]
* provisioner/salt-masterless: states aren't deleted after the run
  anymore. [#265]
* provisioner/salt-masterless: error if any commands exit with a non-zero
  exit status. [#266]

## 0.2.3 (August 7, 2013)

IMPROVEMENTS:

* builder/amazon/all: Added Amazon AMI tag support. [#233]

BUG FIXES:

* core: Absolute/relative filepaths on Windows now work for iso_url
  and other settings. [#240]
* builder/amazon/all: instance info is refreshed while waiting for SSH,
  allowing Packer to see updated IP/DNS info. [#243]

## 0.2.2 (August 1, 2013)

FEATURES:

* New builder: `amazon-chroot` can create EBS-backed AMIs without launching
  a new EC2 instance. This can shave minutes off of the AMI creation process.
  See the docs for more info.
* New provisioner: `salt-masterless` will provision the node using Salt
  without a master.
* The `vmware` builder now works with Workstation 9 on Windows. [#222]
* The `vmware` builder now works with Player 5 on Linux. [#190]

IMPROVEMENTS:

* core: Colors won't be outputted on Windows unless in Cygwin.
* builder/amazon/all: Added `iam_instance_profile` to launch the source
  image with a given IAM profile. [#226]

BUG FIXES:

* builder/virtualbox,vmware: relative paths work properly as URL
  configurations. [#215]
* builder/virtualbox,vmware: fix race condition in deleting the output
  directory on Windows by retrying.

## 0.2.1 (July 26, 2013)

FEATURES:

* New builder: `amazon-instance` can create instance-storage backed
  AMIs.
* VMware builder now works with Workstation 9 on Linux.

IMPROVEMENTS:

* builder/amazon/all: Ctrl-C while waiting for state change works
* builder/amazon/ebs: Can now launch instances into a VPC for added protection. [#210]
* builder/virtualbox,vmware: Add backspace, delete, and F1-F12 keys to the boot
  command.
* builder/virtualbox: massive performance improvements with big ISO files because
  an expensive copy is avoided. [#202]
* builder/vmware: CD is removed prior to exporting final machine. [#198]

BUG FIXES:

* builder/amazon/all: Gracefully handle when AMI appears to not exist
  while AWS state is propogating. [#207]
* builder/virtualbox: Trim carriage returns for Windows to properly
  detect VM state on Windows. [#218]
* core: build names no longer cause invalid config errors. [#197]
* command/build: If any builds fail, exit with non-zero exit status.
* communicator/ssh: SCP exit codes are tested and errors are reported. [#195]
* communicator/ssh: Properly change slash direction for Windows hosts. [#218]

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
  existing artifacts if they exist. [#173]
* You can now log to a file (instead of just stderr) by setting the
  `PACKER_LOG_FILE` environmental variable. [#168]
* Checksums other than MD5 can now be used. SHA1 and SHA256 can also
  be used. See the documentation on `iso_checksum_type` for more info. [#175]

IMPROVEMENTS:

* core: invalid keys in configuration are now considered validation
  errors. [#104]
* core: all builders now share a common SSH connection core, improving
  SSH reliability over all the builders.
* amazon-ebs: Credentials will come from IAM role if available. [#160]
* amazon-ebs: Verify the source AMI is EBS-backed before launching. [#169]
* shell provisioner: the build name and builder type are available in
  the `PACKER_BUILD_NAME` and `PACKER_BUILDER_TYPE` env vars by default,
  respectively. [#154]
* vmware: error if shutdown command has non-zero exit status.

BUG FIXES:

* core: UI messages are now properly prefixed with spaces again.
* core: If SSH connection ends, re-connection attempts will take
  place. [#152]
* virtualbox: "paused" doesn't mean the VM is stopped, improving
  shutdown detection.
* vmware: error if guest IP could not be detected. [#189]

## 0.1.5 (July 7, 2013)

FEATURES:

* "file" uploader will upload files from the machine running Packer to the
  remote machine.
* VirtualBox guest additions URL and checksum can now be specified, allowing
  the VirtualBox builder to have the ability to be used completely offline.

IMPROVEMENTS:

* core: If SCP is not available, a more descriptive error message
  is shown telling the user. [#127]
* shell: Scripts are now executed by default according to their shebang,
  not with `/bin/sh`. [#105]
* shell: You can specify what interpreter you want inline scripts to
  run with `inline_shebang`.
* virtualbox: Delete the packer-made SSH port forwarding prior to
  exporting the VM.

BUG FIXES:

* core: Non-200 response codes on downloads now show proper errors.
  [#141]
* amazon-ebs: SSH handshake is retried. [#130]
* vagrant: The `BuildName` template propery works properly in
  the output path.
* vagrant: Properly configure the provider-specific post-processors so
  things like `vagrantfile_template` work. [#129]
* vagrant: Close filehandles when copying files so Windows can
  rename files. [#100]

## 0.1.4 (July 2, 2013)

FEATURES:

* virtualbox: Can now be built headless with the "Headless" option. [#99]
* virtualbox: <wait5> and <wait10> codes for waiting 5 and 10 seconds
  during the boot sequence, respectively. [#97]
* vmware: Can now be built headless with the "Headless" option. [#99]
* vmware: <wait5> and <wait10> codes for waiting 5 and 10 seconds
  during the boot sequence, respectively. [#97]
* vmware: Disks are defragmented and compacted at the end of the build.
  This can be disabled using "skip_compaction"

IMPROVEMENTS:

* core: Template syntax errors now show line and character number. [#56]
* amazon-ebs: Access key and secret access key default to
  environmental variables. [#40]
* virtualbox: Send password for keyboard-interactive auth. [#121]
* vmware: Send password for keyboard-interactive auth. [#121]

BUG FIXES:

* vmware: Wait until shut down cleans up properly to avoid corrupt
  disk files. [#111]

## 0.1.3 (July 1, 2013)

FEATURES:

* The VMware builder can now upload the VMware tools for you into
  the VM. This is opt-in, you must specify the `tools_upload_flavor`
  option. See the website for more documentation.

IMPROVEMENTS:

* digitalocean: Errors contain human-friendly error messages. [#85]

BUG FIXES:

* core: More plugin server fixes that avoid hangs on OS X 10.7. [#87]
* vagrant: AWS boxes will keep the AMI artifact around. [#55]
* virtualbox: More robust version parsing for uploading guest additions. [#69]
* virtualbox: Output dir and VM name defaults depend on build name,
  avoiding collisions. [#91]
* vmware: Output dir and VM name defaults depend on build name,
  avoiding collisions. [#91]

## 0.1.2 (June 29, 2013)

IMPROVEMENTS:

* core: Template doesn't validate if there are no builders.
* vmware: Delete any VMware files in the VM that aren't necessary for
  it to function.

BUG FIXES:

* core: Plugin servers consider a port in use if there is any
  error listening to it. This fixes I18n issues and Windows. [#58]
* amazon-ebs: Sleep between checking instance state to avoid
  RequestLimitExceeded. [#50]
* vagrant: Rename VirtualBox ovf to "box.ovf" [#64]
* vagrant: VMware boxes have the correct provider type.
* vmware: Properly populate files in artifact so that the Vagrant
  post-processor works. [#63]

## 0.1.1 (June 28, 2013)

BUG FIXES:

* core: plugins listen explicitly on 127.0.0.1, fixing odd hangs. [#37]
* core: fix race condition on verifying checksum of large ISOs which
  could cause panics. [#52]
* virtualbox: `boot_wait` defaults to "10s" rather than 0. [#44]
* virtualbox: if `http_port_min` and max are the same, it will no longer
  panic. [#53]
* vmware: `boot_wait` defaults to "10s" rather than 0. [#44]
* vmware: if `http_port_min` and max are the same, it will no longer
  panic. [#53]

## 0.1.0 (June 28, 2013)

* Initial release
