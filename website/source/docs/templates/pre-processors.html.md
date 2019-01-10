---
description: |
    The pre-processor section within a template configures any pre-processing that
    will be done to images built by the builders. environment. Example usages of
    pre-processors are setting up infrastructure or downloading a iso image and
    uploading it to a cloud provider for use in a builder.
layout: docs
page_title: 'Pre-Processors - Templates'
sidebar_current: 'docs-templates-pre-processors'
---

# Template Pre-Processors

The pre-processor section within a template configures any pre-processing that
will be done to the environment used by the builders. Examples of
pre-processing would be setting up infrastructure or down and uploading images.

Pre-processors are *optional*. If no pre-processors are defined within a
template, then no pre-processing will be done.

This documentation page will cover how to configure a pre-processor in a
template. The specific configuration options available for each pre-processor,
however, must be referenced from the documentation for that specific
pre-processor.

Within a template, a section of pre-processor definitions looks like this:

``` json
{
  "pre-processors": [
    // ... one or more pre-processor definitions here
  ]
}
```

For each pre-processor definition, Packer will run the pre-processors for each
image to build. This means that if you have one pre-processor defined and two
builders defined in a template, the pre-processor will run twice (once for each
builder), by default. There are ways, which will be covered later, to control
what builders pre-processors apply to, if you wish.

## Pre-Processor Definition

Within the `pre-processors` array in a template, there are three ways to define
a pre-processor. There are *simple* definitions, *detailed* definitions, and
*sequence* definitions. Another way to think about this is that the "simple"
and "detailed" definitions are shortcuts for the "sequence" definition.

A **simple definition** is just a string; the name of the pre-processor. An
example is shown below. Simple definitions are used when no additional
configuration is needed for the pre-processor.

``` json
{
  "pre-processors": ["..."]
}
```

A **detailed definition** is a JSON object. It is very similar to a builder or
provisioner definition. It contains a `type` field to denote the type of the
pre-processor, but may also contain additional configuration for the
pre-processor. A detailed definition is used when additional configuration is
needed beyond simply the type for the pre-processor. An example is shown below.

``` json
{
  "pre-processors": [
    {
      "type": "shell-local",
      "command": "echo foo"
    }
  ]
}
```

A **sequence definition** is a JSON array comprised of other **simple** or
**detailed** definitions. The pre-processors defined in the array are run in
order A sequence definition may not contain another sequence definition.
Sequence definitions are used to chain together multiple pre-processors.

As you may be able to imagine, the **simple** and **detailed** definitions are
simply shortcuts for a **sequence** definition of only one element.

## Run on Specific Builds

You can use the `only` or `except` configurations to run a pre-processor only
with specific builds. These two configurations do what you expect: `only` will
only run the pre-processor on the specified builds and `except` will run the
pre-processor on anything other than the specified builds.

An example of `only` being used is shown below, but the usage of `except` is
effectively the same. `only` and `except` can only be specified on "detailed"
configurations. If you have a sequence of pre-processors to run, `only` and
`except` will only affect that single pre-processor in the sequence.

``` json
{
  "type": "vagrant",
  "only": ["virtualbox-iso"]
}
```

The values within `only` or `except` are *build names*, not builder types. If
you recall, build names by default are just their builder type, but if you
specify a custom `name` parameter, then you should use that as the value
instead of the type.
