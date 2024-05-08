package dockercmd

import (
	"bytes"
	"fmt"
	"os/exec"
)

var (
	dockerBin = defaultDockerBin
)

// SetDockerBin ...
func SetDockerBin(bin string) {
	dockerBin = bin
}

// ExecuteCmd ...
func ExecuteCmd(cmd *exec.Cmd) error {
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("execute cmd with error: %v -> %v", err, stderr.String())
	}

	return nil
}

// ExecuteCmdWithStdout ...
func ExecuteCmdWithStdout(cmd *exec.Cmd) (string, error) {
	var stderr, stdout bytes.Buffer
	cmd.Stderr = &stderr
	cmd.Stdout = &stdout
	err := cmd.Run()
	if err != nil {
		return stdout.String(), fmt.Errorf("execute cmd with error: %v -> %v", err, stderr.String())
	}

	return stdout.String(), nil
}
