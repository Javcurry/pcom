package lucas

import (
	"fmt"
	"go/parser"
	"go/token"
	"hago-plat/pcom/lucas/profiler"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"

	"github.com/pkg/errors"
)

// FieldSpec defines arguments info
type FieldSpec struct {
	//Name of argument type
	TypeName            string `json:"typeName"`
	TypeNameWithGoPkg   string `json:"typeNameWithGoPkg"`
	ProtobufTypeName    string `json:"protobufTypeName"`
	TypeGenFromProtobuf string `json:"typeGenFromProtobuf"`
	// FieldName is the name in source code
	// ProtobufFieldName is underscore case of FieldName
	FieldName                string `json:"fieldName"`
	ProtobufFieldName        string `json:"protobufFieldName"`
	FieldNameGenFromProtobuf string `json:"fieldNameGenFromProtobuf"`

	//PkgPath is the path of type. Empty if type is kind of golang builtin type.
	//Package is generated from PkgPath replacing "/" to ".", "-" to "" and "_" to "".
	//
	//Example:
	//   PkgPath : hago-plat/hagonetes/hagonetes_info_center_d/objects/user/service
	//   Package : hagoplat.hagonetes.hagonetesinfocenterd.objects.user.service
	//
	PkgPath   string `json:"pkgPath"`
	Package   string `json:"package"`
	GoPkgName string `json:"goPkgName"`

	//protobuf type//

	//Need static cast for type in .pb.go to original type
	//Example:
	//    int：
	//         type in .proto is int32
	//         type in generated .pb.go is int32
	//         need cast to int in .lucas.go
	NeedCast        bool   `json:"needCast"`
	CastType        string `json:"castType"`
	NeedCastInLucas bool   `json:"needCastInLucas"`

	//FieldNumber protobuf field number
	FieldNumber int `json:"fieldNumber"`

	IsPtr                   bool `json:"isPtr"`
	IsStruct                bool `json:"isStruct"`
	IsStructInPtr           bool `json:"isStructInPtr"`
	IsStructInRepeated      bool `json:"isStructInRepeated"`
	IsStructInMapValue      bool `json:"isStructInMapValue"`
	IsPtrOfStructInMapValue bool `json:"isPtrOfStructInMapValue"`
	IsPtrInRepeated         bool `json:"isPtrInRepeated"`
	IsPtrInMapValue         bool `json:"isPtrInMapValue"`
	IsRepeated              bool `json:"isRepeated"`
	IsMap                   bool `json:"isMap"`

	//Empty when IsMap is false,
	KeyTypeName            string `json:"keyTypeName"`
	KeyTypeNameWithGoPkg   string `json:"keyTypeNameWithGoPkg"`
	KeyPkgPath             string `json:"keyPkgPath"`
	KeyPackage             string `json:"keyPackage"`
	KeyProtobufTypeName    string `json:"keyProtobufTypeName"`
	KeyTypeGenFromProtobuf string `json:"keyTypeGenFromProtobuf"`
	KeyNeedCast            bool   `json:"keyNeedCast"`
	KeyCastType            string `json:"keyCastType"`

	// gogoprotobuf field options
	GoGoProtoFieldOptions    string `json:"goGoProtoFieldOption"`
	HasGoGoProtoFieldOptions bool   `json:"hasGoGoProtoFieldOptions"`

	// if true, message option casttype, castkey and cast value will disable
	DisableGoGoCast bool `json:"unableGoGoCast"`
}

// FieldScanner ...
type FieldScanner struct {
	profile *profiler.Profile
	fieldNumberMgr
}

// NewFieldScanner returns FieldScanner
func NewFieldScanner(profile *profiler.Profile) *FieldScanner {
	f := &FieldScanner{
		profile: profile,
	}

	f.initialize()
	return f
}

// NewFieldSpec returns FieldSpec from Method Input or Output
func (f *FieldScanner) NewFieldSpec(index int, fieldName string,
	aType reflect.Type, currentPkgPath string, disableCast bool) (*FieldSpec, error) {
	var argInfo FieldSpec
	argInfo.FieldNumber = index
	argInfo.DisableGoGoCast = disableCast
	if aType.Kind() == reflect.Ptr {
		argInfo.IsPtr = true
		if aType.Elem().Kind() == reflect.Struct {
			argInfo.IsStructInPtr = true
		} else {
			argInfo.IsStructInPtr = false
		}
	} else {
		argInfo.IsPtr = false
	}

	if aType.Kind() == reflect.Struct {
		argInfo.IsStruct = true
	}

	if aType.Kind() == reflect.Array || aType.Kind() == reflect.Slice {
		argInfo.IsRepeated = true
		if aType.Elem().Kind() == reflect.Struct {
			argInfo.IsStructInRepeated = true
		} else {
			argInfo.IsStructInRepeated = false
		}
		if aType.Elem().Kind() == reflect.Ptr {
			argInfo.IsPtrInRepeated = true
		} else {
			argInfo.IsPtrInRepeated = false
		}
	} else {
		argInfo.IsRepeated = false
	}

	if aType.Kind() == reflect.Map {
		argInfo.IsMap = true
		argInfo.IsStructInMapValue = false
		argInfo.IsPtrInMapValue = false
		if aType.Elem().Kind() == reflect.Struct {
			argInfo.IsStructInMapValue = true
		} else if aType.Elem().Kind() == reflect.Ptr {
			argInfo.IsPtrInMapValue = true
			if aType.Elem().Elem().Kind() == reflect.Struct {
				argInfo.IsPtrOfStructInMapValue = true
			}
		}
	} else {
		argInfo.IsMap = false
	}

	if aType.Kind() == reflect.Func || aType.Kind() == reflect.Chan ||
		aType.Kind() == reflect.UnsafePointer || aType.Kind() == reflect.Uintptr ||
		aType.Kind() == reflect.Complex64 || aType.Kind() == reflect.Complex128 {
		return nil, errors.New("unsupported kind of arg: " + aType.String())
	}
	// todo 通用情况需递归处理
	eType := aType
	if argInfo.IsPtr || argInfo.IsRepeated || argInfo.IsMap {
		eType = aType.Elem()
	}
	if argInfo.IsPtrInRepeated || argInfo.IsPtrInMapValue {
		eType = aType.Elem().Elem()
	}
	argInfo.FieldName = fieldName
	argInfo.ProtobufFieldName = CamelCaseToUnderscore(fieldName)
	argInfo.FieldNameGenFromProtobuf = UnderscoreToCamelCase(argInfo.ProtobufFieldName)

	argInfo.TypeName = eType.Name()
	if len(eType.PkgPath()) != 0 {
		argInfo.TypeNameWithGoPkg = RemoveHyphen(FolderSeparator2Underline(eType.PkgPath())) + "." + eType.Name()
	} else {
		if argInfo.TypeString() == "bytes" {
			argInfo.TypeNameWithGoPkg = "[]byte"
		} else {
			argInfo.TypeNameWithGoPkg = eType.String()
		}
	}
	argInfo.PkgPath = eType.PkgPath()
	argInfo.Package = convertPkgName(argInfo.PkgPath)
	argInfo.ProtobufTypeName = typeName2ProtobufType(eType.Kind().String(), argInfo.TypeString())
	argInfo.NeedCast = typeNeedCast(eType.Kind().String(), argInfo.TypeString())
	argInfo.CastType = joinPkgPathAndType(eType.PkgPath(), eType.Name(), currentPkgPath)

	// 对bytes进行特殊处理
	if argInfo.ProtobufTypeName == "bytes" {
		argInfo.IsRepeated = false
	}

	// fmt.Println("   arg", index, "type name with go pkg: ", eType.String())
	// fmt.Println("   arg", index, "source type name: ", eType.Name())
	// fmt.Println("   arg", index, "pkg: ", eType.PkgPath())
	// fmt.Println("   arg", index, "pb type name: ", argInfo.ProtobufTypeName)
	// fmt.Println("   arg", index, "cast type name: ", argInfo.CastType)

	argInfo.KeyNeedCast = false
	if argInfo.IsMap {
		dType := aType.Key()
		argInfo.KeyTypeName = dType.Name()
		argInfo.KeyTypeNameWithGoPkg = dType.String()
		argInfo.KeyPkgPath = dType.PkgPath()
		argInfo.KeyPackage = convertPkgName(argInfo.PkgPath)
		argInfo.KeyProtobufTypeName = typeName2ProtobufType(dType.Kind().String(), argInfo.KeyTypeString())
		argInfo.KeyNeedCast = typeNeedCast(dType.Kind().String(), argInfo.KeyTypeString())
		argInfo.KeyCastType = joinPkgPathAndType(dType.PkgPath(), dType.Name(), currentPkgPath)

		//fmt.Println("   arg", index, "key name: ",dType.Kind().String())
		//fmt.Println("   arg", index, "key pkg: ", argInfo.KeyTypeString())
	}
	//if argInfo.IsMap {
	//	fmt.Println("need cast:", argInfo.NeedCast, argInfo.KeyNeedCast, eType.Name())
	//}
	if argInfo.NeedCast || (argInfo.IsMap && argInfo.KeyNeedCast) {
		argInfo.NeedCastInLucas = true
		argInfo.TypeGenFromProtobuf = protobufType2typeName(argInfo.ProtobufTypeName)
		if argInfo.IsMap {
			argInfo.KeyTypeGenFromProtobuf = protobufType2typeName(argInfo.KeyProtobufTypeName)
		}
	} else {
		argInfo.NeedCastInLucas = false
	}
	if eType.Name() == "Package" {
		fmt.Println("----------- pkg:", argInfo.Package, "currentPkg:", currentPkgPath)
	}
	if len(argInfo.PkgPath) != 0 &&
		eType.Kind() == reflect.Struct &&
		argInfo.PkgPath != currentPkgPath {
		_, err := ScanResource(eType, f.profile)
		if err != nil {
			return nil, err
		}
	}
	if argInfo.IsMap &&
		aType.Key().Kind() == reflect.Struct &&
		len(argInfo.KeyPkgPath) != 0 &&
		argInfo.KeyPkgPath != currentPkgPath {
		cType := aType.Key()
		_, err := ScanResource(cType, f.profile)
		if err != nil {
			return nil, err
		}
	}
	argInfo.genGoGoProtobufFieldOptions()
	return &argInfo, nil
}

// NewFieldParam is input param of function FieldScanner.NewFieldSpecFromInputOutput
type NewFieldParam struct {
	FieldNumber    int
	Index          int
	FuncName       string
	ServiceName    string
	ServicePKG     string
	ArgType        reflect.Type
	CurrentPKGPath string
	Input          bool
}

// NewFieldSpecFromInputOutput returns FieldSpec from Method Input or Output
func (f *FieldScanner) NewFieldSpecFromInputOutput(param NewFieldParam) (*FieldSpec, error) {
	fs := token.NewFileSet()
	path := filepath.Join(filepath.Dir(projectRoot), param.CurrentPKGPath)
	pkgs, err := parser.ParseDir(fs, path, nil, 0)
	if err != nil {
		return nil, err
	}
	fieldName := FieldName(pkgs[param.ServicePKG], param.Index, param.FuncName, param.ServiceName, param.Input)
	return f.NewFieldSpec(param.FieldNumber, fieldName, param.ArgType, param.CurrentPKGPath, true)
}

// NewFieldSpecFromStructField returns FieldSpec from StructField
func (f *FieldScanner) NewFieldSpecFromStructField(field reflect.StructField, currentPkgPath string) (*FieldSpec, error) {
	if field.Anonymous {
		return nil, errors.New("anonymous field will unfold")
	}

	index, err := f.getIndex(field)
	if err != nil {
		fmt.Println("tag format error", err)
		return nil, err
	}
	return f.NewFieldSpec(index, field.Name, field.Type, currentPkgPath, false)
}

func (f *FieldScanner) getIndex(field reflect.StructField) (int, error) {
	indexString := field.Tag.Get("lucas")
	if len(indexString) != 0 {
		index, err := strconv.Atoi(indexString)
		if err != nil {
			fmt.Println("invalid lucas tag")
			return 0, err
		}
		err = f.validateFieldNumFromTag(index, field.Name)
		return index, err
	}
	index := f.getValidFieldNumber(field.Name)
	return index, nil
}

// TypeString returns type name
func (f *FieldSpec) TypeString() string {
	if len(f.PkgPath) == 0 {
		if f.IsRepeated && f.TypeName == "uint8" {
			return "bytes"
		}
		return f.TypeName
	}
	if _, ok := mapTypeName2ProtobufType[f.TypeName]; ok {
		return f.TypeName
	}
	return f.Package + "." + f.TypeName
}

// KeyTypeString returns key type name when FieldSpec.IsMap is true
func (f *FieldSpec) KeyTypeString() string {
	if len(f.PkgPath) == 0 {
		if f.IsRepeated && f.TypeName == "uint8" {
			return "bytes"
		}
		return f.KeyTypeName
	}
	if _, ok := mapTypeName2ProtobufType[f.KeyTypeName]; ok {
		return f.KeyTypeName
	}
	return f.KeyPackage + "." + f.KeyTypeName
}

func (f *FieldSpec) genGoGoProtobufFieldOptions() {
	var options []string
	if !f.DisableGoGoCast {
		if f.IsMap {
			if f.KeyNeedCast {
				option := fmt.Sprintf("(gogoproto.castkey) = \"%v\"", f.KeyCastType)
				options = append(options, option)
			}
			if f.NeedCast {
				option := fmt.Sprintf("(gogoproto.castvalue) = \"%v\"", f.CastType)
				options = append(options, option)
			}
		} else {
			if f.NeedCast {
				option := fmt.Sprintf("(gogoproto.casttype) = \"%v\"", f.CastType)
				options = append(options, option)
			}
		}
	}
	//if !f.DisableGoGoCast {
	if f.FieldName != UnderscoreToCamelCase(f.ProtobufFieldName) {
		option := fmt.Sprintf("(gogoproto.customname)= \"%v\"", f.FieldName)
		options = append(options, option)
	}
	//}
	if f.IsStruct || f.IsStructInRepeated || f.IsStructInMapValue {
		option := "(gogoproto.nullable) = false"
		options = append(options, option)
	}
	f.GoGoProtoFieldOptions = strings.Join(options, ",")
	if len(f.GoGoProtoFieldOptions) > 0 {
		f.HasGoGoProtoFieldOptions = true
	} else {
		f.HasGoGoProtoFieldOptions = false
	}
}

type fieldNumberMgr struct {
	Counter int
	Markup  map[int]string
}

func (f *fieldNumberMgr) initialize() {
	f.Counter = 1
	f.Markup = make(map[int]string)
}

func (f *fieldNumberMgr) getValidFieldNumber(fieldName string) int {
	index := f.Counter
	for {
		if _, ok := f.Markup[index]; !ok {
			f.Markup[index] = fieldName
			break
		}
		f.Counter++
		index = f.Counter
	}
	return index
}

func (f *fieldNumberMgr) validateFieldNumFromTag(fieldNumber int, fieldName string) error {
	existedFieldName, ok := f.Markup[fieldNumber]
	if ok {
		errMsg := fmt.Sprintf("field number of %v conflict with %v. ", fieldNumber, existedFieldName)
		return errors.New(errMsg)
	}
	f.Markup[fieldNumber] = fieldName
	return nil
}
