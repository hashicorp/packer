---
layout: "docs"
page_title: "Shell Provisioner"
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

The reference of available configuration options is listed below. The only
required element is either "inline" or "script". Every other option is optional.

Exactly _one_ of the following is required:

* `inline` (array of strings) - This is an array of commands to execute.
  The commands are concatenated by newlines and turned into a single file,
  so they are all executed within the same context. This allows you to
  change directories in one command and use something in the directory in
  the next and so on. Inline scripts are the easiest way to pull of simple
  tasks within the machine.

* `script` (string) - The path to a script to upload and execute in the machine.
  This path can be absolute or relative. If it is relative, it is relative
  to the working directory when Packer is executed.

* `scripts` (array of strings) - An array of scripts to execute. The scripts
  will be uploaded and executed in the order specified. Each script is executed
  in isolation, so state such as variables from one script won't carry on to
  the next.

Optional parameters:

* `binary` (boolean) - If true, specifies that the script(s) are binary
   files, and Packer should therefore not convert Windows line endings to
   Unix line endings (if there are any). By default this is false.

* `environment_vars` (array of strings) - An array of key/value pairs
  to inject prior to the execute_command. The format should be
  `key=value`. Packer injects some environmental variables by default
  into the environment, as well, which are covered in the section below.

* `execute_command` (string) - The command to use to execute the script.
  By default this is `chmod +x {{ .Path }}; {{ .Vars }} {{ .Path }}`. The value of this is
  treated as [configuration template](/docs/templates/configuration-templates.html). There are two available variables: `Path`, which is
  the path to the script to run, and `Vars`, which is the list of
  `environment_vars`, if configured.

* `inline_shebang` (string) - The
  [shebang](http://en.wikipedia.org/wiki/Shebang_%28Unix%29) value to use when
  running commands specified by `inline`. By default, this is `/bin/sh`.
  If you're not using `inline`, then this configuration has no effect.

* `remote_path` (string) - The path where the script will be uploaded to
  in the machine. This defaults to "/tmp/script.sh". This value must be
  a writable location and any parent directories must already exist.

* `start_retry_timeout` (string) - The amount of time to attempt to
  _start_ the remote process. By default this is "5m" or 5 minutes. This
  setting exists in order to deal with times when SSH may restart, such as
  a system reboot. Set this to a higher value if reboots take a longer
  amount of time.

## Execute Command Example

To many new users, the `execute_command` is puzzling. However, it provides
an important function: customization of how the command is executed. The
most common use case for this is dealing with **sudo password prompts**.

For example, if the default user of an installed operating system is "packer"
and has the password "packer" for sudo usage, then you'll likely want to
change `execute_command` to be:

```
"echo 'packer' | {{ .Vars }} sudo -E -S sh '{{ .Path }}'"
```

The `-S` flag tells `sudo` to read the password from stdin, which in this
case is being piped in with the value of "packer". The `-E` flag tells `sudo`
to preserve the environment, allowing our environmental variables to work
within the script.

By setting the `execute_command` to this, your script(s) can run with
root privileges without worrying about password prompts.

## Default Environmental Variables

In addition to being able to specify custom environmental variables using
the `environment_vars` configuration, the provisioner automatically
defines certain commonly useful environmental variables:

* `PACKER_BUILD_NAME` is set to the name of the build that Packer is running.
  This is most useful when Packer is making multiple builds and you want to
  distinguish them slightly from a common provisioning script.

* `PACKER_BUILDER_TYPE` is the type of the builder that was used to create
  the machine that the script is running on. This is useful if you want to
  run only certain parts of the script on systems built with certain builders.

## Handling Reboots

Provisioning sometimes involves restarts, usually when updating the operating
system. Packer is able to tolerate restarts via the shell provisioner.

Packer handles this by retrying to start scripts for a period of time
before failing. This allows time for the machine to start up and be ready
to run scripts. The amount of time the provisioner will wait is configured
using `start_retry_timeout`, which defaults to a few minutes.

Sometimes, when executing a command like `reboot`, the shell script will
return and Packer will start executing the next one before SSH actually
quits and the machine restarts. For this, put a long `sleep` after the
reboot so that SSH will eventually be killed automatically:

```
reboot
sleep 60
```

Some OS configurations don't properly kill all network connections on
reboot, causing the provisioner to hang despite a reboot occuring.
In this case, make sure you shut down the network interfaces
on reboot or in your shell script. For example, on Gentoo:

```
/etc/init.d/net.eth0 stop
```

## Troubleshooting

*My shell script doesn't work correctly on Ubuntu*

* On Ubuntu the /bin/sh shell is
[dash](http://en.wikipedia.org/wiki/Debian_Almquist_shell). If your script has
[bash](http://en.wikipedia.org/wiki/Bash_(Unix_shell\)) specific commands in it
then put `#!/bin/bash` at the top of your script. Differences
between dash and bash can be found on the [DashAsBinSh](https://wiki.ubuntu.com/DashAsBinSh) Ubuntu wiki page.

*My shell works when I login but fails with the shell provisioner*

* See the above tip. More than likely your login shell is using /bin/bash
while the provisioner is using /bin/sh.

*My installs hang when using `apt-get` or `yum`*

* Make sure you add a "-y" to the command to prevent it from requiring
user input before proceeding.

*How do I tell what my shell script is doing?*

* Adding a `-x` flag to the shebang at the top of the script (`#!/bin/sh -x`)
will echo the script statements as it is executing.

*My builds don't always work the same*

* Some distributions start the SSH daemon before other core services which
can create race conditions. Your first provisoner can tell the machine to
wait until it completely boots.

<pre class="prettyprint">
    {
       "type": "shell",
       "inline": [ "sleep 10" ]
    }
</pre>
