// Copyright 2018-2025 JDCLOUD.COM
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package core

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"errors"
	"reflect"
	urllib "net/url"
)

var baseRequestFields []string

func init() {
	req := JDCloudRequest{}
	reqType := reflect.TypeOf(req)
	for i := 0; i < reqType.NumField(); i++ {
		baseRequestFields = append(baseRequestFields, reqType.Field(i).Name)
	}
}

type ParameterBuilder interface {
	BuildURL(url string, paramJson []byte) (string, error)
	BuildBody(paramJson []byte) (string, error)
}

func GetParameterBuilder(method string, logger Logger) ParameterBuilder {
	if method == MethodGet || method == MethodDelete || method == MethodHead {
		return &WithoutBodyBuilder{logger}
	} else {
		return &WithBodyBuilder{logger}
	}
}

// WithBodyBuilder supports PUT/POST/PATCH methods.
// It has path and body (json) parameters, but no query parameters.
type WithBodyBuilder struct {
	Logger Logger
}

func (b WithBodyBuilder) BuildURL(url string, paramJson []byte) (string, error) {
	paramMap := make(map[string]interface{})
	err := json.Unmarshal(paramJson, &paramMap)
	if err != nil {
		b.Logger.Log(LogError, err.Error())
		return "", err
	}

	replacedUrl, err := replaceUrlWithPathParam(url, paramMap)
	if err != nil {
		b.Logger.Log(LogError, err.Error())
		return "", err
	}

	encodedUrl, err := encodeUrl(replacedUrl, nil)
	if err != nil {
		return "", err
	}

	b.Logger.Log(LogInfo, "URL=" + encodedUrl)
	return encodedUrl, nil
}

func (b WithBodyBuilder) BuildBody(paramJson []byte) (string, error) {
	paramMap := make(map[string]interface{})
	err := json.Unmarshal(paramJson, &paramMap)
	if err != nil {
		b.Logger.Log(LogError, err.Error())
		return "", err
	}

	// remove base request fields
	for k := range paramMap {
		if includes(baseRequestFields, k) {
			delete(paramMap, k)
		}
	}

	body, _ := json.Marshal(paramMap)
	b.Logger.Log(LogInfo, "Body=", string(body))
	return string(body), nil
}

// WithoutBodyBuilder supports GET/DELETE methods.
// It only builds path and query parameters.
type WithoutBodyBuilder struct {
	Logger Logger
}

func (b WithoutBodyBuilder) BuildURL(url string, paramJson []byte) (string, error) {
	paramMap := make(map[string]interface{})
	err := json.Unmarshal(paramJson, &paramMap)
	if err != nil {
		b.Logger.Log(LogError, err.Error())
		return "", err
	}

	resultUrl, err := replaceUrlWithPathParam(url, paramMap)
	if err != nil {
		b.Logger.Log(LogError, err.Error())
		return "", err
	}

	queryParams := buildQueryParams(paramMap, url)
	encodedUrl, err := encodeUrl(resultUrl, queryParams)
	if err != nil {
		return "", err
	}

	b.Logger.Log(LogInfo, string(paramJson))
	b.Logger.Log(LogInfo, "URL=" + encodedUrl)
	return encodedUrl, nil
}

func (b WithoutBodyBuilder) BuildBody(paramJson []byte) (string, error) {
	return "", nil
}

func replaceUrlWithPathParam(url string, paramMap map[string]interface{}) (string, error) {
	r, _ := regexp.Compile("{[a-zA-Z0-9-_]+}")
	matches := r.FindAllString(url, -1)
	for _, match := range matches {
		field := strings.TrimLeft(match, "{")
		field = strings.TrimRight(field, "}")
		value, ok := paramMap[field]
		if !ok {
			return "", errors.New("Can not find path parameter: " + field)
		}

		valueStr := fmt.Sprintf("%v", value)
		url = strings.Replace(url, match, valueStr, -1)
	}

	return url, nil
}

func buildQueryParams(paramMap map[string]interface{}, url string) urllib.Values {
	values := urllib.Values{}
	accessMap(paramMap, url, "", values)
	return values
}

func accessMap(paramMap map[string]interface{}, url, prefix string, values urllib.Values) {
	for k, v := range paramMap {
		// exclude fields of JDCloudRequest class and path parameters
		if shouldIgnoreField(url, k) {
			continue
		}

		switch e := v.(type) {
		case []interface{}:
			for i, n := range e {
				switch f := n.(type) {
				case map[string]interface{}:
					subPrefix := fmt.Sprintf("%s.%d.", k, i+1)
					accessMap(f, url, subPrefix, values)
				case nil:
				default:
					values.Set(fmt.Sprintf("%s%s.%d", prefix, k, i+1), fmt.Sprintf("%s", n))
				}
			}
		case nil:
		default:
			values.Set(fmt.Sprintf("%s%s", prefix, k), fmt.Sprintf("%v", v))
		}
	}
}

func shouldIgnoreField(url, field string) bool {
	flag := "{" + field + "}"
	if strings.Contains(url, flag) {
		return true
	}

	if includes(baseRequestFields, field) {
		return true
	}

	return false
}

func encodeUrl(requestUrl string, values urllib.Values) (string, error) {
	urlObj, err := urllib.Parse(requestUrl)
	if err != nil {
		return "", err
	}

	urlObj.RawPath = EscapePath(urlObj.Path, false)
	uri := urlObj.EscapedPath()

	if values != nil {
		queryParam := values.Encode()
		// RFC 3986, ' ' should be encoded to 20%, '+' to 2B%
		queryParam = strings.Replace(queryParam, "+", "%20", -1)
		if queryParam != "" {
			uri += "?" + queryParam
		}
	}

	return uri, nil
}
