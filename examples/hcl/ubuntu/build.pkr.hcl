
build {
  name = "ubuntu"
  description = <<EOF
This build creates ubuntu images for ubuntu versions :
* 16.04
* 18.04 
For the following builers :
* virtualbox-iso
* parallels-iso
* vmware-iso
EOF

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

  source "source.vmware-iso.base-ubuntu-amd64" {
    name                    = "18.04"
    iso_url                 = local.iso_url_ubuntu_1804
    iso_checksum            = "file:${local.iso_checksum_url_ubuntu_1804}"
    output_directory        = "vmware_iso_ubuntu_1804_amd64"
    boot_command            = local.ubuntu_1804_boot_command
  }

  provisioner "shell" {
    environment_vars  = [ "HOME_DIR=/home/vagrant" ]
    execute_command   = "echo 'vagrant' | {{.Vars}} sudo -S -E sh -eux '{{.Path}}'"
    expect_disconnect = true
    scripts           = fileset(".", "etc/scripts/*.sh")
  }
}
