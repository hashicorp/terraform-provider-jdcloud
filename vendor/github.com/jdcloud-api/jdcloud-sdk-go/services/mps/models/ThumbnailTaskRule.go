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


type ThumbnailTaskRule struct {

    /* 截图模式 单张: single 多张: multi 平均: average default: single (Optional) */
    Mode *string `json:"mode"`

    /* 是否开启关键帧截图 default: true (Optional) */
    KeyFrame *bool `json:"keyFrame"`

    /* 生成截图的开始时间, mode=average 时不可选. default:0 (Optional) */
    StartTimeInSecond *int `json:"startTimeInSecond"`

    /* 生成截图的结束时间, mode=single/average时不可选, 且不得小于startTimeInSecond. default:-1(代表视频时长) (Optional) */
    EndTimeInSecond *int `json:"endTimeInSecond"`

    /* 截图数量, mode=single时不可选. default:1 (Optional) */
    Count *int `json:"count"`
}
