package lucas

import (
	"bytes"
	"fmt"
	"hago-plat/pcom/lucas/profiler"
	"hago-plat/pcom/lucas/templates"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"text/template"

	"github.com/pkg/errors"
)

// RPCGenerator ...
type RPCGenerator struct {
}

// NewRPCGenerator returns RPCGenerator object
func NewRPCGenerator() *RPCGenerator {
	return &RPCGenerator{}
}

// GenerateProto start generation of rpc server
func (r *RPCGenerator) GenerateProto(model profiler.SpecModel) error {
	// fill in profile
	info, ok := model.(*RPCServiceSpec)
	if !ok {
		return errors.New("model is not kind of *RPCServiceSpec")
	}
	fmt.Println("generating", info.Path, "...")
	// new .proto template
	tmpl, err := template.New("rpc_gen").Parse(templates.RPCTemplate)
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

// GenLucas generates .lucas.go file
func (r *RPCGenerator) GenLucas(model profiler.SpecModel) error {
	// fill in profile
	info, ok := model.(*RPCServiceSpec)
	if !ok {
		return errors.New("model is not kind of *RPCServiceSpec")
	}

	// new .lucas template
	ClientTmpl, err := template.New("lucas_gen").Parse(templates.LucasTemplate)
	if err != nil {
		fmt.Println("template parse fail.", err)
		return err
	}
	_, _ = ClientTmpl.New("response").Parse(templates.ResponseTemplate)
	// open .lucas file
	clientGeneratePath := filepath.Join(projectRoot, genPath, info.Path)
	err = os.MkdirAll(clientGeneratePath, 0755)
	if err != nil {
		fmt.Println("mkdir ", clientGeneratePath, "fail:", err)
		return err
	}
	clientFileName := filepath.Join(clientGeneratePath, lucasFileName)
	clientFile, err := os.OpenFile(clientFileName, os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0755)
	if err != nil {
		fmt.Println("open file", clientFileName, "fail,", err)
		return err
	}

	return ClientTmpl.Execute(clientFile, *info)
}

// ExecProtoc exec protoc to generate pb.gw.go, .pb.go and .micro.go
func (r *RPCGenerator) ExecProtoc(model profiler.SpecModel) error {
	info, ok := model.(*RPCServiceSpec)
	if !ok {
		return errors.New("model is not kind of *ResourceSpec")
	}

	ImportPath := filepath.Join(projectRoot, genPath)
	DotPath := filepath.Join(projectRoot, genPath, info.Path) //relativePath)
	GRPCGatewayConfigPath := filepath.Join(filepath.Dir(projectRoot), info.Path, "http_service.yaml")
	GogoPath := filepath.Join(filepath.Dir(projectRoot), "proto") // todo specialized for go mod
	GRPCPATH := filepath.Join(filepath.Dir(projectRoot), "proto/github.com/grpc-ecosystem/grpc-gateway")
	GeneratedFilePath := filepath.Join(DotPath, "generated.proto")

	var protocCmd string
	if _, err := os.Stat(GRPCGatewayConfigPath); os.IsNotExist(err) {
		protocCmd = fmt.Sprintf("protoc -I %v -I %v -I %v -I %v --gofast_out=plugins=grpc:%v --micro_out=%v %v",
			ImportPath, GRPCPATH, GogoPath, DotPath, ImportPath, ImportPath, GeneratedFilePath)
	} else {
		protocCmd = fmt.Sprintf("protoc -I %v -I %v -I %v -I %v --gofast_out=plugins=grpc:%v --micro_out=%v "+
			"--lucas-grpc-gateway_out=grpc_api_configuration=%v:%v --lucas-swagger_out=json_names_for_fields=true,grpc_api_configuration=%v:%v %v",
			ImportPath, GRPCPATH, GogoPath, DotPath, ImportPath, ImportPath,
			GRPCGatewayConfigPath, ImportPath, GRPCGatewayConfigPath,
			ImportPath, GeneratedFilePath)
	}
	fmt.Println(protocCmd)
	cmd := exec.Command("/bin/bash", "-c", protocCmd)
	stdOutIn, _ := cmd.StdoutPipe()
	stdErrIn, _ := cmd.StderrPipe()
	err := RunCmdAndPrint(cmd, stdOutIn, stdErrIn)
	defer stdOutIn.Close()
	defer stdErrIn.Close()
	if err != nil {
		return err
	}
	// there is a bug in gogoprotobuf, fix it using script
	finalGeneratedPath := filepath.Join(ImportPath, info.Path, "generated.pb.go")
	cmdFix := exec.Command("/bin/bash", "-c", "goimports -w "+finalGeneratedPath)
	fixCmdStdOutIn, _ := cmdFix.StdoutPipe()
	fixCmdStdErrIn, _ := cmdFix.StderrPipe()
	defer fixCmdStdOutIn.Close()
	defer fixCmdStdErrIn.Close()
	err = RunCmdAndPrint(cmdFix, fixCmdStdOutIn, fixCmdStdErrIn)
	if err != nil {
		return err
	}
	gatewayPath := filepath.Join(ImportPath, info.Path, "gateway")
	err = os.MkdirAll(gatewayPath, 0755)
	if err != nil {
		fmt.Println("mkdir", gatewayPath, "fail:", err)
		return err
	}

	if _, err := os.Stat(GRPCGatewayConfigPath); os.IsNotExist(err) {
	} else {
		err = os.Rename(filepath.Join(ImportPath, info.Path, "generated.pb.gw.go"), filepath.Join(gatewayPath, "generated.pb.gw.go"))
		if err != nil {
			fmt.Println("mv fail:", err)
			return err
		}
		err = os.Rename(filepath.Join(ImportPath, info.Path, "generated.swagger.json"), filepath.Join(gatewayPath, "generated.swagger.json"))
		if err != nil {
			fmt.Println("mv fail:", err)
			return err
		}
		err = r.swaggerJSON(gatewayPath, "generated.swagger.json", model)
		if err != nil {
			fmt.Println("generated.swagger.json generate fail:", err)
			return err
		}
	}
	cpcmd := exec.Command("/bin/bash", "-c", fmt.Sprintf("cp %v %v", finalGeneratedPath, filepath.Join(ImportPath, info.Path, "gateway")))
	cpCmdStdOutIn, _ := cpcmd.StdoutPipe()
	cpCmdStdErrIn, _ := cpcmd.StderrPipe()
	defer cpCmdStdOutIn.Close()
	defer cpCmdStdErrIn.Close()
	err = RunCmdAndPrint(cpcmd, cpCmdStdOutIn, cpCmdStdErrIn)
	if err != nil {
		return err
	}
	return nil
}

func (r *RPCGenerator) swaggerJSON(swaggerFilePath, swaggerFileName string, model profiler.SpecModel) error {
	info := model.(*RPCServiceSpec)
	file, err := os.Open(filepath.Join(swaggerFilePath, swaggerFileName))
	if err != nil {
		return err
	}
	var buff []byte
	bw := bytes.NewBuffer(buff)
	tmpl, err := template.New("swagger-go").Parse(templates.SwaggerTemplatePrefix)
	if err != nil {
		return err
	}
	err = tmpl.Execute(bw, info)
	if err != nil {
		return err
	}
	allJSON, _ := ioutil.ReadAll(file)
	fileString := bw.String() + string(allJSON) + templates.SwaggerTemplateSuffix
	err = ioutil.WriteFile(filepath.Join(swaggerFilePath, "generated.swagger.go"), []byte(fileString), 0755)
	if err != nil {
		return err
	}
	return nil
}
