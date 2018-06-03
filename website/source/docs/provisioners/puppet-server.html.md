---
description: |
    The puppet-server Packer provisioner provisions Packer machines with Puppet
    by connecting to a Puppet master.
layout: docs
page_title: 'Puppet Server - Provisioners'
sidebar_current: 'docs-provisioners-puppet-server'
---

# Puppet Server Provisioner

Type: `puppet-server`

The `puppet-server` Packer provisioner provisions Packer machines with Puppet by
connecting to a Puppet master.

-&gt; **Note:** Puppet will *not* be installed automatically by this
provisioner. This provisioner expects that Puppet is already installed on the
machine. It is common practice to use the [shell
provisioner](/docs/provisioners/shell.html) before the Puppet provisioner to do
this.

## Basic Example

The example below is fully functional and expects a Puppet server to be
accessible from your network.

``` json
{
   "type": "puppet-server",
   "extra_arguments": "--test --pluginsync",
   "facter": {
     "server_role": "webserver"
   }
}
```

## Configuration Reference

The reference of available configuration options is listed below.

The provisioner takes various options. None are strictly required. They are
listed below:

-   `client_cert_path` (string) - Path to the directory on your disk that
    contains the client certificate for the node. This defaults to nothing,
    in which case a client cert won't be uploaded.

-   `client_private_key_path` (string) - Path to the directory on your disk that
    contains the client private key for the node. This defaults to nothing, in
    which case a client private key won't be uploaded.

-   `execute_command` (string) - The command-line to execute Puppet. This also has
    various [configuration template variables](/docs/templates/engine.html) available.

-   `extra_arguments` (array of strings) - Additional options to
    pass to the Puppet command. This allows for customization of
    `execute_command` without having to completely replace
    or subsume its contents, making forward-compatible customizations much
    easier to maintain.
    
    This string is lazy-evaluated so one can incorporate logic driven by template variables as
    well as private elements of ExecuteTemplate (see source: provisioner/puppet-server/provisioner.go).
```
[
  {{if ne "{{user environment}}" ""}}--environment={{user environment}}{{end}}
]
```

-   `facter` (object of key/value strings) - Additional
    [facts](https://puppetlabs.com/facter) to make
    available to the Puppet run.

-   `guest_os_type` (string) - The remote host's OS type ('windows' or 'unix') to
    tailor command-line and path separators. (default: unix).

-   `ignore_exit_codes` (boolean) - If true, Packer will ignore failures.

-   `prevent_sudo` (boolean) - On Unix platforms Puppet is typically invoked with `sudo`. If true,
    it will be omitted. (default: false)

-   `puppet_bin_dir` (string) - Path to the Puppet binary. Ideally the program
    should be on the system (unix: `$PATH`, windows: `%PATH%`), but some builders (eg. Docker) do
    not run profile-setup scripts and therefore PATH might be empty or minimal.

-   `puppet_node` (string) - The name of the node. If this isn't set, the fully
    qualified domain name will be used.

-   `puppet_server` (string) - Hostname of the Puppet server. By default
    "puppet" will be used.

-   `staging_dir` (string) - Directory to where uploaded files
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
  puppet agent --onetime --no-daemonize --detailed-exitcodes
	{{if .Debug}}--debug {{end}}
	{{if ne .PuppetServer ""}}--server='{{.PuppetServer}}' {{end}}
	{{if ne .PuppetNode ""}}--certname='{{.PuppetNode}}' {{end}}
	{{if ne .ClientCertPath ""}}--certdir='{{.ClientCertPath}}' {{end}}
	{{if ne .ClientPrivateKeyPath ""}}--privatekeydir='{{.ClientPrivateKeyPath}}' {{end}}
	{{if ne .ExtraArguments ""}}{{.ExtraArguments}} {{end}}
```

The following command is used if guest OS type is windows:

```
cd {{.WorkingDir}} &&
	{{if ne .FacterVars ""}}{{.FacterVars}} && {{end}}
	{{if ne .PuppetBinDir ""}}{{.PuppetBinDir}}/{{end}}
  puppet agent --onetime --no-daemonize --detailed-exitcodes
	{{if .Debug}}--debug {{end}}
	{{if ne .PuppetServer ""}}--server='{{.PuppetServer}}' {{end}}
	{{if ne .PuppetNode ""}}--certname='{{.PuppetNode}}' {{end}}
	{{if ne .ClientCertPath ""}}--certdir='{{.ClientCertPath}}' {{end}}
	{{if ne .ClientPrivateKeyPath ""}}--privatekeydir='{{.ClientPrivateKeyPath}}' {{end}}
	{{if ne .ExtraArguments ""}}{{.ExtraArguments}} {{end}}
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
