---
description: |
    The `packer fix` Packer command takes a template and finds backwards
    incompatible parts of it and brings it up to date so it can be used with the
    latest version of Packer. After you update to a new Packer release, you should
    run the fix command to make sure your templates work with the new release.
layout: docs
page_title: 'Fix - Command-Line'
...

# Command-Line: Fix

The `packer fix` Packer command takes a template and finds backwards
incompatible parts of it and brings it up to date so it can be used with the
latest version of Packer. After you update to a new Packer release, you should
run the fix command to make sure your templates work with the new release.

The fix command will output the changed template to standard out, so you should
redirect standard using standard OS-specific techniques if you want to save it
to a file. For example, on Linux systems, you may want to do this:

    $ packer fix old.json > new.json

If fixing fails for any reason, the fix command will exit with a non-zero exit
status. Error messages appear on standard error, so if you're redirecting
output, you'll still see error messages.

-> **Even when Packer fix doesn't do anything** to the template, the template
will be outputted to standard out. Things such as configuration key ordering and
indentation may be changed. The output format however, is pretty-printed for
human readability.

The full list of fixes that the fix command performs is visible in the help
output, which can be seen via `packer fix -h`.
