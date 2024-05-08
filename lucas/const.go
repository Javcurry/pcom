package lucas

import (
	"os/exec"
	"path/filepath"
)

var (
	goMod       = ""
	projectRoot = ""
	genPath     = ""
)

const (
	genPBFileName = "generated.proto"
	lucasFileName = "generated.lucas.go"

//	lucasClientFileName = "generated.lucas.client.go"
)

func init() {
	cmd := exec.Command("/bin/bash", "-c", "go env GOMOD")
	out, _ := cmd.CombinedOutput()
	goMod = string(out)
	projectRoot = filepath.Dir(goMod)
}
