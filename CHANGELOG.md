## 0.1.2 (unreleased)

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
