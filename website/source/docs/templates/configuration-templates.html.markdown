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

## Name configuration variables<a id="variables"></a>

The variables supported in the naming templates for AMI, snapshots and VM's are:
 
* `.CreateTime` - This will be replaced with the Unix timestamp of when
   the AMI was built.
* `time "TimeZone" "TimeFormat"` - This will be replaced with the time format
as formatted by the [Go Language time formatter](http://golang.org/pkg/time/#pkg-constants).
* `user` - Put in the username of the user running the script.

### Examples
* CreateTime
<pre>
"ami_name": "My Packer AMI {{.CreateTime}}"
</pre>

* ISO 8601 time format
<pre>
  "ami_name": "packer-{{time \"UTC\" \"2006-01-02T15:04:05Z\""}}"
</pre>

* Local time format
<pre>
  "ami_name": "packer-{{time \"Local\" \"2006-01-02T15:04:05Z\""}}"
</pre>

* Username
<pre>
  "ami_name": "packer-{{user}}"
</pre>
