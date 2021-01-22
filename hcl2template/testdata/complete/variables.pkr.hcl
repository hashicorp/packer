
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
  default = ["A", "B", "C"]
}

locals {
  feefoo = "${var.foo}_${var.image_id}"
  data_source = data.amazon-ami.test.string
}


locals {
  standard_tags = {
    Component   = "user-service"
    Environment = "production"
  }

  abc_map = [
    {id = "a"},
    {id = "b"},
    {id = "c"},
  ]
}

local "supersecret" {
  expression = "${var.image_id}-password"
  sensitive = true
}
