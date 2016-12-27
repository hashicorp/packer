---
description: |-
    The Converge Packer provisioner uses Converge modules to provision the machine.
layout: docs
page_title: Converge Provisioner
...

# Converge Provisioner

Type: `converge`

The [Converge](http://converge.aster.is) Packer provisioner uses Converge
modules to provision the machine. It uploads module directories to use as
source, or you can use remote modules.

The provisioner can optionally bootstrap the Converge client/server binary onto
new images.

## Basic Example

The example below is fully functional.

``` {.javascript}
{
  "type": "converge",
  "modules": [
    {
      "module": "https://raw.githubusercontent.com/asteris-llc/converge/master/samples/fileContent.hcl",
      "params": {
        "message": "Hello, Packer!"
      }
    }
  ]
}
```

## Configuration Reference

The reference of available configuration options is listed below. The only
required element is "modules", of which there must be at least one. Every other
option is optional.

- `modules` (array of module specifications) - The root modules to run by
  Converge. See below for the specification.

Optional parameters:

- `bootstrap` (boolean) - Set to allow the provisioner to download the latest
  Converge bootstrap script and the specified `version` of Converge from the
  internet.

- `version` (string) - Set to a [released Converge version](https://github.com/asteris-llc/converge/releases) for bootstrap.

- `module_dirs` (array of directory specifications) - Module directories to
  transfer to the remote host for execution. See below for the specification.

### Modules

Modules control what Converge applies to your system. The `modules` key should
be a list of objects with the following keys. Of these, only `module` is
required.

- `module` (string) - the path (or URL) to the root module.

- `directory` (string) - the directory to run Converge in. If not set, the
  provisioner will use `/tmp` by default.

- `params` (maps of string to string) - parameters to pass into the root module.

### Module Directories

The provisioner can transfer module directories to the remote host for
provisioning. Of these fields, `source` and `destination` are required in every
directory.

- `source` (string) - the path to the folder on the local machine.

- `destination` (string) - the path to the folder on the remote machine. Parent
  directories will not be created; use the shell module to do this.

- `exclude` (array of string) - files and directories to exclude from transfer.
