package main

import (
    _ "os"
    "fmt"
    "log"
    "go/ast"
    "go/parser"
    "go/token"
    "path"
    "flag"
    "bytes"
    "strings"
    "reflect"
    "path/filepath"
    "io/ioutil"
    "text/template"
    "gopkg.in/yaml.v2"
)

type (
    Template struct {
        TemplateImports []string
        TemplateFunc    *template.Template
    }

    TypeVars struct {
        TypeName       string
        TableName      string
        Columns        []string
        UpdateColumns  []string
        PrivateColumns []string
    }
)

func main() {
    var tpath string
    flag.StringVar(&tpath, "t", "gen.yml", "template path")
    flag.Parse();
    if flag.NArg() < 1 {
        panic("path did not specified.")
    }
    Generate(flag.Arg(0), tpath)
}

func Generate(generatePath string, templatePath string) {
    // load gen.yml
    tpath := path.Join(generatePath, templatePath)
    tmap  := LoadTemplate(tpath)

    // gather targets
    gpath := path.Join(generatePath, "*.go")
    files, err := filepath.Glob(gpath)
    if err != nil {
        panic(err)
    }

    // generate _gen.go from source file.
    for _, f := range files {
        var ext = filepath.Ext(f)
        var of  = fmt.Sprintf("%s_gen.go", f[0:len(f)-len(ext)])
        GenerateSourceFile(f, of, tmap)
    }
}

func GenerateSourceFile(inputPath string, outputPath string, tmap map[string]*Template) bool {
    // parse .go
    fset := token.NewFileSet()
    src, err := parser.ParseFile(fset, inputPath, nil, parser.ParseComments)
    if err != nil {
        log.Printf("Parse error %s: %s", inputPath, err.Error())
        return false
    }

    // gather struct type informations
    outputImports := []string{}
    outputFuncs   := []string{}
    for _, decl := range src.Decls {
        switch d := decl.(type) {
        case *ast.GenDecl:
            switch d.Tok {
            case token.TYPE:
                for _, spec := range d.Specs {
                    s := spec.(*ast.TypeSpec)
                    switch t := s.Type.(type) {
                    case *ast.StructType:
                        c := s.Comment.Text()
                        defs := strings.Split(c, ":")
                        if len(defs) != 2 {
                            continue
                        }
                        tableName := strings.Trim(defs[0], " \n")
                        methods   := strings.Split(defs[1], ",")
                        for _, method := range methods {
                            m := strings.Trim(method, " \n")
                            if tmap[m] != nil {
                                genImports, genFunc := GenerateStruct(s, t, tableName, tmap[m])
                                outputImports = append(outputImports, genImports...)
                                outputFuncs   = append(outputFuncs, genFunc)
                            }
                        }
                    case *ast.InterfaceType:
                    default:
                    }
                }
            case token.IMPORT:
            case token.CONST:
            case token.VAR:
            default:
            }
        case *ast.FuncDecl:
        default:
        }
    }

    // output _gen.go
    // TODO
    fmt.Println(outputImports)
    fmt.Println(outputFuncs)
    return true
}

func GenerateStruct(s *ast.TypeSpec, t *ast.StructType, tableName string, temp *Template) (typeImports []string, typeFunc string) {
    // type name and fields
    TypeName       := s.Name.String()
    TableName      := tableName
    Columns        := []string{}
    UpdateColumns  := []string{}
    PrivateColumns := []string{}
    for _, f := range t.Fields.List {
        tag := reflect.StructTag(f.Tag.Value[1:len(f.Tag.Value)-1])
        db  := tag.Get("db")
        if db != "" {
            Columns = append(Columns, db)
            auto    := tag.Get("auto")
            private := tag.Get("private")
            if auto != "true" && private != "true" {
                UpdateColumns = append(UpdateColumns, db)
            }
            if private == "true" {
                PrivateColumns = append(PrivateColumns, db)
            }
        }
    }

    // expand template
    vars := &TypeVars {TypeName, TableName, Columns, UpdateColumns, PrivateColumns}
    buf  := &bytes.Buffer{}
    err  := temp.TemplateFunc.Execute(buf, vars)
    if err != nil {
        panic(err)
    }

    // function results
    typeImports = temp.TemplateImports
    typeFunc    = buf.String()
    return
}

func LoadTemplate(templatePath string) map[string]*Template {
    // read gen.yml
    buf, err := ioutil.ReadFile(templatePath)
    if err != nil {
        panic(err)
    }

    // parse gen.yml
    d := make(map[string]interface{})
    err = yaml.Unmarshal(buf, &d)
    if err != nil {
        panic(err)
    }

    // seek elements
    tmap := make(map[string]*Template)
    for k, v := range d {
        m := v.(map[interface{}]interface{})

        // import
        templateImports := []string{}
        if m["import"] != nil {
            items := m["import"].([]interface{})
            for _, item := range items {
                templateImports = append(templateImports, item.(string))
            }
        }

        // func
        templateFunc, err := template.New(k).Parse(m["func"].(string))
        if err != nil {
            panic(err)
        }

        // push all information into map
        tmap[k] = &Template{templateImports, templateFunc}
    }
    return tmap
}

