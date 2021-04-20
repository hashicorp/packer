package ncloud

import (
	"fmt"
	"log"
	"time"

	"github.com/NaverCloudPlatform/ncloud-sdk-go-v2/services/server"
	"github.com/NaverCloudPlatform/ncloud-sdk-go-v2/services/vserver"
)

func waiterClassicBlockStorageStatus(conn *NcloudAPIClient, blockStorageInstanceNo *string, status string, timeout time.Duration) error {
	reqParams := &server.GetBlockStorageInstanceListRequest{
		BlockStorageInstanceNoList: []*string{blockStorageInstanceNo},
	}

	c1 := make(chan error, 1)

	go func() {
		for {
			blockStorageInstanceList, err := conn.server.V2Api.GetBlockStorageInstanceList(reqParams)
			if err != nil {
				c1 <- err
				return
			}

			if status == "DETAC" && len(blockStorageInstanceList.BlockStorageInstanceList) == 0 {
				c1 <- nil
				return
			}

			blockStorageInstance := blockStorageInstanceList.BlockStorageInstanceList[0]
			code := blockStorageInstance.BlockStorageInstanceStatus.Code
			operationCode := blockStorageInstance.BlockStorageInstanceOperation.Code

			if *code == status && *operationCode == "NULL" {
				c1 <- nil
				return
			}

			log.Println(blockStorageInstance)
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

func waiterVpcBlockStorageStatus(conn *NcloudAPIClient, blockStorageInstanceNo *string, status string, timeout time.Duration) error {
	reqParams := &vserver.GetBlockStorageInstanceListRequest{
		BlockStorageInstanceNoList: []*string{blockStorageInstanceNo},
	}

	c1 := make(chan error, 1)

	go func() {
		for {
			blockStorageInstanceList, err := conn.vserver.V2Api.GetBlockStorageInstanceList(reqParams)
			if err != nil {
				c1 <- err
				return
			}

			if status == "DETAC" && len(blockStorageInstanceList.BlockStorageInstanceList) == 0 {
				c1 <- nil
				return
			}

			blockStorageInstance := blockStorageInstanceList.BlockStorageInstanceList[0]
			code := blockStorageInstance.BlockStorageInstanceStatus.Code
			operationCode := blockStorageInstance.BlockStorageInstanceOperation.Code

			if *code == status && *operationCode == "NULL" {
				c1 <- nil
				return
			}

			log.Println(blockStorageInstance)
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

func waiterClassicDetachedBlockStorage(conn *NcloudAPIClient, serverInstanceNo string, timeout time.Duration) error {
	reqParams := &server.GetBlockStorageInstanceListRequest{
		ServerInstanceNo: &serverInstanceNo,
	}
	c1 := make(chan error, 1)

	go func() {
		for {
			blockStorageInstanceList, err := conn.server.V2Api.GetBlockStorageInstanceList(reqParams)
			if err != nil {
				c1 <- err
				return
			}

			if *blockStorageInstanceList.TotalRows == 1 {
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

func waiterVpcDetachedBlockStorage(conn *NcloudAPIClient, serverInstanceNo string, timeout time.Duration) error {
	reqParams := &vserver.GetBlockStorageInstanceListRequest{
		ServerInstanceNo: &serverInstanceNo,
	}
	c1 := make(chan error, 1)

	go func() {
		for {
			blockStorageInstanceList, err := conn.vserver.V2Api.GetBlockStorageInstanceList(reqParams)
			if err != nil {
				c1 <- err
				return
			}

			if *blockStorageInstanceList.TotalRows == 1 {
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
