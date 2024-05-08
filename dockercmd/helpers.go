package dockercmd

import (
	"strings"
)

// CanonicalImage ...
func CanonicalImage(image string) string {
	return strings.ToLower(image)
}
