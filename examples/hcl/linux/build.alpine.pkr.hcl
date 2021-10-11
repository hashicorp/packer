
build {
  name = "alpine"
  description = <<EOF
This build creates alpine images for versions :
* v3.12
For the following builders :
* virtualbox-iso
EOF

  // the common fields of the source blocks are defined in the
  // source.builder-type.pkr.hcl files, here we only set the fields specific to
  // the different versions of ubuntu.
  source "source.virtualbox-iso.base-alpine-amd64" {
    name                    = "3.12"
    iso_url                 = local.iso_url_alpine_312
    iso_checksum            = "file:${local.iso_checksum_url_alpine_312}"
    output_directory        = "virtualbox_iso_alpine_312_amd64"
    boot_command            = local.alpine_312_floppy_boot_command
    boot_wait               = "10s"
  }

  source "source.vsphere-iso.base-alpine-amd64" {
    name                    = "3.12"
    vm_name                 = "alpine-3.12"
    iso_url                 = local.iso_url_alpine_312
    iso_checksum            = "file:${local.iso_checksum_url_alpine_312}"
    boot_command            = local.alpine_312_floppy_boot_command_vsphere
    boot_wait               = "10s"
  }

  source "source.vmware-iso.esxi-base-alpine-amd64" {
    name                    = "3.12-from-esxi"
    iso_url                 = local.iso_url_alpine_312
    iso_checksum            = "file:${local.iso_checksum_url_alpine_312}"
    boot_command            = local.alpine_312_floppy_boot_command_vsphere
  }

  provisioner "shell" {
    inline = ["echo hi"]
  }
}
