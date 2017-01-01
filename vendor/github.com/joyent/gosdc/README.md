# gosdc

[![wercker status](https://app.wercker.com/status/349ee60ed0afffd99d2b2b354ada5938/s/master "wercker status")](https://app.wercker.com/project/bykey/349ee60ed0afffd99d2b2b354ada5938)

`gosdc` is a Go client for Joyent's SmartDataCenter

<!-- markdown-toc start - Don't edit this section. Run M-x markdown-toc-generate-toc again -->
**Table of Contents**

- [gosdc](#gosdc)
    - [Usage](#usage)
        - [Examples](#examples)
    - [Resources](#resources)
    - [License](#license)

<!-- markdown-toc end -->

## Usage

To create a client
([`*cloudapi.Client`](https://godoc.org/github.com/joyent/gosdc/cloudapi#Client)),
you'll need a few things:

1. your account ID
2. the ID of the key associated with your account
3. your private key material
4. the cloud endpoint you want to use (for example
   `https://us-east-1.api.joyentcloud.com`)

Given these four pieces of information, you can initialize a client with the
following:

```go
package main

import (
	"io/ioutil"
	"log"
	"os"

	"github.com/joyent/gocommon/client"
	"github.com/joyent/gosdc/cloudapi"
	"github.com/joyent/gosign/auth"
)

func client(key, keyId, account, endpoint string) (*cloudapi.Client, error) {
	keyData, err := ioutil.ReadFile(key)
	if err != nil {
		return nil, err
	}
	userAuth, err := auth.NewAuth(account, string(keyData), "rsa-sha256")
	if err != nil {
		return nil, err
	}

	creds := &auth.Credentials{
		UserAuthentication: auth,
		SdcKeyId:           keyId,
		SdcEndpoint:        auth.Endpoint{URL: endpoint},
	}

	return cloudapi.New(client.NewClient(
		creds.SdcEndpoint.URL,
		cloudapi.DefaultAPIVersion,
		creds,
		log.New(os.Stderr, "", log.LstdFlags),
	)), nil
}
```

### Examples

Projects using the gosdc API:

 - [triton-terraform](https://github.com/joyent/triton-terraform)

## Resources

After creating a client, you can manipulate resources in the following ways:

| Resource | Create | Read | Update | Delete | Extra |
|----------|--------|------|--------|--------|-------|
| Datacenters | | [GetDatacenter](https://godoc.org/github.com/joyent/gosdc/cloudapi#Client.GetDatacenter), [ListDatacenters](https://godoc.org/github.com/joyent/gosdc/cloudapi#Client.ListDatacenters) | | | |
| Firewall Rules | [CreateFirewallRule](https://godoc.org/github.com/joyent/gosdc/cloudapi#Client.CreateFirewallRule) | [GetFirewallRule](https://godoc.org/github.com/joyent/gosdc/cloudapi#Client.GetFirewallRule), [ListFirewallRules](https://godoc.org/github.com/joyent/gosdc/cloudapi#Client.ListFirewallRules), [ListmachineFirewallRules](https://godoc.org/github.com/joyent/gosdc/cloudapi#Client.ListMachineFirewallRules) | [UpdateFirewallRule](https://godoc.org/github.com/joyent/gosdc/cloudapi#Client.UpdateFirewallRule), [EnableFirewallRule](https://godoc.org/github.com/joyent/gosdc/cloudapi#Client.EnableFirewallRule), [DisableFirewallRule](https://godoc.org/github.com/joyent/gosdc/cloudapi#Client.DisableFirewallRule) | [DeleteFirewallRule](https://godoc.org/github.com/joyent/gosdc/cloudapi#Client.DeleteFirewallRule) | |
| Instrumentations | [CreateInstrumentation](https://godoc.org/github.com/joyent/gosdc/cloudapi#Client.CreateInstrumentation) | [GetInstrumentation](https://godoc.org/github.com/joyent/gosdc/cloudapi#Client.GetInstrumentation), [ListInstrumentations](https://godoc.org/github.com/joyent/gosdc/cloudapi#Client.ListInstrumentations), [GetInstrumentationHeatmap](https://godoc.org/github.com/joyent/gosdc/cloudapi#Client.GetInstrumentationHeatmap), [GetInstrumentationHeatmapDetails](https://godoc.org/github.com/joyent/gosdc/cloudapi#Client.GetInstrumentationHeatmapDetails), [GetInstrumentationValue](https://godoc.org/github.com/joyent/gosdc/cloudapi#Client.GetInstrumentationValue) | | [DeleteInstrumentation](https://godoc.org/github.com/joyent/gosdc/cloudapi#Client.DeleteInstrumentation) | [DescribeAnalytics](https://godoc.org/github.com/joyent/gosdc/cloudapi#Client.DescribeAnalytics) |
| Keys | [CreateKey](https://godoc.org/github.com/joyent/gosdc/cloudapi#Client.CreateKey) | [GetKey](https://godoc.org/github.com/joyent/gosdc/cloudapi#Client.GetKey), [ListKeys](https://godoc.org/github.com/joyent/gosdc/cloudapi#Client.ListKeys) | | [DeleteKey](https://godoc.org/github.com/joyent/gosdc/cloudapi#Client.DeleteKey) | |
| Machines | [CreateMachine](https://godoc.org/github.com/joyent/gosdc/cloudapi#Client.CreateMachine) | [GetMachine](https://godoc.org/github.com/joyent/gosdc/cloudapi#Client.GetMachine), [ListMachines](https://godoc.org/github.com/joyent/gosdc/cloudapi#Client.ListMachines), [ListFirewallRuleMachines](https://godoc.org/github.com/joyent/gosdc/cloudapi#Client.ListFirewallRuleMachines)  | [RenameMachine](https://godoc.org/github.com/joyent/gosdc/cloudapi#Client.RenameMachine), [ResizeMachine](https://godoc.org/github.com/joyent/gosdc/cloudapi#Client.ResizeMachine) | [DeleteMachine](https://godoc.org/github.com/joyent/gosdc/cloudapi#Client.DeleteMachine) | [CountMachines](https://godoc.org/github.com/joyent/gosdc/cloudapi#Client.CountMachines), [MachineAudit](https://godoc.org/github.com/joyent/gosdc/cloudapi#Client.MachineAudit), [StartMachine](https://godoc.org/github.com/joyent/gosdc/cloudapi#Client.StartMachine), [StartMachineFromSnapshot](https://godoc.org/github.com/joyent/gosdc/cloudapi#Client.StartMachineFromSnapshot), [StopMachine](https://godoc.org/github.com/joyent/gosdc/cloudapi#Client.StopMachine), [RebootMachine](https://godoc.org/github.com/joyent/gosdc/cloudapi#Client.RebootMachine) |
| Machine (Images) | [CreateImageFromMachine](https://godoc.org/github.com/joyent/gosdc/cloudapi#Client.CreateImageFromMachine) | [GetImage](https://godoc.org/github.com/joyent/gosdc/cloudapi#Client.GetImage), [ListImages](https://godoc.org/github.com/joyent/gosdc/cloudapi#Client.ListImages) | | [DeleteImage](https://godoc.org/github.com/joyent/gosdc/cloudapi#Client.DeleteImage) | [ExportImage](https://godoc.org/github.com/joyent/gosdc/cloudapi#Client.ExportImage) |
| Machine (Metadata) | | [GetMachineMetadata](https://godoc.org/github.com/joyent/gosdc/cloudapi#Client.GetMachineMetadata) | [UpdateMachineMetadata](https://godoc.org/github.com/joyent/gosdc/cloudapi#Client.UpdateMachineMetadata) | [DeleteMachineMetadata](https://godoc.org/github.com/joyent/gosdc/cloudapi#Client.DeleteMachineMetadata), [DeleteAllMachineMetadata](https://godoc.org/github.com/joyent/gosdc/cloudapi#Client.DeleteAllMachineMetadata) | |
| Machine (Snapshots) | [CreateMachineSnapshot](https://godoc.org/github.com/joyent/gosdc/cloudapi#Client.CreateMachineSnapshot) | [GetMachineSnapshot](https://godoc.org/github.com/joyent/gosdc/cloudapi#Client.GetMachineSnapshot), [ListMachineSnapshots](https://godoc.org/github.com/joyent/gosdc/cloudapi#Client.ListMachineSnapshots) | | [DeleteMachineSnapshot](https://godoc.org/github.com/joyent/gosdc/cloudapi#Client.DeleteMachineSnapshot) | |
| Machine (Tags) | | [GetMachineTag](https://godoc.org/github.com/joyent/gosdc/cloudapi#Client.GetMachineTag), [ListMachineTags](https://godoc.org/github.com/joyent/gosdc/cloudapi#Client.ListMachineTags) | [AddMachineTags](https://godoc.org/github.com/joyent/gosdc/cloudapi#Client.AddMachineTags), [ReplaceMachineTags](https://godoc.org/github.com/joyent/gosdc/cloudapi#Client.ReplaceMachineTags) | [DeleteMachineTag](https://godoc.org/github.com/joyent/gosdc/cloudapi#Client.DeleteMachineTag), [DeleteMachineTags](https://godoc.org/github.com/joyent/gosdc/cloudapi#Client.DeleteMachineTags) | [EnableFirewallMachine](https://godoc.org/github.com/joyent/gosdc/cloudapi#Client.EnableFirewallMachine), [DisableFirewallMachine](https://godoc.org/github.com/joyent/gosdc/cloudapi#Client.DisableFirewallMachine) |
| Networks | | [GetNetwork](https://godoc.org/github.com/joyent/gosdc/cloudapi#Client.GetNetwork), [ListNetworks](https://godoc.org/github.com/joyent/gosdc/cloudapi#Client.ListNetworks) | | | |
| Packages | | [GetPackage](https://godoc.org/github.com/joyent/gosdc/cloudapi#Client.GetPackage), [ListPackages](https://godoc.org/github.com/joyent/gosdc/cloudapi#Client.ListPackages) | | | |


## Contributing

Report bugs and request features using [GitHub Issues](https://github.com/joyent/gosdc/issues), or contribute code via a [GitHub Pull Request](https://github.com/joyent/gosdc/pulls). Changes will be code reviewed before merging. In the near future, automated tests will be run, but in the meantime please `go fmt`, `go lint`, and test all contributions.


## Developing

This library assumes a Go development environment setup based on [How to Write Go Code](https://golang.org/doc/code.html). Your GOPATH environment variable should be pointed at your workspace directory.

You can now use `go get github.com/joyent/gosdc` to install the repository to the correct location, but if you are intending on contributing back a change you may want to consider cloning the repository via git yourself. This way you can have a single source tree for all Joyent Go projects with each repo having two remotes -- your own fork on GitHub and the upstream origin.

For example if your GOPATH is `~/src/joyent/go` and you're working on multiple repos then that directory tree might look like:

```
~/src/joyent/go/
|_ pkg/
|_ src/
   |_ github.com
      |_ joyent
         |_ gocommon
         |_ gomanta
         |_ gosdc
         |_ gosign
```

### Recommended Setup

```
$ mkdir -p ${GOPATH}/src/github.com/joyent
$ cd ${GOPATH}/src/github.com/joyent
$ git clone git@github.com:<yourname>/gosdc.git

# fetch dependencies
$ git clone git@github.com:<yourname>/gocommon.git
$ git clone git@github.com:<yourname>/gosign.git
$ go get -v -t ./...

# add upstream remote
$ cd gosdc
$ git remote add upstream git@github.com:joyent/gosdc.git
$ git remote -v
origin  git@github.com:<yourname>/gosdc.git (fetch)
origin  git@github.com:<yourname>/gosdc.git (push)
upstream        git@github.com:joyent/gosdc.git (fetch)
upstream        git@github.com:joyent/gosdc.git (push)
```

### Run Tests

You can run the tests either locally or against live Triton. If you want to run the tests locally you'll want to generate an SSH key and pass the appropriate flags to the test harness as shown below.

```
cd ${GOPATH}/src/github.com/joyent/gosdc
ssh-keygen -b 2048 -C "Testing Key" -f test_key.id_rsa -t rsa -P ""
env KEY_NAME=`pwd`/test_key.id_rsa LIVE=false go test ./...
```

### Build the Library

```
cd ${GOPATH}/src/github.com/joyent/gosdc
go build ./...
```

## License

gosdc is licensed under the Mozilla Public License Version 2.0, a copy of which
is available at [LICENSE](LICENSE)
