package cmd

import (
	"fmt"
	"hago-plat/pcom/nameconv"
	"hago-plat/pcom/quintus"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

// NewQuintusCommand quintus 主命令
func NewQuintusCommand() *cobra.Command {
	executor := Executor{}

	cmd := &cobra.Command{
		Use:     "quintus",
		Short:   "quintus is a generator from go template file to the destination file",
		Example: `quintus [Options...] [template filepath]`,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return executor.PreRun(args)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return executor.Run()
		},
	}

	executor.BindFlags(cmd)

	return cmd
}

// Executor 主命令执行逻辑
type Executor struct {
	OutputPath string
	SourcePath string

	namesSlice  []string
	valuesSlice []string

	Names  map[string]quintus.ObjectName
	Values map[string]string
}

// BindFlags 绑定参数
func (e *Executor) BindFlags(cmd *cobra.Command) {
	cmd.Flags().StringVarP(&e.OutputPath, "output", "o", "", "[required] path to output destination file")
	cmd.Flags().StringSliceVarP(&e.namesSlice, "name", "n", nil, "[optional] specified name which can be convert into ExportedCamel/UnexportedCamel/Snake")
	cmd.Flags().StringSliceVarP(&e.valuesSlice, "value", "v", nil, "[optional] custom values for go template")
}

// Validate 校验并重建或转换必要的参数
func (e *Executor) Validate(args []string) error {
	if len(e.OutputPath) == 0 {
		return fmt.Errorf("flag -o or --output required")
	}

	e.Names = map[string]quintus.ObjectName{}
	for _, v := range e.namesSlice {
		splits := strings.Split(v, "=")
		if len(splits) < 2 {
			fmt.Fprintf(os.Stderr, "warning: invalid name of: %v\n", v)
			continue
		}
		e.Names[splits[0]] = quintus.ObjectName{Name: nameconv.Name(splits[1])}
	}

	e.Values = map[string]string{}
	for _, v := range e.valuesSlice {
		splits := strings.Split(v, "=")
		if len(splits) < 2 {
			fmt.Fprintf(os.Stderr, "warning: invalid value of: %v\n", v)
			continue
		}
		e.Values[splits[0]] = splits[1]
	}

	if len(args) == 0 {
		return fmt.Errorf("invalid arguments")
	}

	e.SourcePath = args[0]
	err := e.checkFile(e.SourcePath)
	if err != nil {
		return err
	}

	return nil
}

//func (e *Executor) checkDir(path string) error {
//	f, err := os.Stat(path)
//	if err != nil {
//		return err
//	}
//	if !f.IsDir() {
//		return fmt.Errorf("no such directory: %v", path)
//	}
//
//	return nil
//}

func (e *Executor) checkFile(path string) error {
	f, err := os.Stat(path)
	if err != nil {
		return err
	}
	if !f.Mode().IsRegular() {
		return fmt.Errorf("no such file: %v", path)
	}

	return nil
}

// PreRun ...
func (e *Executor) PreRun(args []string) error {
	err := e.Validate(args)
	if err != nil {
		return err
	}

	return err
}

// Run ...
func (e *Executor) Run() error {
	data := &quintus.Data{
		Names:  e.Names,
		Values: e.Values,
	}
	return quintus.Convert(data, e.SourcePath, e.OutputPath)
}
