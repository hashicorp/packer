---
description: |
    The masterless Puppet Packer provisioner configures Puppet to run on the
    machines by Packer from local modules and manifest files. Modules and manifests
    can be uploaded from your local machine to the remote machine or can simply use
    remote paths (perhaps obtained using something like the shell provisioner).
    Puppet is run in masterless mode, meaning it never communicates to a Puppet
    master.
layout: docs
page_title: 'Puppet (Masterless) Provisioner'
...

# Puppet (Masterless) Provisioner

Type: `puppet-masterless`

The masterless Puppet Packer provisioner configures Puppet to run on the
machines by Packer from local modules and manifest files. Modules and manifests
can be uploaded from your local machine to the remote machine or can simply use
remote paths (perhaps obtained using something like the shell provisioner).
Puppet is run in masterless mode, meaning it never communicates to a Puppet
master.

-&gt; **Note:** Puppet will *not* be installed automatically by this
provisioner. This provisioner expects that Puppet is already installed on the
machine. It is common practice to use the [shell
provisioner](/docs/provisioners/shell.html) before the Puppet provisioner to do
this.

## Basic Example

The example below is fully functional and expects the configured manifest file
to exist relative to your working directory:

``` {.javascript}
{
  "type": "puppet-masterless",
  "manifest_file": "site.pp"
}
```

## Configuration Reference

The reference of available configuration options is listed below.

Required parameters:

-   `manifest_file` (string) - This is either a path to a puppet manifest
    (`.pp` file) *or* a directory containing multiple manifests that puppet will
    apply (the ["main
    manifest"](https://docs.puppetlabs.com/puppet/latest/reference/dirs_manifest.html)).
    These file(s) must exist on your local system and will be uploaded to the
    remote machine.

Optional parameters:

-   `execute_command` (string) - The command used to execute Puppet. This has
    various [configuration template
    variables](/docs/templates/configuration-templates.html) available. See
    below for more information.

-   `extra_arguments` (array of strings) - This is an array of additional options to
    pass to the puppet command when executing puppet. This allows for
    customization of the `execute_command` without having to completely replace
    or include it's contents, making forward-compatible customizations much
    easier.

-   `facter` (object of key/value strings) - Additional
    [facts](https://puppetlabs.com/facter) to make
    available when Puppet is running.

-   `hiera_config_path` (string) - The path to a local file with hiera
    configuration to be uploaded to the remote machine. Hiera data directories
    must be uploaded using the file provisioner separately.

-   `ignore_exit_codes` (boolean) - If true, Packer will never consider the
    provisioner a failure.

-   `manifest_dir` (string) - The path to a local directory with manifests to be
    uploaded to the remote machine. This is useful if your main manifest file
    uses imports. This directory doesn't necessarily contain the
    `manifest_file`. It is a separate directory that will be set as the
    "manifestdir" setting on Puppet.

\~&gt; `manifest_dir` is passed to `puppet apply` as the `--manifestdir` option.
This option was deprecated in puppet 3.6, and removed in puppet 4.0. If you have
multiple manifests you should use `manifest_file` instead.

-   `module_paths` (array of strings) - This is an array of paths to module
    directories on your local filesystem. These will be uploaded to the
    remote machine. By default, this is empty.

-   `prevent_sudo` (boolean) - By default, the configured commands that are
    executed to run Puppet are executed with `sudo`. If this is true, then the
    sudo will be omitted.

-   `staging_directory` (string) - This is the directory where all the
    configuration of Puppet by Packer will be placed. By default this
    is "/tmp/packer-puppet-masterless". This directory doesn't need to exist but
    must have proper permissions so that the SSH user that Packer uses is able
    to create directories and write into this folder. If the permissions are not
    correct, use a shell provisioner prior to this to configure it properly.

-   `working_directory` (string) - This is the directory from which the puppet
    command will be run. When using hiera with a relative path, this option
    allows to ensure that the paths are working properly. If not specified,
    defaults to the value of specified `staging_directory` (or its default value
    if not specified either).

## Execute Command

By default, Packer uses the following command (broken across multiple lines for
readability) to execute Puppet:

``` {.liquid}
cd {{.WorkingDir}} && \
{{.FacterVars}}{{if .Sudo}} sudo -E {{end}}puppet apply \
  --verbose \
  --modulepath='{{.ModulePath}}' \
  {{if ne .HieraConfigPath ""}}--hiera_config='{{.HieraConfigPath}}' {{end}} \
  {{if ne .ManifestDir ""}}--manifestdir='{{.ManifestDir}}' {{end}} \
  --detailed-exitcodes \
  {{.ManifestFile}}
```

This command can be customized using the `execute_command` configuration. As you
can see from the default value above, the value of this configuration can
contain various template variables, defined below:

-   `WorkingDir` - The path from which Puppet will be executed.
-   `FacterVars` - Shell-friendly string of environmental variables used to set
    custom facts configured for this provisioner.
-   `HieraConfigPath` - The path to a hiera configuration file.
-   `ManifestFile` - The path on the remote machine to the manifest file for
    Puppet to use.
-   `ModulePath` - The paths to the module directories.
-   `Sudo` - A boolean of whether to `sudo` the command or not, depending on the
    value of the `prevent_sudo` configuration.

## Default Facts

In addition to being able to specify custom Facter facts using the `facter`
configuration, the provisioner automatically defines certain commonly useful
facts:

-   `packer_build_name` is set to the name of the build that Packer is running.
    This is most useful when Packer is making multiple builds and you want to
    distinguish them in your Hiera hierarchy.

-   `packer_builder_type` is the type of the builder that was used to create the
    machine that Puppet is running on. This is useful if you want to run only
    certain parts of your Puppet code on systems built with certain builders.
