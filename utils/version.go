package utils

import "fmt"

var (
	buildCommit string
	buildDate   string
)

func Version() string {
	version := "v1.0.0"
	if buildCommit != "" {
		version = fmt.Sprintf("%s-%s", version, buildCommit)
	}
	if buildDate != "" {
		version = fmt.Sprintf("%s-%s", version, buildDate)
	}
	return version
}
