#!/usr/bin/env python
import textwrap
import sys
import fileinput


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
        print ""

    for line in sorted(lines, key=lambda s: s.lower()):
        if line.strip() == "":
            continue
        # print "-"*79
        wrapped = textwrap.wrap(line, 79)
        print wrapped[0]
        indented = " ".join([s.strip() for s in wrapped[1:]])
        for iline in textwrap.wrap(indented, 79-4):
            print "    " + iline

    # take care of blank line at end of selection
    sys.stdin.seek(0)
    if sys.stdin.readlines()[-1].strip() == "":
        print ""

"""
    for line in lines:
        do_indent = False
        for wline in textwrap.wrap(line, 79):
            if do_indent:
                print "    ",
            print wline
            do_indent = True
"""
