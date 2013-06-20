---
layout: "docs"
---

# Shell Provisioner

Type: `shell`

The shell provisioner provisions machines built by Packer using shell scripts.
Shell provisioning is the easiest way to get software installed and configured
on a machine.

## Basic Example

The example below is fully functional.

<pre class="prettyprint">
{
  "type": "shell",
  "inline": ["echo foo"]
}
</pre>

## Configuration Reference

The reference of available configuratin options is listed below. The only
required element is either "inline" or "path". Every other option is optional.

Exactly _one_ of the following is required:

* `inline` (array of strings) - This is an array of commands to execute.
  The commands are concatenated by newlines and turned into a single file,
  so they are all executed within the same context. This allows you to
  change directories in one command and use something in the directory in
  the next and so on. Inline scripts are the easiest way to pull of simple
  tasks within the machine.

* `path` (string) - The path to a script to upload and execute in the machine.
  This path can be absolute or relative. If it is relative, it is relative
  to the working directory when Packer is executed.

* `scripts` (array of strings) - An array of scripts to execute. The scripts
  will be uploaded and executed in the order specified. Each script is executed
  in isolation, so state such as variables from one script won't carry on to
  the next.

Optional parameters:

* `execute_command` (string) - The command to use to execute the script.
  By default this is `sh {{ .Path }}`. The value of this is treated as a
  [configuration template](/docs/templates/configuration-templates.html).
  The only available variable in it is `Path` which is the path to the
  script to run.

* `remote_path` (string) - The path where the script will be uploaded to
  in the machine. This defaults to "/tmp/script.sh". This value must be
  a writable location and any parent directories must already exist.
