
variable "preseed_path" {
  type    = string
  default = "preseed.cfg"
}

variable "guest_additions_url" {
  type    = string
  default = ""
}

variable "headless" {
  type    = bool
  default = true
}

locals {
  // fileset lists all files in the http directory as a set, we convert that
  // set to a list of strings and we then take the directory of the first
  // value. This validates that the http directory exists even before starting
  // any builder/provisioner.
  http_directory = dirname(convert(fileset(".", "etc/http/*"), list(string))[0])
  http_directory_content = {
    "/alpine-answers"           = file("${local.http_directory}/alpine-answers"),
    "/alpine-setup.sh"          = file("${local.http_directory}/alpine-setup.sh"),
    "/preseed_hardcoded_ip.cfg" = file("${local.http_directory}/preseed_hardcoded_ip.cfg"),
    "/preseed.cfg"              = file("${local.http_directory}/preseed.cfg"),
  }
}
