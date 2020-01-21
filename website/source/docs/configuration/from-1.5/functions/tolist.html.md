---
layout: "docs"
page_title: "tolist - Functions - Configuration Language"
sidebar_current: "configuration-functions-conversion-tolist"
description: |-
  The tolist function converts a value to a list.
---

# `tolist` Function


`tolist` converts its argument to a list value.

Explicit type conversions are rarely necessary in Packer because it will
convert types automatically where required. Use the explicit type conversion
functions only to normalize types returned in folder outputs.

Pass a _set_ value to `tolist` to convert it to a list. Since set elements are
not ordered, the resulting list will have an undefined order that will be
consistent within a particular run of Packer.

## Examples

```
> tolist(["a", "b", "c"])
[
  "a",
  "b",
  "c",
]
```

Since Packer's concept of a list requires all of the elements to be of the
same type, mixed-typed elements will be converted to the most general type:

```
> tolist(["a", "b", 3])
[
  "a",
  "b",
  "3",
]
```
