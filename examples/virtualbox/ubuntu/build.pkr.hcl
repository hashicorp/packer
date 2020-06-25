
build {
  name = "ubuntu"
  description = <<<EOF
This build creates ubuntu images for ubuntu versions :
* 16.04
* 18.04 
For the following builers :
* virtualbox-iso
EOF
  source "virtualbox-iso.base-ubuntu" {
    iso_url = local.iso_ubuntu_1604
    output_directory = "ubuntu_1604"
  }

  source "virtualbox-iso.base-ubuntu" {
    iso_url = local.iso_ubuntu_1804
    output_directory = "ubuntu_1804"
  }

  provisioner "shell" {
    scripts = [
      "/scripts/update.sh",
      "/../_common/motd.sh",
      "/../_common/sshd.sh",
      "/scripts/networking.sh",
      "/scripts/sudoers.sh",
      "/../_common/vagrant.sh",
      "/scripts/systemd.sh",
      "/../_common/virtualbox.sh",
      "/../ubuntu/scripts/vmware.sh",
      "/../_common/parallels.sh",
      "/scripts/cleanup.sh",
      "/../_common/minimize.sh"
    ]
  }
}
