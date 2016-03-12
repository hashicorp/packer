---
description: |
    Packer Commands are the components of Packer that add functionality to the
    `packer` application. Packer comes with a set of commands out of the box, such
    as `build`. Commands are invoked as `packer <COMMAND>`. Custom commands allow
    you to add new commands to Packer to perhaps perform new functionality.
layout: docs
page_title: Custom Command Development
...

# Custom Command Development

Packer Commands are the components of Packer that add functionality to the
`packer` application. Packer comes with a set of commands out of the box, such
as `build`. Commands are invoked as `packer <COMMAND>`. Custom commands allow
you to add new commands to Packer to perhaps perform new functionality.

Prior to reading this page, it is assumed you have read the page on [plugin
development basics](/docs/extend/developing-plugins.html).

Command plugins implement the `packer.Command` interface and are served using
the `plugin.ServeCommand` function. Commands actually have no control over what
keyword invokes the command with the `packer` binary. The keyword to invoke the
command depends on how the plugin is installed and configured in the core Packer
configuration.

\~&gt; **Warning!** This is an advanced topic. If you're new to Packer, we
recommend getting a bit more comfortable before you dive into writing plugins.

## The Interface

The interface that must be implemented for a command is the `packer.Command`
interface. It is reproduced below for easy reference. The actual interface in
the source code contains some basic documentation as well explaining what each
method should do.

``` {.go}
type Command interface {
  Help() string
  Run(env Environment, args []string) int
  Synopsis() string
}
```

### The "Help" Method

The `Help` method returns long-form help. This help is most commonly shown when
a command is invoked with the `--help` or `-h` option. The help should document
all the available command line flags, purpose of the command, etc.

Packer commands generally follow the following format for help, but it is not
required. You're allowed to make the help look like anything you please.

``` {.text}
Usage: packer COMMAND [options] ARGS...

  Brief one or two sentence about the function of the command.

Options:

  -foo=bar                  A description of the flag.
  -another                  Another description.
```

### The "Run" Method

`Run` is what is called when the command is actually invoked. It is given the
`packer.Environment`, which has access to almost all components of the current
Packer run, such as UI, builders, other plugins, etc. In addition to the
environment, the remaining command line args are given. These command line args
have already been stripped of the command name, so they can be passed directly
into something like the standard Go `flag` package for command-line flag
parsing.

The return value of `Run` is the exit status for the command. If everything ran
successfully, this should be 0. If any errors occurred, it should be any
positive integer.

### The "Synopsis" Method

The `Synopsis` method should return a short single-line description of what the
command does. This is used when `packer` is invoked on its own in order to show
a brief summary of the commands that Packer supports.

The synopsis should be no longer than around 50 characters, since it is already
appearing on a line with other text.
