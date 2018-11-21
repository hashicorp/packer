#!/bin/sh

set -e

echo "mode: atomic" > coverage.txt
counter=0
for package in $(find . -iname '*.go' | grep -v '^./vendor' | xargs -n 1 dirname | sort -n | uniq -c | awk '{print $2}'); do
  out="${counter}.txt"
  go test -v -coverprofile="$out" -covermode=atomic "$package"
  tail -n +2 "$out" >> coverage.txt
  rm "$out"
  counter=$(expr "$counter" + 1)
done
