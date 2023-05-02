variable "windows_user" {
  type = string
}

variable "windows_password" {
  type = string
}

data "file" "user_data_file" {
  contents = templatefile("scripts/enable_winrm.ps", {
    "winrm_user"     = var.windows_user,
    "winrm_password" = var.windows_password,
  })
  destination = "enable_winrm"
  force = true
}

source "amazon-ebs" "windows" {
  region     = var.region
  access_key = var.aws_access_key
  secret_key = var.aws_secret_key

  ami_name       = "windows-app"
  source_ami     = "ami-00b2c40b15619f518" # Windows server 2016 base x86_64
  instance_type  = "m3.medium"
  communicator   = "winrm"
  winrm_username = var.windows_user
  winrm_password = var.windows_password
  user_data_file = data.file.user_data_file.path
}

build {
  sources = ["amazon-ebs.windows"]

  provisioner "powershell" {
    inline = [
      "C:/ProgramData/Amazon/EC2-Windows/Launch/Scripts/InitializeInstance.ps1 -Schedule",
      "C:/ProgramData/Amazon/EC2-Windows/Launch/Scripts/SysprepInstance.ps1 -NoShutdown"
    ]
  }

  // Other provisioners/post-processors
}
