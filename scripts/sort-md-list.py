#!/usr/bin/env python3
# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: BUSL-1.1

"""
sort-md-list.py sorts markdown lists

Use this script to prepare sections of the changelog.

this script expects a bulleted markdown list in stdout

for example

./sort-md-list.py - <<EOF
* builder/amazon: Only delete temporary key if we created one. [GH-4850]
* core: Correctly reject config files which have junk after valid json.
        [GH-4906]
    * builder/azure: Replace calls to panic with error returns. [GH-4846]
* communicator/winrm: Use KeepAlive to keep long-running connections open.  [GH-4952]
EOF

output:

* builder/amazon: Only delete temporary key if we created one. [GH-4850]
* builder/azure: Replace calls to panic with error returns. [GH-4846]
* communicator/winrm: Use KeepAlive to keep long-running connections open.
    [GH-4952]
* core: Correctly reject config files which have junk after valid json.
    [GH-4906]

As you can see, the output is sorted and spaced appropriately.

Limitations:
    * nested lists are not supported
    * must be passed a list to stdin, it does not process the changelog
    * whitespace within the list is elided
"""

import fileinput
import sys
import textwrap


if __name__ == '__main__':
    lines = []
    working = []
    for line in fileinput.input():
        line = line.strip()
        if line.startswith('*'):
            if len(working):
                lines.append( " ".join(working))
            working = [line]
        else:
            working.append(line)
    if len(working):
        lines.append( " ".join(working))

    # take care of blank line at start of selection
    sys.stdin.seek(0)
    if sys.stdin.readlines()[0].strip() == "":
        print()

    for line in sorted(lines, key=lambda s: s.lower()):
        if line.strip() == "":
            continue
        # print "-"*79
        wrapped = textwrap.wrap(line, 79)
        print( wrapped[0] )
        indented = " ".join([s.strip() for s in wrapped[1:]])
        for iline in textwrap.wrap(indented, 79-4):
            print("     " + iline)

    # take care of blank line at end of selection
    sys.stdin.seek(0)
    if sys.stdin.readlines()[-1].strip() == "":
        print()
