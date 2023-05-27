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
	"context"
	"encoding/hex"
	"time"

	"github.com/PaddlePaddle/PaddleDTX/crypto/core/ecdsa"
	"github.com/PaddlePaddle/PaddleDTX/crypto/core/hash"
	"github.com/sirupsen/logrus"

	"github.com/PaddlePaddle/PaddleDTX/xdb/blockchain"
	"github.com/PaddlePaddle/PaddleDTX/xdb/engine/common"
	util "github.com/PaddlePaddle/PaddleDTX/xdb/pkgs/strings"
)

// heartbeat sends heartbeats regularly in order to claim it's alive
func (m *NodeMaintainer) heartbeat(ctx context.Context) {
	pubkey := ecdsa.PublicKeyFromPrivateKey(m.localNode.PrivateKey)

	l := logger.WithField("runner", "heartbeat loop")
	defer l.Info("runner stopped")

	ticker := time.NewTicker(m.heartbeatInterval)
	defer ticker.Stop()

	m.doneHbC = make(chan struct{})
	defer close(m.doneHbC)

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
		// invoke contract
		timestamp := time.Now().UnixNano()
		opt := &blockchain.NodeHeartBeatOptions{
			NodeID:        []byte(pubkey.String()),
			CurrentTime:   timestamp,
			BeginningTime: common.TodayBeginning(timestamp),
		}
		msg, err := util.GetSigMessage(opt)
		if err != nil {
			l.WithError(err).Warn("failed to get the message to sign for heartbeat")
			continue
		}
		sig, err := ecdsa.Sign(m.localNode.PrivateKey, hash.HashUsingSha256([]byte(msg)))
		if err != nil {
			l.WithError(err).Warn("failed to sign heartbeat")
			continue
		}
		opt.Signature = sig[:]

		if err := m.blockchain.Heartbeat(opt); err != nil {
			l.WithError(err).Warn("failed to update heartbeat")
			continue
		}

		l.WithFields(logrus.Fields{
			"target_node": hex.EncodeToString(pubkey[:4]),
			"update_at":   timestamp,
		}).Info("successfully updated heartbeat of node")
	}

}
