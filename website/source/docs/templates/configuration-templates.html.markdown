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
* <code>isotime &#96;&lt;format&gt;&#96;</code> - UTC time in [&lt;format&gt;](http://golang.org/pkg/time/#example_Time_Format) format.
* `timestamp` - The current Unix timestamp in UTC.
* `uuid` - Returns a random UUID.

### isotime Format
Formatting for the function <code>isotime &#96;&lt;format&gt;&#96;</code> uses the magic reference date **Mon Jan 2 15:04:05 -0700 MST 2006**, which breaks down to the following:

 <table border="1" cellpadding="5" width="100%">
 	<tr bgcolor="lightgray">
 		<td></td>
 		<td align="center"><strong>Day of Week</strong></td>
 		<td align="center"><strong>Month</strong></td>
 		<td align="center"><strong>Date</strong></td>
 		<td align="center"><strong>Hour</strong></td>
 		<td align="center"><strong>Minute</strong></td>
 		<td align="center"><strong>Second</strong></td>
 		<td align="center"><strong>Year</strong></td>
 		<td align="center"><strong>Timezone</strong></td>
 	</tr>
 	<tr>
 		<td><strong>Numeric</strong></td>
 		<td align="center">-</td>
 		<td align="center">01</td>
 		<td align="center">02</td>
 		<td align="center">03 (15)</td>
 		<td align="center">04</td>
 		<td align="center">05</td>
 		<td align="center">06</td>
 		<td align="center">-0700</td>
 	</tr>
 	<tr>
 		<td><strong>Textual</strong></td>
 		<td align="center">Monday (Mon)</td>
 		<td align="center">January (Jan)</td>
 		<td align="center">-</td>
 		<td align="center">-</td>
 		<td align="center">-</td>
 		<td align="center">-</td>
 		<td align="center">-</td>
 		<td align="center">MST</td>
 	</tr>
 </table>
 _The values in parentheses are the abbreviated, or 24-hour clock values_
 
 Here are some example formated time, using the above format options:
 
 <pre>
isotime = June 7, 7:22:43pm 2014

{{isotime "2006-01-02"}} = 2014-06-07
{{isotime "Mon 1506"}} = Sat 1914
{{isotime "01-Jan-06 03\_04\_05"}} = 07-Jun-2014 07\_22\_43
{{isotime "Hour15Year200603"}} = Hour19Year201407
</pre>


## Amazon Specific Functions

Specific to Amazon builders:

* ``clean_ami_name`` - AMI names can only contain certain characters. This
  function will replace illegal characters with a '-" character. Example usage
  since ":" is not a legal AMI name is: `{{isotime | clean_ami_name}}`.
