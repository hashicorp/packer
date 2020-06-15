#!/bin/zsh

LAST_RELEASE=$1
DO_PR_CHECK=1

set -o pipefail

is_doc_or_tech_debt_pr(){
    if ! (($+commands[jq])); then
        DO_PR_CHECK=0
        echo "jq not found"
        return 1
    fi
    out=$(cat pull.json | python -m json.tool \
    | jq '[.labels[].name == "docs" or .labels[].name == "tech-debt" or .labels[].name == "website"] | any')
    grep -q true <<< $out
    return $?
}

if [ -z $LAST_RELEASE ]; then
    echo "you need to give the previous release version. prepare_changelog.sh v<version>"
    exit 1
fi

get_prs(){
    # git log v0.10.2...c3861d167533fb797b0fae0c380806625712e5f7 |
    git log HEAD...${LAST_RELEASE} --first-parent --oneline --grep="Merge pull request #[0-9]\+" --grep="(#[0-9]\+)$" |
    grep -o "#\([0-9]\+\)" | awk -F\# '{print $2}' | while read line
    do
        grep -q "GH-${line}" CHANGELOG.md
        if [ $? -ne 0 ]; then
            echo $line
        fi
    done | while read PR_NUM
    do
        if [[ -z "${GITHUB_TOKEN}" ]] || [[ -z "${GITHUB_USERNAME}" ]] ; then
          out=$(curl -fsS "https://api.github.com/repos/hashicorp/packer/issues/${PR_NUM}" -o pull.json)
        else
          # authenticated call
          out=$(curl -u ${GITHUB_USERNAME}:${GITHUB_TOKEN} -fsS "https://api.github.com/repos/hashicorp/packer/issues/${PR_NUM}" -o pull.json)
        fi
        exy="$?"
        if [ $exy -ne 0 ]; then
            echo "bad response from github: manually check PR ${PR_NUM}"
            continue
        fi

        if (($DO_PR_CHECK)) && is_doc_or_tech_debt_pr; then
            echo "Skipping PR ${PR_NUM}: labeled as tech debt, docs or website. (waiting a second so we don't get rate-limited...)"
            continue
        fi
        echo "$(cat pull.json | python -m json.tool | jq '.title') - https://github.com/hashicorp/packer/pull/${PR_NUM}"
    done
}

#is_doc_or_tech_debt_pr 52061111
# is_doc_or_tech_debt_pr 5206 # non-doc pr
#is_doc_or_tech_debt_pr 5434 # doc pr
#echo $?
#exit

# prpid=$!
# trap 'kill -9 ${prpid}; exit' INT TERM

get_prs | while read line; do
    echo $line
    if [[ "$line" =~ "bad" ]]; then
        exit 1
    elif [[ "$line" =~ "Skipping" ]]; then
        sleep 1 # GH will rate limit us if we have several in a row
        continue
    fi
    rm -f pull.json
    vared -ch ok
done


#TODO: just generate it automatically using PR titles and tags
