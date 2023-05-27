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

package common

import (
	"fmt"
	"strconv"
)

// ImportFeaturesForLinReg import linear regression features from file 从文件导入线性回归要素
func ImportFeaturesForLinReg(fileRows [][]string) ([]*DataFeature, error) {
	if fileRows == nil {
		return nil, fmt.Errorf("empty file content")
	}

	// read the first row to get all features 阅读第一行以获取所有功能
	featureNum := len(fileRows[0])
	features := make([]*DataFeature, featureNum)
	for i := 0; i < featureNum; i++ {
		features[i] = new(DataFeature)
		features[i].Sets = make(map[int]float64)
		features[i].FeatureName = fileRows[0][i]
	}

	// read from all rows to get feature values 从所有行中读取以获取特征值
	sample := 0
	for row := 1; row < len(fileRows); row++ {
		for i := 0; i < featureNum; i++ {
			value, err := strconv.ParseFloat(fileRows[row][i], 64)
			if err != nil {
				return nil, fmt.Errorf("failed to parse value, err: %v", err)
			}
			features[i].Sets[sample] = value
		}
		sample++
	}

	return features, nil
}

// ImportFeaturesForLogReg import logic regression features from file, target variable imported as 1 or 0
// - fileRows file rows, first row is feature list
// - label target feature
// - labelName target variable
//从文件导入逻辑回归特征，目标变量导入为 1 或 0
//- 文件行文件行，第一行是功能列表
//- 标签目标特征
//- 标签名称目标变量
func ImportFeaturesForLogReg(fileRows [][]string, label, labelName string) ([]*DataFeature, error) {
	if fileRows == nil {
		return nil, fmt.Errorf("empty file content")
	}

	// read the first row to get all features 阅读第一行以获取所有功能
	featureNum := len(fileRows[0])
	features := make([]*DataFeature, featureNum)
	for i := 0; i < featureNum; i++ {
		features[i] = new(DataFeature)
		features[i].Sets = make(map[int]float64)
		features[i].FeatureName = fileRows[0][i]
	}

	// read from all rows to get feature values 从所有行中读取以获取特征值
	sample := 0
	for row := 1; row < len(fileRows); row++ {
		for i := 0; i < featureNum; i++ {
			if features[i].FeatureName == label {
				// parse target feature variable to 0 or 1
				if fileRows[row][i] == labelName {
					features[i].Sets[sample] = 1.0
				} else {
					features[i].Sets[sample] = 0.0
				}
			} else {
				value, err := strconv.ParseFloat(fileRows[row][i], 64)
				if err != nil {
					return nil, fmt.Errorf("failed to parse value, err: %v", err)
				}
				features[i].Sets[sample] = value
			}
		}

		sample++
	}

	return features, nil
}

// ImportFeaturesForDT import decision tree features from file
// - fileRows file rows, first row is feature list
//从文件导入决策树特征
//- 文件行文件行，第一行是功能列表
func ImportFeaturesForDT(fileRows [][]string) ([]*DTDataFeature, error) {
	if fileRows == nil {
		return nil, fmt.Errorf("empty file content")
	}

	// read the first row to get all features
	featureNum := len(fileRows[0])
	features := make([]*DTDataFeature, featureNum)
	for i := 0; i < featureNum; i++ {
		features[i] = new(DTDataFeature)
		features[i].Sets = make(map[int]string)
		features[i].FeatureName = fileRows[0][i]
	}

	// read from all rows to get feature values  从所有行中读取以获取特征值
	sample := 0
	for row := 1; row < len(fileRows); row++ {
		for i := 0; i < featureNum; i++ {
			features[i].Sets[sample] = fileRows[row][i]
		}
		sample++
	}

	return features, nil
}
