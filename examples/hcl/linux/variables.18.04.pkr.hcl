variable "ubuntu_1804_version" {
  default = "18.04.5"
}

locals {
  iso_url_ubuntu_1804           = "http://cdimage.ubuntu.com/ubuntu/releases/18.04/release/ubuntu-${var.ubuntu_1804_version}-server-amd64.iso"
  iso_checksum_url_ubuntu_1804  = "http://cdimage.ubuntu.com/ubuntu/releases/18.04/release/SHA256SUMS"
  ubuntu_1804_boot_command      = [
    "<esc><wait>",
    "<esc><wait>",
    "<enter><wait>",
    "/install/vmlinuz<wait>",
    " auto<wait>",
    " console-setup/ask_detect=false<wait>",
    " console-setup/layoutcode=us<wait>",
    " console-setup/modelcode=pc105<wait>",
    " debconf/frontend=noninteractive<wait>",
    " debian-installer=en_US.UTF-8<wait>",
    " fb=false<wait>",
    " initrd=/install/initrd.gz<wait>",
    " kbd-chooser/method=us<wait>",
    " keyboard-configuration/layout=USA<wait>",
    " keyboard-configuration/variant=USA<wait>",
    " locale=en_US.UTF-8<wait>",
    " netcfg/get_domain=vm<wait>",
    " netcfg/get_hostname=vagrant<wait>",
    " grub-installer/bootdev=/dev/sda<wait>",
    " noapic<wait>",
    " preseed/url=http://{{ .HTTPIP }}:{{ .HTTPPort }}/${var.preseed_path}<wait>",
    " -- <wait>",
    "<enter><wait>"
  ]
}
