
// {{user `mirror`}}/{{user `mirror_directory`}}/{{user `iso_name`}}

variable "ubuntu_mirror" {
  default = "http://releases.ubuntu.com/releases"
}

variable "ubuntu_variant" {
  default = "server-amd64"
}

variable "ubuntu_1604_version" {
  default = "16.04.6"
}

variable "ubuntu_1804_version" {
  default = "18.04.4"
}

////
// ubuntu 1604
////

locals {
  iso_url_ubuntu_1604      = "${var.ubuntu_mirror}/16.04/ubuntu-${var.ubuntu_1604_version}-${var.ubuntu_variant}.iso"
  iso_checksum_ubuntu_1604 = "${var.ubuntu_mirror}/16.04/SHA256SUMS"

  iso_ubuntu_1604 = "${local.iso_url_ubuntu_1604}?checksum=${local.iso_checksum_ubuntu_1604}"
}

////
// ubuntu 1804
////

locals {
  iso_url_ubuntu_1804      = "${var.ubuntu_mirror}/18.04/ubuntu-${var.ubuntu_1804_version}-${var.ubuntu_variant}.iso"
  iso_checksum_ubuntu_1804 = "${var.ubuntu_mirror}/18.04/SHA256SUMS"

  iso_ubuntu_1804 = "${local.iso_url_ubuntu_1804}?checksum=${local.iso_checksum_ubuntu_1804}"
}
