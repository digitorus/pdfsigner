package main

import (
	"bitbucket.org/digitorus/pdfsigner/cmd"
	"bitbucket.org/digitorus/pdfsigner/version"
)

var (
	// Version of the application set at build time
	Version = "undefined"
	// BuildDate of the application set at build time
	BuildDate string
	// GitCommit of the application set at build time
	GitCommit string
	// GitBranch of the application set at build time
	GitBranch string
)

func main() {
	ver := version.Version{
		Version:   Version,
		BuildDate: BuildDate,
		GitCommit: GitCommit,
		GitBranch: GitBranch,
	}
	cmd.Execute(ver)
}

//
//func main() {
//	// if no flags provided print usage
//	if len(os.Args) == 1 {
//		//usage()
//		return
//	}
//
//
//	log.Println(cfgFile)
//	switch os.Args[1] {
//	case "sign":
//		signCmd()
//	case "verify":
//		//verifyPDF()
//	case "serve":
//	case "watch":
//	default:
//		fmt.Printf("%q is not valid command.\n", os.Args[1])
//		os.Exit(2)
//	}
//
//
//}
//
//func signCmd() {
//	signCommand := flag.NewFlagSet("sign", flag.ExitOnError)
//	switch os.Args[2] {
//	case "pksc11":
//		//pksc11sign(signCommand)
//	case "ssl":
//		//sslSign(signCommand)
//	case "signer":
//		signCmd()
//	default:
//		fmt.Printf("%q is not valid command.\n", os.Args[2])
//		os.Exit(2)
//	}
//
//	signCommand.Parse(os.Args[2:])
//}
//
//var (
//	cfgFile               string
//	signatureApproval     bool
//	signatureType         uint
//	signatureInfoName     string
//	signatureInfoLocation string
//	signatureInfoReason   string
//	signatureInfoContact  string
//	signatureTSAUrl       string
//	signatureTSAUsername  string
//	signatureTSAPassword  string
//	signerNameFlag        string
//	certificateChainPath  string
//	certificatePath       string
//	privateKeyPath        string
//	pksc11LibPath         string
//	pksc11Pass            string
//)
//
//type configSigner struct {
//	Name         string
//	SignerType   string
//	CrtPath      string
//	KeyPath      string
//	LibPath      string
//	Pass         string
//	CrtChainPath string
//	Signature    sign.SignDataSignature
//	TSA          sign.TSA
//}
//
//var signers []configSigner
//
//func parseSignDataFlags(cmd *flag.FlagSet, c configSigner) {
//	cmd.BoolVar(&signatureApproval, "approval", c.Signature.Approval, "Approval")
//	cmd.UintVar(&signatureType, "type", uint(c.Signature.CertType), "Certificate type")
//	cmd.StringVar(&signatureInfoName, "name", c.Signature.Info.Name, "Signature info name")
//	cmd.StringVar(&signatureInfoLocation, "location", c.Signature.Info.Location, "Signature info location")
//	cmd.StringVar(&signatureInfoReason, "reason", c.Signature.Info.Reason, "Signature info reason")
//	cmd.StringVar(&signatureInfoContact, "contact", c.Signature.Info.ContactInfo, "Signature info contact")
//	cmd.StringVar(&signatureTSAUrl, "tsa-url", c.TSA.URL, "TSA url")
//	cmd.StringVar(&signatureTSAUsername, "tsa-username", c.TSA.Username, "TSA username")
//	cmd.StringVar(&signatureTSAPassword, "tsa-password", c.TSA.Password, "TSA password")
//	cmd.StringVar(&certificateChainPath, "chain", c.CrtChainPath, "Certificate chain")
//	cmd.StringVar(&cfgFile, "config", "", "config file (default is $HOME/.pdfsigner.yaml)")
//	cmd.Parse(os.Args[2:])
//}
//
//// initConfig reads in config file and ENV variables if set.
//func initConfig() {
//	log.Print(cfgFile)
//	if cfgFile != "" {
//		// Use config file from the flag.
//		viper.SetConfigFile(cfgFile)
//	} else {
//		// Find home directory.
//		home, err := homedir.Dir()
//		if err != nil {
//			fmt.Println(err)
//			os.Exit(1)
//		}
//
//		// Search config in home directory with name ".pdfsigner" (without extension).
//		viper.AddConfigPath(home)
//		viper.SetConfigName(".pdfsigner")
//	}
//
//	viper.AutomaticEnv() // read in environment variables that match
//	// If a config file is found, read it in.
//	if err := viper.ReadInConfig(); err == nil {
//		fmt.Println("Using config file:", viper.ConfigFileUsed())
//	}
//
//	viper.UnmarshalKey("signer", &signers)
//}
