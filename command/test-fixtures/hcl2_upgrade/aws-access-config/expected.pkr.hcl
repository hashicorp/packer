packer {
  required_version = ">= 1.6.0"
}

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

data "amazon-ami" "autogenerated_1" {
  access_key = "NJDBFASJDbsajhbda5487"
  filters = {
    name                = "ubuntu/images/*/ubuntu-xenial-16.04-amd64-server-*"
    root-device-type    = "ebs"
    virtualization-type = "hvm"
  }
  most_recent = true
  owners      = ["099720109477"]
  region      = "us-west-2"
  secret_key  = "ASEfewdsfAWASTT51874"
}

data "amazon-ami" "autogenerated_2" {
  access_key = "${var.aws_access_key}"
  filters = {
    name                = "ubuntu/images/*/ubuntu-xenial-16.04-amd64-server-*"
    root-device-type    = "ebs"
    virtualization-type = "hvm"
  }
  most_recent = true
  owners      = ["099720109477"]
  region      = "${var.aws_region}"
  secret_key  = "${var.aws_secret_key}"
}

locals { timestamp = regex_replace(timestamp(), "[- TZ:]", "") }

source "amazon-ebs" "autogenerated_1" {
  access_key    = "NJDBFASJDbsajhbda5487"
  ami_name      = "ubuntu-16-04-test-${local.timestamp}"
  region        = "us-west-2"
  secret_key    = "ASEfewdsfAWASTT51874"
  source_ami    = "${data.amazon-ami.autogenerated_1.id}"
  ssh_interface = "session_manager"
  ssh_username  = "ubuntu"
}

source "amazon-ebs" "named_builder" {
  access_key    = "${var.aws_access_key}"
  ami_name      = "ubuntu-16-04-test-${local.timestamp}"
  region        = "${var.aws_region}"
  secret_key    = "${var.aws_secret_key}"
  source_ami    = "${data.amazon-ami.autogenerated_2.id}"
  ssh_interface = "session_manager"
  ssh_username  = "ubuntu"
}

build {
  sources = ["source.amazon-ebs.autogenerated_1", "source.amazon-ebs.named_builder"]

}
