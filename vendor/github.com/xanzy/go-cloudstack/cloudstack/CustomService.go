//
// Copyright 2018, Sander van Harmelen
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
//

package cloudstack

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"strings"
)

type CustomServiceParams struct {
	p map[string]interface{}
}

func (p *CustomServiceParams) toURLValues() url.Values {
	u := url.Values{}
	if p.p == nil {
		return u
	}

	for k, v := range p.p {
		switch t := v.(type) {
		case bool:
			u.Set(k, strconv.FormatBool(t))
		case int:
			u.Set(k, strconv.Itoa(t))
		case int64:
			vv := strconv.FormatInt(t, 10)
			u.Set(k, vv)
		case string:
			u.Set(k, t)
		case []string:
			u.Set(k, strings.Join(t, ", "))
		case map[string]string:
			i := 0
			for kk, vv := range t {
				u.Set(fmt.Sprintf("%s[%d].%s", k, i, kk), vv)
				i++
			}
		}
	}

	return u
}

func (p *CustomServiceParams) SetParam(param string, v interface{}) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p[param] = v
	return
}

func (s *CustomService) CustomRequest(api string, p *CustomServiceParams, result interface{}) error {
	resp, err := s.cs.newRequest(api, p.toURLValues())
	if err != nil {
		return err
	}

	return json.Unmarshal(resp, result)
}
