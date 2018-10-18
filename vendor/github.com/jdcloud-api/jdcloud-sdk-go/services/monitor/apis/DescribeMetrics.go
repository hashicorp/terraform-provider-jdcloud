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

package apis

import (
    "github.com/jdcloud-api/jdcloud-sdk-go/core"
    monitor "github.com/jdcloud-api/jdcloud-sdk-go/services/monitor/models"
)

type DescribeMetricsRequest struct {

    core.JDCloudRequest

    /* 资源的类型 ： 
vm-->云主机
disk-->云硬盘
ip-->公网ip
balance-->负载均衡
database-->云数据库mysql版本
cdn-->京东CDN
redis-->redis云缓存
mongodb-->mongoDB云缓存
storage-->云存储
sqlserver-->云数据库sqlserver版 
nativecontainer-->容器
  */
    ServiceCode string `json:"serviceCode"`
}

/*
 * param serviceCode: 资源的类型 ： 
vm-->云主机
disk-->云硬盘
ip-->公网ip
balance-->负载均衡
database-->云数据库mysql版本
cdn-->京东CDN
redis-->redis云缓存
mongodb-->mongoDB云缓存
storage-->云存储
sqlserver-->云数据库sqlserver版 
nativecontainer-->容器
 (Required)
 *
 * @Deprecated, not compatible when mandatory parameters changed
 */
func NewDescribeMetricsRequest(
    serviceCode string,
) *DescribeMetricsRequest {

	return &DescribeMetricsRequest{
        JDCloudRequest: core.JDCloudRequest{
			URL:     "/metrics",
			Method:  "GET",
			Header:  nil,
			Version: "v1",
		},
        ServiceCode: serviceCode,
	}
}

/*
 * param serviceCode: 资源的类型 ： 
vm-->云主机
disk-->云硬盘
ip-->公网ip
balance-->负载均衡
database-->云数据库mysql版本
cdn-->京东CDN
redis-->redis云缓存
mongodb-->mongoDB云缓存
storage-->云存储
sqlserver-->云数据库sqlserver版 
nativecontainer-->容器
 (Required)
 */
func NewDescribeMetricsRequestWithAllParams(
    serviceCode string,
) *DescribeMetricsRequest {

    return &DescribeMetricsRequest{
        JDCloudRequest: core.JDCloudRequest{
            URL:     "/metrics",
            Method:  "GET",
            Header:  nil,
            Version: "v1",
        },
        ServiceCode: serviceCode,
    }
}

/* This constructor has better compatible ability when API parameters changed */
func NewDescribeMetricsRequestWithoutParam() *DescribeMetricsRequest {

    return &DescribeMetricsRequest{
            JDCloudRequest: core.JDCloudRequest{
            URL:     "/metrics",
            Method:  "GET",
            Header:  nil,
            Version: "v1",
        },
    }
}

/* param serviceCode: 资源的类型 ： 
vm-->云主机
disk-->云硬盘
ip-->公网ip
balance-->负载均衡
database-->云数据库mysql版本
cdn-->京东CDN
redis-->redis云缓存
mongodb-->mongoDB云缓存
storage-->云存储
sqlserver-->云数据库sqlserver版 
nativecontainer-->容器
(Required) */
func (r *DescribeMetricsRequest) SetServiceCode(serviceCode string) {
    r.ServiceCode = serviceCode
}

// GetRegionId returns path parameter 'regionId' if exist,
// otherwise return empty string
func (r DescribeMetricsRequest) GetRegionId() string {
    return ""
}

type DescribeMetricsResponse struct {
    RequestID string `json:"requestId"`
    Error core.ErrorResponse `json:"error"`
    Result DescribeMetricsResult `json:"result"`
}

type DescribeMetricsResult struct {
    Metrics []monitor.MetricDetail `json:"metrics"`
}