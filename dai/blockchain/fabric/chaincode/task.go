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
	"bytes"
	"encoding/json"
	"github.com/PaddlePaddle/PaddleDTX/crypto/core/ecdsa"
	"github.com/PaddlePaddle/PaddleDTX/crypto/core/hash"
	"github.com/PaddlePaddle/PaddleDTX/xdb/errorx"
	"github.com/hyperledger/fabric/core/chaincode/shim"
	pb "github.com/hyperledger/fabric/protos/peer"

	"github.com/PaddlePaddle/PaddleDTX/dai/blockchain"
	pbTask "github.com/PaddlePaddle/PaddleDTX/dai/protos/task"
	util "github.com/PaddlePaddle/PaddleDTX/xdb/pkgs/strings"
)

// PublishTask publishes task 发布任务
func (x *Xdata) PublishTask(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var opt blockchain.PublishFLTaskOptions
	if len(args) < 1 {
		return shim.Error("incorrect arguments. expecting PublishTaskOptions")
	}
	if err := json.Unmarshal([]byte(args[0]), &opt); err != nil {
		return shim.Error(errorx.NewCode(err, errorx.ErrCodeInternal,
			"fail to unmarshal PublishFLTaskOptions").Error())
	}

	// get fltask 获取任务
	t := opt.FLTask
	msg, err := util.GetSigMessage(t)
	if err != nil {
		return shim.Error(errorx.NewCode(err, errorx.ErrCodeInternal, "failed to get the message to sign").Error())
	}
	if err := x.checkSign(opt.Signature, t.Requester, []byte(msg)); err != nil {
		return shim.Error(err.Error())
	}

	t.Status = blockchain.TaskConfirming

	// marshal fltask 元帅飞行任务
	s, err := json.Marshal(t)
	if err != nil {
		return shim.Error(errorx.NewCode(err, errorx.ErrCodeInternal,
			"fail to marshal FLTask").Error())
	}

	// put index-fltask on fabric, judge if index exists  将索引任务放在结构上，判断索引是否存在
	index := packFlTaskIndex(t.TaskID)
	if resp := x.GetValue(stub, []string{index}); len(resp.Payload) != 0 {
		return shim.Error(errorx.New(errorx.ErrCodeAlreadyExists, "duplicated taskID").Error())
	}
	if resp := x.SetValue(stub, []string{index, string(s)}); resp.Status == shim.ERROR {
		return shim.Error(errorx.New(errorx.ErrCodeWriteBlockchain,
			"fail to put index-flTask on fabric: %s", resp.Message).Error())
	}

	// put requester listIndex-fltask on fabric 将请求者列表索引-fltask 放在结构上
	index = packFlTaskListIndex(t)

	if resp := x.SetValue(stub, []string{index, t.TaskID}); resp.Status == shim.ERROR {
		return shim.Error(errorx.New(errorx.ErrCodeWriteBlockchain,
			"fail to put requester listIndex-fltask on fabric: %s", resp.Message).Error())
	}

	//put executor listIndex-fltask on fabric  将执行器列表索引-fltask 放在结构上
	for _, ds := range t.DataSets {
		index := packExecutorTaskListIndex(ds.Executor, t)
		if resp := x.SetValue(stub, []string{index, t.TaskID}); resp.Status == shim.ERROR {
			return shim.Error(errorx.New(errorx.ErrCodeWriteBlockchain,
				"fail to put executor listIndex-fltask on fabric: %s", resp.Message).Error())
		}
		// put requester and executor listIndex-fltask on fabric 将请求者和执行者列表Index-fltask 放在结构上
		index_re := packRequesterExecutorTaskIndex(ds.Executor, t)
		if resp := x.SetValue(stub, []string{index_re, t.TaskID}); resp.Status == shim.ERROR {
			return shim.Error(errorx.New(errorx.ErrCodeWriteBlockchain,
				"fail to put requester and executor listIndex-fltask on fabric: %s", resp.Message).Error())
		}
	}
	return shim.Success([]byte("added"))
}

// ListTask lists tasks  列出任务
func (x *Xdata) ListTask(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) < 1 {
		return shim.Error("invalid arguments. expecting ListTaskOptions")
	}

	// unmarshal opt 解组选项
	var opt blockchain.ListFLTaskOptions
	if err := json.Unmarshal([]byte(args[0]), &opt); err != nil {
		return shim.Error(errorx.NewCode(err, errorx.ErrCodeInternal,
			"failed to unmarshal ListFLTaskOptions").Error())
	}

	var tasks blockchain.FLTasks

	// get fltasks by list_prefix  通过list_prefix获得任务
	prefix, attr := packFlTaskFilter(opt.PubKey, opt.ExecPubKey)
	iterator, err := stub.GetStateByPartialCompositeKey(prefix, attr)

	// defer iter.Close()  推迟迭代。关闭（）
	if err != nil {
		return shim.Error(err.Error())
	}
	defer iterator.Close()
	for iterator.HasNext() {
		queryResponse, err := iterator.Next()
		if err != nil {
			return shim.Error(err.Error())
		}
		if opt.Limit > 0 && int64(len(tasks)) >= opt.Limit {
			break
		}

		t, err := x.getTaskById(stub, string(queryResponse.Value))
		if err != nil {
			return shim.Error(errorx.NewCode(err, errorx.ErrCodeInternal, "failed to get task by id").Error())
		}

		if t.PublishTime < opt.TimeStart || (opt.TimeEnd > 0 && t.PublishTime > opt.TimeEnd) ||
			(opt.Status != "" && t.Status != opt.Status) {
			continue
		}
		tasks = append(tasks, t)

	}

	// marshal tasks 封送任务
	s, err := json.Marshal(tasks)
	if err != nil {
		return shim.Error(errorx.NewCode(err, errorx.ErrCodeInternal,
			"fail to marshal tasks").Error())
	}
	return shim.Success(s)
}

// GetTaskById gets task by id 按 ID 获取任务
func (x *Xdata) GetTaskById(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) < 1 {
		return shim.Error("invalid arguments. missing param: id")
	}
	index := packFlTaskIndex(args[0])
	resp := x.GetValue(stub, []string{index})
	if len(resp.Payload) == 0 {
		return shim.Error(errorx.New(errorx.ErrCodeNotFound, "task not found: %s", resp.Message).Error())
	}
	return shim.Success(resp.Payload)
}

// ConfirmTask is called when Executor confirms task 在执行程序确认任务时调用
func (x *Xdata) ConfirmTask(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	return x.setTaskConfirmStatus(stub, args, true)
}

func (x *Xdata) RejectTask(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	return x.setTaskConfirmStatus(stub, args, false)
}

// setTaskConfirmStatus sets task status as Confirmed or Rejected    将任务状态设置为“已确认”或“已拒绝”  参考xdb fileauth 112
func (x *Xdata) setTaskConfirmStatus(stub shim.ChaincodeStubInterface, args []string, isConfirm bool) pb.Response {
	var opt blockchain.FLTaskConfirmOptions
	if len(args) < 1 {
		return shim.Error("invalid arguments. expecting FLTaskConfirmOptions")
	}
	if err := json.Unmarshal([]byte(args[0]), &opt); err != nil {
		return shim.Error(errorx.NewCode(err, errorx.ErrCodeInternal,
			"fail to unmarshal FLTaskConfirmOptions").Error())
	}
	t, err := x.getTaskById(stub, opt.TaskID)
	if err != nil {
		return shim.Error(err.Error())
	}

	// executor validity check 遗嘱执行人有效性检查
	if ok := x.checkExecutor(opt.Pubkey, t.DataSets); !ok {
		return shim.Error(errorx.New(errorx.ErrCodeParam, "bad param: executor").Error())
	}

	// verify sig 验证签名
	msg, err := util.GetSigMessage(opt)
	if err != nil {
		return shim.Error(errorx.Internal(err, "failed to get the message to sign").Error())
	}
	if err := x.checkSign(opt.Signature, opt.Pubkey, []byte(msg)); err != nil {
		return shim.Error(err.Error())
	}

	// check status 检查状态
	if t.Status != blockchain.TaskConfirming {
		return shim.Error(errorx.New(errorx.ErrCodeParam,
			"confirm task error, taskStatus is not Confirming, taskId: %s, taskStatus: %s", t.TaskID, t.Status).Error())
	}

	isAllConfirm := true
	for index, ds := range t.DataSets {
		if bytes.Equal(ds.Executor, opt.Pubkey) {
			if resp := x.GetValue(stub, []string{ds.DataID}); len(resp.Payload) == 0 {
				return shim.Error(errorx.New(errorx.ErrCodeParam, "bad param:taskId, dataId not exist").Error())
			}

			// judge task is confirmed 裁判任务确认
			if ds.ConfirmedAt > 0 || ds.RejectedAt > 0 {
				return shim.Error(errorx.New(errorx.ErrCodeAlreadyExists, "bad param:taskId, task already confirmed").Error())
			}
			if isConfirm {
				t.DataSets[index].ConfirmedAt = opt.CurrentTime
			} else {
				t.DataSets[index].RejectedAt = opt.CurrentTime
			}
		} else {
			if ds.ConfirmedAt == 0 {
				isAllConfirm = false
			}
		}
	}

	// if all executor nodes confirmed task, task status is ready 如果所有执行器节点都确认了任务，则任务状态为就绪
	if isAllConfirm {
		t.Status = blockchain.TaskReady
	}
	// if one of executor nodes rejected task, task status is rejected 如果其中一个执行程序节点拒绝了任务，则任务状态为“拒绝”
	if !isConfirm {
		t.Status = blockchain.TaskRejected
		t.ErrMessage = opt.RejectReason
	}
	s, err := json.Marshal(t)
	if err != nil {
		return shim.Error(errorx.NewCode(err, errorx.ErrCodeInternal, "fail to marshal FLTask").Error())
	}

	// update index-fltask on fabric 在结构上更新索引任务
	index := packFlTaskIndex(t.TaskID)
	if resp := x.SetValue(stub, []string{index, string(s)}); resp.Status == shim.ERROR {
		return shim.Error(errorx.New(errorx.ErrCodeWriteBlockchain,
			"fail to confirm index-flTask on fabric: %s", resp.Message).Error())
	}
	return shim.Success([]byte("OK"))
}

// StartTask is called when Requester starts task after Executors confirmed
// task status will be updated from 'Ready' to 'ToProcess'
//在执行程序确认后请求者启动任务时调用
//任务状态将从“就绪”更新为“待处理”
func (x *Xdata) StartTask(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var opt blockchain.StartFLTaskOptions
	if len(args) < 1 {
		return shim.Error("incorrect arguments. expecting StartFLTaskOptions")
	}
	// unmarshal opt  解组选项
	if err := json.Unmarshal([]byte(args[0]), &opt); err != nil {
		return shim.Error(errorx.NewCode(err, errorx.ErrCodeInternal,
			"fail to unmarshal StartFLTaskOptions").Error())
	}
	t, err := x.getTaskById(stub, opt.TaskID)
	if err != nil {
		return shim.Error(err.Error())
	}

	// verify sig 验证签名
	msg, err := util.GetSigMessage(opt)
	if err != nil {
		return shim.Error(errorx.Internal(err, "failed to get the message to sign").Error())
	}
	if err := x.checkSign(opt.Signature, t.Requester, []byte(msg)); err != nil {
		return shim.Error(err.Error())
	}
	if t.Status != blockchain.TaskReady && t.Status != blockchain.TaskFailed {
		return shim.Error(errorx.New(errorx.ErrCodeParam,
			"start task error, task status is not Ready or Failed, taskId: %s, taskStatus: %s", t.TaskID, t.Status).Error())
	}

	// update task status 更新任务状态
	t.Status = blockchain.TaskToProcess
	s, err := json.Marshal(t)
	if err != nil {
		return shim.Error(errorx.NewCode(err, errorx.ErrCodeInternal, "fail to marshal FLTask").Error())
	}

	// update index-fltask on fabric 在结构上更新索引任务
	index := packFlTaskIndex(t.TaskID)
	if resp := x.SetValue(stub, []string{index, string(s)}); resp.Status == shim.ERROR {
		return shim.Error(errorx.New(errorx.ErrCodeWriteBlockchain,
			"fail to confirm index-flTask on fabric: %s", resp.Message).Error())
	}
	return shim.Success([]byte("OK"))
}

// ExecuteTask is called when Executor run task 在执行器运行任务时调用
func (x *Xdata) ExecuteTask(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	return x.setTaskExecuteStatus(stub, args, false)
}

// FinishTask is called when task execution finished 在任务执行完成时调用
func (x *Xdata) FinishTask(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	return x.setTaskExecuteStatus(stub, args, true)
}

// setTaskExecuteStatus is called by the Executor when the task is started or when the task has finished
// if the task status is 'ToProcess', update status to 'Processing'
// if the task status is 'Processing', update status to 'Finished' or 'Failed'
//在任务启动或任务完成时由执行程序调用
//如果任务状态为“待处理”，请将状态更新为“正在处理”
//如果任务状态为“正在处理”，请将状态更新为“已完成”或“失败”
func (x *Xdata) setTaskExecuteStatus(stub shim.ChaincodeStubInterface, args []string, isFinish bool) pb.Response {
	var opt blockchain.FLTaskExeStatusOptions
	if len(args) < 1 {
		return shim.Error("invalid arguments. expecting FLTaskExeStatusOptions")
	}
	if err := json.Unmarshal([]byte(args[0]), &opt); err != nil {
		return shim.Error(errorx.NewCode(err, errorx.ErrCodeInternal,
			"fail to unmarshal FLTaskExeStatusOptions").Error())
	}
	t, err := x.getTaskById(stub, opt.TaskID)
	if err != nil {
		return shim.Error(err.Error())
	}

	// executor validity check 遗嘱执行人有效性检查
	if ok := x.checkExecutor(opt.Executor, t.DataSets); !ok {
		return shim.Error(errorx.New(errorx.ErrCodeParam, "bad param:executor").Error())
	}

	// verify sig 验证签名
	msg, err := util.GetSigMessage(opt)
	if err != nil {
		return shim.Error(errorx.Internal(err, "failed to get the message to sign").Error())
	}
	if err := x.checkSign(opt.Signature, opt.Executor, []byte(msg)); err != nil {
		return shim.Error(err.Error())
	}

	if isFinish {
		if t.Status != blockchain.TaskProcessing {
			return shim.Error(errorx.New(errorx.ErrCodeParam,
				"finish task error, task status is not Processing, taskId: %s, taskStatus: %s", t.TaskID, t.Status).Error())
		}

		t.Status = blockchain.TaskFinished
		t.EndTime = opt.CurrentTime
		t.Result = opt.Result

		if opt.ErrMessage != "" {
			t.Status = blockchain.TaskFailed
			t.ErrMessage = opt.ErrMessage
		}
	} else {
		if t.Status != blockchain.TaskToProcess {
			return shim.Error(errorx.New(errorx.ErrCodeParam,
				"execute task error, task status is not ToProcess, taskId: %s, taskStatus: %s", t.TaskID, t.Status).Error())
		}
		t.Status = blockchain.TaskProcessing
		t.StartTime = opt.CurrentTime
	}

	// update task status 更新任务状态
	s, err := json.Marshal(t)
	if err != nil {
		return shim.Error(errorx.NewCode(err, errorx.ErrCodeInternal, "fail to marshal FLTask").Error())
	}

	// update index-fltask on fabric 在结构上更新索引任务
	index := packFlTaskIndex(t.TaskID)
	if resp := x.SetValue(stub, []string{index, string(s)}); resp.Status == shim.ERROR {
		return shim.Error(errorx.New(errorx.ErrCodeWriteBlockchain,
			"fail to set task execute status on fabric: %s", resp.Message).Error())
	}
	return shim.Success([]byte("OK"))
}

// getTaskById gets task details from the blockchain ledger 从区块链账本中获取任务详细信息
func (x *Xdata) getTaskById(stub shim.ChaincodeStubInterface, taskID string) (t blockchain.FLTask, err error) {
	index := packFlTaskIndex(taskID)
	resp := x.GetValue(stub, []string{index})
	if len(resp.Payload) == 0 {
		return t, errorx.New(errorx.ErrCodeNotFound, "the task[%s] not found", taskID, resp.Message)
	}
	if err = json.Unmarshal(resp.Payload, &t); err != nil {
		return t, errorx.NewCode(err, errorx.ErrCodeInternal,
			"fail to unmarshal FlTask")
	}
	return t, nil
}

// checkSign verifies the signature 验证签名
func (x *Xdata) checkSign(sign, owner, mes []byte) (err error) {
	// verify sig 验证签名
	if len(sign) != ecdsa.SignatureLength {
		return errorx.New(errorx.ErrCodeParam, "bad param:signature")
	}
	var pubkey [ecdsa.PublicKeyLength]byte
	var sig [ecdsa.SignatureLength]byte
	copy(pubkey[:], owner)
	copy(sig[:], sign)
	if err := ecdsa.Verify(pubkey, hash.HashUsingSha256(mes), sig); err != nil {
		return errorx.NewCode(err, errorx.ErrCodeBadSignature, "failed to verify signature")
	}
	return nil
}

// checkExecutor used for Executor validity check, only the Executor specified by the Requester can confirm the task
//用于执行者有效性检查，只有请求者指定的执行者才能确认任务
func (x *Xdata) checkExecutor(executor []byte, dataSets []*pbTask.DataForTask) bool {
	for _, ds := range dataSets {
		if bytes.Equal(ds.Executor, executor) {
			return true
		}
	}
	return false
}
