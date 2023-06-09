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

package config

import (
	"strings"

	"github.com/spf13/viper"

	"github.com/PaddlePaddle/PaddleDTX/dai/util/file"
)

var (
	logConf      *Log
	executorConf *ExecutorConf
	cliConf      *ExecutorBlockchainConf
)

// ExecutorConf defines the configuration info required for excutor node startup,
// and convert it to a struct by parsing 'conf/config.toml'.
//定义执行器节点启动所需的配置信息，
//并通过解析“conf/config.toml”将其转换为结构体。
type ExecutorConf struct {
	Name            string // executor node name 执行程序节点名称
	ListenAddress   string // the port on which the executor node is listening 执行程序节点正在侦听的端口
	PublicAddress   string // local grpc host 本地 grpc 主机
	PrivateKey      string // private key 私钥
	PaddleFLAddress string
	PaddleFLRole    int
	KeyPath         string            // key path, include private key and public key 密钥路径，包括私钥和公钥
	HttpServer      *HttpServerConf   // include executor node's httpserver configuration 包括执行程序节点的 HTTP 服务器配置
	Mode            *ExecutorModeConf // the task execution type 任务执行类型
	Mpc             *ExecutorMpcConf
	Storage         *ExecutorStorageConf // model storage and prediction results storage 模型存储和预测结果存储
	Blockchain      *ExecutorBlockchainConf
}

// HttpServerConf defines the configuration required to start the executor node's httpserver
// 'AllowCros' decides whether to allow cross-domain requests, the default is false
//定义启动执行程序节点的 HTTP 服务器所需的配置
//“AllowCros”决定是否允许跨域请求，默认值为 false
type HttpServerConf struct {
	Switch      string
	HttpAddress string
	HttpPort    string
	AllowCros   bool
}

// ExecutorModeConf defines the task execution type, such as proxy-execution or self-execution.
// "Self" is suitable for the executor node and the dataOwner node are the same organization and execute by themselves,
// and the executor node can download sample files from the dataOwner node without permission application.
//定义任务执行类型，例如代理执行或自行执行。
//“Self”适用于执行节点和dataOwner节点是同一个组织，由他们自己执行，
//并且执行器节点可以在没有权限的情况下从 dataOwner 节点下载示例文件应用程序。
type ExecutorModeConf struct {
	Type string
	Self *XuperDBConf
}

// ExecutorMpcConf defines the features of the mpc process 定义 MPC 过程的功能
type ExecutorMpcConf struct {
	TrainTaskLimit   int
	PredictTaskLimit int
	RpcTimeout       int // rpc request timeout between executor nodes 执行程序节点之间的 RPC 请求超时
	TaskLimitTime    int
}

// ExecutorStorageConf defines the storage used by the executor,
// include model storage and prediction results storage, and evaluation storage and live evaluation storage.
// the prediction results storage support 'XuperDB' and 'Local' two storage mode.
//定义执行程序使用的存储，
//包括模型存储和预测结果存储，以及评估存储和实时评估存储。
//预测结果存储支持“XuperDB”和“本地”两种存储模式。
type ExecutorStorageConf struct {
	Type                       string
	LocalModelStoragePath      string
	LocalEvaluationStoragePath string
	LiveEvaluationStoragePath  string // live evaluation results storage path 实时评估结果存储路径
	XuperDB                    *XuperDBConf
	Local                      *PredictLocalConf
}

// XuperDBConf defines the XuperDB's endpoint, used to upload or download files 定义 XuperDB 的端点，用于上传或下载文件
type XuperDBConf struct {
	PrivateKey string
	Host       string
	KeyPath    string
	NameSpace  string
	ExpireTime int64
}

// PredictLocalConf defines the local path of prediction results storage 定义预测结果存储的本地路径
type PredictLocalConf struct {
	LocalPredictStoragePath string
}

// ExecutorBlockchainConf defines the configuration required to invoke blockchain contracts 定义调用区块链合约所需的配置
type ExecutorBlockchainConf struct {
	Type   string
	Xchain *XchainConf
	Fabric *FabricConf
}

type XchainConf struct {
	Mnemonic        string
	ContractName    string
	ContractAccount string
	ChainAddress    string
	ChainName       string
}

type FabricConf struct {
	ConfigFile string
	ChannelID  string
	Chaincode  string
	UserName   string
	OrgName    string
}

// Log defines the storage path of the logs generated by the executor node at runtime 定义执行器节点在运行时生成的日志的存储路径
type Log struct {
	Level string
	Path  string
}

// InitConfig parses configuration file 解析配置文件
func InitConfig(configPath string) error {
	v := viper.New()
	v.SetConfigFile(configPath)
	if err := v.ReadInConfig(); err != nil {
		return err
	}
	logConf = new(Log)
	err := v.Sub("log").Unmarshal(logConf)
	if err != nil {
		return err
	}
	executorConf = new(ExecutorConf)
	err = v.Sub("executor").Unmarshal(executorConf)
	if err != nil {
		return err
	}
	// get the private key , if the private key does not exist, read it from 'keyPath' 获取私钥，如果私钥不存在，则从“keyPath”读取
	if executorConf.PrivateKey == "" {
		privateKeyBytes, err := file.ReadFile(executorConf.KeyPath, file.PrivateKeyFileName)
		if err == nil && len(privateKeyBytes) != 0 {
			executorConf.PrivateKey = strings.TrimSpace(string(privateKeyBytes))
		} else {
			return err
		}
	}
	return nil
}

// InitCliConfig parses client configuration file. if cli's configuration file is not existed, use executor's configuration file. 解析客户端配置文件。如果 CLI 的配置文件不存在，请使用执行程序的配置文件。
func InitCliConfig(configPath string) error {
	v := viper.New()
	v.SetConfigFile(configPath)
	if err := v.ReadInConfig(); err != nil {
		return err
	}
	innerV := v.Sub("blockchain")
	if innerV != nil {
		// If "blockchain" was existed, cli would use the configuration of cli. 如果存在“区块链”，cli将使用cli的配置。
		cliConf = new(ExecutorBlockchainConf)
		err := innerV.Unmarshal(cliConf)
		if err != nil {
			return err
		}
		return nil
	} else {
		// If "blockchain" wasn't existed, use the configuration of the executor.
		//如果“区块链”不存在，请使用执行程序的配置。
		err := InitConfig(configPath)
		if err == nil {
			cliConf = executorConf.Blockchain
		}
		return err
	}
}

// GetExecutorConf returns all configuration of the executor 返回执行器的所有配置
func GetExecutorConf() *ExecutorConf {
	return executorConf
}

// GetLogConf returns log configuration of the executor 返回执行程序的日志配置
func GetLogConf() *Log {
	return logConf
}

// GetCliConf returns blockchain configuration of the executor 返回执行程序的区块链配置
func GetCliConf() *ExecutorBlockchainConf {
	return cliConf
}
