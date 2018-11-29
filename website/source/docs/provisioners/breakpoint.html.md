---
description: |
    The breakpoint provisioner will pause until the user presses "enter" to
    resume the build. This is intended for debugging purposes, and allows you
    to halt at a particular part of the provisioning process.
layout: docs
page_title: 'breakpoint - Provisioners'
sidebar_current: 'docs-provisioners-breakpoint'
---

# File Provisioner

Type: `breakpoint`

The breakpoint provisioner will pause until the user presses "enter" to
resume the build. This is intended for debugging purposes, and allows you
to halt at a particular part of the provisioning process, rather than using the
`-debug` flag, which will instead halt at every step and between every
provisioner.

## Basic Example

``` json
{
  "type": "breakpoint",
  "note": "foo bar baz"
}
```

## Configuration Reference

### Optional

-   `note` (string) - a string to include explaining the purpose or location of
    the breakpoint. For example, you may find it useful to number your
    breakpoints or label them with information about where in the build they
    occur