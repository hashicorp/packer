#!/bin/sh

RETVAL=0

for file in $(find . -name '*.go' -not -path './vendor/*')
do
    if [ -n "$(gofmt -l $file)" ]
    then
        echo "$file does not conform to gofmt rules. Run: gofmt -s -w $file" >&2
        RETVAL=1
    fi
done

exit $RETVAL
