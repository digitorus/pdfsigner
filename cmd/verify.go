package cmd

import (
	"github.com/digitorus/pdfsigner/signer"
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
			result, err := signer.VerifyFile(f)
			if err != nil {
				log.Println("File", f, "Couldn't be verified", err)
			} else if result.Valid {
				log.Println("File", f, "verified successfully")
			} else {
				log.Println("File", f, "is signed but not valid")
			}
		}
	},
}

func init() {
	RootCmd.AddCommand(verifyCmd)
}
