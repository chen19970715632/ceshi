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
	"github.com/xuperchain/xuperchain/core/contractsdk/go/code"

	"github.com/PaddlePaddle/PaddleDTX/dai/blockchain"
	pbTask "github.com/PaddlePaddle/PaddleDTX/dai/protos/task"
	util "github.com/PaddlePaddle/PaddleDTX/xdb/pkgs/strings"
)

// PublishTask publishes task 发布任务
func (x *Xdata) PublishTask(ctx code.Context) code.Response {
	var opt blockchain.PublishFLTaskOptions
	// get opt 获取选项
	p, ok := ctx.Args()["opt"]
	if !ok {
		return code.Error(errorx.New(errorx.ErrCodeParam, "missing param:opt"))
	}
	// unmarshal opt 解组选项
	if err := json.Unmarshal(p, &opt); err != nil {
		return code.Error(errorx.NewCode(err, errorx.ErrCodeInternal,
			"fail to unmarshal PublishFLTaskOptions"))
	}
	// get fltask 获取任务
	t := opt.FLTask
	// get fltask signature msg 获取 FLTASK 签名消息
	msg, err := util.GetSigMessage(t)
	if err != nil {
		return code.Error(errorx.NewCode(err, errorx.ErrCodeInternal, "failed to get the message to sign"))
	}
	if err := x.checkSign(opt.Signature, t.Requester, []byte(msg)); err != nil {
		return code.Error(err)
	}

	t.Status = blockchain.TaskConfirming
	// marshal fltask 元帅飞行任务
	s, err := json.Marshal(t)
	if err != nil {
		return code.Error(errorx.NewCode(err, errorx.ErrCodeInternal,
			"fail to marshal FLTask"))
	}

	// put index-fltask on xchain, judge if index exists 将 index-fltask 放在 xchain 上，判断索引是否存在
	index := packFlTaskIndex(t.TaskID)
	if _, err := ctx.GetObject([]byte(index)); err == nil {
		return code.Error(errorx.New(errorx.ErrCodeAlreadyExists,
			"duplicated taskID"))
	}
	if err := ctx.PutObject([]byte(index), s); err != nil {
		return code.Error(errorx.NewCode(err, errorx.ErrCodeWriteBlockchain,
			"fail to put index-flTask on xchain"))
	}

	// put requester listIndex-fltask on xchain 把请求者列表索引-fltask 放在 xchain 上
	index = packFlTaskListIndex(t)
	if err := ctx.PutObject([]byte(index), []byte(t.TaskID)); err != nil {
		return code.Error(errorx.NewCode(err, errorx.ErrCodeWriteBlockchain,
			"fail to put requester listIndex-fltask on xchain"))
	}
	// put executor listIndex-fltask on xchain 将执行器列表索引-fltask 放在 xchain 上
	for _, ds := range t.DataSets {
		index := packExecutorTaskListIndex(ds.Executor, t)
		if err := ctx.PutObject([]byte(index), []byte(t.TaskID)); err != nil {
			return code.Error(errorx.NewCode(err, errorx.ErrCodeWriteBlockchain,
				"fail to put executor listIndex-fltask on xchain"))
		}
		// put requester and executor listIndex-fltask on xchain 将请求者和执行者列表Index-fltask 放在 xchain 上
		index_re := packRequesterExecutorTaskIndex(ds.Executor, t)
		if err := ctx.PutObject([]byte(index_re), []byte(t.TaskID)); err != nil {
			return code.Error(errorx.NewCode(err, errorx.ErrCodeWriteBlockchain,
				"fail to put requester and executor listIndex-fltask on xchain"))
		}
	}
	return code.OK([]byte("added"))
}

// ListTask lists tasks 列出任务
func (x *Xdata) ListTask(ctx code.Context) code.Response {
	// get opt 获取选项
	p, ok := ctx.Args()["opt"]
	if !ok {
		return code.Error(errorx.New(errorx.ErrCodeParam, "missing param:opt"))
	}
	// unmarshal opt 解组选项
	var opt blockchain.ListFLTaskOptions
	if err := json.Unmarshal(p, &opt); err != nil {
		return code.Error(errorx.NewCode(err, errorx.ErrCodeInternal,
			"fail to unmarshal ListFLTaskOptions"))
	}

	var tasks blockchain.FLTasks

	// get fltasks by list_prefix 通过list_prefix获得任务
	prefix := packFlTaskFilter(opt.PubKey, opt.ExecPubKey)
	iter := ctx.NewIterator(code.PrefixRange([]byte(prefix)))
	defer iter.Close()
	for iter.Next() {
		if opt.Limit > 0 && int64(len(tasks)) >= opt.Limit {
			break
		}
		t, err := x.getTaskById(ctx, string(iter.Value()))
		if err != nil {
			return code.Error(err)
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
		return code.Error(errorx.NewCode(err, errorx.ErrCodeInternal,
			"fail to marshal tasks"))
	}
	return code.OK(s)
}

// GetTaskById gets task by id 按 ID 获取任务
func (x *Xdata) GetTaskById(ctx code.Context) code.Response {
	taskID, ok := ctx.Args()["id"]
	if !ok {
		return code.Error(errorx.New(errorx.ErrCodeParam, "missing param:id"))
	}

	// get fltask by index 按索引获取 fltask
	index := packFlTaskIndex(string(taskID))
	s, err := ctx.GetObject([]byte(index))
	if err != nil {
		return code.Error(errorx.NewCode(err, errorx.ErrCodeNotFound, "task not found"))
	}
	return code.OK(s)
}

// ConfirmTask is called when Executor confirms task 在执行程序确认任务时调用
func (x *Xdata) ConfirmTask(ctx code.Context) code.Response {
	return x.setTaskConfirmStatus(ctx, true)
}

// RejectTask is called when Executor rejects task  在执行程序拒绝任务时调用
func (x *Xdata) RejectTask(ctx code.Context) code.Response {
	return x.setTaskConfirmStatus(ctx, false)
}

// setTaskConfirmStatus sets task status as Confirmed or Rejected 将任务状态设置为“已确认”或“已拒绝”
func (x *Xdata) setTaskConfirmStatus(ctx code.Context, isConfirm bool) code.Response {
	var opt blockchain.FLTaskConfirmOptions
	// get opt 获取选项
	p, ok := ctx.Args()["opt"]
	if !ok {
		return code.Error(errorx.New(errorx.ErrCodeParam, "missing param:opt"))
	}
	if err := json.Unmarshal(p, &opt); err != nil {
		return code.Error(errorx.NewCode(err, errorx.ErrCodeInternal,
			"fail to unmarshal FLTaskConfirmOptions"))
	}
	t, err := x.getTaskById(ctx, opt.TaskID)
	if err != nil {
		return code.Error(err)
	}
	// executor validity check 遗嘱执行人有效性检查
	if ok := x.checkExecutor(opt.Pubkey, t.DataSets); !ok {
		return code.Error(errorx.New(errorx.ErrCodeParam, "bad param: executor"))
	}
	// verify sig 验证签名
	msg, err := util.GetSigMessage(opt)
	if err != nil {
		return code.Error(errorx.NewCode(err, errorx.ErrCodeInternal, "failed to get the message to sign"))
	}
	if err := x.checkSign(opt.Signature, opt.Pubkey, []byte(msg)); err != nil {
		return code.Error(err)
	}

	// check status 检查状态
	if t.Status != blockchain.TaskConfirming {
		return code.Error(errorx.New(errorx.ErrCodeParam,
			"confirm task error, taskStatus is not Confirming, taskId: %s, taskStatus: %s", t.TaskID, t.Status))
	}
	isAllConfirm := true
	for index, ds := range t.DataSets {
		if bytes.Equal(ds.Executor, opt.Pubkey) {
			// judge sample file exists 判断示例文件存在
			if _, err := ctx.GetObject([]byte(ds.DataID)); err != nil {
				return code.Error(errorx.New(errorx.ErrCodeParam, "bad param:taskId, dataId not exist"))
			}
			// judge task is confirmed 裁判任务确认
			if ds.ConfirmedAt > 0 || ds.RejectedAt > 0 {
				return code.Error(errorx.New(errorx.ErrCodeAlreadyUpdate, "bad param:taskId, task already confirmed"))
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
		return code.Error(errorx.NewCode(err, errorx.ErrCodeInternal, "fail to marshal FLTask"))
	}
	// update index-fltask on xchain 在 xchain 上更新索引-FL任务
	index := packFlTaskIndex(t.TaskID)
	if err := ctx.PutObject([]byte(index), s); err != nil {
		return code.Error(errorx.NewCode(err, errorx.ErrCodeWriteBlockchain,
			"fail to confirm index-flTask on xchain"))
	}
	return code.OK([]byte("OK"))
}

// StartTask is called when Requester starts task after Executors confirmed 在执行程序确认后请求者启动任务时调用
// task status will be updated from 'Ready' to 'ToProcess'
func (x *Xdata) StartTask(ctx code.Context) code.Response {
	var opt blockchain.StartFLTaskOptions
	// get opt 获取选项
	p, ok := ctx.Args()["opt"]
	if !ok {
		return code.Error(errorx.New(errorx.ErrCodeParam, "missing param:opt"))
	}
	if err := json.Unmarshal(p, &opt); err != nil {
		return code.Error(errorx.NewCode(err, errorx.ErrCodeInternal,
			"fail to unmarshal StartFLTaskOptions"))
	}
	t, err := x.getTaskById(ctx, opt.TaskID)
	if err != nil {
		return code.Error(err)
	}
	// verify sig 验证签名
	msg, err := util.GetSigMessage(opt)
	if err != nil {
		return code.Error(errorx.NewCode(err, errorx.ErrCodeInternal, "failed to get the message to sign"))
	}
	if err := x.checkSign(opt.Signature, t.Requester, []byte(msg)); err != nil {
		return code.Error(err)
	}
	if t.Status != blockchain.TaskReady && t.Status != blockchain.TaskFailed {
		return code.Error(errorx.New(errorx.ErrCodeParam,
			"start task error, task status is not Ready or Failed, taskId: %s, taskStatus: %s", t.TaskID, t.Status))
	}
	// update task status 更新任务状态
	t.Status = blockchain.TaskToProcess
	s, err := json.Marshal(t)
	if err != nil {
		return code.Error(errorx.NewCode(err, errorx.ErrCodeInternal, "fail to marshal FLTask"))
	}
	// update index-fltask on xchain
	index := packFlTaskIndex(t.TaskID)
	if err := ctx.PutObject([]byte(index), s); err != nil {
		return code.Error(errorx.NewCode(err, errorx.ErrCodeWriteBlockchain,
			"fail to start index-flTask on xchain"))
	}
	return code.OK([]byte("OK"))
}

// ExecuteTask is called when Executor run task 在执行器运行任务时调用
func (x *Xdata) ExecuteTask(ctx code.Context) code.Response {
	return x.setTaskExecuteStatus(ctx, false)
}

// FinishTask is called when task execution finished 在任务执行完成时调用
func (x *Xdata) FinishTask(ctx code.Context) code.Response {
	return x.setTaskExecuteStatus(ctx, true)
}

// setTaskExecuteStatus is called by the Executor when the task is started or when the task has finished
// if the task status is 'ToProcess', update status to 'Processing'
// if the task status is 'Processing', update status to 'Finished' or 'Failed'
//在任务启动或任务完成时由执行程序调用
//如果任务状态为“待处理”，请将状态更新为“正在处理”
//如果任务状态为“正在处理”，请将状态更新为“已完成”或“失败”
func (x *Xdata) setTaskExecuteStatus(ctx code.Context, isFinish bool) code.Response {
	var opt blockchain.FLTaskExeStatusOptions
	// get opt 获取选项
	p, ok := ctx.Args()["opt"]
	if !ok {
		return code.Error(errorx.New(errorx.ErrCodeParam, "missing param:opt"))
	}
	if err := json.Unmarshal(p, &opt); err != nil {
		return code.Error(errorx.NewCode(err, errorx.ErrCodeInternal,
			"fail to unmarshal FLTaskExeStatusOptions"))
	}
	t, err := x.getTaskById(ctx, opt.TaskID)
	if err != nil {
		return code.Error(err)
	}
	// executor validity check 遗嘱执行人有效性检查
	if ok := x.checkExecutor(opt.Executor, t.DataSets); !ok {
		return code.Error(errorx.New(errorx.ErrCodeParam, "bad param:executor"))
	}
	// verify sig 验证签名
	msg, err := util.GetSigMessage(opt)
	if err != nil {
		return code.Error(errorx.NewCode(err, errorx.ErrCodeInternal, "failed to get the message to sign"))
	}
	if err := x.checkSign(opt.Signature, opt.Executor, []byte(msg)); err != nil {
		return code.Error(err)
	}

	if isFinish {
		if t.Status != blockchain.TaskProcessing {
			return code.Error(errorx.New(errorx.ErrCodeParam,
				"finish task error, task status is not Processing, taskId: %s, taskStatus: %s", t.TaskID, t.Status))
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
			return code.Error(errorx.New(errorx.ErrCodeParam,
				"execute task error, task status is not ToProcess, taskId: %s, taskStatus: %s", t.TaskID, t.Status))
		}
		t.Status = blockchain.TaskProcessing
		t.StartTime = opt.CurrentTime
	}

	// update task status 更新任务状态
	s, err := json.Marshal(t)
	if err != nil {
		return code.Error(errorx.NewCode(err, errorx.ErrCodeInternal, "fail to marshal FLTask"))
	}
	// update index-fltask on xchain 在 xchain 上更新索引-FL任务
	index := packFlTaskIndex(t.TaskID)
	if err := ctx.PutObject([]byte(index), s); err != nil {
		return code.Error(errorx.NewCode(err, errorx.ErrCodeWriteBlockchain,
			"fail to set task execute status on xchain"))
	}
	return code.OK([]byte("OK"))
}

// getTaskById gets task details from the blockchain ledger 从区块链账本中获取任务详细信息
func (x *Xdata) getTaskById(ctx code.Context, taskID string) (t blockchain.FLTask, err error) {
	index := packFlTaskIndex(taskID)
	s, err := ctx.GetObject([]byte(index))
	if err != nil {
		return t, errorx.NewCode(err, errorx.ErrCodeNotFound,
			"the task[%s] not found", taskID)
	}

	if err = json.Unmarshal([]byte(s), &t); err != nil {
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

// checkExecutor used for Executor validity check, only the Executor specified by the Requester can confirm the task 用于执行者有效性检查，只有请求者指定的执行者才能确认任务
func (x *Xdata) checkExecutor(executor []byte, dataSets []*pbTask.DataForTask) bool {
	for _, ds := range dataSets {
		if bytes.Equal(ds.Executor, executor) {
			return true
		}
	}
	return false
}
