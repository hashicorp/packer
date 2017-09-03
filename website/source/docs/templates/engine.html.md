---
description: |
    All strings within templates are processed by a common Packer templating
    engine, where variables and functions can be used to modify the value of a
    configuration parameter at runtime.
layout: docs
page_title: 'Template Engine - Templates'
sidebar_current: 'docs-templates-engine'
---

# Template Engine

All strings within templates are processed by a common Packer templating engine,
where variables and functions can be used to modify the value of a
configuration parameter at runtime.

The syntax of templates uses the following conventions:

-   Anything template related happens within double-braces: `{{ }}`.
-   Functions are specified directly within the braces, such as `{{timestamp}}`.
-   Template variables are prefixed with a period and capitalized, such as
    `{{.Variable}}`.

## Functions

Functions perform operations on and within strings, for example the `{{timestamp}}` function can be used in any string to generate
the current timestamp. This is useful for configurations that require unique
keys, such as AMI names. By setting the AMI name to something like `My Packer AMI {{timestamp}}`, the AMI name will be unique down to the second. If you
need greater than one second granularity, you should use `{{uuid}}`, for
example when you have multiple builders in the same template.

Here is a full list of the available functions for reference.

-   `build_name` - The name of the build being run.
-   `build_type` - The type of the builder being used currently.
-   `isotime [FORMAT]` - UTC time, which can be
    [formatted](https://golang.org/pkg/time/#example_Time_Format). See more
    examples below in [the `isotime` format reference](/docs/templates/engine.html#isotime-function-format-reference).
-   `lower` - Lowercases the string.
-   `pwd` - The working directory while executing Packer.
-   `template_dir` - The directory to the template for the build.
-   `timestamp` - The current Unix timestamp in UTC.
-   `uuid` - Returns a random UUID.
-   `upper` - Uppercases the string.
-   `user` - Specifies a user variable.

#### Specific to Amazon builders:

-   `clean_ami_name` - AMI names can only contain certain characters. This
    function will replace illegal characters with a '-" character. Example usage
    since ":" is not a legal AMI name is: `{{isotime | clean_ami_name}}`.

## Template variables

Template variables are special variables automatically set by Packer at build time. Some builders, provisioners and other components have template variables that are available only for that component. Template variables are recognizable because they're prefixed by a period, such as `{{ .Name }}`. For example, when using the [`shell`](/docs/builders/vmware-iso.html) builder template variables are available to customize the [`execute_command`](/docs/provisioners/shell.html#execute_command) parameter used to determine how Packer will run the shell command.

``` liquid
{
    "provisioners": [
        {
            "type": "shell",
            "execute_command": "{{.Vars}} sudo -E -S bash '{{.Path}}'",
            "scripts": [
                "scripts/bootstrap.sh"
            ]
        }
    ]
}
```

The `{{ .Vars }}` and `{{ .Path }}` template variables will be replaced with the list of the environment variables and the path to the script to be executed respectively.

-&gt; **Note:** In addition to template variables, you can specify your own user variables. See the [user variable](/docs/templates/user-variables.html) documentation for more information on user variables.

# isotime Function Format Reference

Formatting for the function `isotime` uses the magic reference date **Mon Jan 2
15:04:05 -0700 MST 2006**, which breaks down to the following:

<table class="table table-bordered table-condensed">
  <thead>
    <tr>
      <th>
      </th>
      <th align="center">
        Day of Week
      </th>
      <th align="center">
        Month
      </th>
      <th align="center">
        Date
      </th>
      <th align="center">
        Hour
      </th>
      <th align="center">
        Minute
      </th>
      <th align="center">
        Second
      </th>
      <th align="center">
        Year
      </th>
      <th align="center">
        Timezone
      </th>
    </tr>
  </thead>
  <tr>
    <th>
      Numeric
    </th>
    <td align="center">
      -   
    </td>
    <td align="center">
      01
    </td>
    <td align="center">
      02
    </td>
    <td align="center">
      03 (15)
    </td>
    <td align="center">
      04
    </td>
    <td align="center">
      05
    </td>
    <td align="center">
      06
    </td>
    <td align="center">
      -0700
    </td>
  </tr>
  <tr>
    <th>
      Textual
    </th>
    <td align="center">
      Monday (Mon)
    </td>
    <td align="center">
      January (Jan)
    </td>
    <td align="center">
      -   
    </td>
    <td align="center">
      -   
    </td>
    <td align="center">
      -   
    </td>
    <td align="center">
      -   
    </td>
    <td align="center">
      -   
    </td>
    <td align="center">
      MST
    </td>
  </tr>
</table>
*The values in parentheses are the abbreviated, or 24-hour clock values*

Note that "-0700" is always formatted into "+0000" because `isotime` is always UTC time.

Here are some example formatted time, using the above format options:

``` liquid
isotime = June 7, 7:22:43pm 2014

{{isotime "2006-01-02"}} = 2014-06-07
{{isotime "Mon 1504"}} = Sat 1922
{{isotime "02-Jan-06 03\_04\_05"}} = 07-Jun-2014 07\_22\_43
{{isotime "Hour15Year200603"}} = Hour19Year201407
```

Please note that double quote characters need escaping inside of templates (in this case, on the `ami_name` value):

``` json
{
  "builders": [
    {
      "type": "amazon-ebs",
      "access_key": "...",
      "secret_key": "...",
      "region": "us-east-1",
      "source_ami": "ami-fce3c696",
      "instance_type": "t2.micro",
      "ssh_username": "ubuntu",
      "ami_name": "packer {{isotime \"2006-01-02\"}}"
    }
  ]
}
```

-&gt; **Note:** See the [Amazon builder](/docs/builders/amazon.html) documentation for more information on how to correctly configure the Amazon builder in this example.
