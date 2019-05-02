---
layout: guides
sidebar_current: isotime-template-function
page_title: Using the isotime template function - Guides
description: |-
  It can be a bit confusing to figure out how to format your isotime using the
  golang reference date string. Here is a small guide and some examples.
---

# Using the Isotime template function with a format string

The way you format isotime in golang is a bit nontraditional compared to how
you may be used to formatting datetime strings.

Full docs and examples for the golang time formatting function can be found
[here](https://golang.org/pkg/time/#example_Time_Format)

However, the formatting basics are worth describing here. From the [golang docs](https://golang.org/pkg/time/#pkg-constants):


>These are predefined layouts for use in Time.Format and time.Parse. The
>reference time used in the layouts is the specific time:
>
>Mon Jan 2 15:04:05 MST 2006
>
>which is Unix time 1136239445. Since MST is GMT-0700, the reference time
>can be thought of as
>
>01/02 03:04:05PM '06 -0700
>
> To define your own format, write down what the reference time would look like
> formatted your way; see the values of constants like ANSIC, StampMicro or
> Kitchen for examples. The model is to demonstrate what the reference time
> looks like so that the Format and Parse methods can apply the same
> transformation to a general time value.


So what does that look like in a Packer template function?

``` json
{
	"variables":
	{
		"myvar": "packer-{{isotime \"2006-01-02 03:04:05\"}}"
	},
	"builders": [
		{
			"type": "null",
			"communicator": "none"
		}
	],
	"provisioners": [
		{
			"type": "shell-local",
			"inline": ["echo {{ user `myvar`}}"]
		}
	]
}
```

You can switch out the variables section above with the following examples to
get different timestamps:

Date only, not time:

```json
	"variables":
	{
		"myvar": "packer-{{isotime \"2006-01-02\"}}"
	},
```

A timestamp down to the millisecond:

```json
	"variables":
	{
		"myvar": "packer-{{isotime \"Jan-_2-15:04:05.000\"}}"
	},
```

Or just the time as it would appear on a digital clock:

```json
	"variables":
	{
		"myvar": "packer-{{isotime \"3:04PM\"}}"
	},
```