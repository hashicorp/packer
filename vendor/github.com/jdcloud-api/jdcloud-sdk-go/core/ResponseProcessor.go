package core

import (
	"net/http"
	"io/ioutil"
	"errors"
	"fmt"
)

type ResponseProcessor interface {
	Process(response *http.Response) ([]byte, error)
}

func GetResponseProcessor(method string) ResponseProcessor {
	if method == MethodHead {
		return &WithoutBodyResponseProcessor{}
	} else {
		return &WithBodyResponseProcessor{}
	}
}

type WithBodyResponseProcessor struct {
}

func (p WithBodyResponseProcessor) Process(response *http.Response) ([]byte, error) {
	defer response.Body.Close()
	return ioutil.ReadAll(response.Body)
}

type WithoutBodyResponseProcessor struct {
}

func (p WithoutBodyResponseProcessor) Process(response *http.Response) ([]byte, error) {
	requestId := response.Header.Get(HeaderJdcloudRequestId)
	if requestId != "" {
		return []byte(fmt.Sprintf(`{"requestId":"%s"}`, requestId)), nil
	}

	return nil, errors.New("can not get requestId in HEAD response")
}