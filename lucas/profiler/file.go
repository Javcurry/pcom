package profiler

// FileDesc is the description for for generated files
type FileDesc struct {
	FileDescBase

	// The specification for generated file
	GenerateSpec SpecModel `json:"generateSpec"`

	NewProto bool
}

// FileDescBase ...
type FileDescBase struct {
	// Path is relative path of the generate project as origin path
	Path string `json:"path"`

	// FileDesc update time
	// 用于判断是否有更新，需重新生成
	UpdateTime int64 `json:"updateTime"`

	// 源代码路径
	SourcePath string `json:"sourcePath"`

	Kind Kind `json:"kind"`
}
