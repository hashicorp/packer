#!/bin/bash
#
# This script only builds the application from source.
set -e

NO_COLOR="\x1b[0m"
OK_COLOR="\x1b[32;01m"
ERROR_COLOR="\x1b[31;01m"
WARN_COLOR="\x1b[33;01m"

# http://stackoverflow.com/questions/4023830/bash-how-compare-two-strings-in-version-format
verify_go () {
    if [[ $1 == $2 ]]; then
        return 0
    fi

    local IFS=.
    local i ver1=($1) ver2=($2)

    for ((i=${#ver1[@]}; i<${#ver2[@]}; i++)); do
        ver1[i]=0
    done

    for ((i=0; i<${#ver1[@]}; i++)); do
        if [[ -z ${ver2[i]} ]]; then
            ver2[i]=0
        fi
        if ((10#${ver1[i]} > 10#${ver2[i]})); then
            echo -e "${ERROR_COLOR}==> Required Go version $1 not installed. Found $2 instead"
            exit 1
        fi
    done
}

GO_MINIMUM_VERSION=1.2
GO_INSTALLED_VERSION=$(go version | cut -d ' ' -f 3)
GO_INSTALLED_VERSION=${GO_INSTALLED_VERSION#"go"}

echo -e "${OK_COLOR}==> Verifying Go"
verify_go $GO_MINIMUM_VERSION $GO_INSTALLED_VERSION

# Get the parent directory of where this script is.
SOURCE="${BASH_SOURCE[0]}"
while [ -h "$SOURCE" ] ; do SOURCE="$(readlink "$SOURCE")"; done
DIR="$( cd -P "$( dirname "$SOURCE" )/.." && pwd )"

# Change into that directory
cd $DIR

# Compile the thing
export XC_ARCH=$(go env GOARCH)
export XC_OS=$(go env GOOS)
./scripts/compile.sh

# Move all the compiled things to the PATH
case $(uname) in
    CYGWIN*)
        GOPATH="$(cygpath $GOPATH)"
        ;;
esac
IFS=: MAIN_GOPATH=( $GOPATH )
cp pkg/${XC_OS}_${XC_ARCH}/* ${MAIN_GOPATH}/bin
cp pkg/${XC_OS}_${XC_ARCH}/* ./bin
