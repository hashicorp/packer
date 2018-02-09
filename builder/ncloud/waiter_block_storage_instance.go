package ncloud

import (
	"fmt"
	"log"
	"time"

	ncloud "github.com/NaverCloudPlatform/ncloud-sdk-go/sdk"
)

func waiterBlockStorageInstanceStatus(conn *ncloud.Conn, blockStorageInstanceNo string, status string, timeout time.Duration) error {
	reqParams := new(ncloud.RequestBlockStorageInstanceList)
	reqParams.BlockStorageInstanceNoList = []string{blockStorageInstanceNo}

	c1 := make(chan error, 1)

	go func() {
		for {
			blockStorageInstanceList, err := conn.GetBlockStorageInstance(reqParams)
			if err != nil {
				c1 <- err
				return
			}

			if status == "DETAC" && len(blockStorageInstanceList.BlockStorageInstance) == 0 {
				c1 <- nil
				return
			}

			code := blockStorageInstanceList.BlockStorageInstance[0].BlockStorageInstanceStatus.Code
			operationCode := blockStorageInstanceList.BlockStorageInstance[0].BlockStorageInstanceOperation.Code

			if code == status && operationCode == "NULL" {
				c1 <- nil
				return
			}

			log.Println(blockStorageInstanceList.BlockStorageInstance[0])
			time.Sleep(time.Second * 5)
		}
	}()

	select {
	case res := <-c1:
		return res
	case <-time.After(timeout):
		return fmt.Errorf("TIMEOUT : block storage instance status is not changed into status %s", status)
	}
}

func waiterDetachedBlockStorageInstance(conn *ncloud.Conn, serverInstanceNo string, timeout time.Duration) error {
	reqParams := new(ncloud.RequestBlockStorageInstanceList)
	reqParams.ServerInstanceNo = serverInstanceNo

	c1 := make(chan error, 1)

	go func() {
		for {
			blockStorageInstanceList, err := conn.GetBlockStorageInstance(reqParams)
			if err != nil {
				c1 <- err
				return
			}

			if blockStorageInstanceList.TotalRows == 1 {
				c1 <- nil
				return
			}

			time.Sleep(time.Second * 5)
		}
	}()

	select {
	case res := <-c1:
		return res
	case <-time.After(timeout):
		return fmt.Errorf("TIMEOUT : attached block storage instance is not detached")
	}
}
