package common

import "encoding/xml"

type CommonResponse struct {
	RequestID     string `xml:"requestId"`
	ReturnCode    int    `xml:"returnCode"`
	ReturnMessage string `xml:"returnMessage"`
}

type ResponseError struct {
	ResponseError xml.Name `xml:"responseError"`
	CommonResponse
}
