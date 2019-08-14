# GoVultr

[![Build Status](https://travis-ci.org/vultr/govultr.svg?branch=master)](https://travis-ci.org/vultr/govultr)
[![GoDoc](https://godoc.org/github.com/vultr/govultr?status.svg)](https://godoc.org/github.com/vultr/govultr)
[![codecov](https://codecov.io/gh/vultr/govultr/branch/master/graph/badge.svg?token=PDJXBc7Rci)](https://codecov.io/gh/vultr/govultr)
[![Go Report Card](https://goreportcard.com/badge/github.com/vultr/govultr)](https://goreportcard.com/report/github.com/vultr/govultr)

The official Vultr Go client - GoVultr allows you to interact with the Vultr V1 API.

## Installation

```sh
go get -u github.com/vultr/govultr
```

## Usage

Vultr uses a PAT (Personal Access token) to interact/authenticate with the APIs. An API Key can be generated and acquired from the API menu in [settings](https://my.vultr.com/settings/#settingsapi).

To instantiate a govultr client you invoke `NewClient()`.

This takes in two parameters:

- \*http.Client
- API Key

You can define your own `http.Client` however if you pass in `nil` then you will be defaulted to use `http.DefaultClient`. For the API key, we recommend you store this as a environment variable and not hard code it.

There are also three optional parameters you may change regarding the client:

- BaseUrl: allows you to override the base URL that Vultr defaults to
- UserAgent: allows you to override the UserAgent that Vultr defaults to
- RateLimit: Vultr currently rate limits how fast you can make calls back to back. This lets you configure if you want a delay in between calls

### Example Client Setup

```go
package main

import (
	"github.com/vultr/govultr"
	"os"
)

func main() {
	apiKey := os.Getenv("VultrAPIKey")

	vultrClient := govultr.NewClient(nil, apiKey)

	// Optional changes
	_ = vultrClient.SetBaseURL("https://api.vultr.com")
	vultrClient.SetUserAgent("mycool-app")
	vultrClient.SetRateLimit(500)
}
```

### Example Usage

Create a VPS

```go
vpsOptions := &govultr.ServerOptions{
	Label:                "awesome-go-app",
	Hostname:             "awesome-go.com",
	EnablePrivateNetwork: true,
	AutoBackups:          true,
	EnableIPV6:           true,
}

// RegionId, VpsPlanID, OsID can be grabbed from their respective API calls
res, err := vultrClient.Server.Create(context.Background(), 1, 201, 1, vpsOptions)

if err != nil {
	fmt.Println(err)
}
```

## Versioning

This project follows [SemVer](http://semver.org/) for versioning. For the versions available, see the [tags on this repository](https://github.com/vultr/govultr/tags).

## Documentation

For detailed information about our V1 API head over to our [API documentation](https://www.vultr.com/api/).

If you want more details about this client's functionality then check out our [GoDoc](https://godoc.org/github.com/vultr/govultr) documentation.

## Contributing

Feel free to send pull requests our way! Please see the [contributing guidelines](CONTRIBUTING.md).

## License

This project is licensed under the MIT License - see the [LICENSE.md](LICENSE) file for details.
