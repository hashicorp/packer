---
description: |
    Packer Provisioners are the components of Packer that install and configure
    software into a running machine prior to turning that machine into an image. An
    example of a provisioner is the shell provisioner, which runs shell scripts
    within the machines.
layout: docs
page_title: Custom Provisioner Development
...

# Custom Provisioner Development

Packer Provisioners are the components of Packer that install and configure
software into a running machine prior to turning that machine into an image. An
example of a provisioner is the [shell
provisioner](/docs/provisioners/shell.html), which runs shell scripts within the
machines.

Prior to reading this page, it is assumed you have read the page on [plugin
development basics](/docs/extend/developing-plugins.html).

Provisioner plugins implement the `packer.Provisioner` interface and are served
using the `plugin.ServeProvisioner` function.

\~&gt; **Warning!** This is an advanced topic. If you're new to Packer, we
recommend getting a bit more comfortable before you dive into writing plugins.

## The Interface

The interface that must be implemented for a provisioner is the
`packer.Provisioner` interface. It is reproduced below for easy reference. The
actual interface in the source code contains some basic documentation as well
explaining what each method should do.

``` {.go}
type Provisioner interface {
  Prepare(...interface{}) error
  Provision(Ui, Communicator) error
}
```

### The "Prepare" Method

The `Prepare` method for each provisioner is called prior to any runs with the
configuration that was given in the template. This is passed in as an array of
`interface{}` types, but is generally `map[string]interface{}`. The prepare
method is responsible for translating this configuration into an internal
structure, validating it, and returning any errors.

For multiple parameters, they should be merged together into the final
configuration, with later parameters overwriting any previous configuration. The
exact semantics of the merge are left to the builder author.

For decoding the `interface{}` into a meaningful structure, the
[mapstructure](https://github.com/mitchellh/mapstructure) library is
recommended. Mapstructure will take an `interface{}` and decode it into an
arbitrarily complex struct. If there are any errors, it generates very human
friendly errors that can be returned directly from the prepare method.

While it is not actively enforced, **no side effects** should occur from running
the `Prepare` method. Specifically, don't create files, don't launch virtual
machines, etc. Prepare's purpose is solely to configure the builder and validate
the configuration.

The `Prepare` method is called very early in the build process so that errors
may be displayed to the user before anything actually happens.

### The "Provision" Method

The `Provision` method is called when a machine is running and ready to be
provisioned. The provisioner should do its real work here.

The method takes two parameters: a `packer.Ui` and a `packer.Communicator`. The
UI can be used to communicate with the user what is going on. The communicator
is used to communicate with the running machine, and is guaranteed to be
connected at this point.

The provision method should not return until provisioning is complete.

## Using the Communicator

The `packer.Communicator` parameter and interface is used to communicate with
running machine. The machine may be local (in a virtual machine or container of
some sort) or it may be remote (in a cloud). The communicator interface
abstracts this away so that communication is the same overall.

The documentation around the [code
itself](https://github.com/mitchellh/packer/blob/master/packer/communicator.go)
is really great as an overview of how to use the interface. You should begin by
reading this. Once you have read it, you can see some example usage below:

``` {.go}
// Build the remote command.
var cmd packer.RemoteCmd
cmd.Command = "echo foo"

// We care about stdout, so lets collect that into a buffer. Since
// we don't set stderr, that will just be discarded.
var stdout bytes.Buffer
cmd.Stdout = &stdout

// Start the command
if err := comm.Start(&cmd); err != nil {
  panic(err)
}

// Wait for it to complete
cmd.Wait()

// Read the stdout!
fmt.Printf("Command output: %s", stdout.String())
```
