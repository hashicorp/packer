---
description: |
    Packer can create machine images for any platform. Packer ships with support for
    a set of platforms, but can be extended through plugins to support any platform.
    This page documents the list of supported image types that Packer supports
    creating.
layout: intro
next_title: 'Packer & the HashiCorp Ecosystem'
next_url: '/intro/hashicorp-ecosystem.html'
page_title: Supported Platforms
prev_url: '/intro/use-cases.html'
...

# Supported Platforms

Packer can create machine images for any platform. Packer ships with support for
a set of platforms, but can be [extended through
plugins](/docs/extend/builder.html) to support any platform. This page documents
the list of supported image types that Packer supports creating.

If you were looking to see what platforms Packer is able to run on, see the page
on [installing Packer](/intro/getting-started/setup.html).

-&gt; **Note:** We're always looking to officially support more target
platforms. If you're interested in adding support for another platform, please
help by opening an issue or pull request within
[GitHub](https://github.com/mitchellh/packer) so we can discuss how to make it
happen.

Packer supports creating images for the following platforms or targets. The
format of the resulting image and any high-level information about the platform
is noted. They are listed in alphabetical order.

-   ***[Amazon EC2 (AMI)](/docs/builders/amazon.html)***. Both EBS-backed and instance-store AMIs within
    [EC2](https://aws.amazon.com/ec2/), optionally distributed to
    multiple regions.

-   ***[Azure Resource Manager (VHD)](/docs/builders/azure-arm.html)***. VHDs for [Azure Resource Manager](https://azure.microsoft.com/en-us/documentation/articles/resource-group-overview/), which can be deployed using the Azure Portal or Azure CLI.

-   ***[CloudStack](/docs/builders/cloudstack.html)***. Images for [CloudStack](https://cloudstack.apache.org/)
    that can be used to start pre-configured CloudStack servers.

-   ***[DigitalOcean](/docs/builders/digitalocean.html)***. Snapshots for
    [DigitalOcean](https://www.digitalocean.com/) that can be used to start a
    pre-configured DigitalOcean instance of any size.

-   ***[Docker](/docs/builders/docker.html)***. Snapshots for [Docker](https://www.docker.io/) that can be used
    to start a pre-configured Docker instance.

-   ***[Google Compute Engine](/docs/builders/googlecompute.html)***. Snapshots for [Google Compute
    Engine](https://cloud.google.com/products/compute-engine) that can be used
    to start a pre-configured Google Compute Engine instance.

-   ***[Hyper-V](/docs/builders/hyperv.html)***. Virtual machines for [Hyper-V](https://www.microsoft.com/en-us/server-cloud/solutions/virtualization.aspx), which are exportable.

-   ***[1&1](/docs/builders/oneandone.html)***. Virtual machines for [1&1](https://www.1and1.com/).

-   ***[OpenStack](/docs/builders/openstack.html)***. Images for [OpenStack](https://www.openstack.org/) that can
    be used to start pre-configured OpenStack servers.

-   ***[Parallels (PVM)](/docs/builders/parallels.html)***. Exported virtual machines for
    [Parallels](https://www.parallels.com/downloads/desktop/), including virtual
    machine metadata such as RAM, CPUs, etc. These virtual machines are portable
    and can be started on any platform Parallels runs on.

-   ***[ProfitBricks](/docs/builders/profitbricks.html)***. Virtual machines for [ProfitBricks](https://www.profitbricks.com).

-   ***[QEMU](/docs/builders/qemu.html)***. Images for [KVM](http://www.linux-kvm.org/) or
    [Xen](http://www.xenproject.org/) that can be used to start pre-configured
    KVM or Xen instances.

-   ***[VirtualBox (OVF)](/docs/builders/virtualbox.html)***. Exported virtual machines for
    [VirtualBox](https://www.virtualbox.org/), including virtual machine
    metadata such as RAM, CPUs, etc. These virtual machines are portable and can
    be started on any platform VirtualBox runs on.

-   ***[VMware (VMX)](/docs/builders/vmware.html)***. Exported virtual machines for
    [VMware](https://www.vmware.com/) that can be run within any desktop products
    such as Fusion, Player, or Workstation, as well as server products such
    as vSphere.

As previously mentioned, these are just the target image types that Packer ships
with out of the box. You can always [extend Packer through
plugins](/docs/extend/builder.html) to support more.
