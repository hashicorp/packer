---
layout: "docs"
page_title: "Chef-Client Provisioner"
description: |-
  The Chef Client Packer provisioner installs and configures software on machines built by Packer using chef-client. Packer configures a Chef client to talk to a remote Chef Server to provision the machine.
---

# Chef Client Provisioner

Type: `chef-client`

The Chef Client Packer provisioner installs and configures software on machines built
by Packer using [chef-client](http://docs.opscode.com/chef_client.html).
Packer configures a Chef client to talk to a remote Chef Server to
provision the machine.

The provisioner will even install Chef onto your machine if it isn't already
installed, using the official Chef installers provided by Opscode.

## Basic Example

The example below is fully functional. It will install Chef onto the
remote machine and run Chef client.

```javascript
{
  "type": "chef-client",
  "server_url": "https://mychefserver.com/"
}
```

Note: to properly clean up the Chef node and client the machine on which
packer is running must have knife on the path and configured globally,
i.e, ~/.chef/knife.rb must be present and configured for the target chef server

## Configuration Reference

The reference of available configuration options is listed below. No
configuration is actually required.

* `chef_environment` (string) - The name of the chef_environment sent to the
  Chef server. By default this is empty and will not use an environment.

* `config_template` (string) - Path to a template that will be used for
  the Chef configuration file. By default Packer only sets configuration
  it needs to match the settings set in the provisioner configuration. If
  you need to set configurations that the Packer provisioner doesn't support,
  then you should use a custom configuration template. See the dedicated
  "Chef Configuration" section below for more details.

* `execute_command` (string) - The command used to execute Chef. This has
  various [configuration template variables](/docs/templates/configuration-templates.html)
  available. See below for more information.

* `install_command` (string) - The command used to install Chef. This has
  various [configuration template variables](/docs/templates/configuration-templates.html)
  available. See below for more information.

* `json` (object) - An arbitrary mapping of JSON that will be available as
  node attributes while running Chef.

* `node_name` (string) - The name of the node to register with the Chef
  Server. This is optional and by default is packer-{{uuid}}.

* `prevent_sudo` (boolean) - By default, the configured commands that are
  executed to install and run Chef are executed with `sudo`. If this is true,
  then the sudo will be omitted.

* `run_list` (array of strings) - The [run list](http://docs.opscode.com/essentials_node_object_run_lists.html)
  for Chef. By default this is empty, and will use the run list sent
  down by the Chef Server.

* `server_url` (string) - The URL to the Chef server. This is required.

* `skip_clean_client` (boolean) - If true, Packer won't remove the client
  from the Chef server after it is done running. By default, this is false.

* `skip_clean_node` (boolean) - If true, Packer won't remove the node
  from the Chef server after it is done running. By default, this is false.

* `skip_install` (boolean) - If true, Chef will not automatically be installed
  on the machine using the Opscode omnibus installers.

* `staging_directory` (string) - This is the directory where all the configuration
  of Chef by Packer will be placed. By default this is "/tmp/packer-chef-client".
  This directory doesn't need to exist but must have proper permissions so that
  the SSH user that Packer uses is able to create directories and write into
  this folder. If the permissions are not correct, use a shell provisioner
  prior to this to configure it properly.

* `client_key` (string) - Path to client key. If not set, this defaults to a file
  named client.pem in `staging_directory`.

* `validation_client_name` (string) - Name of the validation client. If
  not set, this won't be set in the configuration and the default that Chef
  uses will be used.

* `validation_key_path` (string) - Path to the validation key for communicating
  with the Chef Server. This will be uploaded to the remote machine. If this
  is NOT set, then it is your responsibility via other means (shell provisioner,
  etc.) to get a validation key to where Chef expects it.

## Chef Configuration

By default, Packer uses a simple Chef configuration file in order to set
the options specified for the provisioner. But Chef is a complex tool that
supports many configuration options. Packer allows you to specify a custom
configuration template if you'd like to set custom configurations.

The default value for the configuration template is:

```liquid
log_level        :info
log_location     STDOUT
chef_server_url  "{{.ServerUrl}}"
validation_client_name "chef-validator"
{{if ne .ValidationKeyPath ""}}
validation_key "{{.ValidationKeyPath}}"
{{end}}
node_name "{{.NodeName}}"
```

This template is a [configuration template](/docs/templates/configuration-templates.html)
and has a set of variables available to use:

* `NodeName` - The node name set in the configuration.
* `ServerUrl` - The URL of the Chef Server set in the configuration.
* `ValidationKeyPath` - Path to the validation key, if it is set.

## Execute Command

By default, Packer uses the following command (broken across multiple lines
for readability) to execute Chef:

```liquid
{{if .Sudo}}sudo {{end}}chef-client \
  --no-color \
  -c {{.ConfigPath}} \
  -j {{.JsonPath}}
```

This command can be customized using the `execute_command` configuration.
As you can see from the default value above, the value of this configuration
can contain various template variables, defined below:

* `ConfigPath` - The path to the Chef configuration file.
  file.
* `JsonPath` - The path to the JSON attributes file for the node.
* `Sudo` - A boolean of whether to `sudo` the command or not, depending on
  the value of the `prevent_sudo` configuration.

## Install Command

By default, Packer uses the following command (broken across multiple lines
for readability) to install Chef. This command can be customized if you want
to install Chef in another way.

```text
curl -L https://www.opscode.com/chef/install.sh | \
  {{if .Sudo}}sudo{{end}} bash
```

This command can be customized using the `install_command` configuration.

## Folder Permissions

!> The `chef-client` provisioner will chmod the directory with your Chef
keys to 777. This is to ensure that Packer can upload and make use of that
directory. However, once the machine is created, you usually don't
want to keep these directories with those permissions. To change the
permissions on the directories, append a shell provisioner after Chef
to modify them.
