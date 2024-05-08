package profiler

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/pkg/errors"
)

// Profile ...
type Profile struct {
	// generate path
	GeneratePath string

	// profile path
	ProfilePath string

	// FileObjectMap contain all file object description
	// key : FileDesc.SourcePath
	FileObjectMap map[string]*FileDesc

	GeneratorMap map[Kind]Generator
	FactoryMap   map[Kind]SpecModelFactory

	basePath string
	MLAdapter
}

// NewProfile returns Profile object
//     Profile.GeneratePath = base + gen
//     Profile.ProfilePath = GeneratePath + profileDirName
func NewProfile(base, gen string) *Profile {
	p := &Profile{
		GeneratePath: filepath.Join(base, gen),
		ProfilePath:  filepath.Join(base, gen, profileDirName),
		MLAdapter:    &JSONCompiler{},
	}
	p.FileObjectMap = make(map[string]*FileDesc)
	p.GeneratorMap = make(map[Kind]Generator)
	p.FactoryMap = make(map[Kind]SpecModelFactory)
	p.basePath = base
	return p
}

// Save save profile to .lucas
func (p *Profile) Save() error {
	err := os.MkdirAll(p.ProfilePath, 0755)
	if err != nil {
		fmt.Println("save fail: ", err.Error())
		return err
	}

	for path, object := range p.FileObjectMap {
		filePath := filepath.Join(p.ProfilePath, path)
		content, err := p.MarshalToString(object)
		if err != nil {
			fmt.Println("encode profile fail, err:", object.SourcePath)
			return err
		}
		err = os.MkdirAll(filePath, 0755)
		if err != nil {
			fmt.Println("mkdir", filePath, "fail. err:", err.Error())
			return err
		}
		fileName := filepath.Join(filePath, profileName)
		err = func() error {
			file, err := os.OpenFile(fileName, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0755)
			defer func() { _ = file.Close() }()
			if err != nil {
				fmt.Println("open file", filePath, "fail. err:", err.Error())
				return err
			}
			w := bufio.NewWriter(file)
			_, err = w.WriteString(content)
			if err != nil {
				fmt.Println("bufio write buffer", filePath, "fail. err:", err.Error())
				return err
			}
			err = w.Flush()
			if err != nil {
				fmt.Println("write file", filePath, "fail. err:", err.Error())
			}
			return err
		}()
		if err != nil {
			return err
		}

	}
	return nil
}

// Load load profile from .lucas
func (p *Profile) Load() error {
	_, err := os.Stat(p.ProfilePath)
	if os.IsNotExist(err) {
		fmt.Println(p.ProfilePath, "not exists")
		return nil
	}

	err = filepath.Walk(p.ProfilePath, p.profilerFileWalk)
	if err != nil {
		return err
	}
	return nil
}

func (p *Profile) profilerFileWalk(path string, info os.FileInfo, err error) error {
	if err != nil {
		fmt.Println("walk func err:", err)
		return err
	}
	if info.IsDir() {
		return nil
	}
	if !strings.HasSuffix(path, ".json") {
		return nil
	}

	file, err := os.Open(path)
	defer func() { _ = file.Close() }()
	if err != nil {
		fmt.Println("open file", info.Name(), "fail, err:", err)
		return err
	}
	json, err := ioutil.ReadAll(file)
	if err != nil {
		fmt.Println("read file", info.Name(), "fail, err:", err)
		return err
	}
	filePath := filepath.Dir(path)
	relativePath, _ := filepath.Rel(p.ProfilePath, filePath)
	var fileDescBase FileDescBase
	err = p.Unmarshal(json, &fileDescBase)
	if err != nil {
		fmt.Println("unmarshal base", path)
		return err
	}
	fileDesc := &FileDesc{}

	fileDesc.GenerateSpec = p.FactoryMap[fileDescBase.Kind].NewSpecModel()

	p.FileObjectMap[relativePath] = fileDesc
	err = p.Unmarshal(json, p.FileObjectMap[relativePath])
	if err != nil {
		fmt.Println("unmarshal", path)
		return err
	}
	return nil
}

// Set spec info profile
func (p *Profile) Set(path, sourcePath string, spec SpecModel) {
	p.FileObjectMap[path] = &FileDesc{
		FileDescBase: FileDescBase{
			Path:       path,
			UpdateTime: time.Now().UnixNano(),
			SourcePath: sourcePath,
			Kind:       spec.GetKind(),
		},
		GenerateSpec: spec,
	}
}

// Get returns spec from profile
func (p *Profile) Get(path string) (SpecModel, bool) {
	fileSpec, ok := p.FileObjectMap[path]
	if !ok {
		return nil, ok
	}
	return fileSpec.GenerateSpec, ok
}

// RegisterGenerator register generator
func (p *Profile) RegisterGenerator(kind Kind, generator Generator) {
	p.GeneratorMap[kind] = generator
}

// StartGeneration start generation
func (p *Profile) StartGeneration() error {
	for _, fo := range p.FileObjectMap {
		//if !p.IsModified(path) {
		//	fo.NewProto = false
		//	continue
		//} else {
		//	fo.NewProto = true
		//}
		generator, ok := p.GeneratorMap[fo.GenerateSpec.GetKind()]
		if !ok {
			return errors.New("generator not found")
		}
		err := generator.GenerateProto(fo.GenerateSpec)
		if err != nil {
			return err
		}
		err = generator.GenLucas(fo.GenerateSpec)
		if err != nil {
			return err
		}
	}
	err := p.ExecProtoc()
	if err != nil {
		return err
	}
	return nil
}

// ExecProtoc execute protoc in shell
func (p *Profile) ExecProtoc() error {
	for _, fo := range p.FileObjectMap {
		//if !fo.NewProto {
		//	continue
		//}
		generator, ok := p.GeneratorMap[fo.GenerateSpec.GetKind()]
		if !ok {
			return errors.New("generator not found")
		}
		err := generator.ExecProtoc(fo.GenerateSpec)
		if err != nil {
			return err
		}
	}
	return nil
}

// RegisterFactory register desc model factory
func (p *Profile) RegisterFactory(kind Kind, factory SpecModelFactory) {
	p.FactoryMap[kind] = factory
}

//func (p *Profile) IsModified(path string) bool {
//	sourcePath := filepath.Join(filepath.Dir(p.basePath), path)
//	genProtoFilePath := filepath.Join(p.GeneratePath, path, "generated.proto")
//	stat, err := os.Stat(genProtoFilePath)
//	if err != nil {
//		fmt.Println("IsModified open:", err)
//		if os.IsNotExist(err) {
//			return true
//		}
//	}
//
//	lastUpdateTime := stat.ModTime().UnixNano()
//	newestFileUpdateTime := lastUpdateTime
//	files, err := ioutil.ReadDir(sourcePath)
//	if err != nil {
//		fmt.Println("IsModified:", err)
//		return true
//	}
//	for i := 0; i < len(files); i++ {
//		if files[i].ModTime().UnixNano() > lastUpdateTime {
//			newestFileUpdateTime = files[i].ModTime().UnixNano()
//		}
//	}
//	if newestFileUpdateTime > lastUpdateTime {
//		fmt.Println("is modified", path)
//	}
//	return newestFileUpdateTime > lastUpdateTime
//}
