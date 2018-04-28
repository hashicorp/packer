#!/usr/bin/env bash

for f in $@; do
  [ -n "`dos2unix 2>/dev/null < $f | gofmt -s -d`" ] && echo $f
done

# always return success or else 'make' will abort
exit 0
