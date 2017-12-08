---
layout: guides
sidebar_current: guides-packer-on-cicd-build-virtualbox
page_title: Building a VirtualBox Image with Packer in TeamCity
---

# Building a VirtualBox Image with Packer in TeamCity

This guide walks through the process of building a VirtualBox image using Packer on a new TeamCity Agent. Before getting started you should have access to a TeamCity Server.

The Packer VirtualBox builder requires access to VirtualBox, which needs to run on a bare-metal machine as virtualization is generally not supported on cloud instances. This is also true for the [VMWare](https://www.packer.io/docs/builders/vmware.html) and the [QEMU](https://www.packer.io/docs/builders/qemu.html) Packer builders.

## 1. Provision a Bare-metal Machine

The Packer VirtualBox builder requires running on bare-metal (hardware). If you do not have access to a bare-metal machine, we recommend using [Packet.net](https://www.packet.net/) to obtain a new machine. If you are a first time user of Packet.net, the Packet.net team has provided HashiCorp the coupon code `hash25` which you can use for $25 off to test out this guide. You can use a `baremetal_0` for testing, but for regular use the `baremetal_1` instance may be a better option.

There is also a [Packet Provider](https://www.terraform.io/docs/providers/packet/index.html) in Terraform you can use to provision the project and instance.

```hcl
provider "packet" { }

resource "packet_project" "teamcity_agents" {
  name = "TeamCity"
}

resource "packet_device" "agent" {
  hostname         = "teamcity-agent"
  plan             = "baremetal_0"
  facility         = "ams1"
  operating_system = "ubuntu_16_04"
  billing_cycle    = "hourly"
  project_id       = "${packet_project.teamcity_project.id}"
}
```

## 2. Install VirtualBox and TeamCity dependencies

VirtualBox must be installed on the new instance, and TeamCity requires the JDK prior to installation. This guide uses Ubuntu as the Linux distribution, so you may need to adjust these commands for your distribution of choice.

**Install Teamcity Dependencies**

```shell
apt-get upgrade
apt-get install -y zip linux-headers-generic linux-headers-4.13.0-16-generic build-essential openjdk-8-jdk
```

**Install VirtualBox**

```
curl -OL "http://download.virtualbox.org/virtualbox/5.2.2/virtualbox-5.2_5.2.2-119230~Ubuntu~xenial_amd64.deb"
dpkg -i virtualbox-5.2_5.2.2-119230~Ubuntu~xenial_amd64.deb
```

You can also use the [`remote-exec` provisioner](https://www.terraform.io/docs/provisioners/remote-exec.html) in your Terraform configuration to automatically run these commands when provisioning the new instance.

## 3. Install Packer

The TeamCity Agent machine will also need Packer Installed. You can find the latest download link from the [Packer Download](https://www.packer.io/downloads.html) page.

```shell
curl -OL "https://releases.hashicorp.com/packer/1.1.2/packer_1.1.2_linux_amd64.zip"
unzip ./packer_1.1.2_linux_amd64.zip
```

Packer is installed at the `/root/packer` path which is used in subsequent steps. If it is installed elsewhere, take note of the path.

## 4. Install TeamCity Agent

This guide assume you already have a running instance of TeamCity Server. The new TeamCity Agent can be installed by [downloading a zip file and installing manually](https://confluence.jetbrains.com/display/TCD10//Setting+up+and+Running+Additional+Build+Agents#SettingupandRunningAdditionalBuildAgents-InstallingAdditionalBuildAgents), or using [Agent Push](https://confluence.jetbrains.com/display/TCD10//Setting+up+and+Running+Additional+Build+Agents#SettingupandRunningAdditionalBuildAgents-InstallingviaAgentPush). Once it is installed it should appear in TeamCity as a new Agent.

Create a new Agent Pool for agents responsible for VirtualBox Packer builds and assign the new Agent to it.

## 5. Create a New Build in TeamCity

In TeamCity Server create a new build and configure the Version Control Settings to download the Packer build configuration from the VCS repository.

Add one **Build Step: Command Line** to the build.

![TeamCity screenshot: New Build](/assets/images/guides/teamcity_new_build.png)

In the **Script content** field add the following:

```shell
#!/usr/bin/env bash
/root/packer build -only=virtualbox-iso -var "headless=true" ./packer.json
```

This assumes that `packer.json` is the Packer build configuration file in the root path of the VCS repository.

## 6. Run a build in TeamCity

The entire configuration is ready for a new build. Start a new run in TeamCity by pressing “Run”.

The new run should be triggered and the virtual box image will be built.

![TeamCity screenshot: Build log](/assets/images/guides/teamcity_build_log.png)

Once complete, the build status should be updated to complete and successful.

![TeamCity screenshot: Build log complete](/assets/images/guides/teamcity_build_log_complete.png)
