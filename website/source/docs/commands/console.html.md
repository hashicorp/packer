---
description: |
    The `packer console` command allows you to experiment with Packer variable
    interpolations.
layout: docs
page_title: 'packer console - Commands'
sidebar_current: 'docs-commands-console'
---

# `console` Command

The `packer console` command allows you to experiment with Packer variable
interpolations. You may access variables in the Packer config you called the
console with, or provide variables when you call console using the -var or
-var-file command line options.

~> **Note:** `console` is available from version 1.4.2 and above.

Type in the interpolation to test and hit \<enter\> to see the result.

To exit the console, type "exit" and hit \<enter\>, or use Control-C.

``` shell
$ packer console my_template.json
```

The full list of options that the console command will accept is visible in the
help output, which can be seen via `packer console -h`.

## Options

-   `-var` - Set a variable in your packer template. This option can be used
    multiple times. This is useful for setting version numbers for your build.
    example: `-var "myvar=asdf"`

-   `-var-file` - Set template variables from a file.
	example: `-var-file myvars.json`

## REPL commands
-   `help` - displays help text for Packer console.

-   `exit` - exits the console

-   `variables` - prints a list of all variables read into the console from the
    `-var` option, `-var-files` option, and template.

## Usage Examples

Let's say you launch a console using a Packer template `example_template.json`:

```
$ packer console example_template.json
```

You'll be dropped into a prompt that allows you to enter template functions and
see how they're evaluated; for example, if the variable `myvar` is defined in
your example_template's variable section:

```
"variables":{
	"myvar": "asdfasdf"
},
...
```
and you enter `{{user `myvar`}}` in the Packer console, you'll see the value of
myvar:

```
> {{user `myvar`}}
> asdfasdf
```

From there you can test more complicated interpolations:

```
> {{user `myvar`}}-{{timestamp}}
> asdfasdf-1559854396
```

And when you're done using the console, just type "exit" or CTRL-C

```
> exit
$
```

If you'd like to provide a variable or variable files, you'd do this:

```
packer console -var "myvar=fdsafdsa" -var-file myvars.json example_template.json
```

If you don't have specific variables or var files you want to test, and just
want to experiment with a particular template engine, you can do so by simply
calling `packer console` without a template file.

If you'd like to just see a specific single interpolation without launching
the REPL, you can do so by echoing and piping the string into the console
command:

```
$ echo {{timestamp}} | packer console
1559855090
```
