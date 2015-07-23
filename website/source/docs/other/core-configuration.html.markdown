---
description: |
    There are a few configuration settings that affect Packer globally by
    configuring the core of Packer. These settings all have reasonable defaults, so
    you generally don't have to worry about it until you want to tweak a
    configuration. If you're just getting started with Packer, don't worry about
    core configuration for now.
layout: docs
page_title: Core Configuration
...

# Core Configuration

There are a few configuration settings that affect Packer globally by
configuring the core of Packer. These settings all have reasonable defaults, so
you generally don't have to worry about it until you want to tweak a
configuration. If you're just getting started with Packer, don't worry about
core configuration for now.

The default location where Packer looks for this file depends on the platform.
For all non-Windows platforms, Packer looks for `$HOME/.packerconfig`. For
Windows, Packer looks for `%APPDATA%/packer.config`. If the file doesn't exist,
then Packer ignores it and just uses the default configuration.

The location of the core configuration file can be modified by setting the
`PACKER_CONFIG` environmental variable to be the path to another file.

The format of the configuration file is basic JSON.

## Configuration Reference

Below is the list of all available configuration parameters for the core
configuration file. None of these are required, since all have sane defaults.

-   `plugin_min_port` and `plugin_max_port` (integer) - These are the minimum
    and maximum ports that Packer uses for communication with plugins, since
    plugin communication happens over TCP connections on your local host. By
    default these are 10,000 and 25,000, respectively. Be sure to set a fairly
    wide range here, since Packer can easily use over 25 ports on a single run.

-   `builders`, `commands`, `post-processors`, and `provisioners` are objects
    that are used to install plugins. The details of how exactly these are set
    is covered in more detail in the [installing plugins documentation
    page](/docs/extend/plugins.html).
