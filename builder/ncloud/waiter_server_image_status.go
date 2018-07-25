package ncloud

import (
	"fmt"
	"log"
	"time"

	ncloud "github.com/NaverCloudPlatform/ncloud-sdk-go/sdk"
)

func waiterMemberServerImageStatus(conn *ncloud.Conn, memberServerImageNo string, status string, timeout time.Duration) error {
	reqParams := new(ncloud.RequestServerImageList)
	reqParams.MemberServerImageNoList = []string{memberServerImageNo}

	c1 := make(chan error, 1)

	go func() {
		for {
			memberServerImageList, err := conn.GetMemberServerImageList(reqParams)
			if err != nil {
				c1 <- err
				return
			}

			code := memberServerImageList.MemberServerImageList[0].MemberServerImageStatus.Code
			if code == status {
				c1 <- nil
				return
			}

			log.Printf("Status of member server image [%s] is %s\n", memberServerImageNo, code)
			log.Println(memberServerImageList.MemberServerImageList[0])
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
