package profiler

import jsoniter "github.com/json-iterator/go"

// Encoder ...
type Encoder interface {
	Marshal(v interface{}) ([]byte, error)
	MarshalToString(v interface{}) (string, error)
}

// Decoder ...
type Decoder interface {
	Unmarshal(data []byte, v interface{}) error
	UnmarshalFromString(data string, v interface{}) error
}

// MLAdapter ...
type MLAdapter interface {
	Encoder
	Decoder
}

// JSONCompiler ...
type JSONCompiler struct{}

// Marshal ...
func (j *JSONCompiler) Marshal(v interface{}) ([]byte, error) {
	return jsoniter.MarshalIndent(v, "", " ")
}

// MarshalToString ...
func (j *JSONCompiler) MarshalToString(v interface{}) (string, error) {
	json, err := jsoniter.MarshalIndent(v, "", " ")
	return string(json), err
}

// Unmarshal ...
func (j *JSONCompiler) Unmarshal(data []byte, v interface{}) error {
	return jsoniter.Unmarshal(data, v)
}

// UnmarshalFromString ...
func (j *JSONCompiler) UnmarshalFromString(data string, v interface{}) error {
	return jsoniter.UnmarshalFromString(data, v)
}
