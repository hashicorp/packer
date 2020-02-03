---
layout: "docs"
page_title: "v4 - uuid - Functions - Configuration Language"
sidebar_current: "configuration-functions-uuid-uuidv4"
description: |-
  The uuidv4 function generates a unique id.
---

# `uuidv4` Function


`uuidv4` generates a unique identifier string.

The id is a generated and formatted as required by
[RFC 4122 section 4.4](https://tools.ietf.org/html/rfc4122#section-4.4),
producing a Version 4 UUID. The result is a UUID generated only from
pseudo-random numbers.

This function produces a new value each time it is called, and so using it
directly in resource arguments will result in spurious diffs. We do not
recommend using the `uuidv4` function in resource configurations, but it can
be used with care in conjunction with
[the `ignore_changes` lifecycle meta-argument](../../resources.html#ignore_changes).

In most cases we recommend using [the `random` provider](/docs/providers/random/index.html)
instead, since it allows the one-time generation of random values that are
then retained in the Packer [state](/docs/state/index.html) for use by
future operations. In particular,
[`random_id`](/docs/providers/random/r/id.html) can generate results with
equivalent randomness to the `uuidv4` function.

## Examples

```
> uuidv4()
b5ee72a3-54dd-c4b8-551c-4bdc0204cedb
```

## Related Functions

* [`uuidv5`](./uuidv5.html), which generates name-based UUIDs.
