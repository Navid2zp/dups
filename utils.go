package dups

import (
	"github.com/cheggaaa/pb/v3"
	"strings"
)

func CleanPath(path string) string {
	return strings.Replace(path, "\\", "/", -1)
}

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

func createBar(limit int) *pb.ProgressBar {
	tmpl := `{{ blue "Progress:" }} {{ bar . "[" "=" (cycle . ">") "." "]"}} {{speed . | green }} {{percent . | green}}`
	bar := pb.ProgressBarTemplate(tmpl).Start64(int64(limit))
	return bar
}
