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

package xuperdb

import (
	"context"
	"io"
	"time"

	"github.com/PaddlePaddle/PaddleDTX/crypto/core/ecdsa"
	httpclient "github.com/PaddlePaddle/PaddleDTX/xdb/client/http"
)

// Maximum default time for saving predict file results
const DefaultFileRetentionTime = time.Hour * 72

// XuperDB defines xuperdb client
// Only the prediction files supports storing to xuperdb
type XuperDB struct {
	PrivateKey ecdsa.PrivateKey // the private key is the dataOwner node client private key generated by the executor
	Address    string           // the dataOwner node host
	Ns         string           // it defines which namespace in XuperDB prediction file will be stored
	ExpireTime int64            // the expiration time of the files stored in XuperDB
}

// New initiates xuperDB Storage
func New(expireTime int64, ns, host string, privateKey ecdsa.PrivateKey) *XuperDB {
	expiretime := time.Duration(expireTime) * time.Hour
	if expiretime == 0 {
		expiretime = DefaultFileRetentionTime
	}
	return &XuperDB{
		PrivateKey: privateKey,
		Address:    host,
		Ns:         ns,
		ExpireTime: time.Now().UnixNano() + expiretime.Nanoseconds(),
	}
}

// Write stores files in xuperDB, name is prediction task's ID
// return the id of the file stored in xuperdb
func (x *XuperDB) Write(r io.Reader, name string) (string, error) {
	// new xuperdb http client
	client, err := httpclient.New(x.Address)
	if err != nil {
		return "", err
	}
	// FileName is prediction task's ID, e.g. 'f581c9ef-778f-4d15-87ae-26ba6da93b86.csv'
	// Description default setting "store samples"
	opt := httpclient.WriteOptions{
		PrivateKey: x.PrivateKey.String(),

		Namespace:   x.Ns,
		FileName:    name + ".csv",
		ExpireTime:  x.ExpireTime,
		Description: "store samples",
	}
	// request the dataOwner node to upload prediction file
	resp, err := client.Write(context.Background(), r, opt)
	if err != nil {
		return "", err
	}
	return resp.FileID, nil
}

// Read gets files from xuperDB
func (x *XuperDB) Read(fileID string) (io.ReadCloser, error) {
	client, err := httpclient.New(x.Address)
	if err != nil {
		return nil, err
	}

	opt := httpclient.ReadOptions{
		PrivateKey: x.PrivateKey.String(),
		FileID:     fileID,
	}
	// request the dataOwner node to download prediction file
	reader, err := client.Read(context.Background(), opt)
	if err != nil {
		return nil, err
	}
	return reader, nil
}
