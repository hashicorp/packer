package profitbricks

import "fmt"

func mkdcid(name string) string {
	request := CreateDatacenterRequest{
		DCProperties: DCProperties{
			Name:        name,
			Description: "description",
			Location:    "us/las",
		},
	}
	dc := CreateDatacenter(request)
	fmt.Println("===========================")
	fmt.Println("Created a DC " + name)
	fmt.Println("Created a DC id " + dc.Id)
	fmt.Println(dc.Resp.StatusCode)
	fmt.Println("===========================")
	return dc.Id
}

func mklocid() string {
	resp := ListLocations()

	locid := resp.Items[0].Id
	return locid
}

func mksrvid(srv_dcid string) string {
	var req = CreateServerRequest{
		ServerProperties: ServerProperties{
			Name:  "GO SDK test",
			Ram:   1024,
			Cores: 2,
		},
	}
	srv := CreateServer(srv_dcid, req)
	fmt.Println("===========================")
	fmt.Println("Created a server " + srv.Id)
	fmt.Println(srv.Resp.StatusCode)
	fmt.Println("===========================")
	return srv.Id
}

func mknic(lbal_dcid, serverid string) string {
	var request = NicCreateRequest{
		NicProperties{
			Name: "GO SDK Original Nic",
			Lan:  "1",
		},
	}

	resp := CreateNic(lbal_dcid, serverid, request)
	fmt.Println("===========================")
	fmt.Println("created a nic at server " + serverid)

	fmt.Println("created a nic with id " + resp.Id)
	fmt.Println(resp.Resp.StatusCode)
	fmt.Println("===========================")
	return resp.Id
}
