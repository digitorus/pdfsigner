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
	"os"
	"strings"

	"github.com/digitorus/pdfsigner/license"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// licenseCmd represents the license command.
var licenseCmd = &cobra.Command{
	Use:   "license",
	Short: "Check and update license",
}

// licenseInfoCmd represents the license info command.
var licenseSetupCmd = &cobra.Command{
	Use:   "setup",
	Short: "license setup",
	Run: func(cmd *cobra.Command, args []string) {
		// initialize license
		err := initializeLicense()
		if err != nil {
			log.Fatal(err)
		}

		// print license info
		fmt.Print(license.LD.Info())
	},
}

// licenseInfoCmd represents the license info command.
var licenseInfoCmd = &cobra.Command{
	Use:   "info",
	Short: "license info",
	Run: func(cmd *cobra.Command, args []string) {
		// load license
		err := license.Load()
		if err != nil {
			log.Fatal(err)
		}

		// print license info
		fmt.Print(license.LD.Info())
	},
}

func init() {
	RootCmd.AddCommand(licenseCmd)
	licenseCmd.AddCommand(licenseSetupCmd)
	licenseCmd.AddCommand(licenseInfoCmd)
}

// initializeLicense loads the license file with provided path licenseStrConfOrFlag or stdin.
func initializeLicense() error {
	// reading license file name. Info: can't read license directly from stdin because of a darwin 1024 limit.
	var licenseStr string
	if licenseStrConfOrFlag != "" {
		// try to get license from the flag provided
		licenseStr = licenseStrConfOrFlag
	} else {
		// get license from the stdout
		fmt.Fprint(os.Stdout, "Paste your license here:")

		var err error

		licenseStr, err = bufio.NewReader(os.Stdin).ReadString('\n')
		if err != nil {
			return errors.Wrap(err, "")
		}
	}

	licenseBytes := []byte(strings.Replace(strings.TrimSpace(licenseStr), "\n", "", -1))
	// initialize license
	err := license.Initialize(licenseBytes)
	if err != nil {
		return errors.Wrap(err, "")
	}

	return nil
}
