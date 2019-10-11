package ncloud

import (
	"fmt"
	"log"
	"time"

	ncloud "github.com/NaverCloudPlatform/ncloud-sdk-go/sdk"
)

func waiterServerInstanceStatus(conn *ncloud.Conn, serverInstanceNo string, status string, timeout time.Duration) error {
	reqParams := new(ncloud.RequestGetServerInstanceList)
	reqParams.ServerInstanceNoList = []string{serverInstanceNo}

	c1 := make(chan error, 1)

	go func() {
		for {
			serverInstanceList, err := conn.GetServerInstanceList(reqParams)
			if err != nil {
				c1 <- err
				return
			}

			code := serverInstanceList.ServerInstanceList[0].ServerInstanceStatus.Code
			if code == status {
				c1 <- nil
				return
			}

			log.Printf("Status of serverInstanceNo [%s] is %s\n", serverInstanceNo, code)
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
