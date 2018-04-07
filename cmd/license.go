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
	"fmt"
	"log"
	"os"

	"bitbucket.org/digitorus/pdfsigner/license"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

// licenseCmd represents the license command
var licenseCmd = &cobra.Command{
	Use:   "license",
	Short: "Update license",
	Run: func(cmd *cobra.Command, args []string) {

	},
}

// licenseInfoCmd represents the license info command
var licenseSetupCmd = &cobra.Command{
	Use:   "setup",
	Short: "license setup",
	Run: func(cmd *cobra.Command, args []string) {
		err := initializeLicense()
		if err != nil {
			log.Fatal(err)
		}

		fmt.Printf(`Licensed to %s until %s`, license.LD.Email, license.LD.End.Format("2006-01-02"))
	},
}

// licenseInfoCmd represents the license info command
var licenseInfoCmd = &cobra.Command{
	Use:   "info",
	Short: "license info",
	Run: func(cmd *cobra.Command, args []string) {
		err := license.Load()
		if err != nil {
			log.Fatal(err)
		}

		fmt.Printf(`Licensed to %s until %s`, license.LD.Email, license.LD.End.Format("2006-01-02"))
		license.LD.Info()
	},
}

func init() {
	RootCmd.AddCommand(licenseCmd)
	licenseCmd.AddCommand(licenseSetupCmd)
	licenseCmd.AddCommand(licenseInfoCmd)
}

func initializeLicense() error {
	// read license from input
	fmt.Fprint(os.Stdout, "Enter license:")
	licenseBytes, err := bufio.NewReader(os.Stdin).ReadBytes('\n')
	if err != nil {
		return errors.Wrap(err, "")
	}

	// TODO: remove after tests done
	licenseBytes = []byte(license.LicenseB32)

	err = license.Initialize(licenseBytes)
	if err != nil {
		return errors.Wrap(err, "")
	}

	return nil
}
