package lucas

import (
	"fmt"
	"hago-plat/pcom/lucas/profiler"
	"os"
	"os/exec"
	"syscall"
)

// ServiceGenerate is the entry of lucas.
func ServiceGenerate(gen string, service interface{}) error {
	genPath = gen
	profile := profiler.NewProfile(projectRoot, genPath)
	var err error
	defer func() {
		if errExit, ok := err.(*exec.ExitError); ok {
			if status, ok := errExit.Sys().(syscall.WaitStatus); ok {
				os.Exit(status.ExitStatus())
			}
		}
	}()
	profile.RegisterFactory(profiler.Kind(SpecKindRPCService), NewRPCServiceSpec())
	profile.RegisterFactory(profiler.Kind(SpecKindResource), NewResourceType())
	err = profile.Load()
	if err != nil {
		fmt.Println(err)
		return err
	}

	profile.RegisterGenerator(profiler.Kind(SpecKindRPCService), NewRPCGenerator())
	profile.RegisterGenerator(profiler.Kind(SpecKindResource), NewResourceGenerator())
	err = ScanRPCService(service, profile)
	if err != nil {
		fmt.Println(err)
		return err
	}

	err = profile.Save()
	if err != nil {
		fmt.Println(err)
		return err
	}
	return nil
}

// InitGen init generate path of lucas
func InitGen(gen string) {
	genPath = gen
}
