
variable "alpine_password" {
  type = string
  default = "alpine"
}

locals {
  iso_url_alpine_380             = "http://dl-cdn.alpinelinux.org/alpine/v3.8/releases/x86_64/alpine-standard-3.8.0-x86_64.iso"
  iso_checksum_url_alpine_380    = "http://dl-cdn.alpinelinux.org/alpine/v3.8/releases/x86_64/alpine-standard-3.8.0-x86_64.iso.sha256"
  floppy_files = [
    "${local.http_directory}/alpine-answers",
    "${local.http_directory}/alpine-setup.sh"
  ]
  alpine_380_floppy_boot_command = [
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

}
