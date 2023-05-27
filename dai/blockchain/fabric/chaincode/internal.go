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
	compositeKeyNamespace = "\x00"
	minUnicodeRuneValue   = 0

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
	return createCompositeKey(prefixFlTaskIndex, []string{taskID})
}

// packFlTaskListIndex pack index for saving tasks for requester, in time descending order 用于为请求者保存任务的包索引，按时间降序排列
func packFlTaskListIndex(task blockchain.FLTask) string {
	attributes := []string{fmt.Sprintf("%x", task.Requester), fmt.Sprintf("%d", subByInt64Max(task.PublishTime)), task.TaskID}
	return createCompositeKey(prefixFlTaskListIndex, attributes)
}

// packExecutorTaskListIndex pack index for saving tasks that an executor involves, in time descending order 用于保存执行程序涉及的任务的打包索引（按时间降序排列）
func packExecutorTaskListIndex(executor []byte, task blockchain.FLTask) string {
	attributes := []string{fmt.Sprintf("%x", executor), fmt.Sprintf("%d", subByInt64Max(task.PublishTime)), task.TaskID}
	return createCompositeKey(prefixFlTaskListIndex, attributes)
}

func packRequesterExecutorTaskIndex(executor []byte, task blockchain.FLTask) string {
	attributes := []string{fmt.Sprintf("%x", task.Requester), fmt.Sprintf("%x", executor), fmt.Sprintf("%d", subByInt64Max(task.PublishTime)), task.TaskID}
	return createCompositeKey(prefixREFlTaskListIndex, attributes)
}

// packFlTaskFilter pack filter index with public key for searching tasks for requester or executor 使用公钥打包筛选器索引，用于搜索请求者或执行者的任务
func packFlTaskFilter(rPubkey, ePubkey []byte) (string, []string) {
	if len(rPubkey) > 0 && len(ePubkey) > 0 {
		return prefixREFlTaskListIndex, []string{fmt.Sprintf("%x", rPubkey), fmt.Sprintf("%x", ePubkey)}
	} else if len(rPubkey) > 0 {
		return prefixFlTaskListIndex, []string{fmt.Sprintf("%x", rPubkey)}
	} else {
		return prefixFlTaskListIndex, []string{fmt.Sprintf("%x", ePubkey)}
	}
}

// packNodeIndex pack index-id contract key for saving executor node 打包索引 ID 合约密钥以保存执行器节点
func packNodeIndex(nodeID []byte) string {
	return createCompositeKey(prefixNodeIndex, []string{fmt.Sprintf("%x", nodeID)})
}

// packNodeStringIndex pack index-id contract key for saving executor node 打包索引 ID 合约密钥以保存执行器节点
func packNodeStringIndex(nodeID string) string {
	return createCompositeKey(prefixNodeIndex, []string{nodeID})
}

// packNodeNameIndex pack index-name contract key for saving executor node 打包索引名协定密钥以保存执行器节点
func packNodeNameIndex(name string) string {
	return createCompositeKey(prefixNodeNameIndex, []string{name})
}

// packNodeListIndex pack filter for listing executor nodes 用于列出执行程序节点的包筛选器
func packNodeListIndex(node blockchain.ExecutorNode) string {
	attributes := []string{fmt.Sprintf("%d", subByInt64Max(node.RegTime)), fmt.Sprintf("%x", node.ID)}
	return createCompositeKey(prefixNodeListIndex, attributes)
}

func createCompositeKey(objectType string, attributes []string) string {
	ck := compositeKeyNamespace + objectType + string(minUnicodeRuneValue)
	for _, att := range attributes {
		ck += att + string(minUnicodeRuneValue)
	}
	return ck
}
