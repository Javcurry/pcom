package dockercmd

import (
	"os/exec"
)

// BuildOptions ...
type BuildOptions struct {
	Pull  bool
	Image string
}

// Build ...
func Build(dir string, options *BuildOptions) *exec.Cmd {
	args := make([]string, 0, 16)

	args = append(args, "build")
	if options.Pull {
		args = append(args, "--pull") // always attempt to pull a newer version of the image
	}
	if len(options.Image) > 0 {
		args = append(args, "--tag", options.Image)
	}
	args = append(args, dir)

	cmd := exec.Command(dockerBin, args...)
	cmd.Dir = dir
	return cmd
}
