variable "ssh-username" {
  type = string
}

source "amazon-ebs" "linux" {
  region     = var.region
  access_key = var.aws_access_key
  secret_key = var.aws_secret_key

  ami_name      = "linux-app"
  source_ami    = "ami-06e46074ae430fba6" # Amazon Linux 2023 x86-64
  instance_type = "t2.micro"
  communicator  = "ssh"
  ssh_username  = var.ssh-username
  ssh_timeout   = "45s"
}

build {
  sources = ["amazon-ebs.linux"]

  // Other provisioners/post-processors
}
