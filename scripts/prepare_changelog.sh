#!/bin/zsh

LAST_RELEASE=$1
DO_PR_CHECK=1

set -o pipefail

is_doc_pr(){
    if ! (($+commands[jq])); then
        DO_PR_CHECK=0
        echo "jq not found"
        return 1
    fi
    PR_NUM=$1
    out=$(curl -fsS "https://api.github.com/repos/hashicorp/packer/issues/${PR_NUM}" | jq '[.labels[].name == "docs"] | any')
    exy="$?"
    if [ $exy -ne 0 ]; then
        echo "bad response from github"
        exit $exy
    fi
    grep -q true <<< $out
    return $?
}

if [ -z $LAST_RELEASE ]; then
    echo "you need to give the previous release version. prepare_changelog.sh v<version>"
    exit 1
fi

get_prs(){
    # git log --merges v0.10.2...c3861d167533fb797b0fae0c380806625712e5f7 |
    git log --merges HEAD...${LAST_RELEASE} |
    grep -o "Merge pull request #\(\d\+\)" | awk -F\# '{print $2}' | while read line
    do
        grep -q "GH-${line}" CHANGELOG.md
        if [ $? -ne 0 ]; then
            echo $line
        fi
    done | while read PR_NUM
    do
        if (($DO_PR_CHECK)) && is_doc_pr $PR_NUM; then
            continue
        fi
        echo "https://github.com/hashicorp/packer/pull/${PR_NUM}"
    done
}

#is_doc_pr 52061111
# is_doc_pr 5206 # non-doc pr
#is_doc_pr 5434 # doc pr
#echo $?
#exit

# prpid=$!
# trap 'kill -9 ${prpid}; exit' INT TERM

get_prs | while read line; do
    echo $line
    if [[ "$line" =~ "bad" ]]; then
        exit 1
    fi
    vared -ch ok
done


#TODO: just generate it automatically using PR titles and tags
