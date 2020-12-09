#!/bin/bash

# On first try, exits 1; on second try, passes.
if [[ ! -f test-fixtures/file.txt ]] ; then
    echo 'hello' > test-fixtures/file.txt
    exit 1
fi