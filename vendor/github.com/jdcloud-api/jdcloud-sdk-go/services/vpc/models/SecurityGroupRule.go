// Copyright 2018 JDCLOUD.COM
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
// NOTE: This class is auto generated by the jdcloud code generator program.

package models


type SecurityGroupRule struct {

    /* 安全组规则ID (Optional) */
    RuleId string `json:"ruleId"`

    /* 安全组规则方向。0：入规则; 1：出规则 (Optional) */
    Direction int `json:"direction"`

    /* 规则限定协议。300:All; 6:TCP; 17:UDP; 1:ICMP (Optional) */
    Protocol int `json:"protocol"`

    /* 匹配地址前缀 (Optional) */
    AddressPrefix string `json:"addressPrefix"`

    /* 匹配地址协议版本。4：IPv4 (Optional) */
    IpVersion int `json:"ipVersion"`

    /* 规则限定起始传输层端口, 默认1 ，若protocal不是传输层协议，恒为0 (Optional) */
    FromPort int `json:"fromPort"`

    /* 规则限定终止传输层端口, 默认1 ，若protocal不是传输层协议，恒为0 (Optional) */
    ToPort int `json:"toPort"`

    /* 安全组规则创建时间 (Optional) */
    CreatedTime string `json:"createdTime"`

    /* 描述,​ 允许输入UTF-8编码下的全部字符，不超过256字符 (Optional) */
    Description string `json:"description"`
}
