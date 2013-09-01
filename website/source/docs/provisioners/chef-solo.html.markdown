---
layout: "docs"
page_title: "Chef-Solo Provisioner"
---

# Chef Solo Provisioner

Type: `chef-solo`

The Chef solo provisioner installs and configures software on machines built
by Packer using [chef-solo](http://docs.opscode.com/chef_solo.html). Cookbooks
can be uploaded from your local machine to the remote machine or remote paths
can be used.

The provisioner will even install Chef onto your machine if it isn't already
installed, using the official Chef installers provided by Opscode.

## Basic Example

The example below is fully functional and expects cookbooks in the
"cookbooks" directory relative to your working directory.

<pre class="prettyprint">
{
  "type": "chef-solo",
  "cookbook_paths": ["cookbooks"]
}
</pre>

## Configuration Reference

The reference of available configuration options is listed below. No
configuration is actually required, but at least `run_list` is recommended.

* `cookbook_paths` (array of strings) - This is an array of paths to
  "cookbooks" directories on your local filesystem. These will be uploaded
  to the remote machine in the directory specified by the `staging_directory`.
  By default, this is empty.

* `execute_command` (string) - The command used to execute Chef. This has
  various [configuration template variables](/docs/templates/configuration-templates.html)
  available. See below for more information.

* `install_command` (string) - The command used to install Chef. This has
  various [configuration template variables](/docs/templates/configuration-templates.html)
  available. See below for more information.

* `remote_cookbook_paths` (array of string) - A list of paths on the remote
  machine where cookbooks will already exist. These may exist from a previous
  provisioner or step. If specified, Chef will be configured to look for
  cookbooks here. By default, this is empty.

* `json` (object) - An arbitrary mapping of JSON that will be available as
  node attributes while running Chef.

* `prevent_sudo` (boolean) - By default, the configured commands that are
  executed to install and run Chef are executed with `sudo`. If this is true,
  then the sudo will be omitted.

* `run_list` (array of strings) - The [run list](http://docs.opscode.com/essentials_node_object_run_lists.html)
  for Chef. By default this is empty.

* `skip_install` (boolean) - If true, Chef will not automatically be installed
  on the machine using the Opscode omnibus installers.

* `staging_directory` (string) - This is the directory where all the configuration
  of Chef by Packer will be placed. By default this is "/tmp/packer-chef-solo".
  This directory doesn't need to exist but must have proper permissions so that
  the SSH user that Packer uses is able to create directories and write into
  this folder. If the permissions are not correct, use a shell provisioner
  prior to this to configure it properly.

## Execute Command

By default, Packer uses the following command (broken across multiple lines
for readability) to execute Chef:

```
{{if .Sudo}sudo {{end}}chef-solo \
  --no-color \
  -c {{.ConfigPath}} \
  -j {{.JsonPath}}
```

This command can be customized using the `execute_command` configuration.
As you can see from the default value above, the value of this configuration
can contain various template variables, defined below:

* `ConfigPath` - The path to the Chef configuration file.
* `JsonPath` - The path to the JSON attributes file for the node.
* `Sudo` - A boolean of whether to `sudo` the command or not, depending on
  the value of the `prevent_sudo` configuration.

## Install Command

By default, Packer uses the following command (broken across multiple lines
for readability) to install Chef. This command can be customized if you want
to install Chef in another way.

```
curl -L https://www.opscode.com/chef/install.sh | \
  {{if .Sudo}}sudo{{end}} bash
```

This command can be customized using the `install_command` configuration.
