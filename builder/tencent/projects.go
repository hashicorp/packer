package tencent

import (
	"fmt"
	"log"
	"strings"
)

type (
	AcctDescribeProject struct {
		ProjectName string
		ProjectId   string
		CreateTime  string
		CreatorUin  string
		ProjectInfo string
	}

	AcctDescribeProjectResponse struct {
		Code      int
		Message   string
		CodeDesc  string
		Data      []map[string]interface{}
		Error     CVMError
		RequestId string
	}
)

func DescribeProject(c *Config) {
	configInfo := c.CreateVMmap()
	extraParams := c.Keys()
	// Require only Region, so remove unnecessary keys
	for k, _ := range extraParams {
		switch k {
		case CRegion, CSecretId, CSecretKey, CTimestamp, CVersion:
			continue
		default:
			delete(extraParams, k)
		}
	}

	rawresponse := AcctAPICall("DescribeProject", configInfo, extraParams)
	// !!! **** wrap into expected format, if this is not the expected format, it'll be unable to parse
	// so a subtle error will occur, but won't be noticeable !!! ***
	response := []byte(fmt.Sprintf(`{"Response": %s}`, rawresponse))
	var describeProjectResponse AcctDescribeProjectResponse
	DecodeResponse(response, &describeProjectResponse)
	if c.PackerDebug || CloudAPIDebug {
		msg := fmt.Sprintf("DescribeProkect configInfo: %+v\n", configInfo)
		msg = strings.Replace(msg, c.SecretKey, COBFUSCATED, -1)
		log.Print(msg)

		msg = fmt.Sprintf("DescribeProject extraParams: %+v\n", extraParams)
		msg = strings.Replace(msg, c.SecretKey, COBFUSCATED, -1)
		log.Print(msg)

		log.Printf("DescribeProject Raw response: %s\n", string(rawresponse))
		log.Printf("DescribeProject Wrapped response: %s\n", string(response))
		log.Printf("DescribeProject Decoded: %+v\n", describeProjectResponse)
	}
}
