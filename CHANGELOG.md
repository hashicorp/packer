## 0.2.3 (unreleased)

BUG FIXES:

* core: Absolute/relative filepaths on Windows now work for iso_url
  and other settings. [GH-240]

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
