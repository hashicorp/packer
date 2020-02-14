package ncloud

import (
	"fmt"
	"log"
	"time"

	"github.com/NaverCloudPlatform/ncloud-sdk-go-v2/services/server"
)

func waiterServerInstanceStatus(conn *NcloudAPIClient, serverInstanceNo string, status string, timeout time.Duration) error {
	reqParams := new(server.GetServerInstanceListRequest)
	reqParams.ServerInstanceNoList = []*string{&serverInstanceNo}

	c1 := make(chan error, 1)

	go func() {
		for {
			serverInstanceList, err := conn.server.V2Api.GetServerInstanceList(reqParams)
			if err != nil {
				c1 <- err
				return
			}

			code := serverInstanceList.ServerInstanceList[0].ServerInstanceStatus.Code
			if *code == status {
				c1 <- nil
				return
			}

			log.Printf("Status of serverInstanceNo [%s] is %s\n", serverInstanceNo, *code)
			log.Println(serverInstanceList.ServerInstanceList[0])
			time.Sleep(time.Second * 5)
		}
	}()

	select {
	case res := <-c1:
		return res
	case <-time.After(timeout):
		return fmt.Errorf("TIMEOUT : server instance status is not changed into status %s", status)
	}
}
