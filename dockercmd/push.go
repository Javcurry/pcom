package dockercmd

import (
	"os/exec"
)

// Push ...
func Push(image string) *exec.Cmd {
	args := make([]string, 0, 16)
	args = append(args, "push")
	args = append(args, image)

	cmd := exec.Command(dockerBin, args...)
	return cmd
}
