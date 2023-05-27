## Distributed AI
任务执行节点提供的GRPC接口如下：
``` proto
syntax = "proto3";

import "common/common.proto";

package task;
option go_package = "github.com/PaddlePaddle/PaddleDTX/dai/protos/task";

// Cluster defines communication communication between client and server, and communication between cluster members.
service Task {
    // ListTask is provided by Executor server for Executor client to list tasks with filters.
    rpc ListTask(ListTaskRequest)  returns (FLTasks);
    // GetTaskById is provided by Executor server for Executor client to query a task.
    rpc GetTaskById(GetTaskRequest) returns (FLTask);
    // GetPredictResult is provided by Executor server for Executor client to get prediction result.
    rpc GetPredictResult(TaskRequest) returns (PredictResponse);
}

// ListTaskRequest is message sent to Executor server to list tasks
message ListTaskRequest {
    bytes pubKey = 1;  // requester or executor's public key
    string status = 2;
    int64 timeStart = 3;
    int64 timeEnd = 4;
    int64 limit = 5;
}

// DataForTask is a message received from Executor
message DataForTask {
    bytes  owner = 1;  // samples'owner, also dataOwner who confirms executor's authorization application for the use of samples
	bytes  executor = 2; // executor who process tasks
	string dataID = 3; // file id of samples
	string PSILabel = 4 ; // psi lable for vertical learning
	int64  confirmedAt = 5; // task confirm time
	int64  rejectedAt = 6; // task reject time
    string address = 7; // host of Executor 
    bool isTagPart = 8;   
}

// FLTask is a message received from Executor and defines Federated Learning Task based on MPC
message FLTask {
    string iD = 1;
    string name = 2;
    string description = 3;

    bytes requester = 4;
    repeated DataForTask dataSets = 5;
	common.TaskParams algoParam = 6; // fl algorithm related params

	string status = 7;
	string errMessage = 8;
	string result = 9;
	int64 publishTime = 10;
	int64 startTime = 11;
	int64 endTime = 12;
}

// FLTasks is list of FLTasks received from Executor 
message FLTasks {
    repeated FLTask fLTasks = 1;
}

// GetTaskRequest is message sent to Executor server to get a task
message GetTaskRequest {
    string taskID = 1;
}

// PredictResponse is a message received from Executor 
message PredictResponse {
    string taskID = 1;
    bytes payload = 2; 
}

```
