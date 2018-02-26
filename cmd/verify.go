package cmd

import (
	"log"
	"os"

	"bitbucket.org/digitorus/pdfsign/verify"
	"github.com/spf13/cobra"
)

// verifyCmd represents the verify command
var verifyCmd = &cobra.Command{
	Use:   "verify",
	Short: "Verifies PDF signature",
	Long:  `Long multiline description here`,
	Run: func(cmd *cobra.Command, inputFileNames []string) {
		for _, f := range inputFileNames {
			input_file, err := os.Open(f)
			if err != nil {
				log.Fatal("Coudn't open file", f, ",", err)
			}
			defer input_file.Close()

			_, err = verify.Verify(input_file)
			if err != nil {
				log.Println("File", f, "coudn't be verified", err)
			} else {
				log.Println("File", f, "verified")
			}

		}
	},
}

func init() {
	RootCmd.AddCommand(verifyCmd)
}
