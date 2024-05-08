package lucas

import (
	"go/ast"
	"strconv"
)

// FieldName get the input(output) argument name of function from ast.
// index is number of the input(output) argument. funcName represent function name, and serviceName
// represent receiver's name. When input is true, returns input argument name, when input is false
// returns output argument name.
func FieldName(pkg *ast.Package, index int, funcName, serviceName string, input bool) string {
	if !input {
		return "result" + strconv.Itoa(index+1)
	}
	for _, file := range pkg.Files {
		for _, decl := range file.Decls {
			switch decl.(type) {
			case *ast.FuncDecl:
				funcDecl := decl.(*ast.FuncDecl)
				if funcDecl.Name.Name == funcName {
					if funcDecl.Recv != nil {
						switch funcDecl.Recv.List[0].Type.(type) {
						case *ast.StarExpr:
							if funcDecl.Recv.List[0].Type.(*ast.StarExpr).X.(*ast.Ident).Name == serviceName {
								if input {
									field := funcDecl.Type.Params.List[index]
									return field.Names[0].Name
								}
							}
						default:
						}
					}
					continue
				}
			default:
				continue
			}
		}
	}
	return ""
}
