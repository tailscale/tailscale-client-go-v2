package repoanalysis

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"path/filepath"
	"reflect"
	"slices"
	"strconv"
	"strings"

	"tailscale.com/client/tailscale/v2/tools/internal/openapi"
)

type Analyzer struct {
	root         string
	fset         *token.FileSet
	packageFiles map[string]*ast.File
	structs      map[string]structInfo
	operations   []Operation
}

type Operation struct {
	Method         string
	Path           string
	NormalizedPath string
	ClientMethod   string
	File           string
	Line           int
}

type Field struct {
	Path string
	File string
	Line int
}

type Model struct {
	Name   string
	File   string
	Line   int
	Fields []Field
}

type structInfo struct {
	file string
	node *ast.StructType
}

func Analyze(root string) (*Analyzer, error) {
	fset := token.NewFileSet()
	packages, err := parser.ParseDir(fset, root, func(info fs.FileInfo) bool {
		name := info.Name()
		return strings.HasSuffix(name, ".go") && !strings.HasSuffix(name, "_test.go")
	}, parser.SkipObjectResolution)
	if err != nil {
		return nil, fmt.Errorf("parse repository: %w", err)
	}

	if len(packages) == 0 {
		return nil, fmt.Errorf("no Go package found in %s", root)
	}

	var pkg *ast.Package
	if named, ok := packages["tailscale"]; ok {
		pkg = named
	} else {
		for _, candidate := range packages {
			pkg = candidate
			break
		}
	}

	analyzer := &Analyzer{
		root:         root,
		fset:         fset,
		packageFiles: pkg.Files,
		structs:      make(map[string]structInfo),
	}

	for fileName, file := range pkg.Files {
		for _, decl := range file.Decls {
			gen, ok := decl.(*ast.GenDecl)
			if !ok || gen.Tok != token.TYPE {
				continue
			}

			for _, spec := range gen.Specs {
				typeSpec, ok := spec.(*ast.TypeSpec)
				if !ok {
					continue
				}

				structType, ok := typeSpec.Type.(*ast.StructType)
				if !ok {
					continue
				}

				analyzer.structs[typeSpec.Name.Name] = structInfo{
					file: fileName,
					node: structType,
				}
			}
		}
	}

	for _, file := range pkg.Files {
		for _, decl := range file.Decls {
			function, ok := decl.(*ast.FuncDecl)
			if !ok || function.Recv == nil || function.Body == nil {
				continue
			}

			analyzer.operations = append(analyzer.operations, analyzer.extractOperations(function)...)
		}
	}

	slices.SortFunc(analyzer.operations, func(a, b Operation) int {
		if a.Method != b.Method {
			return strings.Compare(a.Method, b.Method)
		}
		if a.Path != b.Path {
			return strings.Compare(a.Path, b.Path)
		}
		if a.ClientMethod != b.ClientMethod {
			return strings.Compare(a.ClientMethod, b.ClientMethod)
		}
		if a.File != b.File {
			return strings.Compare(a.File, b.File)
		}
		return a.Line - b.Line
	})

	return analyzer, nil
}

func (a *Analyzer) Operations() []Operation {
	return append([]Operation(nil), a.operations...)
}

func (a *Analyzer) StructLeafJSONFields(typeName string) ([]Field, error) {
	fields := make(map[string]Field)
	if err := a.collectStructFields(typeName, "", fields, map[string]bool{}); err != nil {
		return nil, err
	}

	out := make([]Field, 0, len(fields))
	for _, field := range fields {
		out = append(out, field)
	}

	slices.SortFunc(out, func(lhs, rhs Field) int {
		if lhs.Path != rhs.Path {
			return strings.Compare(lhs.Path, rhs.Path)
		}
		if lhs.File != rhs.File {
			return strings.Compare(lhs.File, rhs.File)
		}
		return lhs.Line - rhs.Line
	})

	return out, nil
}

func (a *Analyzer) Models() ([]Model, error) {
	models := make([]Model, 0, len(a.structs))
	for name, info := range a.structs {
		if !ast.IsExported(name) || !hasTaggedJSONFields(info.node) {
			continue
		}

		fields, err := a.StructLeafJSONFields(name)
		if err != nil {
			return nil, err
		}

		position := a.fset.Position(info.node.Pos())
		models = append(models, Model{
			Name:   name,
			File:   position.Filename,
			Line:   position.Line,
			Fields: fields,
		})
	}

	slices.SortFunc(models, func(lhs, rhs Model) int {
		return strings.Compare(lhs.Name, rhs.Name)
	})

	return models, nil
}

func (a *Analyzer) extractOperations(function *ast.FuncDecl) []Operation {
	receiverName, receiverType, ok := receiver(function)
	if !ok {
		return nil
	}

	clientMethod := receiverType + "." + function.Name.Name
	position := a.fset.Position(function.Pos())
	operations := make([]Operation, 0)

	ast.Inspect(function.Body, func(node ast.Node) bool {
		call, ok := node.(*ast.CallExpr)
		if !ok {
			return true
		}

		selector, ok := call.Fun.(*ast.SelectorExpr)
		if !ok || selector.Sel.Name != "buildRequest" {
			return true
		}

		ident, ok := selector.X.(*ast.Ident)
		if !ok || ident.Name != receiverName {
			return true
		}

		if len(call.Args) < 3 {
			return true
		}

		method := httpMethod(call.Args[1])
		path, ok := requestPath(call.Args[2])
		if method == "" || !ok {
			return true
		}

		callPosition := a.fset.Position(call.Pos())
		operations = append(operations, Operation{
			Method:         method,
			Path:           path,
			NormalizedPath: openapi.NormalizePath(path),
			ClientMethod:   clientMethod,
			File:           callPosition.Filename,
			Line:           callPosition.Line,
		})

		return true
	})

	if len(operations) == 0 && position.IsValid() {
		return nil
	}

	return operations
}

func receiver(function *ast.FuncDecl) (string, string, bool) {
	if function.Recv == nil || len(function.Recv.List) != 1 {
		return "", "", false
	}

	field := function.Recv.List[0]
	if len(field.Names) != 1 {
		return "", "", false
	}

	receiverName := field.Names[0].Name
	switch typed := field.Type.(type) {
	case *ast.Ident:
		return receiverName, typed.Name, true
	case *ast.StarExpr:
		ident, ok := typed.X.(*ast.Ident)
		if !ok {
			return "", "", false
		}
		return receiverName, ident.Name, true
	default:
		return "", "", false
	}
}

func httpMethod(expr ast.Expr) string {
	switch typed := expr.(type) {
	case *ast.BasicLit:
		if typed.Kind != token.STRING {
			return ""
		}

		value, err := strconv.Unquote(typed.Value)
		if err != nil {
			return ""
		}

		return strings.ToUpper(value)
	case *ast.SelectorExpr:
		ident, ok := typed.X.(*ast.Ident)
		if !ok || ident.Name != "http" {
			return ""
		}

		switch typed.Sel.Name {
		case "MethodGet":
			return "GET"
		case "MethodPost":
			return "POST"
		case "MethodPut":
			return "PUT"
		case "MethodPatch":
			return "PATCH"
		case "MethodDelete":
			return "DELETE"
		default:
			return ""
		}
	default:
		return ""
	}
}

func requestPath(expr ast.Expr) (string, bool) {
	call, ok := expr.(*ast.CallExpr)
	if !ok {
		return "", false
	}

	selector, ok := call.Fun.(*ast.SelectorExpr)
	if !ok {
		return "", false
	}

	var segments []string
	switch selector.Sel.Name {
	case "buildURL":
		segments = []string{"api", "v2"}
	case "buildTailnetURL":
		segments = []string{"api", "v2", "tailnet", "{tailnet}"}
	default:
		return "", false
	}

	for _, arg := range call.Args {
		segments = append(segments, pathSegment(arg))
	}

	return "/" + strings.Join(segments, "/"), true
}

func pathSegment(expr ast.Expr) string {
	literal, ok := expr.(*ast.BasicLit)
	if ok && literal.Kind == token.STRING {
		value, err := strconv.Unquote(literal.Value)
		if err == nil && value != "" {
			return value
		}
	}

	return "{}"
}

func (a *Analyzer) collectStructFields(typeName, prefix string, out map[string]Field, visiting map[string]bool) error {
	if visiting[typeName] {
		return nil
	}

	info, ok := a.structs[typeName]
	if !ok {
		return fmt.Errorf("struct %s not found", typeName)
	}

	if !hasTaggedJSONFields(info.node) {
		if prefix != "" {
			position := a.fset.Position(info.node.Pos())
			out[prefix] = Field{Path: prefix, File: position.Filename, Line: position.Line}
		}
		return nil
	}

	visiting[typeName] = true
	defer delete(visiting, typeName)

	for _, field := range info.node.Fields.List {
		name, ok := jsonFieldName(field)
		if !ok {
			continue
		}

		path := joinPath(prefix, name)
		if a.shouldRecurse(field.Type) {
			if err := a.collectExprFields(field.Type, path, out, visiting); err != nil {
				return err
			}
			continue
		}

		position := a.fset.Position(field.Pos())
		out[path] = Field{
			Path: path,
			File: position.Filename,
			Line: position.Line,
		}
	}

	return nil
}

func (a *Analyzer) collectExprFields(expr ast.Expr, prefix string, out map[string]Field, visiting map[string]bool) error {
	switch typed := expr.(type) {
	case *ast.StarExpr:
		return a.collectExprFields(typed.X, prefix, out, visiting)
	case *ast.ArrayType:
		if a.shouldRecurse(typed.Elt) {
			return a.collectExprFields(typed.Elt, prefix, out, visiting)
		}
	case *ast.MapType:
		if a.shouldRecurse(typed.Value) {
			return a.collectExprFields(typed.Value, prefix, out, visiting)
		}
	case *ast.StructType:
		return a.collectAnonymousStructFields(typed, prefix, out, visiting)
	case *ast.Ident:
		if _, ok := a.structs[typed.Name]; ok {
			return a.collectStructFields(typed.Name, prefix, out, visiting)
		}
	}

	position := a.fset.Position(expr.Pos())
	out[prefix] = Field{
		Path: prefix,
		File: position.Filename,
		Line: position.Line,
	}

	return nil
}

func (a *Analyzer) shouldRecurse(expr ast.Expr) bool {
	switch typed := expr.(type) {
	case *ast.StarExpr:
		return a.shouldRecurse(typed.X)
	case *ast.ArrayType:
		return a.shouldRecurse(typed.Elt)
	case *ast.MapType:
		return a.shouldRecurse(typed.Value)
	case *ast.StructType:
		return hasTaggedJSONFields(typed)
	case *ast.Ident:
		info, ok := a.structs[typed.Name]
		return ok && hasTaggedJSONFields(info.node)
	default:
		return false
	}
}

func (a *Analyzer) collectAnonymousStructFields(structType *ast.StructType, prefix string, out map[string]Field, visiting map[string]bool) error {
	if !hasTaggedJSONFields(structType) {
		position := a.fset.Position(structType.Pos())
		out[prefix] = Field{
			Path: prefix,
			File: position.Filename,
			Line: position.Line,
		}
		return nil
	}

	for _, field := range structType.Fields.List {
		name, ok := jsonFieldName(field)
		if !ok {
			continue
		}

		path := joinPath(prefix, name)
		if a.shouldRecurse(field.Type) {
			if err := a.collectExprFields(field.Type, path, out, visiting); err != nil {
				return err
			}
			continue
		}

		position := a.fset.Position(field.Pos())
		out[path] = Field{
			Path: path,
			File: position.Filename,
			Line: position.Line,
		}
	}

	return nil
}

func hasTaggedJSONFields(structType *ast.StructType) bool {
	for _, field := range structType.Fields.List {
		if _, ok := jsonFieldName(field); ok {
			return true
		}
	}

	return false
}

func jsonFieldName(field *ast.Field) (string, bool) {
	if field.Tag == nil {
		return "", false
	}

	rawTag, err := strconv.Unquote(field.Tag.Value)
	if err != nil {
		return "", false
	}

	tag := reflect.StructTag(rawTag).Get("json")
	if tag == "" {
		return "", false
	}

	name := strings.Split(tag, ",")[0]
	if name == "" || name == "-" {
		return "", false
	}

	return name, true
}

func joinPath(prefix, name string) string {
	if prefix == "" {
		return name
	}

	return prefix + "." + name
}

func RelativePath(root, file string) string {
	relative, err := filepath.Rel(root, file)
	if err != nil {
		return file
	}

	return relative
}
