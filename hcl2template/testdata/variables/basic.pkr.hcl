
variables {
    key = "value"
    my_secret = "foo"
    image_name = "foo-image-{{user `my_secret`}}"
}

variable "image_id" {
  type = string
  default = "image-id-default"
}

variable "port" {
  type = number
  default = 42
}

variable "availability_zone_names" {
  type    = list(string)
  default = ["us-west-1a"]
  description = <<POTATO
Describing is awesome ;D
POTATO
}

variable "super_secret_password" {
  type     = string
  sensitive = true
description = <<IMSENSIBLE
Handle with care plz
IMSENSIBLE
  default = null
}

locals {
  service_name = "forum"
  owner        = "Community Team"
}

local "supersecret" {
  sensitive = true
  expression = "secretvar"
}
