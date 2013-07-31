---
layout: "docs"
page_title: "Amazon AMI Builder"
---

# Amazon AMI Builder

Packer is able to create Amazon AMIs. To achieve this, Packer comes with
multiple builders depending on the strategy you want to use to build the
AMI. Packer supports the following builders at the moment:

* [amazon-ebs](/docs/builders/amazon-ebs.html) - Create EBS-backed AMIs
  by launching a source instance and re-packaging it into a new AMI after
  provisioning. If in doubt, use this builder, which is the easiest to get
  started with.

* [amazon-instance](/docs/builders/amazon-instance.html) - Create
  instance-store AMIs by launching and provisioning a source instance, then
  rebundling it and uploading it to S3.

* [amazon-chroot](/docs/builders/amazon-chroot.html) - Create EBS-backed AMIs
  from an existing EC2 instance by mounting the root device and using a
  [Chroot](http://en.wikipedia.org/wiki/Chroot) environment to provision
  that device. This is an **advanced builder and should not be used by
  newcomers**. However, it is also the fastest way to build an EBS-backed
  AMI since no new EC2 instance needs to be launched.

<div class="alert alert-block alert-info">
<strong>Don't know which builder to use?</strong> If in doubt, use the
<a href="/docs/builders/amazon-ebs.html">amazon-ebs builder</a>. It is
much easier to use and Amazon generally recommends EBS-backed images nowadays.
</div>
