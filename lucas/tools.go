package lucas

import (
	"context"
	"io"
	"os"
	"os/exec"
	"reflect"
	"strings"
	"unicode"
)

var mapTypeName2ProtobufType = map[string]string{
	"float64": "double",
	"float32": "float",
	"int":     "int32",
	"int8":    "int32",
	"int16":   "int32",
	"int32":   "int32",
	"int64":   "int64",
	"uint":    "uint32",
	"uint8":   "uint32",
	"uint16":  "uint32",
	"uint32":  "uint32",
	"uint64":  "uint64",
	"bool":    "bool",
	"string":  "string",
	"[]byte":  "bytes",
}

var mapProtobufType2TypeName = map[string]string{
	"double": "float64",
	"float":  "float32",
	"int32":  "int32",
	"int64":  "int64",
	"uint32": "uint32",
	"uint64": "uint64",
	"bool":   "bool",
	"string": "string",
	"bytes":  "[]byte",
}

var mapBasicTypeNameNeedCast = map[string]string{
	"int":    "int32",
	"int8":   "int32",
	"int16":  "int32",
	"uint":   "uint32",
	"uint8":  "uint32",
	"uint16": "uint32",
}

func convertPkgName(pkgPath string) string {
	out := FolderSeparator2Dot(pkgPath)
	out = removeUnderline(out)
	return RemoveHyphen(out)
}

func joinPkgPathAndType(pkgPath, typeName, currentPkgPath string) string {
	if len(pkgPath) > 0 {
		if convertPkgName(pkgPath) == currentPkgPath {
			return typeName
		}
		return pkgPath + "." + typeName
	}
	return typeName
}

// FolderSeparator2Dot ...
func FolderSeparator2Dot(in string) string {
	return strings.Replace(in, "/", ".", -1)
}

// FolderSeparator2Underline ...
func FolderSeparator2Underline(in string) string {
	return strings.Replace(in, "/", "_", -1)
}

func removeUnderline(in string) string {
	return strings.Replace(in, "_", "", -1)
}

// RemoveHyphen ...
func RemoveHyphen(in string) string {
	return strings.Replace(in, "-", "", -1)
}

func isContext(t reflect.Type) bool {
	return t.Implements(reflect.TypeOf(new(context.Context)).Elem())
}

func typeName2ProtobufType(kindString, typeName string) string {
	if typeName == "bytes" {
		return typeName
	}
	if pbType, ok := mapTypeName2ProtobufType[kindString]; ok {
		return pbType
	}
	return typeName
}

func protobufType2typeName(rotobufTypeName string) string {
	return mapProtobufType2TypeName[rotobufTypeName]
}

func typeNeedCast(kindString, typeName string) bool {
	if kindString == "struct" {
		return false
	}
	if typeName == "bytes" {
		return false
	}
	_, ok := mapTypeName2ProtobufType[typeName]
	if !ok {
		return true
	}
	_, ok = mapBasicTypeNameNeedCast[typeName]
	if ok {
		return true
	}
	if kindString != typeName {
		return true
	}
	return false
}

// CamelCaseToUnderscore converts from camel case form to underscore separated form.
// Ex.: MyFunc => my_func
//      APIVersion => api_version
func CamelCaseToUnderscore(str string) string {
	var output []rune
	var segment []rune
	prevStrIsUpper := false
	prevStrIsNumber := false
	for i, r := range str {
		if unicode.IsUpper(r) &&
			(!prevStrIsUpper || i < len(str)-1 && unicode.IsLower(rune(str[i+1])) ||
				prevStrIsNumber && unicode.IsUpper(r)) {
			output = addSegment(output, segment)
			segment = nil
		}
		if unicode.IsUpper(r) {
			prevStrIsUpper = true
		} else if unicode.IsLower(r) {
			prevStrIsUpper = false
		}
		if unicode.IsNumber(r) {
			prevStrIsNumber = true
		} else if !unicode.IsNumber(r) {
			prevStrIsNumber = false
		}
		segment = append(segment, unicode.ToLower(r))
	}
	output = addSegment(output, segment)
	return string(output)
}

func addSegment(inrune, segment []rune) []rune {
	if len(segment) == 0 {
		return inrune
	}
	if len(inrune) != 0 {
		inrune = append(inrune, '_')
	}
	inrune = append(inrune, segment...)
	return inrune
}

// UnderscoreToCamelCase converts from underscore separated form to camel case form.
// Ex.: my_func => MyFunc
func UnderscoreToCamelCase(s string) string {
	return strings.Replace(strings.Title(strings.Replace(strings.ToLower(s), "_", " ", -1)), " ", "", -1)
}

// UnCapitalize 首字母小写
func UnCapitalize(s string) string {
	var output []rune
	for i, r := range s {
		if i == 0 && unicode.IsUpper(r) {
			o := unicode.ToLower(r)
			output = append(output, o)
		} else {
			output = append(output, r)
		}
	}
	return string(output)
}

// RunCmdAndPrint ...
func RunCmdAndPrint(cmd *exec.Cmd, stdoutPipe io.ReadCloser, stderrPipe io.ReadCloser) error {
	err := cmd.Start()
	if err != nil {
		return err
	}
	go func() {
		_, _ = io.Copy(os.Stdout, stdoutPipe)
	}()

	go func() {
		_, _ = io.Copy(os.Stderr, stderrPipe)
	}()
	err = cmd.Wait()
	if err != nil {
		return err
	}
	return nil
}
