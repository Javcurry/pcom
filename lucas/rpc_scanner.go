package lucas

import (
	"fmt"
	"hago-plat/pcom/lucas/profiler"
	"reflect"
	"strings"

	"github.com/pkg/errors"
)

// RPCScanner ...
type RPCScanner struct {
	RPCInfo *RPCServiceSpec
	profile *profiler.Profile
}

// NewRPCScanner ...
func NewRPCScanner(profile *profiler.Profile) *RPCScanner {
	scanner := &RPCScanner{
		RPCInfo: &RPCServiceSpec{},
		profile: profile,
	}
	scanner.RPCInfo.Kind = SpecKindRPCService
	scanner.RPCInfo.ImportsPKGs = make(map[string]string)
	scanner.RPCInfo.LucasImportPKGs = make(map[string]importPKG)
	return scanner
}

// Scan ...
func (s *RPCScanner) Scan(resource interface{}) (*RPCServiceSpec, error) {
	tType := reflect.TypeOf(resource)

	err := s.scanPkg(tType)
	if err != nil {
		return nil, err
	}
	fmt.Println("scanning: ", s.RPCInfo.Path)

	err = s.scanService(tType)
	if err != nil {
		return nil, err
	}
	err = s.scanMethods(tType)
	if err != nil {
		return nil, err
	}

	return s.RPCInfo, nil
}

/**
 * Scan package info
 */
func (s *RPCScanner) scanPkg(tType reflect.Type) error {
	pkgPath := ""

	if tType.Kind() == reflect.Ptr {
		pkgPath = tType.Elem().PkgPath()
	} else {
		pkgPath = tType.PkgPath()
	}

	s.RPCInfo.Path = pkgPath
	s.RPCInfo.Package = convertPkgName(pkgPath)

	index := strings.LastIndex(pkgPath, "/") + 1
	if index < 0 {
		index = 0
	}
	s.RPCInfo.GoPKGName = pkgPath[index:]
	return nil
}

/**
 * Scan service info
 */
func (s *RPCScanner) scanService(tType reflect.Type) error {
	name := ""
	if tType.Kind() == reflect.Ptr {
		name = tType.Elem().Name()

	} else {
		name = tType.Name()
	}

	// fmt.Println("service name: ", name)
	s.RPCInfo.ServiceName = name
	s.RPCInfo.ServiceNameLowerCase = UnCapitalize(name)
	pkgPathTrimService := strings.TrimSuffix(s.RPCInfo.Path, "/service")
	indexBeforeObjects := strings.LastIndex(pkgPathTrimService, "/")
	if indexBeforeObjects < 0 {
		indexBeforeObjects = 0
	}
	// maybe has bug
	s.RPCInfo.LowerCaseObjectsName = UnCapitalize(UnderscoreToCamelCase(pkgPathTrimService[indexBeforeObjects+1:]))
	return nil
}

/**
 * Scan Method info
 */
func (s *RPCScanner) scanMethods(tType reflect.Type) error {
	for i := 0; i < tType.NumMethod(); i++ {
		funcInfo := tType.Method(i)
		err := s.scanSingleMethod(funcInfo)
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *RPCScanner) scanSingleMethod(funcInfo reflect.Method) error {
	var rpcFuncInfo FuncSpec
	funcType := funcInfo.Type
	if funcType.NumIn() <= 1 {
		return errors.New("unsupported method type with 0 input")
	}

	rpcFuncInfo.FuncName = funcInfo.Name
	// fmt.Println("func name: ", funcInfo.Name)
	err := s.scanInput(&rpcFuncInfo, funcType)
	if err != nil {
		return err
	}

	err = s.scanOutput(&rpcFuncInfo, funcType)
	if err != nil {
		return err
	}
	s.RPCInfo.Functions = append(s.RPCInfo.Functions, rpcFuncInfo)
	return nil
}

func (s *RPCScanner) scanInput(rpcFuncInfo *FuncSpec, funcType reflect.Type) error {
	minArgNum := 2
	startNum := 1
	if isContext(funcType.In(startNum)) {
		rpcFuncInfo.WithCtx = true
		minArgNum++
		startNum++
	} else {
		rpcFuncInfo.WithCtx = false
	}

	// 仅当ctx以外参数数目为1， 且该参数类型为struct或struct的指针时，不用把参数聚合成XXRequest
	if funcType.NumIn() == minArgNum &&
		(funcType.In(minArgNum-1).Kind() == reflect.Struct ||
			(funcType.In(minArgNum-1).Kind() == reflect.Ptr &&
				funcType.In(minArgNum-1).Elem().Kind() == reflect.Struct)) {
		rpcFuncInfo.AggregateRequest = false
	} else {
		rpcFuncInfo.AggregateRequest = true
	}
	fieldGenerator := NewFieldScanner(s.profile)
	if !rpcFuncInfo.AggregateRequest {
		argType := funcType.In(minArgNum - 1)
		param := NewFieldParam{
			FieldNumber:    minArgNum - 1,
			Index:          minArgNum - 2,
			FuncName:       rpcFuncInfo.FuncName,
			ServiceName:    s.RPCInfo.ServiceName,
			ServicePKG:     s.RPCInfo.GoPKGName,
			ArgType:        argType,
			CurrentPKGPath: s.RPCInfo.Path,
			Input:          true,
		}
		fieldSpec, err := fieldGenerator.NewFieldSpecFromInputOutput(param)
		if err != nil {
			return err
		}
		s.appendArgsImports(fieldSpec)
		rpcFuncInfo.RequestDefinition = fieldSpec
	} else {
		var requestType aggregateArgs
		requestType.FuncName = rpcFuncInfo.FuncName
		requestType.Suffix = aggregateArgsSuffixRequest
		for i := startNum; i < funcType.NumIn(); i++ {
			argType := funcType.In(i)
			param := NewFieldParam{
				FieldNumber:    i - minArgNum + 2,
				Index:          i - 1,
				FuncName:       rpcFuncInfo.FuncName,
				ServiceName:    s.RPCInfo.ServiceName,
				ServicePKG:     s.RPCInfo.GoPKGName,
				ArgType:        argType,
				CurrentPKGPath: s.RPCInfo.Path,
				Input:          true,
			}
			fieldSpec, err := fieldGenerator.NewFieldSpecFromInputOutput(param)
			if err != nil {
				return err
			}
			s.appendArgsImports(fieldSpec)
			requestType.Args = append(requestType.Args, *fieldSpec)
		}
		rpcFuncInfo.RequestDefinition = &FieldSpec{
			TypeName:          rpcFuncInfo.FuncName + aggregateArgsSuffixRequest,
			TypeNameWithGoPkg: rpcFuncInfo.FuncName + aggregateArgsSuffixRequest,
			ProtobufTypeName:  rpcFuncInfo.FuncName + aggregateArgsSuffixRequest,
		}

		rpcFuncInfo.AggregatedRequestDefinition = requestType
		s.RPCInfo.AggregatedArgsDefinitions = append(s.RPCInfo.AggregatedArgsDefinitions, requestType)
	}
	return nil
}

func (s *RPCScanner) scanOutput(rpcFuncInfo *FuncSpec, funcType reflect.Type) error {
	minReturnNum := 1
	if funcType.NumOut() < minReturnNum {
		return errors.New("return number 0")
	}

	lastReturn := funcType.Out(funcType.NumOut() - 1)
	if !lastReturn.Implements(reflect.TypeOf(new(error)).Elem()) {
		return errors.New("last return is not error")
	}

	// 仅当除error外返回值数目为1， 且该参数类型为struct或struct的指针时，不用把参数聚合成XXRequest
	if funcType.NumOut() == minReturnNum+1 &&
		(funcType.Out(0).Kind() == reflect.Struct ||
			(funcType.Out(0).Kind() == reflect.Ptr &&
				funcType.Out(0).Elem().Kind() == reflect.Struct)) {
		rpcFuncInfo.AggregateResponse = false
	} else {
		rpcFuncInfo.AggregateResponse = true
	}

	fieldGenerator := NewFieldScanner(s.profile)
	if !rpcFuncInfo.AggregateResponse {
		argType := funcType.Out(0)
		param := NewFieldParam{
			FieldNumber:    1,
			Index:          1,
			FuncName:       rpcFuncInfo.FuncName,
			ServiceName:    s.RPCInfo.ServiceName,
			ServicePKG:     s.RPCInfo.GoPKGName,
			ArgType:        argType,
			CurrentPKGPath: s.RPCInfo.Path,
		}
		fieldSpec, err := fieldGenerator.NewFieldSpecFromInputOutput(param)
		if err != nil {
			return err
		}
		s.appendArgsImports(fieldSpec)
		rpcFuncInfo.ResponseDefinition = fieldSpec
	} else {
		var responseType aggregateArgs
		responseType.FuncName = rpcFuncInfo.FuncName
		responseType.Suffix = aggregateArgsSuffixResponse
		for i := 0; i < funcType.NumOut()-1; i++ {
			argType := funcType.Out(i)
			param := NewFieldParam{
				FieldNumber:    i + 1,
				Index:          i,
				FuncName:       rpcFuncInfo.FuncName,
				ServiceName:    s.RPCInfo.ServiceName,
				ServicePKG:     s.RPCInfo.GoPKGName,
				ArgType:        argType,
				CurrentPKGPath: s.RPCInfo.Path,
				Input:          false,
			}
			field, err := fieldGenerator.NewFieldSpecFromInputOutput(param)
			if err != nil {
				return err
			}
			s.appendArgsImports(field)
			responseType.Args = append(responseType.Args, *field)
		}
		rpcFuncInfo.ResponseDefinition = &FieldSpec{
			TypeName:          rpcFuncInfo.FuncName + aggregateArgsSuffixResponse,
			TypeNameWithGoPkg: rpcFuncInfo.FuncName + aggregateArgsSuffixResponse,
			ProtobufTypeName:  rpcFuncInfo.FuncName + aggregateArgsSuffixResponse,
		}
		rpcFuncInfo.AggregatedResponseDefinition = responseType
		s.RPCInfo.AggregatedArgsDefinitions = append(s.RPCInfo.AggregatedArgsDefinitions, responseType)
	}
	return nil

}

func (s *RPCScanner) appendArgsImports(field *FieldSpec) {
	if len(field.PkgPath) != 0 {
		if !field.NeedCast {
			s.RPCInfo.ImportsPKGs[field.PkgPath] = field.PkgPath
		}
		s.RPCInfo.LucasImportPKGs[field.PkgPath] = importPKG{
			PKGName: field.PkgPath,
			Alias:   RemoveHyphen(FolderSeparator2Underline(field.PkgPath)),
		}
	}
}
