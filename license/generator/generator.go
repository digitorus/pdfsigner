package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/digitorus/pdfsigner/license"
	"github.com/digitorus/pdfsigner/license/ratelimiter"
	log "github.com/sirupsen/logrus"
)

type LimitFlag struct {
	limits []*ratelimiter.Limit
}

func (lf *LimitFlag) String() string {
	result := []string{}
	for _, l := range lf.limits {
		result = append(result, fmt.Sprintf("%d:%s", l.MaxCount, l.IntervalStr))
	}
	return strings.Join(result, ",")
}

func (lf *LimitFlag) Set(value string) error {
	parts := strings.Split(value, ":")
	if len(parts) != 2 {
		return fmt.Errorf("limit must be in format 'count:duration', got %s", value)
	}

	var maxCount int
	if _, err := fmt.Sscanf(parts[0], "%d", &maxCount); err != nil {
		return fmt.Errorf("invalid count in limit: %s", err)
	}

	intervalStr := parts[1]

	lf.limits = append(lf.limits, &ratelimiter.Limit{
		MaxCount:    maxCount,
		IntervalStr: intervalStr,
	})

	return nil
}

func main() {
	var (
		genNewKey        = flag.Bool("new-key", false, "Generate a new key pair (private and public)")
		genHMACKey       = flag.Bool("new-hmac", false, "Generate a new HMAC key")
		name             = flag.String("name", "", "License name")
		email            = flag.String("email", "", "License email")
		duration         = flag.Duration("duration", 24*365*time.Hour, "License duration")
		maxWatchers      = flag.Int("max-watchers", 1, "Maximum number of directory watchers")
		privateKeyEnvVar = flag.String("key-env", "PDFSIGNER_LICENSE_PRIVATE_KEY", "Environment variable containing the private key")
	)

	limitFlag := LimitFlag{}
	flag.Var(&limitFlag, "limit", "Rate limit in format 'count:duration' (e.g., '100:1m'). Can be specified multiple times.")

	flag.Parse()

	// Generate a new key if requested
	if *genNewKey {
		privateKey, publicKey, err := license.GenerateKeyPair()
		if err != nil {
			log.Fatalf("Failed to generate key pair: %v", err)
		}
		fmt.Println("Private Key:", privateKey)
		fmt.Println("Public Key:", publicKey)
		fmt.Println("\nIMPORTANT: Keep the private key secure and use the public key for license validation.")

		return
	}

	// Generate a new HMAC key if requested
	if *genHMACKey {
		hmacKeyStr, err := license.GenerateRandomHMACKey()
		if err != nil {
			log.Fatalf("Failed to generate HMAC key: %v", err)
		}

		fmt.Println("HMAC Key:", hmacKeyStr)

		return
	}

	// Get private key from environment variable
	privateKeyStr := os.Getenv(*privateKeyEnvVar)
	if privateKeyStr == "" {
		log.Fatalf("Environment variable %s not set or empty", *privateKeyEnvVar)
	}

	// Validate required arguments for license generation
	if *name == "" || *email == "" {
		log.Fatal("Name and email are required for license generation")
	}

	// Extract public key from private key
	publicKey, err := license.GetPublicKeyFromPrivate(privateKeyStr)
	if err != nil {
		log.Fatalf("Failed to extract public key: %v", err)
	}

	// Set default limits if none provided
	limits := limitFlag.limits
	if len(limits) == 0 {
		limits = []*ratelimiter.Limit{
			{MaxCount: 60, IntervalStr: "1h"},
		}
	}

	// Create license data struct
	licenseData := license.LicenseData{
		Name:                 *name,
		Email:                *email,
		End:                  time.Now().Add(*duration),
		Limits:               limits,
		MaxDirectoryWatchers: *maxWatchers,
	}

	// Generate license
	licenseStr, err := license.GenerateLicense(privateKeyStr, licenseData)
	if err != nil {
		log.Fatalf("Failed to generate license: %v", err)
	}

	fmt.Println("License Data:", licenseStr)
	fmt.Printf("Licensed to %s <%s> until %s\n",
		licenseData.Name, licenseData.Email, licenseData.End.Format("2006-01-02"))
	fmt.Printf("Max directory watchers: %d\n", licenseData.MaxDirectoryWatchers)
	fmt.Println("Public Key:", publicKey)

	// Print limits
	fmt.Println("Limits:")
	for _, limit := range licenseData.Limits {
		fmt.Printf("  %d per %s\n", limit.MaxCount, limit.IntervalStr)
	}
}
