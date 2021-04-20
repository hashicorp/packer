package ncloud

import (
	"fmt"
	"log"
	"time"

	"github.com/NaverCloudPlatform/ncloud-sdk-go-v2/services/server"
	"github.com/NaverCloudPlatform/ncloud-sdk-go-v2/services/vserver"
)

func waiterClassicMemberServerImageStatus(conn *NcloudAPIClient, memberServerImageNo string, status string, timeout time.Duration) error {
	reqParams := &server.GetMemberServerImageListRequest{
		MemberServerImageNoList: []*string{&memberServerImageNo},
	}

	c1 := make(chan error, 1)

	go func() {
		for {
			memberServerImageList, err := conn.server.V2Api.GetMemberServerImageList(reqParams)
			if err != nil {
				c1 <- err
				return
			}

			code := memberServerImageList.MemberServerImageList[0].MemberServerImageStatus.Code
			if *code == status {
				c1 <- nil
				return
			}

			log.Printf("Status of member server image [%s] is %s\n", memberServerImageNo, *code)
			time.Sleep(time.Second * 5)
		}
	}()

	select {
	case res := <-c1:
		return res
	case <-time.After(timeout):
		return fmt.Errorf("TIMEOUT : member server image status is not changed into status %s", status)
	}
}

func waiterVpcMemberServerImageStatus(conn *NcloudAPIClient, memberServerImageNo string, status string, timeout time.Duration) error {
	reqParams := &vserver.GetMemberServerImageInstanceDetailRequest{
		MemberServerImageInstanceNo: &memberServerImageNo,
	}

	c1 := make(chan error, 1)

	go func() {
		for {
			memberServerImageList, err := conn.vserver.V2Api.GetMemberServerImageInstanceDetail(reqParams)
			if err != nil {
				c1 <- err
				return
			}

			code := memberServerImageList.MemberServerImageInstanceList[0].MemberServerImageInstanceStatus.Code
			if *code == status {
				c1 <- nil
				return
			}

			log.Printf("Status of member server image [%s] is %s\n", memberServerImageNo, *code)
			time.Sleep(time.Second * 5)
		}
	}()

	select {
	case res := <-c1:
		return res
	case <-time.After(timeout):
		return fmt.Errorf("TIMEOUT : member server image status is not changed into status %s", status)
	}
}
