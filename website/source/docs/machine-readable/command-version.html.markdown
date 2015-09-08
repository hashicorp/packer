---
description: |
    These are the machine-readable types that exist as part of the output of
    `packer version`.
layout: 'docs\_machine\_readable'
page_title: 'Command: version - Machine-Readable Reference'
...

# Version Command Types

These are the machine-readable types that exist as part of the output of
`packer version`.

<dl>
<dt>
version (1)
</dt>
<dd>
    <p>The version number of Packer running.</p>

    <p>
    <strong>Data 1: version</strong> - The version of Packer running,
    only including the major, minor, and patch versions. Example:
    "0.2.4".
    </p>

</dd>
<dt>
version-commit (1)
</dt>
<dd>
    <p>The SHA1 of the Git commit that built this version of Packer.</p>

    <p>
    <strong>Data 1: commit SHA1</strong> - The SHA1 of the commit.
    </p>

</dd>
<dt>
version-prerelease (1)
</dt>
<dd>
    <p>
    The prerelease tag (if any) for the running version of Packer. This
    can be "beta", "dev", "alpha", etc. If this is empty, you can assume
    it is a release version running.
    </p>

    <p>
    <strong>Data 1: prerelease name</strong> - The name of the
    prerelease tag.
    </p>

</dd>
</dl>
