package cmd

import (
	"os"

	"github.com/digitorus/pdfsign/verify"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// verifyCmd represents the verify command.
var verifyCmd = &cobra.Command{
	Use:   "verify",
	Short: "Verify PDF signature",
	Run: func(cmd *cobra.Command, inputFileNames []string) {
		if len(inputFileNames) < 1 {
			log.Fatal("no files provided")
		}

		for _, f := range inputFileNames {
			input_file, err := os.Open(f)
			if err != nil {
				log.Fatal("Couldn't open file", f, ",", err)
			}
			defer func() { _ = input_file.Close() }()

			// See the comment on the equivalent call in signer/signer.go for
			// why the deprecated top-level helper is kept over the fluent API.
			_, err = verify.VerifyFile(input_file) //nolint:staticcheck
			if err != nil {
				log.Println("File", f, "Couldn't be verified", err)
			} else {
				log.Println("File", f, "verified successfully")
			}

		}
	},
}

func init() {
	RootCmd.AddCommand(verifyCmd)
}
