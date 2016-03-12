---
description: |
    Within the template, the push section configures how a template can be pushed to
    a remote build service.
layout: docs
page_title: 'Templates: Push'
...

# Templates: Push

Within the template, the push section configures how a template can be
[pushed](/docs/command-line/push.html) to a remote build service.

Push configuration is responsible for defining what files are required to build
this template, what the name of build configuration is in the build service,
etc.

The only build service that Packer can currently push to is
[Atlas](https://atlas.hashicorp.com) by HashiCorp. Support for other build
services will come in the form of plugins in the future.

Within a template, a push configuration section looks like this:

``` {.javascript}
{
  "push": {
    // ... push configuration here
  }
}
```

## Configuration Reference

There are many configuration options available for the builder. They are
segmented below into two categories: required and optional parameters. Within
each category, the available configuration keys are alphabetized.

### Required

-   `name` (string) - Name of the build configuration in the build service. If
    this doesn't exist, it will be created (by default). Note that the name
    cannot contain dots. `[a-zA-Z0-9-_/]+` are safe.

### Optional

-   `address` (string) - The address of the build service to use. By default
    this is `https://atlas.hashicorp.com`.

-   `base_dir` (string) - The base directory of the files to upload. This will
    be the current working directory when the build service executes
    your template. This path is relative to the template.

-   `include` (array of strings) - Glob patterns to include relative to the
    `base_dir`. If this is specified, only files that match the include pattern
    are included.

-   `exclude` (array of strings) - Glob patterns to exclude relative to the
    `base_dir`.

-   `token` (string) - An access token to use to authenticate to the
    build service.

-   `vcs` (boolean) - If true, Packer will detect your VCS (if there is one) and
    only upload the files that are tracked by the VCS. This is useful for
    automatically excluding ignored files. This defaults to false.

## Examples

A push configuration section with minimal options:

``` {.javascript}
{
  "push": {
    "name": "hashicorp/precise64"
  }
}
```

A push configuration specifying Packer to inspect the VCS and list individual
files to include:

``` {.javascript}
{
  "push": {
    "name": "hashicorp/precise64",
    "vcs": true,
    "include": [
      "other_file/outside_of.vcs"
    ]
  }
}
```
