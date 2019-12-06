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

// Credential is used to sign the request,
// AccessKey and SecretKey could be found in JDCloud console
type Credential struct {
	AccessKey string
	SecretKey string
}

func NewCredentials(accessKey, secretKey string) *Credential {
	return &Credential{accessKey, secretKey}
}
