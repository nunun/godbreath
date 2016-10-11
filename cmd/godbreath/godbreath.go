package main

import (
    "fmt"
    "go/ast"
    "go/parser"
    "go/token"
)

func main() {
    fset := token.NewFileSet()
    file, err := parser.ParseFile(fset, "user.go", nil, parser.ParseComments)
    if err != nil {
        panic(err)
    }

    for _, decl := range file.Decls {
        switch d := decl.(type) {
        case *ast.GenDecl:
            switch d.Tok {
            case token.IMPORT:
                fmt.Println("### import")
                for _, spec := range d.Specs {
                    s := spec.(*ast.ImportSpec)
                    fmt.Println(s.Path.Value)
                }

            case token.TYPE:
                fmt.Println("### type")
                for _, spec := range d.Specs {
                    s := spec.(*ast.TypeSpec)
                    fmt.Println(s.Name)
                    fmt.Println(s.Comment.Text()) // Struct Comment
                    switch t := s.Type.(type) {
                    case *ast.InterfaceType:
                        for _, m := range t.Methods.List {
                            fmt.Println(m)
                        }
                    case *ast.StructType:
                        for _, f := range t.Fields.List {
                            fmt.Println(f)
                            fmt.Println(f.Tag) // Field Tag
                            fmt.Println(f.Comment.Text()) // Field Comment
                        }
                    default:
                        fmt.Println(3, t)
                    }
                }
            case token.CONST:
            case token.VAR:
            default:
            }

        case *ast.FuncDecl:
            fmt.Println("### function")
            fmt.Println(d.Name)
            if d.Recv != nil {
                fmt.Println(d.Recv.List[0].Type)
            }
            if d.Type.Params != nil && d.Type.Params.NumFields() > 0 {
                fmt.Println("##### args")
                for _, p := range d.Type.Params.List {
                    fmt.Println(p.Type, p.Names)
                }
            }
            if d.Type.Results != nil && d.Type.Results.NumFields() > 0 {
                fmt.Println("##### returns")
                for _, r := range d.Type.Results.List {
                    fmt.Println(r.Type, r.Names)
                }
            }

        default:
        }

        fmt.Println()
    }
}
