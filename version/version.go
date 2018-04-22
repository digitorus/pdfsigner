package version

// Version represents a git version of the application
type Version struct {
	Version   string `json:"version"`
	BuildDate string `json:"build_date"`
	GitCommit string `json:"git_commit"`
	GitBranch string `json:"git_branch"`
}
