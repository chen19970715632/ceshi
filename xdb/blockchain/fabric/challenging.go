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

package fabric

import (
	"encoding/json"
	"strconv"

	"github.com/PaddlePaddle/PaddleDTX/xdb/blockchain"
	"github.com/PaddlePaddle/PaddleDTX/xdb/errorx"
)

// ListChallengeRequests lists challenge requests on chain
// filter condition may include fileOwner, storageNode, fileID, challenge status, time period and items number limit
func (f *Fabric) ListChallengeRequests(opt *blockchain.ListChallengeOptions) ([]blockchain.Challenge, error) {

	opts, err := json.Marshal(*opt)
	if err != nil {
		return nil, errorx.NewCode(err, errorx.ErrCodeInternal, "failed to marshal ListChallengeOptions")
	}

	s, err := f.QueryContract([][]byte{opts}, "ListChallengeRequests")
	if err != nil {
		return nil, err
	}
	var cs []blockchain.Challenge
	if err = json.Unmarshal(s, &cs); err != nil {
		return nil, errorx.NewCode(err, errorx.ErrCodeInternal,
			"failed to unmarshal Challenges")
	}
	return cs, nil
}

// ChallengeRequest sets a dataowner's challenge request on chain
func (f *Fabric) ChallengeRequest(opt *blockchain.ChallengeRequestOptions) error {

	opts, err := json.Marshal(*opt)
	if err != nil {
		return errorx.NewCode(err, errorx.ErrCodeInternal, "failed to marshal ChallengeRequestOptions")
	}

	if _, err = f.InvokeContract([][]byte{opts}, "ChallengeRequest"); err != nil {
		return err
	}
	return nil
}

// ChallengeAnswer sets a storage node's challenge answer onto blockchain
func (f *Fabric) ChallengeAnswer(opt *blockchain.ChallengeAnswerOptions) ([]byte, error) {
	opts, err := json.Marshal(*opt)
	if err != nil {
		return nil, errorx.NewCode(err, errorx.ErrCodeInternal, "failed to marshal ChallengeAnswerOptions")
	}

	resp, err := f.InvokeContract([][]byte{opts}, "ChallengeAnswer")
	if err != nil {
		return nil, err
	}
	return resp, nil
}

// GetChallengeByID gets a challenge by challengeID
func (f *Fabric) GetChallengeByID(id string) (blockchain.Challenge, error) {
	var c blockchain.Challenge
	s, err := f.QueryContract([][]byte{[]byte(id)}, "GetChallengeByID")
	if err != nil {
		return c, err
	}

	if err = json.Unmarshal(s, &c); err != nil {
		return c, errorx.NewCode(err, errorx.ErrCodeInternal, "failed to unmarshal Challenge")
	}

	return c, nil
}

// GetChallengeNum gets challenge number given filter condition
// filter condition may include storageNode, challenge status, time period
func (f *Fabric) GetChallengeNum(opt *blockchain.GetChallengeNumOptions) (uint64, error) {

	opts, err := json.Marshal(*opt)
	if err != nil {
		return 0, errorx.NewCode(err, errorx.ErrCodeInternal, "failed to marshal GetChallengeNumOptions")
	}

	s, err := f.QueryContract([][]byte{opts}, "GetChallengeNum")
	if err != nil {
		return 0, err
	}

	num, err := strconv.ParseUint(string(s), 10, 64)
	if err != nil {
		return 0, errorx.NewCode(err, errorx.ErrCodeInternal,
			"failed to parse contract response to number")
	}
	return num, nil
}
