package profiler

// Kind type int
type Kind int

// SpecModel ...
type SpecModel interface {
	GetKind() Kind
}

// SpecModelFactory ...
type SpecModelFactory interface {
	NewSpecModel() SpecModel
}

// Generator ...
type Generator interface {
	GenerateProto(model SpecModel) error
	GenLucas(model SpecModel) error
	ExecProtoc(model SpecModel) error
}
