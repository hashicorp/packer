#!/bin/sh

set -eux

export PACKER_ACC=1

RETVAL=0

for file in $(find ./ -type f ! -path "./vendor/*" -name '*.go' -print)
do
    if [[ -n "$(gofmt -l "$file")" ]]
    then
        echo -e "$file does not conform to gofmt rules. Run: gofmt -s -w $file"
        RETVAL=1
    fi
done

exit $RETVAL

go test -v -count 1 -timeout 20m ./driver ./iso ./clone