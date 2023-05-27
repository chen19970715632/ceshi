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

package pairing

import (
	"math/big"

	bls12_381_ecc "github.com/consensys/gnark-crypto/ecc/bls12-381"
)

var (
	g1Gen bls12_381_ecc.G1Affine
	g2Gen bls12_381_ecc.G2Affine
	order *big.Int
)

// PrivateKey pairing based challenge private key 基于配对的质询私钥
type PrivateKey struct {
	X *big.Int
}

// PublicKey pairing based challenge public key 基于配对的质询公钥
type PublicKey struct {
	P *bls12_381_ecc.G2Affine
}

// CalculateSigmaIParams parameters required to calculate sigma_i for each segment 计算每个段的sigma_i所需的参数
type CalculateSigmaIParams struct {
	Content []byte      // file content 文件内容
	Index   *big.Int    // file segment index 文件段索引
	RandomV *big.Int    // a random V 随机 V
	RandomU *big.Int    // a random U 随机 U
	Privkey *PrivateKey // client private key 客户端私钥
	Round   int64       // challenge round 挑战赛
}

// ProofParams parameters required to generate proof 生成证明所需的参数
type ProofParams struct {
	Content       [][]byte                  // file contents 文件内容
	Indices       []*big.Int                // {i} index list {i} 索引列表
	RandomVs      []*big.Int                // {v_i} random challenge number list {v_i} 随机挑战号码列表
	Sigmas        []*bls12_381_ecc.G1Affine // {sigma_i} list in storage {sigma_i} 存储中的列表
	RandThisRound []byte                    // random number for this challenge round 本轮挑战赛的随机数
}

// VerifyParams parameters required to verify a proof 验证证明所需的参数
type VerifyParams struct {
	Sigma    *bls12_381_ecc.G1Affine // sigma in proof 证明中的西格玛
	Mu       *bls12_381_ecc.G1Affine // mu in proof 亩在证明
	RandomV  *big.Int                // a random V 随机 V
	RandomU  *big.Int                // a random U 一个随机的U
	Indices  []*big.Int              // {i} index list {i} 索引列表
	RandomVs []*big.Int              // {v_i} random challenge number list {v_i} 随机挑战号码列表
	Pubkey   *PublicKey              // client public key 客户端公钥
}

// PrivateKeyToByte convert private key to byes 将私钥转换为再见
func PrivateKeyToByte(privkey *PrivateKey) []byte {
	return privkey.X.Bytes()
}

// PrivateKeyFromByte retrieve private key from byes 从再见中检索私钥
func PrivateKeyFromByte(privkey []byte) *PrivateKey {
	x := new(big.Int).SetBytes(privkey)
	return &PrivateKey{
		X: x,
	}
}

// PublicKeyToByte convert public key to byes 将公钥转换为再见
func PublicKeyToByte(pubkey *PublicKey) []byte {
	return pubkey.P.Marshal()
}

// PublicKeyFromByte retrieve public key from byes 从再见中检索公钥
func PublicKeyFromByte(pubkey []byte) (*PublicKey, error) {
	pub := new(bls12_381_ecc.G2Affine)
	if err := pub.Unmarshal(pubkey); err != nil {
		return nil, err
	}
	return &PublicKey{
		P: pub,
	}, nil
}

// CalculateSigmaIParamsFromBytes retrieve CalculateSigmaIParams from bytes 从字节中检索计算西格玛IParams
func CalculateSigmaIParamsFromBytes(content, index, randomV, randomU, privkey []byte, round int64) CalculateSigmaIParams {
	return CalculateSigmaIParams{
		Content: content,
		Index:   new(big.Int).SetBytes(index),
		RandomV: new(big.Int).SetBytes(randomV),
		RandomU: new(big.Int).SetBytes(randomU),
		Privkey: PrivateKeyFromByte(privkey),
		Round:   round,
	}
}

// G1ToByte convert G1 point to byes 转换 G1 点 为 再见
func G1ToByte(sigma *bls12_381_ecc.G1Affine) []byte {
	return sigma.Marshal()
}

// G1FromByte retrieve G1 point from bytes 从字节中检索 G1 点
func G1FromByte(sigma []byte) (*bls12_381_ecc.G1Affine, error) {
	s := new(bls12_381_ecc.G1Affine)
	if err := s.Unmarshal(sigma); err != nil {
		return nil, err
	}
	return s, nil
}

// G1sFromBytes retrieve G1 point list from bytes 从字节中检索 G1 点列表
func G1sFromBytes(gs [][]byte) ([]*bls12_381_ecc.G1Affine, error) {
	var ret []*bls12_381_ecc.G1Affine
	for _, g := range gs {
		g1 := new(bls12_381_ecc.G1Affine)
		if err := g1.Unmarshal(g); err != nil {
			return nil, err
		}
		ret = append(ret, g1)
	}
	return ret, nil
}

// IntListToBytes convert bit int list to bytes 将位整数列表转换为字节
func IntListToBytes(intList []*big.Int) [][]byte {
	var ret [][]byte
	for _, n := range intList {
		ret = append(ret, n.Bytes())
	}
	return ret
}

// IntListFromBytes retrieve bit int list from bytes 从字节中检索位整数列表
func IntListFromBytes(intList [][]byte) []*big.Int {
	var ret []*big.Int
	for _, n := range intList {
		ret = append(ret, new(big.Int).SetBytes(n))
	}
	return ret
}

// ProofParamsFromBytes retrieve ProofParams from bytes 从字节中检索证明参数
func ProofParamsFromBytes(content, indices, randVs, sigmas [][]byte, rand []byte) (ProofParams, error) {
	s, err := G1sFromBytes(sigmas)
	if err != nil {
		return ProofParams{}, err
	}
	return ProofParams{
		Content:       content,
		Indices:       IntListFromBytes(indices),
		RandomVs:      IntListFromBytes(randVs),
		Sigmas:        s,
		RandThisRound: rand,
	}, nil
}

// VerifyParamsFromBytes retrieve VerifyParams from bytes 从字节中检索验证参数
func VerifyParamsFromBytes(sigma, mu, randV, randU, pubkey []byte, indices, randVs [][]byte) (VerifyParams, error) {
	s, err := G1FromByte(sigma)
	if err != nil {
		return VerifyParams{}, err
	}
	m, err := G1FromByte(mu)
	if err != nil {
		return VerifyParams{}, err
	}
	pub, err := PublicKeyFromByte(pubkey)
	if err != nil {
		return VerifyParams{}, err
	}

	return VerifyParams{
		Sigma:    s,
		Mu:       m,
		RandomV:  new(big.Int).SetBytes(randV),
		RandomU:  new(big.Int).SetBytes(randU),
		Indices:  IntListFromBytes(indices),
		RandomVs: IntListFromBytes(randVs),
		Pubkey:   pub,
	}, nil
}
