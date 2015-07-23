---
description: |
    If you are or were a user of Veewee, then there is an official tool called
    veewee-to-packer that will convert your Veewee definition into an equivalent
    Packer template. Even if you're not a Veewee user, Veewee has a large library of
    templates that can be readily used with Packer by simply converting them.
layout: docs
page_title: Convert Veewee Definitions to Packer Templates
...

# Veewee-to-Packer

If you are or were a user of [Veewee](https://github.com/jedi4ever/veewee), then
there is an official tool called
[veewee-to-packer](https://github.com/mitchellh/veewee-to-packer) that will
convert your Veewee definition into an equivalent Packer template. Even if
you're not a Veewee user, Veewee has a [large
library](https://github.com/jedi4ever/veewee/tree/master/templates) of templates
that can be readily used with Packer by simply converting them.

## Installation and Usage

Since Veewee itself is a Ruby project, so too is the veewee-to-packer
application so that it can read the Veewee configurations. Install it using
RubyGems:

``` {.text}
$ gem install veewee-to-packer
...
```

Once installed, usage is easy! Just point `veewee-to-packer` at the
`definition.rb` file of any template. The converter will output any warnings or
messages about the conversion. The example below converts a CentOS template:

``` {.text}
$ veewee-to-packer templates/CentOS-6.4/definition.rb
Success! Your Veewee definition was converted to a Packer
template! The template can be found in the `template.json` file
in the output directory: output

Please be sure to run `packer validate` against the new template
to verify settings are correct. Be sure to `cd` into the directory
first, since the template has relative paths that expect you to
use it from the same working directory.
```

***Voila!*** By default, `veewee-to-packer` will output a template that contains
a builder for both VirtualBox and VMware. You can use the `-only` flag on
`packer build` to only build one of them. Otherwise you can use the `--builder`
flag on `veewee-to-packer` to only output specific builder configurations.

## Limitations

None, really. The tool will tell you if it can't convert a part of a template,
and whether that is a critical error or just a warning. Most of Veewee's
functions translate perfectly over to Packer. There are still a couple missing
features in Packer, but they're minimal.

## Bugs

If you find any bugs, please report them to the [veewee-to-packer issue
tracker](https://github.com/mitchellh/veewee-to-packer). I haven't been able to
exhaustively test every Veewee template, so there are certainly some edge cases
out there.
