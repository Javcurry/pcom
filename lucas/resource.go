package lucas

import (
	"fmt"
	"hago-plat/pcom/lucas/profiler"
	"path/filepath"
	"reflect"
	"sort"
)

// ResourceSpec specifies resource
type ResourceSpec struct {
	// MessageInfoMap
	// key : message name
	MessageInfoMap map[string]*MessageInfo `json:"messageInfoMap"`
	SpecBase
}

// NewResourceType returns ResourceSpec implement
func NewResourceType() *ResourceSpec {
	rs := &ResourceSpec{
		SpecBase: SpecBase{
			Kind:            SpecKindResource,
			ImportsPKGs:     make(map[string]string),
			LucasImportPKGs: make(map[string]importPKG),
		},
		MessageInfoMap: make(map[string]*MessageInfo),
	}
	return rs
}

// GetKind  implement SpecModel interface
func (r *ResourceSpec) GetKind() profiler.Kind {
	return profiler.Kind(r.Kind)
}

// NewSpecModel implement SpecModel interface
func (r *ResourceSpec) NewSpecModel() profiler.SpecModel {
	return NewResourceType()
}

// MessageInfo specifies message
type MessageInfo struct {
	// message name
	Name string `json:"name"`

	// fields
	FieldMap  map[string]FieldSpec `json:"fieldMap"`
	FieldList []FieldSpec          `json:"fieldList"`
}

// SortFieldMapToFieldList sort data in FieldMap by FieldNumber
// and put it into FieldList
func (m *MessageInfo) SortFieldMapToFieldList() {
	if m == nil {
		return
	}
	m.FieldList = []FieldSpec{}
	m.FieldList = make([]FieldSpec, 0, len(m.FieldMap))
	for _, v := range m.FieldMap {
		m.FieldList = append(m.FieldList, v)
	}
	sort.Slice(m.FieldList, func(i, j int) bool {
		return m.FieldList[i].FieldNumber < m.FieldList[j].FieldNumber
	})
}

// ScanResource starts resource scan
func ScanResource(resource reflect.Type, profile *profiler.Profile) (*ResourceSpec, error) {
	// scan
	resInfo, err := NewResourceScanner(resource, profile).Scan(resource)
	if err != nil {
		fmt.Println("scan resource", resource.Name(), "fail", err)
		return nil, err
	}
	resInfoOld, found := profile.Get(resInfo.Path)
	if found {
		for k, v := range resInfoOld.(*ResourceSpec).MessageInfoMap {
			resInfo.MessageInfoMap[k] = v
		}
	}
	profile.Set(resInfo.Path, filepath.Join(filepath.Dir(projectRoot), resInfo.Path), resInfo)
	return resInfo, nil
}
