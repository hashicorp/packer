---
layout: "docs"
page_title: "Puppet (Masterless) Provisioner"
---

# Puppet (Masterless) Provisioner

Type: `puppet-masterless`

The masterless Puppet provisioner configures Puppet to run on the machines
by Packer from local modules and manifest files. Modules and manifests
can be uploaded from your local machine to the remote machine or can simply
use remote paths (perhaps obtained using something like the shell provisioner).
Puppet is run in masterless mode, meaning it never communicates to a Puppet
master.

<div class="alert alert-info alert-block">
<strong>Note that Puppet will <em>not</em> be installed automatically
by this provisioner.</strong> This provisioner expects that Puppet is already
installed on the machine. It is common practice to use the
<a href="/docs/provisioners/shell.html">shell provisioner</a> before the
Puppet provisioner to do this.
</div>

## Basic Example

The example below is fully functional and expects the configured manifest
file to exist relative to your working directory:

<pre class="prettyprint">
{
  "type": "puppet-masterless",
  "manifest_file": "site.pp"
}
</pre>

## Configuration Reference

The reference of available configuration options is listed below.

Required parameters:

* `manifest_file` (string) - The manifest file for Puppet to use in order
  to compile and run a catalog. This file must exist on your local system
  and will be uploaded to the remote machine.

Optional parameters:

* `execute_command` (string) - The command used to execute Puppet. This has
  various [configuration template variables](/docs/templates/configuration-templates.html)
  available. See below for more information.

* `facter` (object, string keys and values) - Additonal
  [facts](http://puppetlabs.com/puppet/related-projects/facter) to make
  available when Puppet is running.

* `hiera_config_path` (string) - The path to a local file with hiera
  configuration to be uploaded to the remote machine. Hiera data directories
  must be uploaded using the file provisioner separately.

* `manifest_dir` (string) - The path to a local directory with manifests
  to be uploaded to the remote machine. This is useful if your main
  manifest file uses imports. This directory doesn't necessarily contain
  the `manifest_file`. It is a separate directory that will be set as
  the "manifestdir" setting on Puppet.

* `module_paths` (array of strings) - This is an array of paths to module
  directories on your local filesystem. These will be uploaded to the remote
  machine. By default, this is empty.

* `prevent_sudo` (boolean) - By default, the configured commands that are
  executed to run Puppet are executed with `sudo`. If this is true,
  then the sudo will be omitted.

* `staging_directory` (string) - This is the directory where all the configuration
  of Puppet by Packer will be placed. By default this is "/tmp/packer-puppet-masterless".
  This directory doesn't need to exist but must have proper permissions so that
  the SSH user that Packer uses is able to create directories and write into
  this folder. If the permissions are not correct, use a shell provisioner
  prior to this to configure it properly.

## Execute Command

By default, Packer uses the following command (broken across multiple lines
for readability) to execute Puppet:

```
{{.FacterVars}}{{if .Sudo}} sudo -E {{end}}puppet apply \
  --verbose \
  --modulepath='{{.ModulePath}}' \
  {{if ne .HieraConfigPath ""}}--hiera_config='{{.HieraConfigPath}}' {{end}} \
  {{if ne .ManifestDir ""}}--manifestdir='{{.ManifestDir}}' {{end}} \
  --detailed-exitcodes \
  {{.ManifestFile}}
```

This command can be customized using the `execute_command` configuration.
As you can see from the default value above, the value of this configuration
can contain various template variables, defined below:

* `FacterVars` - Shell-friendly string of environmental variables used
  to set custom facts configured for this provisioner.
* `HieraConfigPath` - The path to a hiera configuration file.
* `ManifestFile` - The path on the remote machine to the manifest file
  for Puppet to use.
* `ModulePath` - The paths to the module directories.
* `Sudo` - A boolean of whether to `sudo` the command or not, depending on
  the value of the `prevent_sudo` configuration.

