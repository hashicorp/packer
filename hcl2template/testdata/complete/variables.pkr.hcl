
variables {
    foo = "value"
    // my_secret = "foo"
    // image_name = "foo-image-{{user `my_secret`}}"
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
  default = ["a", "b", "c"]
}

locals {
  feefoo = "${var.foo}_${var.image_id}"
}
