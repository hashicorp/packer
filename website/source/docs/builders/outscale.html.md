---
description: |
    Packer is able to create Outscale Machine Images (OMIs). To achieve this, Packer comes with
    multiple builders depending on the strategy you want to use to build the OMI.
layout: docs
page_title: 'Outscale OMI - Builders'
sidebar_current: 'docs-builders-outscale'
---

# Outscale OMI Builder

Packer is able to create Outscale OMIs. To achieve this, Packer comes with
multiple builders depending on the strategy you want to use to build the OMI.
Packer supports the following builders at the moment:

- [osc-bsu](/docs/builders/osc-bsu.html) - Create BSU-backed OMIs by
    launching a source OMI and re-packaging it into a new OMI after
    provisioning. If in doubt, use this builder, which is the easiest to get
    started with.

- [osc-chroot](/docs/builders/osc-chroot.html) - Create EBS-backed OMIs
    from an existing OUTSCALE VM by mounting the root device and using a
    [Chroot](https://en.wikipedia.org/wiki/Chroot) environment to provision
    that device. This is an **advanced builder and should not be used by
    newcomers**. However, it is also the fastest way to build an EBS-backed OMI
    since no new OUTSCALE VM needs to be launched.

- [osc-bsusurrogate](/docs/builders/osc-bsusurrogate.html) - Create BSU-backed OMIs from scratch. Works similarly to the `chroot` builder but does
    not require running in Outscale VM. This is an **advanced builder and should not be
    used by newcomers**.

-&gt; **Don't know which builder to use?** If in doubt, use the [osc-bsu
builder](/docs/builders/osc-bsu.html). It is much easier to use and Outscale generally recommends BSU-backed images nowadays.

# Outscale BSU Volume Builder

Packer is able to create Outscale BSU Volumes which are preinitialized with a filesystem and data.

- [osc-bsuvolume](/docs/builders/osc-bsuvolume.html) - Create EBS volumes by launching a source OMI with block devices mapped. Provision the VM, then destroy it, retaining the EBS volumes.

## Authentication

The OUTSCALE provider offers a flexible means of providing credentials for authentication. The following methods are supported, in this order, and explained below:

- Static credentials
- Environment variables
- Shared credentials file
- Outscale Role

### Static Credentials

Static credentials can be provided in the form of an access key id and secret.
These look like:

``` json
{
    "access_key": "AKIAIOSFODNN7EXAMPLE",
    "secret_key": "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY",
    "region": "us-east-1",
    "type": "osc-bsu",
    "oapi_custom_endpoint": "outscale.com/oapi/latest"
}
```

### Environment variables

You can provide your credentials via the `OUTSCALE_ACCESSKEYID` and
`OUTSCALE_SECRETKEYID`, environment variables, representing your Outscale Access
Key and Outscale Secret Key, respectively. The `OUTSCALE_REGION` and
`OUTSCALE_OAPI_URL` environment variables are also used, if applicable:

Usage:

    $ export OUTSCALE_ACCESSKEYID="anaccesskey"
    $ export OUTSCALE_SECRETKEYID="asecretkey"
    $ export OUTSCALE_REGION="eu-west-2"
    $ packer build packer.json

### Checking that system time is current

Outscale uses the current time as part of the [request signing
process](http://docs.aws.osc.com/general/latest/gr/sigv4_signing.html). If
your system clock is too skewed from the current time, your requests might
fail. If that's the case, you might see an error like this:

    ==> osc-bsu: Error querying OMI: AuthFailure: OUTSCALE was not able to validate the provided access credentials

If you suspect your system's date is wrong, you can compare it against
<http://www.time.gov/>. On Linux/OS X, you can run the `date` command to get
the current time. If you're on Linux, you can try setting the time with ntp by
running `sudo ntpd -q`.
