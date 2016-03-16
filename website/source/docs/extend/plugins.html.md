---
description: |
    Packer Plugins allow new functionality to be added to Packer without modifying
    the core source code. Packer plugins are able to add new commands, builders,
    provisioners, hooks, and more. In fact, much of Packer itself is implemented by
    writing plugins that are simply distributed with Packer. For example, all the
    commands, builders, provisioners, and more that ship with Packer are implemented
    as Plugins that are simply hardcoded to load with Packer.
layout: docs
page_title: 'Packer Plugins - Extend Packer'
...

# Packer Plugins

Packer Plugins allow new functionality to be added to Packer without modifying
the core source code. Packer plugins are able to add new commands, builders,
provisioners, hooks, and more. In fact, much of Packer itself is implemented by
writing plugins that are simply distributed with Packer. For example, all the
commands, builders, provisioners, and more that ship with Packer are implemented
as Plugins that are simply hardcoded to load with Packer.

This page will cover how to install and use plugins. If you're interested in
developing plugins, the documentation for that is available the [developing
plugins](/docs/extend/developing-plugins.html) page.

Because Packer is so young, there is no official listing of available Packer
plugins. Plugins are best found via Google. Typically, searching "packer plugin
*x*" will find what you're looking for if it exists. As Packer gets older, an
official plugin directory is planned.

## How Plugins Work

Packer plugins are completely separate, standalone applications that the core of
Packer starts and communicates with.

These plugin applications aren't meant to be run manually. Instead, Packer core
executes these plugin applications in a certain way and communicates with them.
For example, the VMware builder is actually a standalone binary named
`packer-builder-vmware`. The next time you run a Packer build, look at your
process list and you should see a handful of `packer-` prefixed applications
running.

## Installing Plugins

The easiest way to install a plugin is to name it correctly, then place it in
the proper directory. To name a plugin correctly, make sure the binary is named
`packer-TYPE-NAME`. For example, `packer-builder-amazon-ebs` for a "builder"
type plugin named "amazon-ebs". Valid types for plugins are down this page more.

Once the plugin is named properly, Packer automatically discovers plugins in the
following directories in the given order. If a conflicting plugin is found
later, it will take precedence over one found earlier.

1.  The directory where `packer` is, or the executable directory.

2.  `~/.packer.d/plugins` on Unix systems or `%APPDATA%/packer.d/plugins`
    on Windows.

3.  The current working directory.

The valid types for plugins are:

-   `builder` - Plugins responsible for building images for a specific platform.

-   `command` - A CLI sub-command for `packer`.

-   `post-processor` - A post-processor responsible for taking an artifact from
    a builder and turning it into something else.

-   `provisioner` - A provisioner to install software on images created by
    a builder.
