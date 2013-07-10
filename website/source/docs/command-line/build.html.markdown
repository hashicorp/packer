---
layout: "docs"
---

# Command-Line: Build

The `packer build` command takes a template and runs all the builds within
it in order to generate a set of artifacts. The various builds specified within
a template are executed in parallel, unless otherwise specified. And the
artifacts that are created will be outputted at the end of the build.

## Options

* `-debug` - Disables parallelization and enables debug mode. Debug mode flags
  the builders that they should output debugging information. The exact behavior
  of debug mode is left to the builder. In general, builders usually will stop
  between each step, waiting for keyboard input before continuing. This will allow
  the user to inspect state and so on.

* `-except=foo,bar,baz` - Builds all the builds except those with the given
  comma-separated names. Build names by default are the names of their builders,
  unless a specific `name` attribute is specified within the configuration.

* `-only=foo,bar,baz` - Only build the builds with the given comma-separated
  names. Build names by default are the names of their builders, unless a
  specific `name` attribute is specified within the configuration.
