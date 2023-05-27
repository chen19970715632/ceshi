// Copyright (c) 2021 PaddlePaddle Authors. All Rights Reserved.
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"fmt"
	"math"

	"github.com/PaddlePaddle/PaddleDTX/dai/blockchain"
)

const (
	prefixFlTaskIndex       = "index_fltask"
	prefixFlTaskListIndex   = "index_fltask_list"
	prefixREFlTaskListIndex = "index_re_fltask_list"
	prefixNodeIndex         = "index_executor_node"
	prefixNodeNameIndex     = "index_executor_name"
	prefixNodeListIndex     = "index_executor_node_list"
)

// subByInt64Max return maxInt64 - N 返回 maxInt64 - N
func subByInt64Max(n int64) int64 {
	return math.MaxInt64 - n
}

// packFlTaskIndex pack task index for saving task using taskID 打包任务索引以使用任务 ID 保存任务
func packFlTaskIndex(taskID string) string {
	return fmt.Sprintf("%s/%s", prefixFlTaskIndex, taskID)
}

// packFlTaskListIndex pack index for saving tasks for requester, in time descending order 用于为请求者保存任务的包索引，按时间降序排列
func packFlTaskListIndex(task blockchain.FLTask) string {
	return fmt.Sprintf("%s/%x/%d/%s", prefixFlTaskListIndex, task.Requester, subByInt64Max(task.PublishTime), task.TaskID)
}

// packExecutorTaskListIndex pack index for saving tasks that an executor involves, in time descending order 用于保存执行程序涉及的任务的打包索引（按时间降序排列）
func packExecutorTaskListIndex(executor []byte, task blockchain.FLTask) string {
	return fmt.Sprintf("%s/%x/%d/%s", prefixFlTaskListIndex, executor, subByInt64Max(task.PublishTime), task.TaskID)
}

// packRequesterExecutorTaskIndex pack index for saving tasks for requester and executor, in time descending order 用于保存请求者和执行者任务的包索引（按时间降序排列）
func packRequesterExecutorTaskIndex(executor []byte, task blockchain.FLTask) string {
	return fmt.Sprintf("%s/%x/%x/%d/%s", prefixREFlTaskListIndex, task.Requester, executor, subByInt64Max(task.PublishTime), task.TaskID)
}

// packFlTaskFilter pack filter index with public key for searching tasks for requester or executor 使用公钥打包筛选器索引，用于搜索请求者或执行者的任务
func packFlTaskFilter(rPubkey, ePubkey []byte) string {
	// If requester and executor public key not empty 如果请求者和执行者公钥不为空
	filter := prefixFlTaskListIndex + "/"
	if len(rPubkey) > 0 && len(ePubkey) > 0 {
		filter = prefixREFlTaskListIndex + "/" + fmt.Sprintf("%x/%x", rPubkey, ePubkey)
	} else if len(rPubkey) > 0 {
		filter += fmt.Sprintf("%x/", rPubkey)
	} else {
		filter += fmt.Sprintf("%x/", ePubkey)
	}
	return filter
}

// packNodeIndex pack index-id contract key for saving executor node 打包索引 ID 合约密钥以保存执行器节点
func packNodeIndex(nodeID []byte) string {
	return fmt.Sprintf("%s/%x", prefixNodeIndex, nodeID)
}

// packNodeStringIndex pack index-id contract key for saving executor node 打包索引 ID 合约密钥以保存执行器节点
func packNodeStringIndex(nodeID string) string {
	return fmt.Sprintf("%s/%s", prefixNodeIndex, nodeID)
}

// packNodeNameIndex pack index-name contract key for saving executor node 打包索引名协定密钥以保存执行器节点
func packNodeNameIndex(name string) string {
	return fmt.Sprintf("%s/%s", prefixNodeNameIndex, name)
}

// packNodeListIndex pack filter for listing executor nodes 用于列出执行程序节点的包筛选器
func packNodeListIndex(node blockchain.ExecutorNode) string {
	return fmt.Sprintf("%s/%d/%x", prefixNodeListIndex, subByInt64Max(node.RegTime), node.ID)
}
