
build {
  name = "ubuntu"
  description = <<EOF
This build creates ubuntu images for ubuntu versions :
* 16.04
* 18.04 
For the following builers :
* virtualbox-iso
EOF

  source "source.virtualbox-iso.base-ubuntu-amd64" {
    iso_url                 = local.iso_url_ubuntu_1604
    iso_checksum            = local.iso_checksum_ubuntu_1604
    output_directory        = "ubuntu_1604"
    boot_command            = local.ubuntu_1604_boot_command
    hard_drive_interface    = "sata"
    virtualbox_version_file = ".vbox_version"
    guest_additions_path    = "VBoxGuestAdditions_{{.Version}}.iso"
    guest_additions_url     = var.guest_additions_url
  }

  // source "source.virtualbox-iso.base-ubuntu-amd64" {
  //   iso_url          = local.iso_url_ubuntu_1804
  //   iso_checksum     = local.iso_checksum_ubuntu_1804
  //   output_directory = "ubuntu_1804"
  // }

  provisioner "shell" {
    environment_vars  = [ "HOME_DIR=/home/vagrant" ]
    execute_command   = "echo 'vagrant' | {{.Vars}} sudo -S -E sh -eux '{{.Path}}'"
    expect_disconnect = true
    scripts           = fileset(".", "etc/scripts/*.sh")

  // /../_common/minimize.sh"
  }
}
