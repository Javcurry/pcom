package lucas

import (
	"hago-plat/pcom/lucas/profiler"
	"path/filepath"
)

// RPCServiceSpec specifies an rpc service
type RPCServiceSpec struct {
	/**
	 * Grpc service Name
	 */
	ServiceName          string `json:"serviceName"`
	ServiceNameLowerCase string `json:"serviceNameLowerCase"`
	LowerCaseObjectsName string `json:"lowerCaseObjectName"`

	/**
	 * rpc functions
	 */
	Functions []FuncSpec `json:"functions"`

	/**
	 * The aggregated arguments definitions
	 */
	AggregatedArgsDefinitions []aggregateArgs `json:"aggregatedArgsDefinitions"`

	/**
	 * Resources that is in the same package as rpc service
	 */
	ResourceInfos []MessageInfo `json:"resourceInfos"`

	SpecBase
}

// NewRPCServiceSpec returns implement of RPCServiceSpec
func NewRPCServiceSpec() *RPCServiceSpec {
	return &RPCServiceSpec{}
}

// GetKind implement SpecModel interface
func (r *RPCServiceSpec) GetKind() profiler.Kind {
	return profiler.Kind(r.Kind)
}

// NewSpecModel implement SpecModel interface
func (r *RPCServiceSpec) NewSpecModel() profiler.SpecModel {
	return &RPCServiceSpec{}
}

// FuncSpec specifies an function
type FuncSpec struct {
	/*
	 * Function Name
	 */
	FuncName string `json:"funcName"`

	/*
	 * Args and response Aggregate info one message
	 */
	AggregateRequest  bool `json:"aggregateRequest"`
	AggregateResponse bool `json:"aggregateResponse"`

	/*
	 * rpc request and response definition (in pb)
	 */
	RequestDefinition  *FieldSpec `json:"requestDefinition"`
	ResponseDefinition *FieldSpec `json:"responseDefinition"`

	/**
	 * The aggregated arguments definitions (in lucas)
	 */
	AggregatedRequestDefinition  aggregateArgs `json:"aggregatedRequestDefinition"`
	AggregatedResponseDefinition aggregateArgs `json:"aggregatedResponseDefinition"`

	/**
	 * Function has argument type of context.Context
	 */
	WithCtx bool `json:"withCtx"`
}

const (
	aggregateArgsSuffixRequest  = "Request"
	aggregateArgsSuffixResponse = "Response"
)

type aggregateArgs struct {
	FuncName string      `json:"funcName"`
	Suffix   string      `json:"suffix"`
	Args     []FieldSpec `json:"args"`
}

// ScanRPCService starts rpc service scan
func ScanRPCService(service interface{}, profile *profiler.Profile) error {
	// scan
	scanner := NewRPCScanner(profile)
	rpcInfo, err := scanner.Scan(service)
	if err != nil {
		return err
	}
	profile.Set(rpcInfo.Path, filepath.Join(filepath.Dir(projectRoot), rpcInfo.Path), rpcInfo)
	return err
}
