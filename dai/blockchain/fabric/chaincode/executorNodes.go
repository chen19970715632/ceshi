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
	"github.com/hyperledger/fabric/core/chaincode/shim"
	pb "github.com/hyperledger/fabric/protos/peer"

	"github.com/PaddlePaddle/PaddleDTX/dai/blockchain"
)

// RegisterExecutorNode registers Executor node 注册执行器节点
func (x *Xdata) RegisterExecutorNode(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var opt blockchain.AddNodeOptions
	if len(args) < 1 {
		return shim.Error("invalid arguments. expecting AddNodeOptions")
	}
	// unmarshal opt 解组选项
	if err := json.Unmarshal([]byte(args[0]), &opt); err != nil {
		return shim.Error(errorx.NewCode(err, errorx.ErrCodeInternal,
			"failed to unmarshal AddNodeOptions").Error())
	}

	// get node 获取节点
	n := opt.Node
	// marshal node 元组节点
	s, err := json.Marshal(n)
	if err != nil {
		return shim.Error(errorx.NewCode(err, errorx.ErrCodeInternal,
			"failed to marshal Node").Error())
	}

	// check node.Name, the length of the executor node name is 4-16 characters
	// only support lowercase letters and numbers
	//检查节点。名称，执行节点名称的长度为4-16个字符
	//	仅支持小写字母和数字
	if ok, _ := regexp.MatchString("^[a-z0-9]{4,16}", n.Name); !ok {
		return shim.Error(errorx.New(errorx.ErrCodeParam,
			"bad param, nodeName only supports numbers and lowercase letters with a length of 4-16").Error())
	}

	// get the message to sign 获取要签名的消息
	msg, err := util.GetSigMessage(opt)
	if err != nil {
		return shim.Error(errorx.Internal(err, "failed to get the message to sign").Error())
	}

	// verify sig 验证签名
	if err := x.checkSign(opt.Signature, n.ID, []byte(msg)); err != nil {
		return shim.Error(err.Error())
	}

	// put index-node on fabric, judge if index exists 将索引节点放在结构上，判断索引是否存在
	index := packNodeIndex(n.ID)
	if resp := x.GetValue(stub, []string{index}); len(resp.Payload) != 0 {
		return shim.Error(errorx.New(errorx.ErrCodeAlreadyExists,
			"duplicated nodeID").Error())
	}
	if resp := x.SetValue(stub, []string{index, string(s)}); resp.Status == shim.ERROR {
		return shim.Error(errorx.New(errorx.ErrCodeWriteBlockchain,
			"failed to put index-Node on chain: %s", resp.Message).Error())
	}

	// put index-nodeName on fabric 将索引节点名称放在结构上
	index = packNodeNameIndex(n.Name)
	if resp := x.GetValue(stub, []string{index}); len(resp.Payload) != 0 {
		return shim.Error(errorx.New(errorx.ErrCodeAlreadyExists,
			"duplicated nodeName").Error())
	}
	if resp := x.SetValue(stub, []string{index, string(s)}); resp.Status == shim.ERROR {
		return shim.Error(errorx.New(errorx.ErrCodeWriteBlockchain, "failed to put index-NodeName on chain: %s",
			resp.Message).Error())
	}

	// put listIndex-node on fabric 将列表索引节点放在结构上
	index = packNodeListIndex(n)
	if resp := x.SetValue(stub, []string{index, string(s)}); resp.Status == shim.ERROR {
		return shim.Error(errorx.New(errorx.ErrCodeWriteBlockchain, "failed to put listIndex-Node on chain: %s",
			resp.Message).Error())
	}
	return shim.Success(nil)
}

// ListExecutorNodes gets all Executor nodes  获取所有执行程序节点
func (x *Xdata) ListExecutorNodes(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var nodes blockchain.ExecutorNodes
	iterator, err := stub.GetStateByPartialCompositeKey(prefixNodeListIndex, []string{})
	if err != nil {
		return shim.Error(err.Error())
	}
	defer iterator.Close()
	for iterator.HasNext() {
		queryResponse, err := iterator.Next()
		if err != nil {
			return shim.Error(err.Error())
		}

		var node blockchain.ExecutorNode
		if err := json.Unmarshal(queryResponse.Value, &node); err != nil {
			return shim.Error(errorx.NewCode(err, errorx.ErrCodeInternal,
				"failed to unmarshal node").Error())
		}
		nodes = append(nodes, node)
	}

	// marshal nodes  封送节点
	s, err := json.Marshal(nodes)
	if err != nil {
		return shim.Error(errorx.NewCode(err, errorx.ErrCodeInternal,
			"failed to marshal nodes").Error())
	}
	return shim.Success(s)
}

// GetExecutorNodeByID gets Executor node by ID 按 ID 获取执行器节点
func (x *Xdata) GetExecutorNodeByID(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) < 1 {
		return shim.Error("invalid arguments. expecting nodeID")
	}
	index := packNodeStringIndex(args[0])
	resp := x.GetValue(stub, []string{index})
	if len(resp.Payload) == 0 {
		return shim.Error(errorx.New(errorx.ErrCodeNotFound, "node not found: %s", resp.Message).Error())
	}
	return shim.Success(resp.Payload)
}

// GetExecutorNodeByName gets Executor node by NodeName  通过节点名称获取执行器节点
func (x *Xdata) GetExecutorNodeByName(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) < 1 {
		return shim.Error("invalid arguments. expecting name")
	}
	index := packNodeNameIndex(args[0])
	resp := x.GetValue(stub, []string{index})
	if len(resp.Payload) == 0 {
		return shim.Error(errorx.New(errorx.ErrCodeNotFound,
			"node not found: %s", resp.Message).Error())
	}
	return shim.Success(resp.Payload)
}
