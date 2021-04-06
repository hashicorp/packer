
variable "aws_access_key" {
  type      = string
  default   = ""
  sensitive = true
}

variable "aws_region" {
  type = string
}

variable "aws_secret_key" {
  type      = string
  default   = ""
  sensitive = true
}

local "password" {
  sensitive  = true
  expression = "${var.aws_secret_key}-${var.aws_access_key}"
}

locals {
  aws_secondary_region = "${var.aws_region}"
}
