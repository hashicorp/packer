---
description: |
    The VSphere Packer builder is able to create VSphere virtual machines for use with
    any VSphere product.
layout: docs
page_title: VSphere Builder
...

# VSphere Builder

The VSphere Packer builder is able to create VSphere virtual machines for use with
any VSphere product using the VSphere API.

Packer actually comes with multiple builders able to create VSphere machines,
depending on the strategy you want to use to build the image. Packer supports
the following VSphere builders:

-   [VSphere-iso](/docs/builders/VSphere-iso.html) - Starts from an ISO file,
    creates a brand new VSphere VM, installs an OS, provisions software within
    the OS, then exports that machine to create an image. This is best for
    people who want to start from scratch.

-   [VSphere-vm](/docs/builders/VSphere-vm.html) - This builder clone an
    existing VSphere machine, runs provisioners on top of that
    VM, and exports that machine to create an image. This is best if you have an
    existing VSphere VM you want to use as the source. As an additional benefit,
    you can feed the artifact of this builder back into Packer to iterate on
    a machine.


-&gt; **Note:** Packer supports ESXi 5.1 and above.



When using a remote VSphere Hypervisor, the builder still downloads the ISO and
various files locally, and uploads these to the remote machine. 

Packer also requires VNC to issue boot commands during a build, which may be
disabled on some remote VSphere Hypervisors. Please consult the appropriate
documentation on how to update VSphere Hypervisor's firewall to allow these
connections.

Before using a remote vSphere Hypervisor, you need to enable GuestIPHack by
running the following command:

``` {.text}
esxcli system settings advanced set -o /Net/GuestIPHack -i 1
```


**NOTE:** Due to licences restriction (full API availability), this builder can only works with licenced VSphere product (Vcenter or ESXi). For using free ESXi, the only solution is to use the vmware builder that use direct ssh connection rather than API calls.
