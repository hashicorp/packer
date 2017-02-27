---
description: |
    The `puppet-server` Packer provisioner provisions Packer machines with Puppet by
    connecting to a Puppet master.
layout: docs
page_title: Puppet Server Provisioner
...

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

``` {.javascript}
{
   "type": "puppet-server",
   "options": "--test --pluginsync",
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

-   `facter` (object of key/value strings) - Additional Facter facts to make
    available to the Puppet run.

-   `ignore_exit_codes` (boolean) - If true, Packer will never consider the
    provisioner a failure.

-   `options` (string) - Additional command line options to pass to
    `puppet agent` when Puppet is run.

-   `prevent_sudo` (boolean) - By default, the configured commands that are
    executed to run Puppet are executed with `sudo`. If this is true, then the
    sudo will be omitted.

-   `puppet_node` (string) - The name of the node. If this isn't set, the fully
    qualified domain name will be used.

-   `puppet_server` (string) - Hostname of the Puppet server. By default
    "puppet" will be used.

-   `staging_dir` (string) - This is the directory where all the
    configuration of Puppet by Packer will be placed. By default this
    is /tmp/packer-puppet-server. This directory doesn't need to exist but
    must have proper permissions so that the SSH user that Packer uses is able
    to create directories and write into this folder. If the permissions are not
    correct, use a shell provisioner prior to this to configure it properly.

-   `puppet_bin_dir` (string) - The path to the directory that contains the puppet
    binary for running `puppet agent`. Usually, this would be found via the `$PATH`
    or `%PATH%` environment variable, but some builders (notably, the Docker one) do
    not run profile-setup scripts, therefore the path is usually empty.

-   `execute_command` (string) - This is optional. The command used to execute Puppet. This has
    various [configuration template
    variables](/docs/templates/configuration-templates.html) available. See
    below for more information. By default, Packer uses the following command:

```
{{.FacterVars}} {{if .Sudo}} sudo -E {{end}} \
  {{if ne .PuppetBinDir \"\"}}{{.PuppetBinDir}}/{{end}}puppet agent --onetime --no-daemonize \
  {{if ne .PuppetServer \"\"}}--server='{{.PuppetServer}}' {{end}} \
  {{if ne .Options \"\"}}{{.Options}} {{end}} \
  {{if ne .PuppetNode \"\"}}--certname={{.PuppetNode}} {{end}} \
  {{if ne .ClientCertPath \"\"}}--certdir='{{.ClientCertPath}}' {{end}} \
  {{if ne .ClientPrivateKeyPath \"\"}}--privatekeydir='{{.ClientPrivateKeyPath}}' \
  {{end}} --detailed-exitcodes
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
