package cmd

import (
	"bufio"
	"fmt"
	"hago-plat/pcom/lucas"
	"html/template"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

// ReflectMain 生成reflect扫描main函数描述
type ReflectMain struct {
	PKGConfigs []ReflectMainPKGConfig
}

// ReflectMainPKGConfig 每个包的设置
type ReflectMainPKGConfig struct {
	ServiceName     string
	ServicePKG      string
	ServicePKGAlias string
	GenPathRel      string
}

// ScanServiceConfig 扫描service包，并记录service包内reflect所需内容
func (r *ReflectMain) ScanServiceConfig(serviceName, path string) error {
	genPathRel, err := filepath.Rel(projectRoot, genPath)
	if err != nil {
		return err
	}
	config := ReflectMainPKGConfig{
		ServiceName: serviceName,
		ServicePKG:  strings.TrimPrefix(filepath.Dir(path), filepath.Dir(projectRoot)+"/"),
		GenPathRel:  genPathRel,
	}
	config.ServicePKGAlias = lucas.RemoveHyphen(lucas.FolderSeparator2Underline(config.ServicePKG))
	r.PKGConfigs = append(r.PKGConfigs, config)
	return nil
}

// ScanServices walker函数，用于寻找并扫描Service包
func (r *ReflectMain) ScanServices(path string, info os.FileInfo, err error) error {
	if err != nil {
		return err
	}
	if info.IsDir() || !strings.HasSuffix(path, ".go") {
		return nil
	}
	serviceName, err := findService(path)
	if err != nil {
		fmt.Println(err)
		return err
	}
	if len(serviceName) == 0 {
		return nil
	}

	err = r.ScanServiceConfig(serviceName, path)
	if err != nil {
		return err
	}
	return nil
}

// GenMainAndRun start lucas generation
func (r *ReflectMain) GenMainAndRun() error {
	reflectGenPath := getReflectGenPath()
	err := os.MkdirAll(reflectGenPath, 0755)
	if err != nil && !os.IsExist(err) {
		return err
	}
	file, err := os.OpenFile(filepath.Join(reflectGenPath, "hagopl_h7s_docker_agent_d.go"),
		os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0755)
	defer func() { _ = file.Close() }()
	if err != nil {
		return err
	}
	err = os.Chdir(reflectGenPath)
	if err != nil {
		return err
	}
	tmpl, err := template.New("main").Parse(mainTemplate)
	if err != nil {
		return err
	}
	err = tmpl.Execute(file, r)
	if err != nil {
		return err
	}
	shcmd := exec.Command("/bin/bash", "-c", "go run hagopl_h7s_docker_agent_d.go")
	stdOutIn, _ := shcmd.StdoutPipe()
	stdErrIn, _ := shcmd.StderrPipe()
	defer stdOutIn.Close()
	defer stdOutIn.Close()

	err = lucas.RunCmdAndPrint(shcmd, stdOutIn, stdErrIn)
	if err != nil {
		return err
	}
	return nil
}

// RemoveReflectMain 清理
func (r *ReflectMain) RemoveReflectMain() error {
	reflectGenPath := getReflectGenPath()
	return os.RemoveAll(reflectGenPath)
}

func getReflectGenPath() string {
	return filepath.Join(genPath, profileDirName, reflectGenDir)
}

func findService(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}

	reg, err := regexp.Compile("type[ ]+[A-Z]+[a-zA-Z0-9]*Service[ ]+struct")
	if err != nil {
		return "", err
	}
	reader := bufio.NewReader(file)
	serviceName := ""
	for {
		line, err := reader.ReadString('\n')
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", err
		}
		if reg.MatchString(line) {
			reg2, err := regexp.Compile("[A-Z]+[a-zA-Z0-9]*Service")
			if err != nil {
				return "", err
			}
			serviceName = reg2.FindString(line)
			break
		}
	}
	return serviceName, nil
}
