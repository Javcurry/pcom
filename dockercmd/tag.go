package dockercmd

import (
	"os/exec"
)

// Tag ...
func Tag(srcImage string, dstImage string) *exec.Cmd {
	args := make([]string, 0, 16)
	args = append(args, "tag")
	args = append(args, srcImage)
	args = append(args, dstImage)

	cmd := exec.Command(dockerBin, args...)
	return cmd
}
