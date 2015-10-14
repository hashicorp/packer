---
description: |
    These are the machine-readable types that exist as part of the output of
    `packer inspect`.
layout: 'docs\_machine\_readable'
page_title: 'Command: inspect - Machine-Readable Reference'
...

# Inspect Command Types

These are the machine-readable types that exist as part of the output of
`packer inspect`.

<dl>
<dt>
template-variable (3)
</dt>
<dd>
    <p>
    A <a href="/docs/templates/user-variables.html">user variable</a>
    defined within the template.
    </p>

    <p>
    <strong>Data 1: name</strong> - Name of the variable.
    </p>

    <p>
    <strong>Data 2: default</strong> - The default value of the
    variable.
    </p>

    <p>
    <strong>Data 3: required</strong> - If non-zero, then this variable
    is required.
    </p>

</dd>
<dt>
template-builder (2)
</dt>
<dd>
    <p>
    A builder defined within the template
    </p>

    <p>
    <strong>Data 1: name</strong> - The name of the builder.
    </p>
    <p>
    <strong>Data 2: type</strong> - The type of the builder. This will
    generally be the same as the name unless you explicitly override
    the name.
    </p>

</dd>
<dt>
template-provisioner (1)
</dt>
<dd>
    <p>
    A provisioner defined within the template. Multiple of these may
    exist. If so, they are outputted in the order they would run.
    </p>

    <p>
    <strong>Data 1: name</strong> - The name/type of the provisioner.
    </p>

</dd>
</dl>
