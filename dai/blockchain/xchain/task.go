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

package xchain

import (
	"encoding/json"

	"github.com/PaddlePaddle/PaddleDTX/xdb/errorx"

	"github.com/PaddlePaddle/PaddleDTX/dai/blockchain"
)

// PublishTask publishes task on xchain 在xchain上发布任务
func (x *XChain) PublishTask(opt *blockchain.PublishFLTaskOptions) error {
	opts, err := json.Marshal(*opt)
	if err != nil {
		return errorx.NewCode(err, errorx.ErrCodeInternal,
			"fail to marshal PublishFLTaskOptions")
	}
	args := map[string]string{
		"opt": string(opts),
	}
	mName := "PublishTask"
	if _, err = x.InvokeContract(args, mName); err != nil {
		return err
	}
	return nil
}

// ListTask lists tasks from xchain 列出来自 Xchain 的任务
func (x *XChain) ListTask(opt *blockchain.ListFLTaskOptions) (blockchain.FLTasks, error) {
	var ts blockchain.FLTasks

	opts, err := json.Marshal(*opt)
	if err != nil {
		return ts, errorx.NewCode(err, errorx.ErrCodeInternal,
			"fail to marshal ListFLTaskOptions")
	}
	args := map[string]string{
		"opt": string(opts),
	}
	mName := "ListTask"
	s, err := x.QueryContract(args, mName)
	if err != nil {
		return ts, err
	}
	if err = json.Unmarshal(s, &ts); err != nil {
		return ts, errorx.NewCode(err, errorx.ErrCodeInternal,
			"fail to unmarshal FLTasks")
	}
	return ts, nil
}

// GetTaskById gets task by id 按 ID 获取任务
func (x *XChain) GetTaskById(id string) (blockchain.FLTask, error) {
	var t blockchain.FLTask
	args := map[string]string{
		"id": id,
	}
	mName := "GetTaskById"
	s, err := x.QueryContract(args, mName)
	if err != nil {
		return t, err
	}

	if err = json.Unmarshal([]byte(s), &t); err != nil {
		return t, errorx.NewCode(err, errorx.ErrCodeInternal,
			"fail to unmarshal File")
	}
	return t, nil
}

// ConfirmTask is called when Executor confirms task 在执行程序确认任务时调用
func (x *XChain) ConfirmTask(opt *blockchain.FLTaskConfirmOptions) error {
	return x.setTaskConfirmStatus(opt, true)
}

// RejectTask is called when Executor rejects task 在执行程序拒绝任务时调用
func (x *XChain) RejectTask(opt *blockchain.FLTaskConfirmOptions) error {
	return x.setTaskConfirmStatus(opt, false)
}

// setTaskConfirmStatus is used by the Executor to confirm or reject the task
// if isConfirm is false, update the task status from 'Confirming' to 'Rejected'
//由执行者用于确认或拒绝任务
//如果 isConfirm 为 false，请将任务状态从“正在确认”更新为“已拒绝”
func (x *XChain) setTaskConfirmStatus(opt *blockchain.FLTaskConfirmOptions, isConfirm bool) error {
	opts, err := json.Marshal(*opt)
	if err != nil {
		return errorx.NewCode(err, errorx.ErrCodeInternal,
			"fail to marshal FLTaskConfirmOptions")
	}
	args := map[string]string{
		"opt": string(opts),
	}
	mName := "RejectTask"
	if isConfirm {
		mName = "ConfirmTask"
	}
	if _, err := x.InvokeContract(args, mName); err != nil {
		return err
	}
	return nil
}

// StartTask is called when Requester starts task after all Executors confirmed 当请求者在所有执行程序确认后启动任务时调用
func (x *XChain) StartTask(opt *blockchain.StartFLTaskOptions) error {
	opts, err := json.Marshal(*opt)
	if err != nil {
		return errorx.NewCode(err, errorx.ErrCodeInternal,
			"fail to marshal StartFLTaskOptions")
	}
	args := map[string]string{
		"opt": string(opts),
	}
	mName := "StartTask"
	if _, err := x.InvokeContract(args, mName); err != nil {
		return err
	}
	return nil
}

// ExecuteTask is called when Executor run task 在执行器运行任务时调用
func (x *XChain) ExecuteTask(opt *blockchain.FLTaskExeStatusOptions) error {
	return x.setTaskExecuteStatus(opt, false)
}

// FinishTask is called when task execution finished 在任务执行完成时调用
func (x *XChain) FinishTask(opt *blockchain.FLTaskExeStatusOptions) error {
	return x.setTaskExecuteStatus(opt, true)
}

// setTaskExecuteStatus updates task status when the Executor starts running the task or finished the task
// call the contract's 'ExecuteTask' or 'FinishTask' method
//当执行程序开始运行任务或完成任务时更新任务状态
//调用合约的“ExecuteTask”或“FinishTask”方法
func (x *XChain) setTaskExecuteStatus(opt *blockchain.FLTaskExeStatusOptions, isFinish bool) error {
	opts, err := json.Marshal(*opt)
	if err != nil {
		return errorx.NewCode(err, errorx.ErrCodeInternal,
			"fail to marshal FLTaskExeStatusOptions")
	}
	args := map[string]string{
		"opt": string(opts),
	}
	mName := "ExecuteTask"
	if isFinish {
		mName = "FinishTask"
	}
	if _, err := x.InvokeContract(args, mName); err != nil {
		return err
	}
	return nil
}
