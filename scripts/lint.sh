#!/usr/bin/env bash

CHANGED_FILES=$(git diff --name-status `git merge-base origin/master HEAD`...HEAD | grep '^A.*\.go$'| awk '{print $2}')
if [ ! -z "${CHANGED_FILES}" ]; then
  echo $CHANGED_FILES | xargs -n1 golangci-lint run
fi
