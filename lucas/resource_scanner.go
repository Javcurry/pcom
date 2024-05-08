package lucas

import (
	"fmt"
	"hago-plat/pcom/lucas/profiler"
	"reflect"
	"strings"
)

// ResourceScanner ...
type ResourceScanner struct {
	ResInfo *ResourceSpec
	profile *profiler.Profile
	*FieldScanner
}

// NewResourceScanner returns ResourceScanner object
func NewResourceScanner(resType reflect.Type, profile *profiler.Profile) *ResourceScanner {
	resScanner := &ResourceScanner{
		profile:      profile,
		FieldScanner: NewFieldScanner(profile),
	}

	spec, found := profile.Get(resType.PkgPath())
	if found {
		var ok bool
		resScanner.ResInfo, ok = spec.(*ResourceSpec)
		if !ok {
			resScanner.ResInfo = NewResourceType()
		}
	} else {
		resScanner.ResInfo = NewResourceType()
	}

	return resScanner
}

// Scan return scan result of ResourceSpec
func (r *ResourceScanner) Scan(resType reflect.Type) (*ResourceSpec, error) {
	err := r.scanPkg(resType)
	if err != nil {
		fmt.Println("scan resource pkg fail", err)
		return nil, err
	}

	//fmt.Println("resource type:", resType)
	var messageInfo MessageInfo
	if _, ok := r.ResInfo.MessageInfoMap[resType.Name()]; !ok {
		messageInfo.FieldMap = make(map[string]FieldSpec)
		messageInfo.Name = resType.Name()
		r.ResInfo.MessageInfoMap[resType.Name()] = &messageInfo
	}
	err = r.scanFields(resType.Name(), resType)
	if err != nil {
		fmt.Println("scan fields pkg fail", err)
		return nil, err
	}

	r.ResInfo.MessageInfoMap[resType.Name()].SortFieldMapToFieldList()

	//fmt.Printf("resource scan Path: %#v\n", r.ResInfo.Path)
	// profile

	return r.ResInfo, nil
}

/**
 * Scan package info
 */
func (r *ResourceScanner) scanPkg(tType reflect.Type) error {
	pkgPath := ""

	if tType.Kind() == reflect.Ptr {
		pkgPath = tType.Elem().PkgPath()
	} else {
		pkgPath = tType.PkgPath()
	}

	r.ResInfo.Path = pkgPath
	r.ResInfo.Package = convertPkgName(pkgPath)

	index := strings.LastIndex(pkgPath, "/") + 1
	if index < 0 {
		index = 0
	}
	r.ResInfo.GoPKGName = pkgPath[index:]
	return nil
}

func (r *ResourceScanner) scanFields(messageName string, tType reflect.Type) error {
	if tType.Kind() == reflect.Struct {
		for i := 0; i < tType.NumField(); i++ {
			if tType.Field(i).Anonymous {
				err := r.scanFields(messageName, tType.Field(i).Type)
				if err != nil {
					return err
				}
			} else {
				err := r.scanSingleField(messageName, tType.Field(i))
				if err != nil {
					return err
				}
			}
		}
	} else {
		fmt.Println("into non-struct. name:", tType.Name(), "pkg:", tType.PkgPath(), "kind:", tType.Kind(), "messageName:", messageName)
	}
	return nil
}

func (r *ResourceScanner) scanSingleField(messageName string, structField reflect.StructField) error {
	// fmt.Println("field ", structField.Name, "type:", structField.Type, structField.Type.PkgPath(), "currentPkg:", r.ResInfo.Package)
	fieldSpec, err := r.NewFieldSpecFromStructField(structField, r.ResInfo.Package)
	// fmt.Println("field pkg", fieldSpec.PkgPath, fieldSpec.Package)
	if err != nil {
		fmt.Println(err)
		return err
	}
	r.appendArgsImports(fieldSpec)
	var messageInfo *MessageInfo
	if info, ok := r.ResInfo.MessageInfoMap[messageName]; ok {
		messageInfo = info
	}
	messageInfo.FieldMap[structField.Name] = *fieldSpec
	r.ResInfo.MessageInfoMap[messageName] = messageInfo
	return nil
}

func (r *ResourceScanner) appendArgsImports(field *FieldSpec) {
	if len(field.PkgPath) != 0 && field.Package != r.ResInfo.Package &&
		(field.IsStruct || field.IsStructInPtr || field.IsStructInRepeated || field.IsStructInMapValue || field.IsPtrOfStructInMapValue) {
		r.ResInfo.ImportsPKGs[field.PkgPath] = field.PkgPath
	}
}
