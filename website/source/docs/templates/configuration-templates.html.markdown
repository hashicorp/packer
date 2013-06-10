---
layout: "docs"
---

# Configuration Templates

Certain configuration parameters within templates are themselves a
type of "template." These are not Packer templates, but text templates,
where variables can be used to modify the value of a configuration parameter
during runtime.

For example, the `ami_name` configuration for the [AMI builder](/docs/builders/amazon-ebs.html)
is a template. An example value may be `My Packer AMI {{.CreateTime}}`. At
runtime, this will be turned into `My Packer AMI 1370900368`, where the
"CreateTime" variable was replaced with the Unix timestamp of when the
AMI was actually created.

This sort of templating is pervasive throughout Packer. Instead of documenting
the templating syntax in each location, it is documented once here so
you know how to use it.

<div class="alert alert-info">
<strong>For advanced users:</strong> The templates are actually parsed and executed
using Go's <a href="http://golang.org/pkg/text/template/">text/template</a>
package. It therefore supports the complete template syntax.
</div>

## Syntax

99% of the time all you'll need within configuration templates are variables.
Variables are accessed by using `{{.VARIABLENAME}}`. The "." prefixing the
variable name signals that you're accessing a variable on the root
template object. All template directives go between braces. Here is a piece
of a VMware VMX template that uses variables:

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
