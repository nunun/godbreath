package main

import (
    _ "fmt"
    _ "go/ast"
    _ "go/parser"
    _ "go/token"
    "os"
    "path"
    "io/ioutil"
    "text/template"
    "gopkg.in/yaml.v2"
)

type (
    TypeDecl struct {
        TypeName string
    }
)

func main() {
    Generate("test")
}

func Generate(generatePath string) {
    tpath := path.Join(generatePath, "template.yml")
    tmap  := LoadTemplate(tpath)

    d := &TypeDecl {"MyName"}
    err := tmap["Insert"].Execute(os.Stdout, d)
    if err != nil {
        panic(err)
    }
}

func GenerateFile(filePath string, tmap map[string]*template.Template) {
    // TODO
}

func LoadTemplate(templatePath string) map[string]*template.Template {
    buf, err := ioutil.ReadFile(templatePath)
    if err != nil {
        panic(err)
    }

    d := make(map[string]string)
    err = yaml.Unmarshal(buf, &d)
    if err != nil {
        panic(err)
    }

    tmap := make(map[string]*template.Template)
    for k, v := range d {
        t, err := template.New(k).Parse(v)
        if err != nil {
            panic(err)
        }
        tmap[k] = t
    }
    return tmap
}

//fset := token.NewFileSet()
//file, err := parser.ParseFile(fset, "test/user.go", nil, parser.ParseComments)
//if err != nil {
//    panic(err)
//}
//for _, decl := range file.Decls {
//    switch d := decl.(type) {
//    case *ast.GenDecl:
//        switch d.Tok {
//        case token.IMPORT:
//            fmt.Println("### import")
//            for _, spec := range d.Specs {
//                s := spec.(*ast.ImportSpec)
//                fmt.Println(s.Path.Value)
//            }
//        case token.TYPE:
//            fmt.Println("### type")
//            for _, spec := range d.Specs {
//                s := spec.(*ast.TypeSpec)
//                fmt.Println(s.Name)
//                fmt.Println(s.Comment.Text()) // Struct Comment
//                switch t := s.Type.(type) {
//                case *ast.InterfaceType:
//                    for _, m := range t.Methods.List {
//                        fmt.Println(m)
//                    }
//                case *ast.StructType:
//                    for _, f := range t.Fields.List {
//                        fmt.Println(f)
//                        fmt.Println(f.Tag) // Field Tag
//                        fmt.Println(f.Comment.Text()) // Field Comment
//                    }
//                default:
//                    fmt.Println(3, t)
//                }
//            }
//        case token.CONST:
//        case token.VAR:
//        default:
//        }
//    case *ast.FuncDecl:
//        fmt.Println("### function")
//        fmt.Println(d.Name)
//        if d.Recv != nil {
//            fmt.Println(d.Recv.List[0].Type)
//        }
//        if d.Type.Params != nil && d.Type.Params.NumFields() > 0 {
//            fmt.Println("##### args")
//            for _, p := range d.Type.Params.List {
//                fmt.Println(p.Type, p.Names)
//            }
//        }
//        if d.Type.Results != nil && d.Type.Results.NumFields() > 0 {
//            fmt.Println("##### returns")
//            for _, r := range d.Type.Results.List {
//                fmt.Println(r.Type, r.Names)
//            }
//        }
//    default:
//    }
//    fmt.Println()
//}

