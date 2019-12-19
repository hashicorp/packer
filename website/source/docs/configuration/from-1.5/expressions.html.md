---
layout: "docs"
page_title: "Expressions - Configuration Language"
sidebar_current: configuration-expressions
description: |-
  HCL allows the use of expressions to access data exported
  by resources and to transform and combine that data to produce other values.
---

# Expressions

_Expressions_ are used to refer to or compute values within a configuration.
The simplest expressions are just literal values, like `"hello"` or `5`, but
HCL also allows more complex expressions such as references to data exported by
resources, arithmetic, conditional evaluation, and a number of built-in
functions.

Expressions can be used in a number of places in HCL, but some contexts limit
which expression constructs are allowed, such as requiring a literal value of a
particular type or forbidding. Each language feature's documentation describes
any restrictions it places on expressions.

The rest of this page describes all of the features of Packer's
expression syntax.

## Types and Values

The result of an expression is a _value_. All values have a _type_, which
dictates where that value can be used and what transformations can be
applied to it.

HCL uses the following types for its values:

* `string`: a sequence of Unicode characters representing some text, like
  `"hello"`.
* `number`: a numeric value. The `number` type can represent both whole
  numbers like `15` and fractional values like `6.283185`.
* `bool`: either `true` or `false`. `bool` values can be used in conditional
  logic.
* `list` (or `tuple`): a sequence of values, like
  `["us-west-1a", "us-west-1c"]`. Elements in a list or tuple are identified by
  consecutive whole numbers, starting with zero.
* `map` (or `object`): a group of values identified by named labels, like
  `{name = "Mabel", age = 52}`.

Strings, numbers, and bools are sometimes called _primitive types._
Lists/tuples and maps/objects are sometimes called _complex types,_ _structural
types,_ or _collection types._

Finally, there is one special value that has _no_ type:

* `null`: a value that represents _absence_ or _omission._ If you set an
  argument of a source or module to `null`, Packer behaves as though you
  had completely omitted it — it will use the argument's default value if it has
  one, or raise an error if the argument is mandatory. `null` is most useful in
  conditional expressions, so you can dynamically omit an argument if a
  condition isn't met.

### Advanced Type Details

In most situations, lists and tuples behave identically, as do maps and objects.
Whenever the distinction isn't relevant, the Packer documentation uses each
pair of terms interchangeably (with a historical preference for "list" and
"map").

However, module authors and provider developers should understand the
differences between these similar types (and the related `set` type), since they
offer different ways to restrict the allowed values for input variables and
source arguments.

### Type Conversion

Expressions are most often used to set values for the arguments of resources and
child modules. In these cases, the argument has an expected type and the given
expression must produce a value of that type.

Where possible, Packer automatically converts values from one type to
another in order to produce the expected type. If this isn't possible, Packer
will produce a type mismatch error and you must update the configuration with a
more suitable expression.

Packer automatically converts number and bool values to strings when needed.
It also converts strings to numbers or bools, as long as the string contains a
valid representation of a number or bool value.

* `true` converts to `"true"`, and vice-versa
* `false` converts to `"false"`, and vice-versa
* `15` converts to `"15"`, and vice-versa

## Literal Expressions

A _literal expression_ is an expression that directly represents a particular
constant value. Packer has a literal expression syntax for each of the value
types described above:

* Strings are usually represented by a double-quoted sequence of Unicode
  characters, `"like this"`. There is also a "heredoc" syntax for more complex
  strings. String literals are the most complex kind of literal expression in
  Packer, and have additional documentation on this page:
    * See [String Literals](#string-literals) below for information about escape
      sequences and the heredoc syntax.
    * See [String Templates](#string-templates) below for information about
      interpolation and template directives.
* Numbers are represented by unquoted sequences of digits with or without a
  decimal point, like `15` or `6.283185`.
* Bools are represented by the unquoted symbols `true` and `false`.
* The null value is represented by the unquoted symbol `null`.
* Lists/tuples are represented by a pair of square brackets containing a
  comma-separated sequence of values, like `["a", 15, true]`.

    List literals can be split into multiple lines for readability, but always
    require a comma between values. A comma after the final value is allowed,
    but not required. Values in a list can be arbitrary expressions.
* Maps/objects are represented by a pair of curly braces containing a series of
  `<KEY> = <VALUE>` pairs:

    ```hcl
    {
      name = "John"
      age  = 52
    }
    ```

    Key/value pairs can be separated by either a comma or a line break. Values
    can be arbitrary expressions. Keys are strings; they can be left unquoted if
    they are a valid [identifier](./syntax.html#identifiers), but must be quoted
    otherwise. You can use a non-literal expression as a key by wrapping it in
    parentheses, like `(var.business_unit_tag_name) = "SRE"`.

## References to Named Values

Packer makes one named values available. 

The following named values are available:

* `source.<SOURCE TYPE>.<NAME>` is an object representing a
  [source](./sources.html) of the given type
  and name.

## String Literals

HCL has two different syntaxes for string literals. The
most common is to delimit the string with quote characters (`"`), like
`"hello"`. In quoted strings, the backslash character serves as an escape
sequence, with the following characters selecting the escape behavior:

| Sequence     | Replacement                                                                   |
| ------------ | ----------------------------------------------------------------------------- |
| `\n`         | Newline                                                                       |
| `\r`         | Carriage Return                                                               |
| `\t`         | Tab                                                                           |
| `\"`         | Literal quote (without terminating the string)                                |
| `\\`         | Literal backslash                                                             |
| `\uNNNN`     | Unicode character from the basic multilingual plane (NNNN is four hex digits) |
| `\UNNNNNNNN` | Unicode character from supplementary planes (NNNNNNNN is eight hex digits)    |

The alternative syntax for string literals is the so-called Here Documents or
"heredoc" style, inspired by Unix shell languages. This style allows multi-line
strings to be expressed more clearly by using a custom delimiter word on a line
of its own to close the string:

```hcl
<<EOF
hello
world
EOF
```

The `<<` marker followed by any identifier at the end of a line introduces the
sequence. Packer then processes the following lines until it finds one that
consists entirely of the identifier given in the introducer. In the above
example, `EOF` is the identifier selected. Any identifier is allowed, but
conventionally this identifier is in all-uppercase and begins with `EO`, meaning
"end of". `EOF` in this case stands for "end of text".

The "heredoc" form shown above requires that the lines following be flush with
the left margin, which can be awkward when an expression is inside an indented
block:

```hcl
block {
  value = <<EOF
hello
world
EOF
}
```

To improve on this, Packer also accepts an _indented_ heredoc string variant
that is introduced by the `<<-` sequence:

```hcl
block {
  value = <<-EOF
  hello
    world
  EOF
}
```

In this case, Packer analyses the lines in the sequence to find the one
with the smallest number of leading spaces, and then trims that many spaces
from the beginning of all of the lines, leading to the following result:

```
hello
  world
```

Backslash sequences are not interpreted in a heredoc string expression.
Instead, the backslash character is interpreted literally.

In both quoted and heredoc string expressions, Packer supports template
sequences that begin with `${` and `%{`. These are described in more detail
in the following section. To include these sequences _literally_ without
beginning a template sequence, double the leading character: `$${` or `%%{`.

## String Templates

Within quoted and heredoc string expressions, the sequences `${` and `%{` begin
_template sequences_. Templates let you directly embed expressions into a string
literal, to dynamically construct strings from other values.
