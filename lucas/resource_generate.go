package lucas

import (
	"errors"
	"fmt"
	"hago-plat/pcom/lucas/profiler"
	"hago-plat/pcom/lucas/templates"
	"os"
	"os/exec"
	"path/filepath"
	"text/template"
)

// ResourceGenerator ...
type ResourceGenerator struct {
	// profile *profiler.Profile
}

// NewResourceGenerator returns RPCGenerator object
func NewResourceGenerator() *ResourceGenerator {
	return &ResourceGenerator{
		// profile: profile,
	}
}

// GenerateProto implements Generator interface and generates resource's
// .proto file
func (r *ResourceGenerator) GenerateProto(model profiler.SpecModel) error {
	// fill in profile
	info, ok := model.(*ResourceSpec)
	if !ok {
		return errors.New("model is not kind of *ResourceSpec")
	}
	fmt.Println("generating", info.Path, "...")
	// new .proto template
	tmpl, err := template.New("resource_gen").Parse(templates.ResourceTemplate)
	if err != nil {
		fmt.Println("template parse fail.", err)
		return err
	}

	// open .proto file
	generatePath := filepath.Join(projectRoot, genPath, info.Path)
	err = os.MkdirAll(generatePath, 0755)
	if err != nil {
		fmt.Println("mkdir ", generatePath, "fail:", err)
		return err
	}
	fileName := filepath.Join(generatePath, genPBFileName)
	file, err := os.OpenFile(fileName, os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0755)
	defer func() { _ = file.Close() }()
	if err != nil {
		fmt.Println("open file", fileName, "fail,", err)
		return err
	}
	// fmt.Println("import pkgs:", info.ImportsPKGs)

	// execute .proto file generation
	return tmpl.Execute(file, *info)

}

// GenLucas implements Generator interface and do nothing, because resource
// doesn't has rpc and no need to generate .lucas file
func (r *ResourceGenerator) GenLucas(model profiler.SpecModel) error {
	return nil
}

// ExecProtoc implements Generator interface and executes protoc command
func (r *ResourceGenerator) ExecProtoc(model profiler.SpecModel) error {
	info, ok := model.(*ResourceSpec)
	if !ok {
		return errors.New("model is not kind of *ResourceSpec")
	}

	ImportPath := filepath.Join(projectRoot, genPath)
	DotPath := filepath.Join(projectRoot, genPath, info.Path) //relativePath)
	OutPath := filepath.Dir(projectRoot)
	GogoPath := filepath.Join(filepath.Dir(projectRoot), "proto") // todo specialized for go mod
	GeneratedFilePath := filepath.Join(DotPath, "generated.proto")

	protocCmd := fmt.Sprintf("protoc -I %v -I %v -I %v --gofast_out=plugins=grpc:%v %v",
		ImportPath, GogoPath, DotPath, OutPath, GeneratedFilePath)
	fmt.Println(protocCmd)
	cmd := exec.Command("/bin/bash", "-c", protocCmd)
	stdOutIn, _ := cmd.StdoutPipe()
	stdErrIn, _ := cmd.StderrPipe()
	defer stdOutIn.Close()
	defer stdErrIn.Close()
	err := RunCmdAndPrint(cmd, stdOutIn, stdErrIn)
	if err != nil {
		return err
	}
	// there is a bug in gogoprotobuf, fix it using script
	finalGeneratedPath := filepath.Join(filepath.Dir(projectRoot), info.Path, "generated.pb.go")
	// cmdFixImport := fmt.Sprintf(FixImportShellScript, finalGeneratedPath, finalGeneratedPath, finalGeneratedPath)
	cmdFix := exec.Command("/bin/bash", "-c", "goimports -w "+finalGeneratedPath)
	output, err := cmdFix.CombinedOutput()
	if err != nil {
		fmt.Println("bash err:", string(output))
		return err
	}
	return nil
}
