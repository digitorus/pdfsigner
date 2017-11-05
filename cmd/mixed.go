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
	"github.com/spf13/cobra"
)

// mixedCmd represents the mixed command
var mixedCmd = &cobra.Command{
	Use:   "mixed",
	Short: "A brief description of your command",
	Long:  `A longer description that spans multiple lines`,
	Run: func(cmd *cobra.Command, serviceNames []string) {
		for _, n := range serviceNames {
			service := getConfigServiceByName(n)
			if service.Type == "watch" {
				c := getConfigSignerByName(service.Signer)

				switch c.Type {
				case "pem":
					c.SignData.SetPEM(c.CrtPath, c.KeyPath, c.CrtChainPath)
				case "pksc11":
					c.SignData.SetPKSC11(c.LibPath, c.Pass, c.CrtChainPath)
				}

				watch(c.SignData, service.In, service.Out)
			} else if service.Type == "serve" {
				// handle serve
			}

		}
	},
}

func init() {
	RootCmd.AddCommand(mixedCmd)
}
