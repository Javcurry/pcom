package cmd

var mainTemplate = `package main

import ({{" "}}
{{- range $index, $conf := .PKGConfigs}}
	{{$conf.ServicePKGAlias}} "{{$conf.ServicePKG}}"
{{- end}}
	"hago-plat/pcom/lucas"
    "os"
)

func main() {
    var err error
{{- range $index, $conf := .PKGConfigs}}
	err = lucas.ServiceGenerate("{{$conf.GenPathRel}}", &{{$conf.ServicePKGAlias}}.{{$conf.ServiceName}}{})
    if err != nil {
        os.Exit(2)
    }
{{- end}}
}`
