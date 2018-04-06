#!/bin/bash

# downloads the packfile of branch from a github repo.

set -e

if [ "$#" -ne 3 ]; then
    echo "Illegal number of arguments." > /dev/stderr
    echo > /dev/stderr
    echo "usage:" > /dev/stderr
    echo -e "\tgetpackfile <user> <repo> <branch>" > /dev/stderr
    exit 1
fi

user=$1
repo=$2
branch=$3

if [ -d /tmp/${repo} ] ; then
    echo "/tmp/${repo} exits, delete it and try again." > /dev/stderr
    exit 1
fi

pushd /tmp
git clone --branch ${branch} --single-branch http://github.com/${user}/${repo}.git
cd ${repo}
git gc
popd

cp /tmp/${repo}/.git/objects/pack/*.pack ./${user}-${repo}.pack
du -h ./${user}-${repo}.pack
