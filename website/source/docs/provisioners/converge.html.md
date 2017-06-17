---
description: |
    The converge Packer provisioner uses Converge modules to provision the
    machine.
layout: docs
page_title: 'Converge - Provisioners'
sidebar_current: 'docs-provisioners-converge'
---

# Converge Provisioner

Type: `converge`

The [Converge](http://converge.aster.is) Packer provisioner uses Converge
modules to provision the machine. It uploads module directories to use as
source, or you can use remote modules.

The provisioner can optionally bootstrap the Converge client/server binary onto
new images.

## Basic Example

The example below is fully functional.

``` json
{
  "type": "converge",
  "module": "https://raw.githubusercontent.com/asteris-llc/converge/master/samples/fileContent.hcl",
  "params": {
    "message": "Hello, Packer!"
  }
}
```

## Configuration Reference

The reference of available configuration options is listed below. The only
required element is "module". Every other option is optional.

-   `module` (string) - Path (or URL) to the root module that Converge will apply.

Optional parameters:

-   `bootstrap` (boolean, defaults to false) - Set to allow the provisioner to
    download the latest Converge bootstrap script and the specified `version` of
    Converge from the internet.

-   `version` (string) - Set to a [released Converge version](https://github.com/asteris-llc/converge/releases) for bootstrap.

-   `module_dirs` (array of directory specifications) - Module directories to
    transfer to the remote host for execution. See below for the specification.

-   `working_directory` (string) - The directory that Converge will change to
    before execution.

-   `params` (maps of string to string) - parameters to pass into the root module.

-   `execute_command` (string) - the command used to execute Converge. This has
    various
    [configuration template variables](/docs/templates/engine.html) available.

-   `prevent_sudo` (bool) - stop Converge from running with adminstrator
    privileges via sudo

-   `bootstrap_command` (string) - the command used to bootstrap Converge. This
    has various
    [configuration template variables](/docs/templates/engine.html) available.

-   `prevent_bootstrap_sudo` (bool) - stop Converge from bootstrapping with
    administrator privileges via sudo

### Module Directories

The provisioner can transfer module directories to the remote host for
provisioning. Of these fields, `source` and `destination` are required in every
directory.

-   `source` (string) - the path to the folder on the local machine.

-   `destination` (string) - the path to the folder on the remote machine. Parent
    directories will not be created; use the shell module to do this.

-   `exclude` (array of string) - files and directories to exclude from transfer.

### Execute Command

By default, Packer uses the following command (broken across multiple lines for readability) to execute Converge:

``` liquid
cd {{.WorkingDirectory}} && \
{{if .Sudo}}sudo {{end}}converge apply \
  --local \
  --log-level=WARNING \
  --paramsJSON '{{.ParamsJSON}}' \
  {{.Module}}
```

This command can be customized using the `execute_command` configuration. As you
can see from the default value above, the value of this configuration can
contain various template variables:

-   `WorkingDirectory` - `directory` from the configuration.
-   `Sudo` - the opposite of `prevent_sudo` from the configuration.
-   `ParamsJSON` - The unquoted JSONified form of `params` from the configuration.
-   `Module` - `module` from the configuration.

### Bootstrap Command

By default, Packer uses the following command to bootstrap Converge:

``` liquid
curl -s https://get.converge.sh | {{if .Sudo}}sudo {{end}}sh {{if ne .Version ""}}-s -- -v {{.Version}}{{end}}
```

This command can be customized using the `bootstrap_command` configuration. As you
can see from the default values above, the value of this configuration can
contain various template variables:

-   `Sudo` - the opposite of `prevent_bootstrap_sudo` from the configuration.
-   `Version` - `version` from the configuration.
