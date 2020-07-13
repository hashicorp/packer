
build {
  name = "alpine"
  description = <<EOF
This build creates alpine images for versions :
* v3.8
For the following builers :
* virtualbox-iso
EOF

  // the common fields of the source blocks are defined in the
  // source.builder-type.pkr.hcl files, here we only set the fields specific to
  // the different versions of ubuntu.
  source "source.virtualbox-iso.base-alpine-amd64" {
    name                    = "3.8"
    iso_url                 = local.iso_url_alpine_380
    iso_checksum            = "file:${local.iso_checksum_url_alpine_380}"
    output_directory        = "virtualbox_iso_alpine_380_amd64"
    boot_command            = local.alpine_380_floppy_boot_command
    floppy_files            = local.floppy_files
    boot_wait               = "10s"
  }

  provisioner "shell" {
    inline = ["echo hi"]
  }
}
