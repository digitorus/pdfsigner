// Copyright © 2017 NAME HERE <EMAIL ADDRESS>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"bitbucket.org/digitorus/pdfsigner/priority_queue"
	"bitbucket.org/digitorus/pdfsigner/signer"
	"github.com/spf13/cobra"
)

// multiCmd represents the mixed command
var multiCmd = &cobra.Command{
	Use:   "mixed",
	Short: "A brief description of your command",
	Long:  `A longer description that spans multiple lines`,
	Run: func(cmd *cobra.Command, serviceNames []string) {
		setupServices(serviceNames)
	},
}

type job struct {
	ID   string
	file string // tmp or real file location
}

func setupServices(serviceNames []string) {
	for _, n := range serviceNames {
		// get service by name
		service := getConfigServiceByName(n)

		// get signer by name
		c := getConfigSignerByName(service.Signer)

		// set sign data
		switch c.Type {
		case "pem":
			c.SignData.SetPEM(c.CrtPath, c.KeyPath, c.CrtChainPath)
		case "pksc11":
			c.SignData.SetPKSC11(c.LibPath, c.Pass, c.CrtChainPath)
		}

		// create signer
		s := qSigner{
			pq:       priority_queue.NewPriorityQueue(10),
			signData: c.SignData,
		}

		// set jobs
		if service.Type == "watch" {

			watch(service.In, func(filePath string) {
				//c.SignData, service.In, service.Out
				i := priority_queue.Item{
					Value: job{
						file: filePath,
					},
					Priority: priority_queue.LowPriority,
				}
				s.pq.Push(i)
			})
		} else if service.Type == "serve" {
			// handle serve
		}

		signers = append(signers)
	}
}

var signers []qSigner

type qSigner struct {
	pq       *priority_queue.PriorityQueue
	signData signer.SignData
}


func init() {
	RootCmd.AddCommand(multiCmd)
}
