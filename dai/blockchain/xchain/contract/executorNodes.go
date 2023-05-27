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
	"encoding/json"
	"regexp"

	"github.com/PaddlePaddle/PaddleDTX/xdb/errorx"
	util "github.com/PaddlePaddle/PaddleDTX/xdb/pkgs/strings"
	"github.com/xuperchain/xuperchain/core/contractsdk/go/code"

	"github.com/PaddlePaddle/PaddleDTX/dai/blockchain"
)

// RegisterExecutorNode registers Executor node 注册执行器节点
func (x *Xdata) RegisterExecutorNode(ctx code.Context) code.Response {
	var opt blockchain.AddNodeOptions
	// get opt 获取选项
	p, ok := ctx.Args()["opt"]
	if !ok {
		return code.Error(errorx.New(errorx.ErrCodeParam, "missing param:node"))
	}
	// unmarshal opt 解组选项
	if err := json.Unmarshal(p, &opt); err != nil {
		return code.Error(errorx.NewCode(err, errorx.ErrCodeInternal,
			"fail to unmarshal AddNodeOptions"))
	}
	// get node 获取节点
	node := opt.Node
	// marshal node 元组节点
	s, err := json.Marshal(node)
	if err != nil {
		return code.Error(errorx.NewCode(err, errorx.ErrCodeInternal,
			"fail to marshal Node"))
	}
	// check node.Name, the length of the executor node name is 4-16 characters
	// only support lowercase letters and numbers
	//检查节点。名称，执行节点名称的长度为4-16个字符
	//	仅支持小写字母和数字
	if ok, _ := regexp.MatchString("^[a-z0-9]{4,16}", node.Name); !ok {
		return code.Error(errorx.New(errorx.ErrCodeParam,
			"bad param, nodeName only supports numbers and lowercase letters with a length of 4-16"))
	}

	// get the message to sign 获取要签名的消息
	msg, err := util.GetSigMessage(opt)
	if err != nil {
		return code.Error(errorx.Internal(err, "failed to get the message to sign"))
	}
	// verify sig 验证签名
	if err := x.checkSign(opt.Signature, node.ID, []byte(msg)); err != nil {
		return code.Error(err)
	}

	// put index-node on xchain, judge if index exists 将索引节点放在 xchain 上，判断索引是否存在
	index := packNodeIndex(node.ID)
	if _, err := ctx.GetObject([]byte(index)); err == nil {
		return code.Error(errorx.New(errorx.ErrCodeAlreadyExists,
			"duplicated nodeID"))
	}
	if err := ctx.PutObject([]byte(index), s); err != nil {
		return code.Error(errorx.NewCode(err, errorx.ErrCodeWriteBlockchain,
			"fail to put index-Node on xchain"))
	}
	// put index-nodeName on xchain 将索引节点名称放在 xchain 上
	index = packNodeNameIndex(node.Name)
	if _, err := ctx.GetObject([]byte(index)); err == nil {
		return code.Error(errorx.New(errorx.ErrCodeAlreadyExists,
			"duplicated nodeName"))
	}
	if err := ctx.PutObject([]byte(index), s); err != nil {
		return code.Error(errorx.NewCode(err, errorx.ErrCodeWriteBlockchain,
			"fail to put index-NodeName on xchain"))
	}

	// put listIndex-node on xchain 把列表索引节点放在xchain上
	index = packNodeListIndex(node)
	if err := ctx.PutObject([]byte(index), s); err != nil {
		return code.Error(errorx.NewCode(err, errorx.ErrCodeWriteBlockchain,
			"fail to put listIndex-Node on xchain"))
	}
	return code.OK([]byte("added"))
}

// ListExecutorNodes gets all Executor nodes 获取所有执行程序节点
func (x *Xdata) ListExecutorNodes(ctx code.Context) code.Response {
	var nodes blockchain.ExecutorNodes

	// get data nodes by list_prefix 按list_prefix获取数据节点
	prefix := prefixNodeListIndex
	iter := ctx.NewIterator(code.PrefixRange([]byte(prefix)))
	defer iter.Close()

	for iter.Next() {
		var node blockchain.ExecutorNode
		if err := json.Unmarshal(iter.Value(), &node); err != nil {
			return code.Error(errorx.NewCode(err, errorx.ErrCodeInternal,
				"fail to unmarshal node"))
		}
		nodes = append(nodes, node)
	}
	// marshal nodes 封送节点
	s, err := json.Marshal(nodes)
	if err != nil {
		return code.Error(errorx.NewCode(err, errorx.ErrCodeInternal,
			"fail to marshal nodes"))
	}
	return code.OK(s)
}

// GetExecutorNodeByID gets Executor node by ID 按 ID 获取执行器节点
func (x *Xdata) GetExecutorNodeByID(ctx code.Context) code.Response {
	// get id 获取标识
	nodeID, ok := ctx.Args()["id"]
	if !ok {
		return code.Error(errorx.New(errorx.ErrCodeParam, "missing param:id"))
	}

	// get node by index 按索引获取节点
	index := packNodeStringIndex(string(nodeID))
	s, err := ctx.GetObject([]byte(index))
	if err != nil {
		return code.Error(errorx.NewCode(err, errorx.ErrCodeNotFound, "node not found"))
	}
	return code.OK(s)
}

// GetExecutorNodeByName gets Executor node by NodeName 通过节点名称获取执行器节点
func (x *Xdata) GetExecutorNodeByName(ctx code.Context) code.Response {
	// get name 获取名称
	nodeName, ok := ctx.Args()["name"]
	if !ok {
		return code.Error(errorx.New(errorx.ErrCodeParam, "missing param:name"))
	}

	// get node by index-nodeName 按索引节点名称获取节点
	index := packNodeNameIndex(string(nodeName))
	s, err := ctx.GetObject([]byte(index))
	if err != nil {
		return code.Error(errorx.NewCode(err, errorx.ErrCodeNotFound, "node not found"))
	}
	return code.OK(s)
}
