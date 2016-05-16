#!/bin/bash
set -e

go test -v ./...

mkdir -p .cover
go list ./... | xargs -I% bash -c 'name="%"; go test -covermode=count % --coverprofile=.cover/${name//\//_} '
echo "mode: count" > profile.cov
cat .cover/* | grep -v mode >> profile.cov
rm -rf .cover

go tool cover -func=profile.cov
