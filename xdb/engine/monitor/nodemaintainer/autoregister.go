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

package nodemaintainer

import (
	"strconv"
	"time"

	"github.com/PaddlePaddle/PaddleDTX/crypto/core/ecdsa"
	"github.com/PaddlePaddle/PaddleDTX/crypto/core/hash"
	"github.com/sirupsen/logrus"

	"github.com/PaddlePaddle/PaddleDTX/xdb/blockchain"
	"github.com/PaddlePaddle/PaddleDTX/xdb/errorx"
	util "github.com/PaddlePaddle/PaddleDTX/xdb/pkgs/strings"
)

// autoRegister storage-node automatically register in blockchain
func (m *NodeMaintainer) autoRegister() error {
	logrus.WithField("module", "autoRegister")

	pubkey := ecdsa.PublicKeyFromPrivateKey(m.localNode.PrivateKey)
	node, err := m.blockchain.GetNode([]byte(pubkey.String()))
	if err == nil && node.Online {
		logrus.Info("node already registered on blockchain")
		return nil
	} else if err == nil && !node.Online {
		// get sig message
		nonce := time.Now().UnixNano()
		msg, err := util.GetSigMessage(map[string]string{
			"nodeID": pubkey.String(),
			"nonce":  strconv.FormatInt(nonce, 10),
		})
		if err != nil {
			return errorx.Internal(err, "failed to get the message to sign")
		}
		// generate signature
		sig, err := ecdsa.Sign(m.localNode.PrivateKey, hash.HashUsingSha256([]byte(msg)))
		if err != nil {
			return errorx.Wrap(err, "failed to sign File")
		}
		nodeOpts := &blockchain.NodeOperateOptions{
			NodeID:    []byte(pubkey.String()),
			Nonce:     nonce,
			Signature: sig[:],
		}
		err = m.blockchain.NodeOnline(nodeOpts)
		if err != nil {
			logrus.Error("node failed online on  blockchain")
		}
		logrus.Info("node online")
		return err
	} else if errorx.Is(err, errorx.ErrCodeNotFound) {
		timestamp := time.Now().UnixNano()
		opt := blockchain.AddNodeOptions{
			Node: blockchain.Node{
				ID:       []byte(pubkey.String()),
				Name:     m.localNode.Name,
				Address:  m.localNode.Address,
				Online:   true,
				RegTime:  timestamp,
				UpdateAt: timestamp,
			},
		}
		msg, err := util.GetSigMessage(opt)
		if err != nil {
			return errorx.Internal(err, "failed to get the message to sign")
		}
		sig, err := ecdsa.Sign(m.localNode.PrivateKey, hash.HashUsingSha256([]byte(msg)))
		if err != nil {
			return errorx.Wrap(err, "failed to sign node")
		}
		opt.Signature = sig[:]
		if err := m.blockchain.AddNode(&opt); err != nil {
			logrus.Error("failed to register node automatically")
			return errorx.Wrap(err, "failed to register node automatically")
		}
		logrus.WithFields(logrus.Fields{
			"node_id":   pubkey.String(),
			"online_at": timestamp,
		}).Info("successfully registered node on blockchain")

		return nil
	} else {
		logrus.Errorf("failed to read blockchain: %v", err)
		return errorx.Wrap(err, "failed to read blockchain")
	}
}
