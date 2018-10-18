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


type DwInstance struct {

    /* 实例名  */
    InstanceName string `json:"instanceName"`

    /* 实例描述 (Optional) */
    Comments string `json:"comments"`

    /* 实例属主 (Optional) */
    InstanceOwnerName string `json:"instanceOwnerName"`

    /* 实例所属区域 (Optional) */
    Area string `json:"area"`

    /* 实例别名（在页面显示） (Optional) */
    Uname string `json:"uname"`

    /* 实例创建时间 (Optional) */
    CreateTime string `json:"createTime"`
}