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

package blockchain

import (
	pbCom "github.com/PaddlePaddle/PaddleDTX/dai/protos/common"
	pbTask "github.com/PaddlePaddle/PaddleDTX/dai/protos/task"
)

const (
	/* Define Task Status stored in Contract */ /* 定义存储在合同中的任务状态 */
	TaskConfirming                              = "Confirming" // waiting for Executors to confirm 等待执行人确认
	TaskReady                                   = "Ready"      // has been confirmed by all Executors, and ready to start 已得到所有执行人的确认，并准备开始
	TaskToProcess                               = "ToProcess"  // has started, and waiting to be precessed 已经开始，正在等待处理
	TaskProcessing                              = "Processing" // under process, that's during training or predicting 在处理中，即在训练或预测期间
	TaskFinished                                = "Finished"   // task finished 任务已完成
	TaskFailed                                  = "Failed"     // task failed 任务失败
	TaskRejected                                = "Rejected"   // task rejected by one of the Executors 任务被其中一个执行者拒绝

	/* Define Task Type stored in Contract */ /* 定义存储在合约中的任务类型 */
	TaskTypeTrain                             = "train"   // training task 训练任务
	TaskTypePredict                           = "predict" // prediction task 预测任务

	/* Define Algorithms stored in Contract */ /* 定义存储在合约中的算法 */
	AlgorithmVLine                             = "linear-vl"       // linear regression with multiple variables in vertical federated learning 垂直联邦学习中具有多个变量的线性回归
	AlgorithmVLog                              = "logistic-vl"     // logistic regression with multiple variables in vertical federated learning 垂直联邦学习中具有多个变量的逻辑回归
	AlgorithmVDnn                              = "dnn-paddlefl-vl" // dnn implemented using paddlefl 使用 Paddlefl 实现的 DNN

	/* Define Regularization stored in Contract */ /* 定义存储在合约中的正则化 */
	RegModeL1                                      = "l1" // L1-norm L1-范数
	RegModeL2                                      = "l2" // L2-norm L2-范数

	/* Define the maximum number of task list query */ /* 定义任务列表查询的最大数量 */
	TaskListMaxNum                                     = 100
)

// VlAlgorithmListName the mapping of vertical algorithm name and value 垂直算法名称和值的映射
var VlAlgorithmListName = map[string]pbCom.Algorithm{
	AlgorithmVLine: pbCom.Algorithm_LINEAR_REGRESSION_VL,
	AlgorithmVLog:  pbCom.Algorithm_LOGIC_REGRESSION_VL,
	AlgorithmVDnn:  pbCom.Algorithm_DNN_PADDLEFL_VL,
}

// VlAlgorithmListValue the mapping of vertical algorithm value and name
// key is the int value of the algorithm
//垂直算法值和名称的映射
//key 是算法的 int 值
var VlAlgorithmListValue = map[pbCom.Algorithm]string{
	pbCom.Algorithm_LINEAR_REGRESSION_VL: AlgorithmVLine,
	pbCom.Algorithm_LOGIC_REGRESSION_VL:  AlgorithmVLog,
	pbCom.Algorithm_DNN_PADDLEFL_VL:      AlgorithmVDnn,
}

// TaskTypeListName the mapping of train task type name and value
// key is the task type name of the training task or prediction task
//训练任务类型名称和值的映射
//key 是训练任务或预测任务的任务类型名称
var TaskTypeListName = map[string]pbCom.TaskType{
	TaskTypeTrain:   pbCom.TaskType_LEARN,
	TaskTypePredict: pbCom.TaskType_PREDICT,
}

// TaskTypeListValue the mapping of train task type value and name
// key is the int value of the training task or prediction task
//训练任务类型值和名称的映射
//key 是训练任务或预测任务的 int 值
var TaskTypeListValue = map[pbCom.TaskType]string{
	pbCom.TaskType_LEARN:   TaskTypeTrain,
	pbCom.TaskType_PREDICT: TaskTypePredict,
}

// RegModeListName the mapping of train regMode name and value 列车注册模式名称和值的映射
var RegModeListName = map[string]pbCom.RegMode{
	RegModeL1: pbCom.RegMode_Reg_Lasso,
	RegModeL2: pbCom.RegMode_Reg_Ridge,
}

// RegModeListValue the mapping of train regMode value and name 列车注册表模式值和名称的映射
var RegModeListValue = map[pbCom.RegMode]string{
	pbCom.RegMode_Reg_Lasso: RegModeL1,
	pbCom.RegMode_Reg_Ridge: RegModeL2,
}

// FLInfo used to parse the content contained in the extra field of the file on the chain,
// only files that can be parsed can be used for task training or prediction
//用于解析链上文件额外字段中包含的内容，
//只有可以解析的文件才能用于任务训练或预测
type FLInfo struct {
	FileType  string `json:"fileType"`  // file type, only supports "csv" 文件类型，仅支持“csv”
	Features  string `json:"features"`  // feature list 功能列表
	TotalRows int64  `json:"totalRows"` // total number of samples 样品总数
}

// ExecutorNode has access to samples with which to train models or to predict,
//  and starts task that multi parties execute synchronically
//可以访问用于训练模型或预测的样本，
//并启动多方同步执行的任务
type ExecutorNode struct {
	ID              []byte `json:"id"`
	Name            string `json:"name"`
	Address         string `json:"address"`     // local grpc host 本地 grpc 主机
	HttpAddress     string `json:"httpAddress"` // local http host 本地 HTTP 主机
	PaddleFLAddress string `json:"paddleFLAddress"`
	PaddleFLRole    int    `json:"paddleFLRole"`
	RegTime         int64  `json:"regTime"` // node registering time 节点注册时间
}

type ExecutorNodes []ExecutorNode

// FLTask defines Federated Learning Task based on MPC 定义基于 MPC 的联邦学习任务
type FLTask *pbTask.FLTask

type FLTasks []*pbTask.FLTask

// PublishFLTaskOptions contains parameters for publishing tasks 发布FLTaskOptions包含用于发布任务的参数
type PublishFLTaskOptions struct {
	FLTask    FLTask `json:"fLTask"`
	Signature []byte `json:"signature"`
}

// StartFLTaskOptions contains parameters for the requester to start tasks 包含请求者启动任务的参数
type StartFLTaskOptions struct {
	TaskID    string `json:"taskID"`
	Signature []byte `json:"signature"`
}

// ListFLTaskOptions contains parameters for listing tasks
// support listing tasks a requester published or tasks an executor involved
//包含用于列出任务的参数
//支持列出请求者发布的任务或执行者涉及的任务
type ListFLTaskOptions struct {
	PubKey     []byte `json:"pubKey"`    // requester's public key 请求者的公钥
	ExecPubKey []byte `json:"exePubKey"` // executor's public key 执行人的公钥
	Status     string `json:"status"`    // task status 任务状态
	TimeStart  int64  `json:"timeStart"` // task publish time period, only task published after TimeStart and before TimeEnd will be listed 任务发布时间段，仅列出在 TimeStart 之后和 TimeEnd 之前发布的任务
	TimeEnd    int64  `json:"timeEnd"`
	Limit      int64  `json:"limit"` // limit number of tasks in list request, default 'all' 限制列表请求中的任务数，默认为“全部”
}

// FLTaskConfirmOptions contains parameters for confirming task 包含用于确认任务的参数
type FLTaskConfirmOptions struct {
	Pubkey       []byte `json:"pubkey"` // one of the task executor's public key 任务执行者的公钥之一
	TaskID       string `json:"taskID"`
	RejectReason string `json:"rejectReason"` // reason of the rejected task 任务被拒绝的原因
	CurrentTime  int64  `json:"currentTime"`  // time when confirming task 确认任务的时间

	Signature []byte `json:"signature"` // executor's signature 遗嘱执行人的签名
}

// FLTaskExeStatusOptions contains parameters for updating executing task 包含用于更新执行任务的参数
type FLTaskExeStatusOptions struct {
	Executor    []byte `json:"executor"`
	TaskID      string `json:"taskID"`
	CurrentTime int64  `json:"currentTime"` // task execute start time or finish time 任务执行开始时间或完成时间

	Signature []byte `json:"signature"`

	ErrMessage string `json:"errMessage"` // for failed task 对于失败的任务
	Result     string `json:"result"`     // for finished task 对于已完成的任务
}

// AddNodeOptions contains parameters for adding node of Executor 包含用于添加执行器节点的参数
type AddNodeOptions struct {
	Node      ExecutorNode `json:"node"`
	Signature []byte       `json:"signature"`
}
