# Building Windows Images with AWS Session Manager

This example shows how to bootstrap a Windows EC2 instance with SSH and the AWS Systems Manager (SSM) agent so Packer can connect over SSM.

## Prerequisites

- AWS account with permissions for EC2, IAM, and SSM
- Packer >= 1.9
- An IAM instance profile that includes `AmazonSSMManagedInstanceCore`

## Steps

1. Launch a Windows base AMI in a private subnet with the SSM agent preinstalled (Windows Server AMIs include it by default).
2. Attach an instance profile that allows SSM.
3. Install OpenSSH Server on the instance (user data or manual) and open port 22 only if you also want SSH fallback.
4. Configure Packer to use the `winrm` or `ssh` communicator with SSM as the connection transport.

## Packer configuration sketch

```hcl
source "amazon-ebs" "windows" {
  ami           = "ami-WINDOWS_BASE"
  instance_type = "t3.large"
  region        = "us-east-1"

  communicator = "ssh"
  ssh_username = "Administrator"

  # Use the amazon plugin SSM session manager integration
  ssh_interface = "session_manager"
  iam_instance_profile = "packer-ssm-profile"
}

build {
  sources = ["source.amazon-ebs.windows"]

  provisioner "powershell" {
    inline = ["Write-Host 'Hello from Packer over SSM'"]
  }
}
```

## Verify SSM connectivity

Before running Packer, confirm the instance appears as Online in the AWS Systems Manager console (Fleet Manager > Managed nodes).

## Learn more

- [Packer Amazon builder](https://developer.hashicorp.com/packer/plugins/builders/amazon/ebs)
- [AWS Session Manager](https://docs.aws.amazon.com/systems-manager/latest/userguide/session-manager.html)
