#!/bin/bash

if [[ ! -f file.txt ]] ; then
    echo 'hello' > file.txt
    exit 1
fi