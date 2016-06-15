---
description: |
    The Chef Client Packer provisioner installs and configures software on machines
    built by Packer using chef-client. Packer configures a Chef client to talk to a
    remote Chef Server to provision the machine.
layout: docs
page_title: 'Chef-Client Provisioner'
...

# Chef Client Provisioner

Type: `chef-client`

The Chef Client Packer provisioner installs and configures software on machines
built by Packer using [chef-client](https://docs.chef.io/chef_client.html).
Packer configures a Chef client to talk to a remote Chef Server to provision the
machine.

The provisioner will even install Chef onto your machine if it isn't already
installed, using the official Chef installers provided by Chef.

## Basic Example

The example below is fully functional. It will install Chef onto the remote
machine and run Chef client.

``` {.javascript}
{
  "type": "chef-client",
  "server_url": "https://mychefserver.com/"
}
```

Note: to properly clean up the Chef node and client the machine on which packer
is running must have knife on the path and configured globally, i.e,
\~/.chef/knife.rb must be present and configured for the target chef server

## Configuration Reference

The reference of available configuration options is listed below. No
configuration is actually required.

-   `chef_environment` (string) - The name of the chef\_environment sent to the
    Chef server. By default this is empty and will not use an environment.

-   `config_template` (string) - Path to a template that will be used for the
    Chef configuration file. By default Packer only sets configuration it needs
    to match the settings set in the provisioner configuration. If you need to
    set configurations that the Packer provisioner doesn't support, then you
    should use a custom configuration template. See the dedicated "Chef
    Configuration" section below for more details.

-   `encrypted_data_bag_secret_path` (string) - The path to the file containing
    the secret for encrypted data bags. By default, this is empty, so no
    secret will be available.

-   `execute_command` (string) - The command used to execute Chef. This has
    various [configuration template
    variables](/docs/templates/configuration-templates.html) available. See
    below for more information.

-   `guest_os_type` (string) - The target guest OS type, either "unix" or
    "windows". Setting this to "windows" will cause the provisioner to use
     Windows friendly paths and commands. By default, this is "unix".

-   `install_command` (string) - The command used to install Chef. This has
    various [configuration template
    variables](/docs/templates/configuration-templates.html) available. See
    below for more information.

-   `json` (object) - An arbitrary mapping of JSON that will be available as
    node attributes while running Chef.

-   `knife_command` (string) - The command used to run Knife during node clean-up. This has
    various [configuration template
    variables](/docs/templates/configuration-templates.html) available. See
    below for more information.

-   `node_name` (string) - The name of the node to register with the
    Chef Server. This is optional and by default is packer-{{uuid}}.

-   `prevent_sudo` (boolean) - By default, the configured commands that are
    executed to install and run Chef are executed with `sudo`. If this is true,
    then the sudo will be omitted. This has no effect when guest_os_type is
    windows.

-   `run_list` (array of strings) - The [run
    list](http://docs.chef.io/essentials_node_object_run_lists.html) for Chef.
    By default this is empty, and will use the run list sent down by the
    Chef Server.

-   `server_url` (string) - The URL to the Chef server. This is required.

-   `skip_clean_client` (boolean) - If true, Packer won't remove the client from
    the Chef server after it is done running. By default, this is false.

-   `skip_clean_node` (boolean) - If true, Packer won't remove the node from the
    Chef server after it is done running. By default, this is false.

-   `skip_install` (boolean) - If true, Chef will not automatically be installed
    on the machine using the Chef omnibus installers.

-   `ssl_verify_mode` (string) - Set to "verify\_none" to skip validation of
    SSL certificates. If not set, this defaults to "verify\_peer" which validates
    all SSL certifications.

-   `staging_directory` (string) - This is the directory where all the
    configuration of Chef by Packer will be placed. By default this is
    "/tmp/packer-chef-client" when guest_os_type unix and
    "$env:TEMP/packer-chef-client" when windows. This directory doesn't need to
    exist but must have proper permissions so that the user that Packer uses is
    able to create directories and write into this folder. By default the
    provisioner will create and chmod 0777 this directory.

-   `client_key` (string) - Path to client key. If not set, this defaults to a
    file named client.pem in `staging_directory`.

-   `validation_client_name` (string) - Name of the validation client. If not
    set, this won't be set in the configuration and the default that Chef uses
    will be used.

-   `validation_key_path` (string) - Path to the validation key for
    communicating with the Chef Server. This will be uploaded to the
    remote machine. If this is NOT set, then it is your responsibility via other
    means (shell provisioner, etc.) to get a validation key to where Chef
    expects it.

## Chef Configuration

By default, Packer uses a simple Chef configuration file in order to set the
options specified for the provisioner. But Chef is a complex tool that supports
many configuration options. Packer allows you to specify a custom configuration
template if you'd like to set custom configurations.

The default value for the configuration template is:

``` {.liquid}
log_level        :info
log_location     STDOUT
chef_server_url  "{{.ServerUrl}}"
client_key       "{{.ClientKey}}"
{{if ne .EncryptedDataBagSecretPath ""}}
encrypted_data_bag_secret "{{.EncryptedDataBagSecretPath}}"
{{end}}
{{if ne .ValidationClientName ""}}
validation_client_name "{{.ValidationClientName}}"
{{else}}
validation_client_name "chef-validator"
{{end}}
{{if ne .ValidationKeyPath ""}}
validation_key "{{.ValidationKeyPath}}"
{{end}}
node_name "{{.NodeName}}"
{{if ne .ChefEnvironment ""}}
environment "{{.ChefEnvironment}}"
{{end}}
{{if ne .SslVerifyMode ""}}
ssl_verify_mode :{{.SslVerifyMode}}
{{end}}
```

This template is a [configuration
template](/docs/templates/configuration-templates.html) and has a set of
variables available to use:

-   `ChefEnvironment` - The Chef environment name.
-   `EncryptedDataBagSecretPath` - The path to the secret key file to decrypt
     encrypted data bags.
-   `NodeName` - The node name set in the configuration.
-   `ServerUrl` - The URL of the Chef Server set in the configuration.
-   `SslVerifyMode` - Whether Chef SSL verify mode is on or off.
-   `ValidationClientName` - The name of the client used for validation.
-   `ValidationKeyPath` - Path to the validation key, if it is set.

## Execute Command

By default, Packer uses the following command (broken across multiple lines for
readability) to execute Chef:

``` {.liquid}
{{if .Sudo}}sudo {{end}}chef-client \
  --no-color \
  -c {{.ConfigPath}} \
  -j {{.JsonPath}}
```

When guest_os_type is set to "windows", Packer uses the following command to
execute Chef. The full path to Chef is required because the PATH environment
variable changes don't immediately propogate to running processes.

``` {.liquid}
c:/opscode/chef/bin/chef-client.bat \
  --no-color \
  -c {{.ConfigPath}} \
  -j {{.JsonPath}}
```

This command can be customized using the `execute_command` configuration. As you
can see from the default value above, the value of this configuration can
contain various template variables, defined below:

-   `ConfigPath` - The path to the Chef configuration file.
-   `JsonPath` - The path to the JSON attributes file for the node.
-   `Sudo` - A boolean of whether to `sudo` the command or not, depending on the
    value of the `prevent_sudo` configuration.

## Install Command

By default, Packer uses the following command (broken across multiple lines for
readability) to install Chef. This command can be customized if you want to
install Chef in another way.

``` {.text}
curl -L https://www.chef.io/chef/install.sh | \
  {{if .Sudo}}sudo{{end}} bash
```

When guest_os_type is set to "windows", Packer uses the following command to
install the latest version of Chef:

``` {.text}
powershell.exe -Command "(New-Object System.Net.WebClient).DownloadFile('http://chef.io/chef/install.msi', 'C:\\Windows\\Temp\\chef.msi');Start-Process 'msiexec' -ArgumentList '/qb /i C:\\Windows\\Temp\\chef.msi' -NoNewWindow -Wait"
```

This command can be customized using the `install_command` configuration.

## Knife Command

By default, Packer uses the following command (broken across multiple lines for
readability) to execute Chef:

``` {.liquid}
{{if .Sudo}}sudo {{end}}knife \
  {{.Args}} \
  {{.Flags}}
```

When guest_os_type is set to "windows", Packer uses the following command to
execute Chef. The full path to Chef is required because the PATH environment
variable changes don't immediately propogate to running processes.

``` {.liquid}
c:/opscode/chef/bin/knife.bat \
  {{.Args}} \
  {{.Flags}}
```

This command can be customized using the `knife_command` configuration. As you
can see from the default value above, the value of this configuration can
contain various template variables, defined below:

-   `Args` - The command arguments that are getting passed to the Knife command.
-   `Flags` - The command flags that are getting passed to the Knife command..
-   `Sudo` - A boolean of whether to `sudo` the command or not, depending on the
    value of the `prevent_sudo` configuration.


## Folder Permissions

!&gt; The `chef-client` provisioner will chmod the directory with your Chef keys
to 777. This is to ensure that Packer can upload and make use of that directory.
However, once the machine is created, you usually don't want to keep these
directories with those permissions. To change the permissions on the
directories, append a shell provisioner after Chef to modify them.

## Examples

### Chef Client Local Mode

The following example shows how to run the `chef-cilent` provisioner in local
mode, while passing a `run_list` using a variable.

**Local environment variables**

    # Machines Chef directory
    export PACKER_CHEF_DIR=/var/chef-packer
    # Comma separated run_list
    export PACKER_CHEF_RUN_LIST="recipe[apt],recipe[nginx]"
    ...

**Packer variables**

Set the necessary Packer variables using environment variables or provide a [var
file](/docs/templates/user-variables.html).

``` {.liquid}
"variables": {
  "chef_dir": "{{env `PACKER_CHEF_DIR`}}",
  "chef_run_list": "{{env `PACKER_CHEF_RUN_LIST`}}",
  "chef_client_config_tpl": "{{env `PACKER_CHEF_CLIENT_CONFIG_TPL`}}",
  "packer_chef_bootstrap_dir": "{{env `PACKER_CHEF_BOOTSTRAP_DIR`}}" ,
  "packer_uid": "{{env `PACKER_UID`}}",
  "packer_gid": "{{env `PACKER_GID`}}"
}
```

**Setup the** `chef-client` **provisioner**

Make sure we have the correct directories and permissions for the `chef-client`
provisioner. You will need to bootstrap the Chef run by providing the necessary
cookbooks using Berkshelf or some other means.

``` {.liquid}
{
  "type": "file",
  "source": "{{user `packer_chef_bootstrap_dir`}}",
  "destination": "/tmp/bootstrap"
},
{
  "type": "shell",
  "inline": [
    "sudo mkdir -p {{user `chef_dir`}}",
    "sudo mkdir -p /tmp/packer-chef-client",
    "sudo chown {{user `packer_uid`}}.{{user `packer_gid`}} /tmp/packer-chef-client",
    "sudo sh /tmp/bootstrap/bootstrap.sh"
  ]
},
{
  "type": "chef-client",
  "server_url": "http://localhost:8889",
  "config_template": "{{user `chef_client_config_tpl`}}/client.rb.tpl",
  "skip_clean_node": true,
  "skip_clean_client": true,
  "run_list": "{{user `chef_run_list`}}"
}
```
