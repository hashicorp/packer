---
description: |
    User variables allow your templates to be further configured with variables from
    the command-line, environment variables, or files. This lets you parameterize
    your templates so that you can keep secret tokens, environment-specific data,
    and other types of information out of your templates. This maximizes the
    portability and shareability of the template.
layout: docs
page_title: User Variables in Templates
...

# User Variables

User variables allow your templates to be further configured with variables from
the command-line, environment variables, or files. This lets you parameterize
your templates so that you can keep secret tokens, environment-specific data,
and other types of information out of your templates. This maximizes the
portability and shareability of the template.

Using user variables expects you know how [configuration
templates](/docs/templates/configuration-templates.html) work. If you don't know
how configuration templates work yet, please read that page first.

## Usage

User variables must first be defined in a `variables` section within your
template. Even if you want a variable to default to an empty string, it must be
defined. This explicitness makes it easy for newcomers to your template to
understand what can be modified using variables in your template.

The `variables` section is a simple key/value mapping of the variable name to a
default value. A default value can be the empty string. An example is shown
below:

``` {.javascript}
{
  "variables": {
    "aws_access_key": "",
    "aws_secret_key": ""
  },

  "builders": [{
    "type": "amazon-ebs",
    "access_key": "{{user `aws_access_key`}}",
    "secret_key": "{{user `aws_secret_key`}}",
    // ...
  }]
}
```

In the above example, the template defines two variables: `aws_access_key` and
`aws_secret_key`. They default to empty values. Later, the variables are used
within the builder we defined in order to configure the actual keys for the
Amazon builder.

If the default value is `null`, then the user variable will be *required*. This
means that the user must specify a value for this variable or template
validation will fail.

Using the variables is extremely easy. Variables are used by calling the user
function in the form of <code>{{user \`variable\`}}</code>. This function can be
used in *any value* within the template, in builders, provisioners, *anything*.
The user variable is available globally within the template.

## Environment Variables

Environment variables can be used within your template using user variables.
The `env` function is available *only* within the default value of a user
variable, allowing you to default a user variable to an environment variable.
An example is shown below:

``` {.javascript}
{
  "variables": {
    "my_secret": "{{env `MY_SECRET`}}",
  },

  // ...
}
```

This will default "my\_secret" to be the value of the "MY\_SECRET" environment
variable (or the empty string if it does not exist).

-&gt; **Why can't I use environment variables elsewhere?** User variables are
the single source of configurable input to a template. We felt that having
environment variables used *anywhere* in a template would confuse the user
about the possible inputs to a template. By allowing environment variables
only within default values for user variables, user variables remain as the
single source of input to a template that a user can easily discover using
`packer inspect`.

-&gt; **Why can't I use `~` for home variable?** `~` is an special variable
that is evaluated by shell during a variable expansion. As packer doesn't run
inside a shell, it won't expand `~`.

## Setting Variables

Now that we covered how to define and use variables within a template, the next
important point is how to actually set these variables. Packer exposes two
methods for setting variables: from the command line or from a file.

### From the Command Line

To set variables from the command line, the `-var` flag is used as a parameter
to `packer build` (and some other commands). Continuing our example above, we
could build our template using the command below. The command is split across
multiple lines for readability, but can of course be a single line.

``` {.text}
$ packer build \
    -var 'aws_access_key=foo' \
    -var 'aws_secret_key=bar' \
    template.json
```

As you can see, the `-var` flag can be specified multiple times in order to set
multiple variables. Also, variables set later on the command-line override
earlier set variables if it has already been set.

Finally, variables set from the command-line override all other methods of
setting variables. So if you specify a variable in a file (the next method
shown), you can override it using the command-line.

### From a File

Variables can also be set from an external JSON file. The `-var-file` flag reads
a file containing a basic key/value mapping of variables to values and sets
those variables. The JSON file is simple:

``` {.javascript}
{
  "aws_access_key": "foo",
  "aws_secret_key": "bar"
}
```

It is a single JSON object where the keys are variables and the values are the
variable values. Assuming this file is in `variables.json`, we can build our
template using the following command:

``` {.text}
$ packer build -var-file=variables.json template.json
```

The `-var-file` flag can be specified multiple times and variables from multiple
files will be read and applied. As you'd expect, variables read from files
specified later override a variable set earlier if it has already been set.

And as mentioned above, no matter where a `-var-file` is specified, a `-var`
flag on the command line will always override any variables from a file.

# Recipes

## Making a provisioner step conditional on the value of a variable

There is no specific syntax in Packer templates for making a provisioner
step conditional, depending on the value of a variable. However, you may
be able to do this by referencing the variable within a command that
you execute. For example, here is how to make a `shell-local`
provisioner only run if the `do_nexpose_scan` variable is non-empty.

``` {.javascript}
{
  "type": "shell-local",
  "command": "if [ ! -z \"{{user `do_nexpose_scan`}}\" ]; then python -u trigger_nexpose_scan.py; fi"
}
```

## Using HOME Variable

In order to use `$HOME` variable, you can create a `home` variable in packer:

``` {.javascript}
"variables": {
  "home": "{{env `HOME`}}"
}
```

And this will be available to be used in the rest of the template, ie:

``` {.javascript}
{
  "builders": [{
    "type":"google",
    "account_file": "{{ user `home` }}/.secrets/gcp-{{ user `env` }}.json"
  }]
}
```
