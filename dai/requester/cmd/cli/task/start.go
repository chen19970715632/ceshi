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

package task

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	requestClient "github.com/PaddlePaddle/PaddleDTX/dai/requester/client"
	"github.com/PaddlePaddle/PaddleDTX/dai/util/file"
)

// startTaskByIDCmd starts a confirmed task
var startTaskByIDCmd = &cobra.Command{
	Use:   "start",
	Short: "start the confirmed task",
	Run: func(cmd *cobra.Command, args []string) {
		client, err := requestClient.GetRequestClient(configPath)
		if err != nil {
			fmt.Printf("GetRequestClient failed: %v\n", err)
			return
		}
		if privateKey == "" {
			privateKeyBytes, err := file.ReadFile(keyPath, file.PrivateKeyFileName)
			if err != nil {
				fmt.Printf("Read privateKey failed, err: %v\n", err)
				return
			}
			privateKey = strings.TrimSpace(string(privateKeyBytes))
		}

		if err := client.StartTask(privateKey, id); err != nil {
			fmt.Printf("StartTask failed：%v\n", err)
			return
		}
		fmt.Println("OK")
	},
}

func init() {
	rootCmd.AddCommand(startTaskByIDCmd)

	startTaskByIDCmd.Flags().StringVarP(&privateKey, "privkey", "k", "", "requester private key hex string")
	startTaskByIDCmd.Flags().StringVarP(&keyPath, "keyPath", "", "./reqkeys", "key path")
	startTaskByIDCmd.Flags().StringVarP(&id, "id", "i", "", "id of task to start, but only Ready and Failed tasks can be started")

	startTaskByIDCmd.MarkFlagRequired("id")
}
