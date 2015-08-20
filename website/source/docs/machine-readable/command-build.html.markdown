---
description: |
    These are the machine-readable types that exist as part of the output of
    `packer build`.
layout: 'docs\_machine\_readable'
page_title: 'Command: build - Machine-Readable Reference'
...

# Build Command Types

These are the machine-readable types that exist as part of the output of
`packer build`.

<dl>
<dt>
artifact (&gt;= 2)
</dt>
<dd>
    <p>
    Information about an artifact of the targeted item. This is a
    fairly complex (but uniform!) machine-readable type that contains
    subtypes. The subtypes are documented within this page in the
    syntax of "artifact subtype: SUBTYPE". The number of arguments within
    that subtype is in addition to the artifact args.
    </p>

    <p>
    <strong>Data 1: index</strong> - The zero-based index of the
    artifact being described. This goes up to "artifact-count" (see
    below).
    </p>
    <p>
    <strong>Data 2: subtype</strong> - The subtype that describes
    the remaining arguments. See the documentation for the
    subtype docs throughout this page.
    </p>
    <p>
    <strong>Data 3..n: subtype data</strong> - Zero or more additional
    data points related to the subtype. The exact count and meaning
    of this subtypes comes from the subtype documentation.
    </p>

</dd>
<dt>
artifact-count (1)
</dt>
<dd>
    <p>
    The number of artifacts associated with the given target. This
    will always be outputted _before_ any other artifact information,
    so you're able to know how many upcoming artifacts to look for.
    </p>

    <p>
    <strong>Data 1: count</strong> - The number of artifacts as
    a base 10 integer.
    </p>

</dd>
<dt>
artifact subtype: builder-id (1)
</dt>
<dd>
    <p>
    The unique ID of the builder that created this artifact.
    </p>

    <p>
    <strong>Data 1: id</strong> - The unique ID of the builder.
    </p>

</dd>
<dt>
artifact subtype: end (0)
</dt>
<dd>
    <p>
    The last machine-readable output line outputted for an artifact.
    This is a sentinel value so you know that no more data related to
    the targetted artifact will be outputted.
    </p>

</dd>
<dt>
artifact subtype: file (2)
</dt>
<dd>
    <p>
    A single file associated with the artifact. There are 0 to
    "files-count" of these entries to describe every file that is
    part of the artifact.
    </p>

    <p>
    <strong>Data 1: index</strong> - Zero-based index of the file.
    This goes from 0 to "files-count" minus one.
    </p>

    <p>
    <strong>Data 2: filename</strong> - The filename.
    </p>

</dd>
<dt>
artifact subtype: files-count (1)
</dt>
<dd>
    <p>
    The number of files associated with this artifact. Not all
    artifacts have files associated with it.
    </p>

    <p>
    <strong>Data 1: count</strong> - The number of files.
    </p>

</dd>
<dt>
artifact subtype: id (1)
</dt>
<dd>
    <p>
    The ID (if any) of the artifact that was built. Not all artifacts
    have associated IDs. For example, AMIs built have IDs associated
    with them, but VirtualBox images do not. The exact format of the ID
    is specific to the builder.
    </p>

    <p>
    <strong>Data 1: id</strong> - The ID of the artifact.
    </p>

</dd>
<dt>
artifact subtype: nil (0)
</dt>
<dd>
    <p>
    If present, this means that the artifact was nil, or that the targeted
    build completed successfully but no artifact was created.
    </p>

</dd>
<dt>
artifact subtype: string (1)
</dt>
<dd>
    <p>
    The human-readable string description of the artifact provided by
    the artifact itself.
    </p>

    <p>
    <strong>Data 1: string</strong> - The string output for the artifact.
    </p>

</dd>
<dt>
error-count (1)
</dt>
<dd>
    <p>
    The number of errors that occurred during the build. This will
    always be outputted before any errors so you know how many are coming.
    </p>

    <p>
    <strong>Data 1: count</strong> - The number of build errors as
    a base 10 integer.
    </p>

</dd>
<dt>
error (1)
</dt>
<dd>
    <p>
    A build error that occurred. The target of this output will be
    the build that had the error.
    </p>

    <p>
    <strong>Data 1: error</strong> - The error message as a string.
    </p>

</dd>
</dl>
