package templates

// ResourceTemplate 资源pb文件模板
var ResourceTemplate = `syntax = "proto3";
package {{.Package}};

import "github.com/gogo/protobuf/gogoproto/gogo.proto";
{{range $index, $pkg := .ImportsPKGs -}}
import "{{$pkg}}/generated.proto";
{{end }}

option (gogoproto.protosizer_all) = true;
option (gogoproto.sizer_all) = false;
option go_package = "{{.GoPKGName}}";

{{range $name, $messageInfo := .MessageInfoMap}}
message {{$name}} {
    option (gogoproto.goproto_getters) = false;
    option (gogoproto.typedecl) = false;
    option (gogoproto.goproto_unrecognized) = false;
{{- range $index, $fieldSpec := $messageInfo.FieldList}}
{{- if $fieldSpec.IsRepeated }}
    repeated {{$fieldSpec.ProtobufTypeName}} {{$fieldSpec.ProtobufFieldName}} = {{$fieldSpec.FieldNumber}} {{with $fieldSpec.GoGoProtoFieldOptions}}[{{$fieldSpec.GoGoProtoFieldOptions}}]{{end}};
{{- else if $fieldSpec.IsMap }}
    map<{{$fieldSpec.KeyProtobufTypeName}}, {{$fieldSpec.ProtobufTypeName}}> {{$fieldSpec.ProtobufFieldName}} = {{$fieldSpec.FieldNumber}} {{with $fieldSpec.GoGoProtoFieldOptions}}[{{$fieldSpec.GoGoProtoFieldOptions}}]{{end}};
{{- else }}
    {{$fieldSpec.ProtobufTypeName}} {{$fieldSpec.ProtobufFieldName}} = {{$fieldSpec.FieldNumber}} {{with $fieldSpec.GoGoProtoFieldOptions}}[{{$fieldSpec.GoGoProtoFieldOptions}}]{{end}};
{{- end }}
{{- end}}
}
{{end}}
`
