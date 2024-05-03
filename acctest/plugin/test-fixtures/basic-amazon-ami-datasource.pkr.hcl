packer {
  required_plugins {
    amazon = {
      version = "~>1"
      source = "github.com/hashicorp/amazon"
    }
  }
}

data "amazon-ami" "test" {
  filters = {
    name                = "ubuntu/images/hvm-ssd/ubuntu-focal-20.04-amd64-server-*"
    root-device-type    = "ebs"
    virtualization-type = "hvm"
  }
  region = "us-west-2"
  most_recent = true
  owners      = ["099720109477"]
}

source "amazon-ebs" "basic-example" {
  region = "us-west-2"
  source_ami = data.amazon-ami.test.id
  ami_name =  "packer-amazon-ami-test"
  communicator  = "ssh"
  instance_type = "t2.micro"
  ssh_username  = "ubuntu"
  skip_create_ami = true
}

build {
  sources = [
    "source.amazon-ebs.basic-example"
  ]
}
