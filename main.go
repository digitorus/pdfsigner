package main

import (
	"github.com/digitorus/pdfsigner/cmd"
	"github.com/digitorus/pdfsigner/version"
)

var (
	// Version of the application set at build time.
	Version = "undefined"
	// BuildDate of the application set at build time.
	BuildDate string
	// GitCommit of the application set at build time.
	GitCommit string
	// GitBranch of the application set at build time.
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
