#!/bin/sh -e

(cd driver && ./test.sh "$@")
(cd clone && ./test.sh "$@")
(cd iso && ./test.sh "$@")