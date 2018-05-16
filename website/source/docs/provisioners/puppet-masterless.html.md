---
description: |
    The masterless Puppet Packer provisioner configures Puppet to run on the
    machines by Packer from local modules and manifest files. Modules and
    manifests can be uploaded from your local machine to the remote machine or can
    simply use remote paths. Puppet is run in masterless mode, meaning it never
    communicates to a Puppet master.
layout: docs
page_title: 'Puppet Masterless - Provisioners'
sidebar_current: 'docs-provisioners-puppet-masterless'
---

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
to exist relative to your working directory.

``` json
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

-   `execute_command` (string) - The command-line to execute Puppet. This also has
    various [configuration template variables](/docs/templates/engine.html) available.

-   `extra_arguments` (array of strings) - Additional options to
    pass to the Puppet command. This allows for customization of  
    `execute_command` without having to completely replace
    or subsume its contents, making forward-compatible customizations much
    easier to maintain.
    
    This string is lazy-evaluated so one can incorporate logic driven by template variables as 
    well as private elements of ExecuteTemplate (see source: provisioner/puppet-masterless/provisioner.go).
```
[
  {{if ne "{{user environment}}" ""}}--environment={{user environment}}{{end}},
  {{if ne ".ModulePath" ""}}--modulepath="{{.ModulePath}}{{.ModulePathJoiner}}$(puppet config print {{if ne "{{user `environment`}}" ""}}--environment={{user `environment`}}{{end}} modulepath)"{{end}}
]
```

-   `facter` (object of key:value strings) - Additional
    [facts](https://puppetlabs.com/facter) to make
    available to the Puppet run.

-   `guest_os_type` (string) - The remote host's OS type ('windows' or 'unix') to
    tailor command-line and path separators. (default: unix).

-   `hiera_config_path` (string) - Local path to self-contained Hiera
    data to be uploaded. NOTE: If you need data directories
    they must be previously transferred with a File provisioner.

-   `ignore_exit_codes` (boolean) - If true, Packer will ignore failures.

-   `manifest_dir` (string) - Local directory with manifests to be
    uploaded. This is useful if your main manifest uses imports, but the
    directory might not contain the `manifest_file` itself.

~&gt; `manifest_dir` is passed to Puppet as `--manifestdir` option.
This option was deprecated in puppet 3.6, and removed in puppet 4.0. If you have
multiple manifests you should use `manifest_file` instead.

-   `module_paths` (array of strings) - Array of local module
    directories to be uploaded.

-   `prevent_sudo` (boolean) - On Unix platforms Puppet is typically invoked with `sudo`. If true,
    it will be omitted. (default: false)

-   `puppet_bin_dir` (string) - Path to the Puppet binary. Ideally the program
    should be on the system (unix: `$PATH`, windows: `%PATH%`), but some builders (eg. Docker) do
    not run profile-setup scripts and therefore PATH might be empty or minimal.

-   `staging_directory` (string) - Directory to where uploaded files
    will be placed (unix: "/tmp/packer-puppet-masterless",
    windows: "%SYSTEMROOT%/Temp/packer-puppet-masterless").
    It doesn't need to pre-exist, but the parent must have permissions sufficient
    for the account Packer connects as to create directories and write files.
    Use a Shell provisioner to prepare the way if needed.

-   `working_directory` (string) - Directory from which `execute_command` will be run.
    If using Hiera files with relative paths, this option can be helpful. (default: `staging_directory`)

## Execute Command

By default, Packer uses the following command (broken across multiple lines for
readability) to execute Puppet:

```
cd {{.WorkingDir}} &&
	{{if ne .FacterVars ""}}{{.FacterVars}} {{end}}
	{{if .Sudo}}sudo -E {{end}}
	{{if ne .PuppetBinDir ""}}{{.PuppetBinDir}}/{{end}}
  puppet apply --detailed-exitcodes
	{{if .Debug}}--debug {{end}}
	{{if ne .ModulePath ""}}--modulepath='{{.ModulePath}}' {{end}}
	{{if ne .HieraConfigPath ""}}--hiera_config='{{.HieraConfigPath}}' {{end}}
	{{if ne .ManifestDir ""}}--manifestdir='{{.ManifestDir}}' {{end}}
	{{if ne .ExtraArguments ""}}{{.ExtraArguments}} {{end}}
	{{.ManifestFile}}
```

The following command is used if guest OS type is windows:

```
cd {{.WorkingDir}} &&
	{{if ne .FacterVars ""}}{{.FacterVars}} && {{end}}
	{{if ne .PuppetBinDir ""}}{{.PuppetBinDir}}/{{end}}
  puppet apply --detailed-exitcodes
	{{if .Debug}}--debug {{end}}
	{{if ne .ModulePath ""}}--modulepath='{{.ModulePath}}' {{end}}
	{{if ne .HieraConfigPath ""}}--hiera_config='{{.HieraConfigPath}}' {{end}}
	{{if ne .ManifestDir ""}}--manifestdir='{{.ManifestDir}}' {{end}}
	{{if ne .ExtraArguments ""}}{{.ExtraArguments}} {{end}}
	{{.ManifestFile}}
```

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
