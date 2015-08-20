---
description: |
    These are the machine-readable types that can appear in almost any
    machine-readable output and are provided by Packer core itself.
layout: 'docs\_machine\_readable'
page_title: 'General Types - Machine-Readable Reference'
...

# General Types

These are the machine-readable types that can appear in almost any
machine-readable output and are provided by Packer core itself.

<dl>
<dt>
ui (2)
</dt>
<dd>
    <p>
    Specifies the output and type of output that would've normally
    gone to the console if Packer were running in human-readable
    mode.
    </p>

    <p>
    <strong>Data 1: type</strong> - The type of UI message that would've
    been outputted. Can be "say", "message", or "error".
    </p>
    <p>
    <strong>Data 2: output</strong> - The UI message that would have
    been outputted.
    </p>

</dd>
</dl>
