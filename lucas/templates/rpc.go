package templates

// RPCTemplate rpc 的pb 文件模板
var RPCTemplate = `syntax = "proto3";
package {{.Package}};

import "github.com/gogo/protobuf/gogoproto/gogo.proto";
import "protoc-gen-swagger/options/annotations.proto";
{{range $index, $pkg := .ImportsPKGs -}}
import "{{$pkg}}/generated.proto";
{{end }}

option (gogoproto.protosizer_all) = true;
option (gogoproto.sizer_all) = false;
option go_package = "{{.GoPKGName}}";

option (grpc.gateway.protoc_gen_swagger.options.openapiv2_swagger) = {
    security_definitions: {
        security: {
            key: "XHagoToken";
            value: {
                type: TYPE_API_KEY;
                in: IN_HEADER;
                name: "X-Hago-Token";
            }
        }
    },
    security: {
        security_requirement: {
            key: "XHagoToken";
            value: {};
        }
    },
    consumes: [
        "application/json",
        "application/octet-stream"
    ]
};

{{range $index, $argDef := .AggregatedArgsDefinitions}}
message {{$argDef.FuncName}}{{$argDef.Suffix}} {
{{- range $index2, $arg := $argDef.Args }}
{{- if $arg.IsRepeated }}
    repeated {{$arg.ProtobufTypeName}} {{$arg.ProtobufFieldName}} = {{$arg.FieldNumber}}{{" "}};
{{- else if $arg.IsMap }}
    map<{{$arg.KeyProtobufTypeName}}, {{$arg.ProtobufTypeName}}> {{$arg.ProtobufFieldName}} = {{$arg.FieldNumber}}{{" "}};
{{- else }}
    {{$arg.ProtobufTypeName}} {{$arg.ProtobufFieldName}} = {{$arg.FieldNumber}} 
    {{- if or $arg.IsStruct $arg.IsStructInPtr}} [(gogoproto.nullable) = false, (gogoproto.embed) = true] {{- end}};
{{- end }}
{{- end }}
}
{{end}}
service {{.ServiceName}} {
{{- range $index, $func := .Functions}}
    rpc {{$func.FuncName}} ({{$func.RequestDefinition.ProtobufTypeName}}) returns ({{$func.ResponseDefinition.ProtobufTypeName}});
{{- end }}
}
`
