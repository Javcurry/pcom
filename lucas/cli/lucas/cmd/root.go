package cmd

import (
	"fmt"
	"hago-plat/pcom/lucas"
	"hago-plat/pcom/lucas/profiler"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"syscall"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:     "lucas -g GEN_PATH -p SOURCE_PATH",
	Short:   "lucas is a generator for GRPC + HTTP + go-micro framework",
	Example: `./lucas -g ~/hago-plat/hago-plat/gen -p ~/hago-plat/hago-plat/hagonetes/hagopl_h7s_info_center_d`,
	Run:     RootCmdRun,
}

var (
	projectRoot = ""
	genPath     = ""
	srcPath     = ""
)

const (
	reflectGenDir  = "gen_reflect"
	profileDirName = ".lucas"
)

func init() {
	cmd := exec.Command("/bin/bash", "-c", "go env GOMOD")
	out, _ := cmd.CombinedOutput()
	if len(out) == 0 {
		fmt.Println("no GOMOD")
		os.Exit(1)
	}
	projectRoot = filepath.Dir(string(out))
	rootCmd.PersistentFlags().StringVarP(&genPath, "gen_path", "g", "", "project root path")
	rootCmd.PersistentFlags().StringVarP(&srcPath, "source_path", "p", "", "relative path to scan")
	_ = rootCmd.MarkPersistentFlagRequired("gen_path")
	_ = rootCmd.MarkPersistentFlagRequired("source_path")
}

// Execute run lucas commands
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func cleanGeneratedGoFiles() error {
	fmt.Println("cleaning generated files...")
	_, err := os.Stat(genPath)
	if os.IsNotExist(err) {
		return nil
	}
	err = filepath.Walk(genPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			sourceRelativePath, err := filepath.Rel(genPath, path)
			if err != nil {
				return err
			}
			sourcePath := filepath.Join(filepath.Dir(projectRoot), sourceRelativePath)
			_, err = os.Stat(filepath.Join(sourcePath, "generated.pb.go"))
			if err == nil {
				// 只清理-p 参数路径下的generated.pb.go
				if strings.HasPrefix(sourcePath, srcPath) {
					fmt.Println("cleaning", filepath.Join(sourcePath, "generated.pb.go"))
					err = os.RemoveAll(filepath.Join(sourcePath, "generated.pb.go"))
					if err != nil {
						return err
					}
				}
			}
		} else {
			reg, err := regexp.Compile("generated.*.go")
			if err != nil {
				return err
			}
			found := reg.MatchString(path)
			if found {
				fmt.Println("cleaning", path)
				err := os.Remove(path)
				if err != nil {
					return err
				}
			}
		}
		return nil
	})
	return err
}

func checkExists(path string) bool {
	file, err := os.Open(path)
	defer func() { _ = file.Close() }()
	return err == nil
}

// RootCmdRun runs root cmd
func RootCmdRun(cmd *cobra.Command, args []string) {
	var err error
	defer func() {
		if err != nil {
			fmt.Println(err)
			if errExit, ok := err.(*exec.ExitError); ok {
				if status, ok := errExit.Sys().(syscall.WaitStatus); ok {
					os.Exit(status.ExitStatus())
				}
			}
			os.Exit(2)
		}
	}()
	genPath, err = filepath.Abs(genPath)
	if err != nil {
		fmt.Println("gen_path:", err)
		return
	}
	pwd, err := os.Getwd()
	if err != nil {
		return
	}
	if !checkExists(projectRoot) {
		err = fmt.Errorf("project root not found: %v", projectRoot)
		fmt.Println(err)
		return
	}
	srcPath, err = filepath.Abs(srcPath)
	if err != nil {
		fmt.Println("source_path:", err)
		return
	}
	if !checkExists(srcPath) {
		err = fmt.Errorf("source path not found: %v", projectRoot)
		fmt.Println(err)
		return
	}
	// to prevent build error
	err = cleanGeneratedGoFiles()
	if err != nil {
		fmt.Println(err)
		return
	}
	reflectConf := ReflectMain{}
	err = filepath.Walk(srcPath, reflectConf.ScanServices)
	if err != nil {
		return
	}
	err = reflectConf.GenMainAndRun()
	if err != nil {
		return
	}
	_ = reflectConf.RemoveReflectMain()
	gen, err := filepath.Rel(projectRoot, genPath)
	if err != nil {
		return
	}
	lucas.InitGen(gen)
	profile := profiler.NewProfile(projectRoot, gen)
	profile.RegisterFactory(profiler.Kind(lucas.SpecKindRPCService), lucas.NewRPCServiceSpec())
	profile.RegisterFactory(profiler.Kind(lucas.SpecKindResource), lucas.NewResourceType())
	err = profile.Load()
	if err != nil {
		return
	}
	profile.RegisterGenerator(profiler.Kind(lucas.SpecKindRPCService), lucas.NewRPCGenerator())
	profile.RegisterGenerator(profiler.Kind(lucas.SpecKindResource), lucas.NewResourceGenerator())

	err = os.Chdir(pwd)
	if err != nil {
		fmt.Println("cd")
		return
	}
	err = profile.StartGeneration()
	if err != nil {
		return
	}
	// 清除生成过程中中间文件
	err = os.RemoveAll(filepath.Join(genPath, profileDirName))
	if err != nil {
		fmt.Println(err)
		return
	}
}
