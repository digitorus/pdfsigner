package cmd

import (
	"errors"
	"fmt"
	"os"

	"github.com/digitorus/pdfsign/verify"
	"github.com/spf13/cobra"
)

// verifyCmd represents the verify command.
var verifyCmd = &cobra.Command{
	Use:   "verify",
	Short: "Verify PDF signature",
	RunE: func(cmd *cobra.Command, inputFileNames []string) error {
		if len(inputFileNames) < 1 {
			return errors.New("no files provided")
		}

		for _, f := range inputFileNames {
			input_file, err := os.Open(f)
			if err != nil {
				return fmt.Errorf("couldn't open file %w", err)
			}
			defer input_file.Close()

			response, err := verify.File(input_file)
			if err != nil {
				return fmt.Errorf("couldn't verify file %w", err)
			}

			if response.Error != "" {
				return errors.New(response.Error)
			} else {
				cmd.Print("File verified successfully")
			}
		}

		return nil
	},
}

func init() {
	RootCmd.AddCommand(verifyCmd)
}
