---
description: |
    Pre-baked machine images have a lot of advantages, but most have been unable to
    benefit from them because images have been too tedious to create and manage.
    There were either no existing tools to automate the creation of machine images
    or they had too high of a learning curve. The result is that, prior to Packer,
    creating machine images threatened the agility of operations teams, and
    therefore aren't used, despite the massive benefits.
layout: intro
next_title: Packer Use Cases
next_url: '/intro/use-cases.html'
page_title: 'Why Use Packer?'
prev_url: '/intro/index.html'
...

# Why Use Packer?

Pre-baked machine images have a lot of advantages, but most have been unable to
benefit from them because images have been too tedious to create and manage.
There were either no existing tools to automate the creation of machine images
or they had too high of a learning curve. The result is that, prior to Packer,
creating machine images threatened the agility of operations teams, and
therefore aren't used, despite the massive benefits.

Packer changes all of this. Packer is easy to use and automates the creation of
any type of machine image. It embraces modern configuration management by
encouraging you to use a framework such as Chef or Puppet to install and
configure the software within your Packer-made images.

In other words: Packer brings pre-baked images into the modern age, unlocking
untapped potential and opening new opportunities.

## Advantages of Using Packer

***Super fast infrastructure deployment***. Packer images allow you to launch
completely provisioned and configured machines in seconds, rather than several
minutes or hours. This benefits not only production, but development as well,
since development virtual machines can also be launched in seconds, without
waiting for a typically much longer provisioning time.

***Multi-provider portability***. Because Packer creates identical images for
multiple platforms, you can run production in AWS, staging/QA in a private cloud
like OpenStack, and development in desktop virtualization solutions such as
VMware or VirtualBox. Each environment is running an identical machine image,
giving ultimate portability.

***Improved stability***. Packer installs and configures all the software for a
machine at the time the image is built. If there are bugs in these scripts,
they'll be caught early, rather than several minutes after a machine is
launched.

***Greater testability***. After a machine image is built, that machine image
can be quickly launched and smoke tested to verify that things appear to be
working. If they are, you can be confident that any other machines launched from
that image will function properly.

Packer makes it extremely easy to take advantage of all these benefits.

What are you waiting for? Let's get started!
