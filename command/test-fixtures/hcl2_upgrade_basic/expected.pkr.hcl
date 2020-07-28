variable "location" {
  default = "westus"
}

variable "secret_account" {
  default   = "dark_vader"
  sensitive = true
}

variable "secret_password" {
  default   = "42"
  sensitive = true
}

# "timestamp" template function replacement
locals { timestamp = regex_replace(timestamp(), "[- TZ:]", "") }

source "amazon-ebs" "1" {
  ami_description              = "Ubuntu 16.04 LTS - expand root partition"
  ami_name                     = "ubuntu-16.04 test ${local.timestamp}"
  encrypt_boot                 = true
  launch_block_device_mappings = [{ delete_on_termination = true, device_name = "/dev/sda1", volume_size = 48, volume_type = "gp2" }]
  region                       = "eu-west-1"
  source_ami_filter            = { filters = { name = "ubuntu/images/*/ubuntu-xenial-16.04-amd64-server-*", root-device-type = "ebs", virtualization-type = "hvm" }, most_recent = true, owners = ["099720109477"] }
  spot_instance_types          = ["t2.small", "t2.medium", "t2.large"]
  spot_price                   = "0.0075"
  ssh_username                 = "ubuntu"
}

build {
  sources = ["source.amazon-ebs.1"]

  provisioner "shell" {
    inline = ["echo ${var.secret_account}", "echo ${build.ID}", "echo ${build.SSHPrivateKey}", "sleep 100000"]
  }
  provisioner "shell-local" {
    inline = ["sleep 100000"]
  }
  post-processor "amazon-import" {
    format         = "vmdk"
    license_type   = "BYOL"
    region         = "eu-west-3"
    s3_bucket_name = "hashicorp.adrien"
    tags           = { Description = "packer amazon-import ${local.timestamp}" }
  }
  post-processors {
    post-processor "artifice" {
      files = ["path/something.ova"]
    }
    post-processor "amazon-import" {
      license_type   = "BYOL"
      s3_bucket_name = "hashicorp.adrien"
      tags           = { Description = "packer amazon-import ${local.timestamp}" }
    }
  }
}
