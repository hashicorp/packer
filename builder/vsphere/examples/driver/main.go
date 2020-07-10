package main

import (
	"context"
	"fmt"

	"github.com/hashicorp/packer/builder/vsphere/driver"
)

func main() {
	d, err := driver.NewDriver(context.TODO(), &driver.ConnectConfig{
		VCenterServer:      "vcenter.vsphere65.test",
		Username:           "root",
		Password:           "jetbrains",
		InsecureConnection: true,
	})
	if err != nil {
		panic(err)
	}

	ds, err := d.FindDatastore("", "esxi-1.vsphere65.test")
	if err != nil {
		panic(err)
	}

	fmt.Println(ds.Name())
}
