package main

import (
    "os"
    "fmt"
    "log"
    "go/ast"
    "go/parser"
    "go/token"
    "path"
    "flag"
    "bytes"
    "bufio"
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
        TypeName          string
        TableName         string
        TableColumns      []string
        NonLabeledColumns []string
        LabeledColumns    map[string][]string
        ToFieldName       map[string]string
    }
)

var (
    help    bool
    silent  bool
    verbose bool
    tpath   string
)

func main() {
    flag.BoolVar(&help,    "h", false, "show help")
    flag.BoolVar(&silent,  "s", false, "silent mode")
    flag.BoolVar(&verbose, "v", false, "verbose mode")
    flag.StringVar(&tpath, "t", "",    "template path (default=<generate dir>/gen.yaml)")
    flag.Parse();
    if help {
        flag.PrintDefaults()
        return
    }
    gpath := "."
    if flag.NArg() >= 1 {
        gpath = flag.Arg(0)
    }
    if (tpath == "") {
        tpath = path.Join(gpath, "gen.yml")
    }
    Generate(gpath, tpath)
}

func Generate(generatePath string, templatePath string) {
    // load gen.yml
    tmap := LoadTemplate(tpath)

    // gather targets
    gpath := path.Join(generatePath, "*.go")
    files, err := filepath.Glob(gpath)
    if err != nil {
        panic(err)
    }

    // generate _gen.go from source file.
    cnt := 0
    for _, f := range files {
        if strings.HasSuffix(f, "_gen.go") {
            continue // 生成ファイルは処理しない。
        }
        ext  := filepath.Ext(f)
        of   := fmt.Sprintf("%s_gen.go", f[0:len(f)-len(ext)])
        done := GenerateSourceFile(f, of, tmap)
        if done {
            cnt++;
        }
    }
    if !silent {
        fmt.Println("[godbreath]", cnt, "file(s) generated.")
    }
}

func GenerateSourceFile(inputPath string, outputPath string, tmap map[string]*Template) bool {
    if verbose {
        fmt.Println(inputPath, "...");
    }

    // parse .go
    fset := token.NewFileSet()
    src, err := parser.ParseFile(fset, inputPath, nil, parser.ParseComments)
    if err != nil {
        log.Printf("Parse error %s: %s", inputPath, err.Error())
        return false
    }

    // gather struct type informations
    outputPackage := src.Name.Name
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
                        if d.Doc == nil {
                            if verbose {
                                fmt.Println(" !! type '", s.Name.String(), "' has no doc comment.")
                            }
                            continue
                        }
                        c := d.Doc.List[len(d.Doc.List) - 1].Text
                        defs := strings.Split(c, ":")
                        if len(defs) != 2 {
                            if verbose {
                                fmt.Println(" !! type '", s.Name.String(), "' has no table name. (", c, ")")
                            }
                            continue
                        }
                        defs[0]    = strings.Trim(defs[0], " /")
                        defs[1]    = strings.Trim(defs[1], " \n")
                        tableName := defs[0]
                        methods   := strings.Split(defs[1], ",")
                        for _, method := range methods {
                            m := strings.Trim(method, " \n")
                            if tmap[m] != nil {
                                genImports, genFunc := GenerateStruct(s, t, tableName, tmap[m])
                                outputImports = append(outputImports, genImports...)
                                outputFuncs   = append(outputFuncs, genFunc)
                            } else {
                                if verbose {
                                    fmt.Println(" ?? method '" + m + "' not found.")
                                }
                                continue;
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


    // empty?
    if len(outputFuncs) <= 0 {
        // remove generate file if exists.
        _, err = os.Stat(outputPath)
        if err == nil {
            os.Remove(outputPath)
            if verbose {
                fmt.Println("Removed", outputPath)
            }
        }
        return false
    }

    // unique imports array
    set := make(map[string]bool, 0)
    for _, item := range outputImports {
        set[item] = true
    }
    outputImports = []string{}
    for k, _ := range set {
        outputImports = append(outputImports, k)
    }

    // output _gen.go
    fp := NewFile(outputPath)
    defer fp.Close()
    writer := bufio.NewWriter(fp)
    fmt.Fprintf(writer, "package %s\n", outputPackage)
    for _, item := range outputImports {
        _, err := fmt.Fprintf(writer, "import \"%s\"\n", item)
        if err != nil {
            panic(err)
        }
    }
    for _, item := range outputFuncs {
        _, err := fmt.Fprintf(writer, "%s\n", item)
        if err != nil {
            panic(err)
        }
    }
    writer.Flush()
    if verbose {
        fmt.Println(" ->", outputPath)
    }
    return true
}

func GenerateStruct(s *ast.TypeSpec, t *ast.StructType, tableName string, temp *Template) (typeImports []string, typeFunc string) {
    // type name and fields
    TypeName          := s.Name.String()
    TableName         := tableName
    TableColumns      := []string{}
    NonLabeledColumns := []string{}
    LabeledColumns    := map[string][]string{}
    ToFieldName       := map[string]string{}
    for _, f := range t.Fields.List {
        tag := reflect.StructTag(f.Tag.Value[1:len(f.Tag.Value)-1])
        db  := tag.Get("db")
        if db != "" {
            TableColumns = append(TableColumns, db)
            dblabel := tag.Get("dblabel")
            if dblabel == "" {
                NonLabeledColumns = append(NonLabeledColumns, db)
            } else {
                dblabels := strings.Split(dblabel, ",")
                for _, label := range dblabels {
                    label := strings.Trim(label, " ")
                    if LabeledColumns[label] == nil {
                        LabeledColumns[label] = []string {}
                    }
                    LabeledColumns[label] = append(LabeledColumns[label], db)
                }
            }
            ToFieldName[db] = f.Names[0].Name
        }
    }

    // expand template
    vars := &TypeVars {TypeName, TableName, TableColumns, NonLabeledColumns, LabeledColumns, ToFieldName}
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

        // template functions
        funcMap := template.FuncMap {
            "q": func(s string) string {
                return "\"" + s + "\""
            },
            "joinq": func(s []string) string {
                if len(s) <= 0 { return "" }
                return "\"" + strings.Join(s, "\", \"") + "\"";
            },
        }

        // import
        templateImports := []string{}
        if m["import"] != nil {
            items := m["import"].([]interface{})
            for _, item := range items {
                templateImports = append(templateImports, item.(string))
            }
        }

        // func
        templateFunc, err := template.New(k).Funcs(funcMap).Parse(m["func"].(string))
        if err != nil {
            panic(err)
        }


        // push all information into map
        tmap[k] = &Template{templateImports, templateFunc}
    }
    return tmap
}


func NewFile(fn string) *os.File {
    fp, err := os.OpenFile(fn, os.O_WRONLY | os.O_CREATE | os.O_TRUNC, 0644)
    if err != nil {
        panic(err)
    }
    return fp
}

