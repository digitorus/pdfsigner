// Copyright © 2018 NAME HERE <EMAIL ADDRESS>
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
	"bufio"
	"log"
	"os"

	"bitbucket.org/digitorus/pdfsigner/db"
	"bitbucket.org/digitorus/pdfsigner/license"
	"github.com/spf13/cobra"
)

// licenseCmd represents the license command
var licenseCmd = &cobra.Command{
	Use:   "license",
	Short: "Update license",
	Run: func(cmd *cobra.Command, args []string) {
		err := loadDB()
		if err != nil {
			err := readStdIn()
			if err != nil {
				log.Fatal(err)
			}
		}
	},
}

func init() {
	RootCmd.AddCommand(licenseCmd)
}

func readStdIn() error {
	// read license from input
	licenseBytes, err := bufio.NewReader(os.Stdin).ReadBytes('\n')
	if err != nil {
		return err
	}

	// load license data
	err = license.ExtractLicense(licenseBytes)
	if err != nil {
		return err
	}

	// save license to db
	err = db.SaveByKey("license", licenseBytes)
	if err != nil {
		return err
	}
}

func loadDB() error {
	lic, err := license.LoadLicense()
	if err != nil {
		return err
	}

	// load license data
	err = license.ExtractLicense(lic)
	if err != nil {
		return err
	}

	return nil
}
