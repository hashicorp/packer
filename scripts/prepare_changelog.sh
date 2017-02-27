#!/bin/zsh


LAST_RELEASE=$1

if [ -z $LAST_RELEASE ]; then
    echo "you need to give the previous release version. prepare_changelog.sh v<version>"
    exit 1
fi


# git log --merges v0.10.2...c3861d167533fb797b0fae0c380806625712e5f7 |
git log --merges HEAD...${LAST_RELEASE} |
grep -o "Merge pull request #\(\d\+\)" | awk -F\# '{print $2}' | while read line
do
    grep -q "GH-${line}" CHANGELOG.md
    if [ $? -ne 0 ]; then
        echo $line
    fi
done | while read line
do
    echo "https://github.com/mitchellh/packer/pull/${line}"
    #TODO get tags. ignore docs
    echo $line
    vared -ch ok
done


#TODO: just generate it automatically using PR titles and tags
