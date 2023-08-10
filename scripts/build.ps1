# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: BUSL-1.1

<#
    .Synopsis
    Build script for Packer.

    .Description
    Build script for Packer for all supported platforms and architectures.
    By default the following OSs and architectures are targeted.

    OS:
     * linux
     * darwin
     * windows
     * freebsd
     * openbsd

    Architecture:
     * 386
     * amd64
     * arm

    If the environment variable PACKER_DEV is defined, then the OS and
    architecture of the go binary in the path is used.

    The built binary is stamped with the current version number of Packer,
    the latest git commit, and +CHANGES if there are any outstanding
    changes in the current repository, e.g.

      Packer v0.10.1.dev (3c736322ba3a5fcb3a4e92394011a2e56f396da6+CHANGES)

    The build artifacts for the current OS and architecture are copied to
    bin and $GOPATH\bin.

    .Example
    .\scripts\build.ps1
#>

# This script builds the application from source for multiple platforms.

# Get the parent directory of where this script is.
$DIR = [System.IO.Path]::GetDirectoryName($PSScriptRoot)

# Change into that directory
Push-Location $DIR | Out-Null

# Get the git commit
$GIT_COMMIT = $(git.exe rev-parse HEAD)
git.exe status --porcelain | Out-Null
if ($LastExitCode -eq 0) {
    $GIT_DIRTY = "+CHANGES"
}

# If its dev mode, only build for ourself
if (Test-Path env:PACKER_DEV) {
    $XC_OS=$(go.exe env GOOS)
    $XC_ARCH=$(go.exe env GOARCH)
} else {
    if (Test-Path env:XC_ARCH) {
        $XC_ARCH = $(Get-Content env:XC_ARCH)
    } else {
        $XC_ARCH="386 amd64 arm arm64 ppc64le"
    }

    if (Test-Path env:XC_OS) {
        $XC_OS = $(Get-Content env:XC_OS)
    } else {
        $XC_OS = "linux darwin windows freebsd openbsd solaris"
    }
}

# Delete the old dir
echo "==> Removing old directory..."
Remove-Item -Recurse -ErrorAction Ignore -Force "bin\"
Remove-Item -Recurse -ErrorAction Ignore -Force "pkg\"
New-Item -Type Directory -Name bin | Out-Null

# Delete the old dir
echo "==> Building..."
gox.exe `
  -os="${XC_OS}" `
  -arch="${XC_ARCH}" `
  -ldflags "-X github.com/hashicorp/packer/version.GitCommit=${GIT_COMMIT}${GIT_DIRTY}" `
  -output "pkg/{{.OS}}_{{.Arch}}/packer" `
  .

if ($LastExitCode -ne 0) {
    exit 1
}

# Move all the compiled things to the $GOPATH/bin
$GOPATH=$(go.exe env GOPATH)

# Copy our OS/Arch to the bin/ directory
echo "==> Copying binaries for this platform..."
Get-ChildItem ".\pkg\$(go env GOOS)_$(go env GOARCH)\" `
  |? { !($_.PSIsContainer) } `
  |% {
      Copy-Item $_.FullName "bin\"
      Copy-Item $_.FullName "${GOPATH}\bin\"
  }

# Done!
echo "`r`n==> Results:"
Get-ChildItem bin\
