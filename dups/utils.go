package dups

import (
	"github.com/cheggaaa/pb/v3"
	"strings"
)

// CleanPath replaces \ with / in a path
func CleanPath(path string) string {
	return strings.Replace(path, "\\", "/", -1)
}

// GetAlgorithm matches the given string to one of the supported algorithms
// Returns md5 if a match wasn't found
func GetAlgorithm(al string) string {
	switch strings.ToLower(al) {
	case MD5:
		return MD5
	case SHA256:
		return SHA256
	case XXHash:
		return XXHash
	}
	return MD5
}

// createBar creates a new progress bar with a custom template
func createBar(limit int) *pb.ProgressBar {
	tmpl := `{{ blue "Progress:" }} {{ bar . "[" "=" (cycle . ">") "." "]"}} {{speed . | green }} {{percent . | green}}`
	bar := pb.ProgressBarTemplate(tmpl).Start64(int64(limit))
	return bar
}
