---
layout: "docs"
page_title: "Configuration Templates"
---

# Configuration Templates

All strings within templates are processed by a common Packer templating
engine, where variables and functions can be used to modify the value of
a configuration parameter at runtime.

For example, the `{{timestamp}}` function can be used in any string to
generate the current timestamp. This is useful for configurations that require
unique keys, such as AMI names. By setting the AMI name to something like
`My Packer AMI {{timestamp}}`, the AMI name will be unique down to the second.

In addition to globally available functions like timestamp shown before,
some configurations have special local variables that are available only
for that configuration. These are recognizable because they're prefixed by
a period, such as `{{.Name}}`.

The complete syntax is covered in the next section, followed by a reference
of globally available functions.

## Syntax

The syntax of templates is extremely simple. Anything template related
happens within double-braces: `{{ }}`. Variables are prefixed with a period
and capitalized, such as `{{.Variable}}` and functions are just directly
within the braces, such as `{{timestamp}}`.

Here is an example from the VMware VMX template that shows configuration
templates in action:

<pre>
.encoding = "UTF-8"
displayName = "{{ .Name }}"
guestOS = "{{ .GuestOS }}"
</pre>

In this case, the "Name" and "GuestOS" variables will be replaced, potentially
resulting in a VMX that looks like this:

<pre>
.encoding = "UTF-8"
displayName = "packer"
guestOS = "otherlinux"
</pre>

## Global Functions

While some configuration settings have local variables specific to only that
configuration, a set of functions are available globally for use in _any string_
in Packer templates. These are listed below for reference.

* `pwd` - The working directory while executing Packer.
* `isotime` - UTC time in RFC-3339 format.
* `timestamp` - The current Unix timestamp in UTC.
* `uuid` - Returns a random UUID.

## Amazon Specific Functions

Specific to Amazon builders:

* ``clean_ami_name`` - AMI names can only contain certain characters. This
  function will replace illegal characters with a '-" character. Example usage
  since ":" is not a legal AMI name is: `{{isotime | clean_ami_name}}`.
