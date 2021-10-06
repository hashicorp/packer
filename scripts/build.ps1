#!/usr/bin/pwsh

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

# Move all the compiled things to the $GOPATH/bin
$GOPATH=$(go.exe env GOPATH)

# Delete the old dir
echo "==> Building..."
$ALL_XC_ARCH="386", "amd64", "arm",  "arm64", "ppc64le", "mips", "mips64", "mipsle", "mipsle64", "s390x"
$ALL_XC_OS="linux", "darwin", "windows", "freebsd", "openbsd", "solaris"
$SKIPPED_OSARCH="darwin/arm",
"freebsd/arm",
"freebsd/arm64",
"darwin/386",
"solaris/386",
"windows/arm",
"solaris/arm",
"windows/arm64",
"solaris/arm64",
"darwin/ppc64le",
"windows/ppc64le",
"freebsd/ppc64le",
"openbsd/ppc64le",
"solaris/ppc64le",
"darwin/mips",
"windows/mips",
"freebsd/mips",
"openbsd/mips",
"solaris/mips",
"darwin/mips64",
"windows/mips64",
"freebsd/mips64",
"solaris/mips64",
"darwin/mipsle",
"windows/mipsle",
"freebsd/mipsle",
"openbsd/mipsle",
"openbsd/mipsle",
"solaris/mipsle",
"linux/mipsle64",
"darwin/mipsle64",
"windows/mipsle64",
"freebsd/mipsle64",
"openbsd/mipsle64",
"solaris/mipsle64",
"darwin/s390x",
"windows/s390x",
"freebsd/s390x",
"openbsd/s390x",
"solaris/s390x"

# build for everything
ForEach($arch in $ALL_XC_ARCH)
{
    ForEach($os in $ALL_XC_OS)
    {
        $OS_ARCH="$os/$arch"
        write-host "Building for $OS_ARCH"
        if ($SKIPPED_OSARCH -contains $OS_ARCH) {
            write-host "Found in skip list, skipping: $OS_ARCH"
            continue
        }
        $OS_ARCH=$os + "_" + $arch
        $BUILD_OUT="pkg/$OS_ARCH/packer"
        go build `
          -ldflags "-X github.com/hashicorp/packer/version.GitCommit=${GIT_COMMIT}${GIT_DIRTY}" `
          -output $BUILD_OUT `
          .
    }
}

if ($LastExitCode -ne 0) {
    exit 1
}


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
