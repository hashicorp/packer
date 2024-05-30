#!/bin/zsh
# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: BUSL-1.1


LAST_RELEASE=$1

set -o pipefail

if [ -z $LAST_RELEASE ]; then
    echo "you need to give the previous release version. prepare_changelog.sh v<version>"
    exit 1
fi

if [ -z "$(which jq)" ]; then
    echo "jq command not found"
    return 1
fi

if [ -z "$(which jq)" ]; then
    echo "gh command not found"
    return 1
fi

get_prs(){
   release_time="$(gh release view --json "createdAt" --jq '.createdAt' ${LAST_RELEASE})"
   gh pr list -s merged -S "merged:>=$release_time -label:documentation -label:automated -label:tech-debt -label:website -label:legal -label:docs -author:hc-github-team-packer" --json "number" --jq '.[]|.number' \
   | while read line
    do
        if grep -q "GH-${line}" CHANGELOG.md; then
            continue
        fi
        echo $line
    done | while read PR_NUM
    do
        out=$(gh pr view ${PR_NUM} --json "title,labels,url" > pull.json)
        if [ "$?" -ne 0 ]; then
            echo "bad response from github: manually check PR ${PR_NUM}"
            continue
        fi

        echo "$(jq -r '.title' < pull.json) - [GH-${PR_NUM}](https://github.com/hashicorp/packer/pull/${PR_NUM})"
        rm -f pull.json
    done
}

get_prs | while read line; do
    echo $line
    if [[ "$line" =~ "bad" ]]; then
        exit 1
    fi
    echo "Press enter to continue with next entry."
    vared -ch ok
done

#TODO: just generate it automatically using PR titles and tags
