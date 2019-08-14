# Contributing to `govultr`

We would love to get your feedback, thoughts, and overall improvements to `govultr`!

## Overview

- All code should run through `go fmt`
- All code **must be tested**
- All types, structs, and funcs **must be documented** for GoDocs

## Getting started

GoVultr supports `go modules` so you can pull down the repo outside of your `$GOPATH`.

You can also run:
`go get -u github.com/vultr/govultr`

## Testing

We aim to have as much code coverage as possible.

To run tests locally:

```sh
go test .
```

If you want to get more information on your local unit tests. You can run the following:

```sh
go test -v -coverprofile cover.out
go tool cover -html=cover.out
```

Upon opening a pull request we have CodeCov checks to make sure that code coverage meets a minimum requirement. In addition to CodeCov we have Travis CI that will run your unit tests on each pull request as well.

## Versioning

GoVultr follows [SemVer](http://semver.org/) for versioning. New functionality will result in a increment to the minor version and bug fixes will result in a increment to the patch version.

## Releases

Releases of new versions are done as independent pull requests and will be made by the maintainers.

To release a new version of `govultr` we must do the following:

- Update version number in `govultr.go` to reflect the new release version
- Make the appropriate updates to `CHANGELOG.md`. This should include the:
  - Version,
  - List of fix/features with accompanying pull request ID
  - Description of each fix/feature

```
## v0.0.1 (2019-05-05)

### Fixes
* Fixed random bug #12

### Features
* BareMetalServer functionality #13
```

- Submit a pull request with the changes above.
- Once the pull request is merged in, create a new tag and publish.
