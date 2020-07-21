
variable "alpine_password" {
  type = string
  default = "alpine"
}

locals {
  iso_url_alpine_312             = "http://dl-cdn.alpinelinux.org/alpine/v3.12/releases/x86_64/alpine-virt-3.12.0-x86_64.iso"
  iso_checksum_url_alpine_312    = "http://dl-cdn.alpinelinux.org/alpine/v3.12/releases/x86_64/alpine-virt-3.12.0-x86_64.iso.sha256"
  floppy_files_alpine = [
    "${local.http_directory}/alpine-answers",
    "${local.http_directory}/alpine-setup.sh"
  ]

  alpine_312_floppy_boot_command = [
    "root<enter><wait>",
    "mount -t vfat /dev/fd0 /media/floppy<enter><wait>",
    "setup-alpine -f /media/floppy/alpine-answers<enter>",
    "<wait5>",
    "${var.alpine_password}<enter>",
    "${var.alpine_password}<enter>",
    "<wait5>",
    "y<enter>",
    "<wait40s>",
    "reboot<enter>",
    "<wait20s>",
    "root<enter>",
    "${var.alpine_password}<enter><wait>",
    "mount -t vfat /dev/fd0 /media/floppy<enter><wait>",
    "/media/floppy/alpine-setup.sh<enter>",
  ]

  floppy_files_alpine_vsphere = [
    "${local.http_directory}/alpine-vsphere-answers",
    "${local.http_directory}/alpine-setup.sh"
  ]

  alpine_312_floppy_boot_command_vsphere = [
    "root<enter><wait1s>",
    "mount -t vfat /dev/fd0 /media/floppy<enter><wait1s>",
    "setup-alpine -f /media/floppy/alpine-vsphere-answers<enter><wait3s>",
    "${var.alpine_password}<enter>",
    "${var.alpine_password}<enter>",
    "<wait6s>",
    "y<enter>",
    "<wait12s>",
    "reboot<enter>",
    "<wait12s>",
    "root<enter>",
    "${var.alpine_password}<enter><wait>",
    "mount -t vfat /dev/fd0 /media/floppy<enter><wait>",
    "/media/floppy/alpine-setup.sh<enter>",
    "<wait55s>",
  ]

}
