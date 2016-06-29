# Go SDK

The ProfitBricks Client Library for [Go](https://www.golang.org/) provides you with access to the ProfitBricks REST API. It is designed for developers who are building applications in Go.

This guide will walk you through getting setup with the library and performing various actions against the API.

# Table of Contents
* [Concepts](#concepts)
* [Getting Started](#getting-started)
* [Installation](#installation)
* [How to: Create Data Center](#how-to-create-data-center)
* [How to: Delete Data Center](#how-to-delete-data-center)
* [How to: Create Server](#how-to-create-server)
* [How to: List Available Images](#how-to-list-available-images)
* [How to: Create Storage Volume](#how-to-create-storage-volume)
* [How to: Update Cores and Memory](#how-to-update-cores-and-memory)
* [How to: Attach or Detach Storage Volume](#how-to-attach-or-detach-storage-volume)
* [How to: List Servers, Volumes, and Data Centers](#how-to-list-servers-volumes-and-data-centers)
* [Example](#example)
* [Return Types](#return-types)
* [Support](#support)


# Concepts

The Go SDK wraps the latest version of the ProfitBricks REST API. All API operations are performed over SSL and authenticated using your ProfitBricks portal credentials. The API can be accessed within an instance running in ProfitBricks or directly over the Internet from any application that can send an HTTPS request and receive an HTTPS response.

# Getting Started

Before you begin you will need to have [signed-up](https://www.profitbricks.com/signup) for a ProfitBricks account. The credentials you setup during sign-up will be used to authenticate against the API.

Install the Go language from: [Go Installation](https://golang.org/doc/install)

The `GOPATH` environment variable specifies the location of your Go workspace. It is likely the only environment variable you'll need to set when developing Go code. This is an example of pointing to a workspace configured underneath your home directory:

```
mkdir -p ~/go/bin
export GOPATH=~/go
export GOBIN=$GOPATH/bin
export PATH=$PATH:$GOBIN
```

# Installation

The following go command will download `profitbricks-sdk-go` to your configured `GOPATH`:

```go
go get "github.com/profitbricks/profitbricks-sdk-go"
```

The source code of the package will be located at:

	$GOBIN\src\profitbricks-sdk-go

Create main package file *example.go*:

```go
package main

import (
	"fmt"
)

func main() {
}
```

Import GO SDK:

```go
import(
	"github.com/profitbricks/profitbricks-sdk-go"
)
```

Add your credentials for connecting to ProfitBricks:

```go
profitbricks.SetAuth("username", "password")
```

Set depth:

```go
profitbricks.SetDepth("5")
```

Depth controls the amount of data returned from the REST server ( range 1-5 ). The larger the number the more information is returned from the server. This is especially useful if you are looking for the information in the nested objects.

**Caution**: You will want to ensure you follow security best practices when using credentials within your code or stored in a file.

# How To's

## How To: Create Data Center

ProfitBricks introduces the concept of Data Centers. These are logically separated from one another and allow you to have a self-contained environment for all servers, volumes, networking, snapshots, and so forth. The goal is to give you the same experience as you would have if you were running your own physical data center.

The following code example shows you how to programmatically create a data center:

```go
request := profitbricks.CreateDatacenterRequest{
    DCProperties: profitbricks.DCProperties{
			Name:        "test",
			Description: "description",
			Location:    "us/lasdev",
	},
}

response := profitbricks.CreateDatacenter(request)
```

## How To: Delete Data Center

You will want to exercise a bit of caution here. Removing a data center will destroy all objects contained within that data center -- servers, volumes, snapshots, and so on.

The code to remove a data center is as follows. This example assumes you want to remove previously data center:

```go
profitbricks.DeleteDatacenter(response.Id)
```

## How To: Create Server

The server create method has a list of required parameters followed by a hash of optional parameters. The optional parameters are specified within the "options" hash and the variable names match the [REST API](https://devops.profitbricks.com/api/rest/) parameters.

The following example shows you how to create a new server in the data center created above:

```go
request = CreateServerRequest{
	ServerProperties: ServerProperties{
		Name:  "go01",
		Ram:   1024,
		Cores: 2,
	},
}

server := CreateServer(datacenter.Id, req)
```

## How To: List Available Images

A list of disk and ISO images are available from ProfitBricks for immediate use. These can be easily viewed and selected. The following shows you how to get a list of images. This list represents both CDROM images and HDD images.

```go
images := profitbricks.ListImages()
```

This will return a [collection](#Collection) object

## How To: Create Storage Volume

ProfitBricks allows for the creation of multiple storage volumes that can be attached and detached as needed. It is useful to attach an image when creating a storage volume. The storage size is in gigabytes.

```go
volumerequest := CreateVolumeRequest{
	VolumeProperties: VolumeProperties{
		Size:        1,
		Name:        "Volume Test",
		ImageId: "imageid",
		Type: "HDD",
		SshKey: []string{"hQGOEJeFL91EG3+l9TtRbWNjzhDVHeLuL3NWee6bekA="},
	},
}

storage := CreateVolume(datacenter.Id, volumerequest)
```

## How To: Update Cores and Memory

ProfitBricks allows users to dynamically update cores, memory, and disk independently of each other. This removes the restriction of needing to upgrade to the next size available size to receive an increase in memory. You can now simply increase the instances memory keeping your costs in-line with your resource needs.

Note: The memory parameter value must be a multiple of 256, e.g. 256, 512, 768, 1024, and so forth.

The following code illustrates how you can update cores and memory:

```go
serverupdaterequest := profitbricks.ServerProperties{
	Cores: 1,
	Ram:   256,
}

resp := PatchServer(datacenter.Id, server.Id, serverupdaterequest)
```

## How To: Attach or Detach Storage Volume

ProfitBricks allows for the creation of multiple storage volumes. You can detach and reattach these on the fly. This allows for various scenarios such as re-attaching a failed OS disk to another server for possible recovery or moving a volume to another location and spinning it up.

The following illustrates how you would attach and detach a volume and CDROM to/from a server:

```go
profitbricks.AttachVolume(datacenter.Id, server.Id, volume.Id)
profitbricks.AttachCdrom(datacenter.Id, server.Id, images.Items[0].Id)

profitbricks.DetachVolume(datacenter.Id, server.Id, volume.Id)
profitbricks.DetachCdrom(datacenter.Id, server.Id, images.Items[0].Id)
```

## How To: List Servers, Volumes, and Data Centers

Go SDK provides standard functions for retrieving a list of volumes, servers, and datacenters.

The following code illustrates how to pull these three list types:

```go
volumes := profitbricks.ListVolumes(datacenter.Id)

servers := profitbricks.ListServers(datacenter.Id)

datacenters := profitbricks.ListDatacenters()
```

## Example

```go
package main

import (
	"fmt"
	"github.com/profitbricks/profitbricks-sdk-go"
)

func main() {

	//Sets username and password
	profitbricks.SetAuth("username", "password")
	//Sets depth.
	profitbricks.SetDepth(5)

	dcrequest := profitbricks.CreateDatacenterRequest{
		DCProperties: profitbricks.DCProperties{
			Name:        "test",
			Description: "description",
			Location:    "us/lasdev",
		},
	}

	datacenter := profitbricks.CreateDatacenter(dcrequest)

	serverrequest := profitbricks.CreateServerRequest{
		ServerProperties: profitbricks.ServerProperties{
			Name:  "go01",
			Ram:   1024,
			Cores: 2,
		},
	}

    server := profitbricks.CreateServer(datacenter.Id, serverrequest)

	images := profitbricks.ListImages()

	fmt.Println(images.Items)

	volumerequest := profitbricks.CreateVolumeRequest{
		VolumeProperties: profitbricks.VolumeProperties{
			Size:        1,
			Name:        "Volume Test",
			LicenceType: "LINUX",
		},
	}

	storage := profitbricks.CreateVolume(datacenter.Id, volumerequest)

	serverupdaterequest := profitbricks.ServerProperties{
		Name:  "go01renamed",
		Cores: 1,
		Ram:   256,
	}

    resp := profitbricks.PatchServer(datacenter.Id, server.Id, serverupdaterequest)

	volumes := profitbricks.ListVolumes(datacenter.Id)
	servers := profitbricks.ListServers(datacenter.Id)
	datacenters := profitbricks.ListDatacenters()

	profitbricks.DeleteServer(datacenter.Id, server.Id)
	profitbricks.DeleteDatacenter(datacenter.Id)
}
```

# Return Types

## Resp struct
* 	Resp is the struct returned by all REST request functions

```go
type Resp struct {
    Req        *http.Request
    StatusCode int
    Headers    http.Header
    Body       []byte
}
```

## Instance struct
* 	`Get`, `Create`, and `Patch` functions all return an Instance struct.
*	A Resp struct is embedded in the Instance struct.
*	The raw server response is available as `Instance.Resp.Body`.

```go
type Instance struct {
    Id_Type_Href
    MetaData   StringMap           `json:"metaData"`
    Properties StringIfaceMap      `json:"properties"`
    Entities   StringCollectionMap `json:"entities"`
    Resp       Resp                `json:"-"`
}
```

## Collection struct
* 	Collection Structs contain Instance arrays. 
* 	List functions return Collections

```go
type Collection struct {
    Id_Type_Href
    Items []Instance `json:"items,omitempty"`
    Resp  Resp       `json:"-"`
}
```

# Support
You are welcome to contact us with questions or comments at [ProfitBricks DevOps Central](https://devops.profitbricks.com/). Please report any issues via [GitHub's issue tracker](https://github.com/profitbricks/profitbricks-sdk-go/issues).