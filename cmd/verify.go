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
	Run: func(cmd *cobra.Command, args []string) {
		input_file, err := os.Open(inputFileNameFlag)
		if err != nil {
			log.Fatal(err)
		}
		defer input_file.Close()

		resp, err := verify.Verify(input_file)
		log.Println(resp)
		if err != nil {
			log.Fatalln(err)
		}
	},
}

func init() {
	RootCmd.AddCommand(verifyCmd)
	verifyCmd.PersistentFlags().StringVar(&inputFileNameFlag, "in", "", "Help here")
}
