package dockercmd

import (
	"os/exec"
)

// PullOptions ...
type PullOptions struct {
}

// Pull ...
func Pull(image string, options *PullOptions) *exec.Cmd {
	args := make([]string, 0, 16)
	args = append(args, "pull")
	args = append(args, image)

	cmd := exec.Command(dockerBin, args...)
	return cmd
}
