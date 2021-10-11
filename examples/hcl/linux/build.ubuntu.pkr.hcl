
build {
  name = "ubuntu"
  description = <<EOF
This build creates ubuntu images for ubuntu versions :
* 16.04
* 18.04
For the following builders :
* virtualbox-iso
* parallels-iso
* vmware-iso
* qemu
* vsphere-iso
EOF

  // the common fields of the source blocks are defined in the
  // source.builder-type.pkr.hcl files, here we only set the fields specific to
  // the different versions of ubuntu.
  source "source.virtualbox-iso.base-ubuntu-amd64" {
    name                    = "16.04"
    iso_url                 = local.iso_url_ubuntu_1604
    iso_checksum            = "file:${local.iso_checksum_url_ubuntu_1604}"
    output_directory        = "virtualbox_iso_ubuntu_1604_amd64"
    boot_command            = local.ubuntu_1604_boot_command
    boot_wait               = "10s"
  }

  source "source.virtualbox-iso.base-ubuntu-amd64" {
    name                    = "18.04"
    iso_url                 = local.iso_url_ubuntu_1804
    iso_checksum            = "file:${local.iso_checksum_url_ubuntu_1804}"
    output_directory        = "virtualbox_iso_ubuntu_1804_amd64"
    boot_command            = local.ubuntu_1804_boot_command
    boot_wait               = "5s"
  }

  source "source.parallels-iso.base-ubuntu-amd64" {
    name                    = "16.04"
    iso_url                 = local.iso_url_ubuntu_1604
    iso_checksum            = "file:${local.iso_checksum_url_ubuntu_1604}"
    output_directory        = "parallels_iso_ubuntu_1604_amd64"
    boot_command            = local.ubuntu_1604_boot_command
  }

  source "source.parallels-iso.base-ubuntu-amd64" {
    name                    = "18.04"
    iso_url                 = local.iso_url_ubuntu_1804
    iso_checksum            = "file:${local.iso_checksum_url_ubuntu_1804}"
    output_directory        = "parallels_iso_ubuntu_1804_amd64"
    boot_command            = local.ubuntu_1804_boot_command
  }

  source "source.vmware-iso.base-ubuntu-amd64" {
    name                    = "16.04"
    iso_url                 = local.iso_url_ubuntu_1604
    iso_checksum            = "file:${local.iso_checksum_url_ubuntu_1604}"
    output_directory        = "vmware_iso_ubuntu_1604_amd64"
    boot_command            = local.ubuntu_1604_boot_command
  }

  source "source.vmware-vmx.base-ubuntu-amd64" {
    name        = "16.04"
    source_path = "vmware_iso_ubuntu_1604_amd64/packer-base-ubuntu-amd64.vmx"
  }

  source "source.vmware-iso.base-ubuntu-amd64" {
    name                    = "18.04"
    iso_url                 = local.iso_url_ubuntu_1804
    iso_checksum            = "file:${local.iso_checksum_url_ubuntu_1804}"
    output_directory        = "vmware_iso_ubuntu_1804_amd64"
    boot_command            = local.ubuntu_1804_boot_command
  }

  source "source.vmware-iso.esxi-base-ubuntu-amd64" {
    name                    = "16.04-from-esxi"
    iso_url                 = local.iso_url_ubuntu_1604
    iso_checksum            = "file:${local.iso_checksum_url_ubuntu_1604}"
    output_directory        = "vmware_iso_ubuntu_1604_amd64_from_esxi"
    boot_command            = local.ubuntu_1604_boot_command
  }

  source "source.qemu.base-ubuntu-amd64" {
    name                    = "16.04"
    iso_url                 = local.iso_url_ubuntu_1604
    iso_checksum            = "file:${local.iso_checksum_url_ubuntu_1604}"
    output_directory        = "qemu_iso_ubuntu_1604_amd64"
    boot_command            = local.ubuntu_1604_boot_command
  }

  source "source.qemu.base-ubuntu-amd64" {
    name                    = "18.04"
    iso_url                 = local.iso_url_ubuntu_1804
    iso_checksum            = "file:${local.iso_checksum_url_ubuntu_1804}"
    output_directory        = "qemu_iso_ubuntu_1804_amd64"
    boot_command            = local.ubuntu_1804_boot_command
  }

  source "source.vsphere-iso.base-ubuntu-amd64" {
    name                    = "16.04"
    vm_name                 = "ubuntu-16.04"
    iso_url                 = local.iso_url_ubuntu_1604
    iso_checksum            = "file:${local.iso_checksum_url_ubuntu_1604}"
  }

  provisioner "shell" {
    environment_vars  = [ "HOME_DIR=/home/vagrant" ]
    execute_command   = "echo 'vagrant' | {{.Vars}} sudo -S -E sh -eux '{{.Path}}'"
    expect_disconnect = true
    // fileset will list files in etc/scripts sorted in an alphanumerical way.
    scripts           = fileset(".", "etc/scripts/*.sh")
  }
}
