---
layout: guides
sidebar_current: use-packer-with-comment
page_title: Use jq and Packer to comment your templates - Guides
description: |-
  You can add detailed comments beyond the root-level underscore-prefixed field
  supported by Packer, and remove them using jq.
---

#How to use jq to strip unsupported comments from a Packer template

One of the biggest complaints we get about packer is that json doesn't use comments. We're in the process of moving to HCL2, the same config language used by Terraform, which does allow comments. But in the meantime, you can add detailed comments beyond the root-level underscore-prefixed field supported by Packer, and remove them using jq.

Let's say we have a file named commented_template.json

``` json
{
  	"_comment": [
	    "this is",
	    "a multi-line",
	    "comment"
  	],
    "builders": [
        {
            "_comment": "this is a comment inside a builder",
            "type": "null",
            "communicator": "none"
        }
    ],
    "_comment": "this is a root level comment",
    "provisioners": [
        {
          "_comment": "this is a different comment",
          "type": "shell",
          "_comment": "this is yet another comment",
          "inline": ["echo hellooooo"]
        }
    ]
}
```

```sh
jq 'walk(if type == "object" then del(._comment) else . end)' commented_template.json > uncommented_template.json
```

will produce a new file containing:

```json
{
  "builders": [
    {
      "type": "null",
      "communicator": "none"
    }
  ],
  "provisioners": [
    {
      "type": "shell",
      "inline": [
        "echo hellooooo"
      ]
    }
  ]
}
```

Once you've got your uncommented file, you can call `packer build` on it like
you normally would.

## The walk function
If your install of jq does not have the walk function and you get an error like

```
jq: error: walk/1 is not defined at <top-level>,
```

You can create a file `~/.jq` and add the [walk function](https://github.com/stedolan/jq/blob/ad9fc9f559e78a764aac20f669f23cdd020cd943/src/builtin.jq#L255-L262) to it by hand.
