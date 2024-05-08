package lucas

import (
	"hago-plat/pcom/lucas/profiler"
)

// SpecKind type Kind
type SpecKind profiler.Kind

// SpecKind
const (
	SpecKindInvalid SpecKind = iota
	SpecKindRPCService
	SpecKindResource
)

// SpecBase ...
type SpecBase struct {
	// Path is the path to generate .proto.
	// Package is generated from Path replacing "/" to ".", "-" to "" and "_" to "".
	// GoPKGName is the package of generate .go.
	//
	// Example:
	//    Path     : hago-plat/hagonetes/api/user/model
	//    Package  : hagoplat.hagonetes.api.user.model
	//    GoPKGName: model
	Path      string `json:"path"`
	Package   string `json:"package"`
	GoPKGName string `json:"goPkgName"`

	/*
	 * imports pkgs paths
	 */
	ImportsPKGs     map[string]string    `json:"importsPkgs"`
	LucasImportPKGs map[string]importPKG `json:"lucasImportPkGs"`

	Kind SpecKind `json:"kind"`
}

type importPKG struct {
	PKGName string `json:"pkgName"`
	Alias   string `json:"alias"`
}
