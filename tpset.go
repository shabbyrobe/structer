package structer

import (
	"fmt"
	"go/ast"
	"go/build"
	"go/importer"
	"go/types"
	"os"
	"path/filepath"
	"strings"
)

var (
	BuildContext = build.Default
)

// Interface checks
var (
	_ types.ImporterFrom = &TypePackageSet{}
)

type Config struct {
	IncludeTests bool
}

type option func(*TypePackageSet)

func CaptureErrors(h func(error)) option {
	return func(t *TypePackageSet) {
		t.TypesConfig.Error = h
	}
}

// TypePackageSet collects information about the types in all of the
// imported packages.
//
type TypePackageSet struct {
	Config Config

	TypePackages map[string]*types.Package
	ASTPackages  *ASTPackageSet
	Infos        map[string]types.Info

	// List of files extracted from go/build.Package, indexed by package name.
	// This is the list of files that would be built for the package on this
	// machine if "go build" or "go install" was run.
	BuiltFiles map[string][]string

	DefaultImporter types.Importer

	TypesConfig types.Config
	Objects     map[TypeName]types.Object
	Kinds       map[string]PackageKind

	// According to the types.Error documentation: "A "soft" error is an error
	// that still permits a valid interpretation of a package (such as 'unused
	// variable'); "hard" errors may lead to unpredictable
	// behavior if ignored."
	//
	// This defaults to "true" as types considers certain errors that are likelier
	// to be present before a code generation step as "hard" errors. Set this to
	// false to fail on hard errors too.
	AllowHardTypesError bool

	Log Log
}

func NewTypePackageSet(opts ...option) *TypePackageSet {
	tps := &TypePackageSet{
		ASTPackages:     NewASTPackageSet(),
		TypePackages:    make(map[string]*types.Package),
		Infos:           make(map[string]types.Info),
		DefaultImporter: importer.Default(),
		BuiltFiles:      make(map[string][]string),
		Objects:         make(map[TypeName]types.Object),
		Kinds:           make(map[string]PackageKind),
	}
	tps.AllowHardTypesError = true
	tps.TypesConfig.IgnoreFuncBodies = false
	tps.TypesConfig.DisableUnusedImportCheck = true
	tps.TypesConfig.Error = func(err error) {
		wlog(tps.Log, LogTypeSet, LogTypesConfigError, err.Error())
	}
	tps.TypesConfig.Importer = tps
	for _, o := range opts {
		o(tps)
	}
	return tps
}

func (t *TypePackageSet) FindImportPath(pkg string, typeName string) (path string, err error) {
	var tpkg *types.Package
	tpkg, err = t.Import(pkg)
	if err != nil {
		return
	}
	if tpkg == nil {
		err = fmt.Errorf("import path not found for pkg %s, type %s", pkg, typeName)
		return
	}

	npkg := strings.SplitN(typeName, ".", 2)
	if len(npkg) == 1 {
		return pkg + "." + typeName, nil
	}

	ipath := ""
	for _, imprt := range tpkg.Imports() {
		if imprt.Name() == npkg[0] {
			ipath = imprt.Path()
			break
		}
	}
	if ipath != "" {
		if len(npkg) > 1 {
			return ipath + "." + npkg[1], nil
		} else {
			return pkg + "." + typeName, nil
		}
	}

	err = fmt.Errorf("import path not found for pkg %s, type %s", pkg, typeName)
	return
}

func (t *TypePackageSet) LocalPackageFromType(name TypeName) (string, error) {
	return t.LocalPackage(name.PackagePath)
}

func (t *TypePackageSet) LocalPackage(packageName string) (string, error) {
	p := t.TypePackages[packageName]
	if p == nil {
		return "", fmt.Errorf("package not found %s", packageName)
	}
	return p.Name(), nil
}

func (t *TypePackageSet) LocalImportName(name TypeName, relPkg string) (string, error) {
	if name.IsBuiltin() {
		return name.Name, nil
	}

	tloc, err := t.LocalPackage(name.PackagePath)
	if err != nil {
		return "", err
	}
	if tloc == "main" {
		if name.PackagePath != relPkg {
			return "", fmt.Errorf("Attempted to import main relative to package %q", relPkg)
		}
	}

	rloc, err := t.LocalPackage(relPkg)
	if err != nil {
		return "", err
	}
	if tloc == rloc {
		return name.Name, nil
	} else {
		return fmt.Sprintf("%s.%s", tloc, name.Name), nil
	}
}

func (t *TypePackageSet) ExtractSource(name TypeName) ([]byte, error) {
	def := t.Objects[name]
	if def == nil {
		return nil, fmt.Errorf("could not find def for %s", name)
	}

	pkg := def.Pkg().Path()
	node := t.ASTPackages.FindNodeByPackagePathPos(pkg, def.Pos())
	if node == nil {
		return nil, fmt.Errorf("no ast node for %s.%s", pkg, name)
	}

	pos := t.ASTPackages.FileSet.Position(node.Pos()).Offset
	end := t.ASTPackages.FileSet.Position(node.End()).Offset

	posn := t.ASTPackages.FileSet.PositionFor(node.Pos(), false)
	astPkg := t.ASTPackages.Packages[pkg]
	if astPkg == nil {
		return nil, fmt.Errorf("no ast pkg for %s", pkg)
	}

	contents, ok := astPkg.Contents[filepath.Base(posn.Filename)]
	if !ok {
		return nil, fmt.Errorf("no contents for %s.%s", pkg, name)
	}

	return contents[pos:end], nil
}

// ExtractConsts extracts all constants that satisfy the supplied type from
// the same package.
//
// Go provides no language-level idea of an Enum constraint, so structer provides
// the IsEnum interface. If the type satisfies IsEnum, it is assumed that the
// range of constants expresses the limit of all possible values for that type.
// Otherwise, it is just a bag of constants.
//
func (t *TypePackageSet) ExtractConsts(name TypeName, includeUnexported bool) (*Consts, error) {
	def := t.Objects[name]
	if def == nil {
		return nil, fmt.Errorf("could not find def for %s", name)
	}

	named, ok := def.Type().(*types.Named)
	if !ok {
		return nil, fmt.Errorf("type %s must be *types.Named, found %T", name, def.Type())
	}

	consts := &Consts{
		Type:       name,
		Underlying: ExtractTypeName(def.Type().Underlying()),
		IsEnum:     isEnum(def, named),
	}

	for n, o := range t.Infos[name.PackagePath].Defs {
		if o == nil {
			continue
		}
		if !name.IsType(o.Type()) {
			continue
		}
		if !includeUnexported && !n.IsExported() {
			continue
		}
		if cns, ok := o.(*types.Const); ok {
			consts.Values = append(consts.Values, &ConstValue{
				Name:  NewTypeName(name.PackagePath, cns.Name()),
				Value: cns.Val(),
			})
		}
	}
	return consts, nil
}

func (t *TypePackageSet) FilePackage(file string) (PackageKind, string, error) {
	goSrcPath := filepath.Join(BuildContext.GOPATH, "src")

	dir := filepath.Dir(file)

	// FIXME: dir MUST be a child of GOPATH. Unfortunately, they didn't bother
	// to provide a reasonable alternative, pointer, suggestion, anything, when
	// they deprecated filepath.HasPrefix.

	dirParts := splitPath(dir)

	// Is it a vendor package?
	j := len(dirParts) - 1
	for ; j >= 0; j-- {
		if dirParts[j] == "vendor" {
			return VendorPackage, filepath.Join(dirParts[j+1:]...), nil
		}
	}

	// Is it a SystemPackage?
	goRootParts := splitPath(filepath.Join(BuildContext.GOROOT, "src"))
	isSystem := len(goRootParts) > 0
	if len(goRootParts) <= len(dirParts) {
		// This check is grotesque. filepath.HasPrefix is deprecated, there's no
		// other way to check if a path is a child of another I can see.
		for i := 0; i < len(goRootParts); i++ {
			s1, err := os.Stat(filepath.Join(goRootParts[:i+1]...))
			if err != nil {
				return "", "", err
			}
			s2, err := os.Stat(filepath.Join(dirParts[:i+1]...))
			if err != nil {
				return "", "", err
			}
			if !os.SameFile(s1, s2) {
				isSystem = false
				break
			}
		}
	}
	if isSystem {
		return SystemPackage, filepath.Join(dirParts[len(goRootParts):]...), nil
	}

	// Is it a UserPackage?
	goSrcParts := splitPath(goSrcPath)
	isUser := len(goSrcPath) > 0
	if len(goSrcParts) <= len(dirParts) {
		for i := 0; i < len(goSrcParts); i++ {
			s1, err := os.Stat(filepath.Join(goSrcParts[:i+1]...))
			if err != nil {
				return "", "", err
			}
			s2, err := os.Stat(filepath.Join(goSrcParts[:i+1]...))
			if err != nil {
				return "", "", err
			}
			if !os.SameFile(s1, s2) {
				isUser = false
				break
			}
		}
	}

	if isUser {
		return UserPackage, filepath.Join(dirParts[len(goSrcParts):]...), nil
	}

	return NoPackage, "", nil
}

func (t *TypePackageSet) ResolvePath(path, srcDir string) (PackageKind, string, error) {
	// Is it a VendorPackage?
	goSrcPath := filepath.Join(BuildContext.GOPATH, "src")

	// FIXME: srcDir MUST be a child of GOPATH. Unfortunately, they didn't bother
	// to provide a reasonable alternative, pointer, suggestion, anything, when
	// they deprecated filepath.HasPrefix.

	cur := srcDir
	last := cur
	for cur != goSrcPath {
		vendorDir := filepath.Join(cur, "vendor")
		info, _ := os.Stat(vendorDir)
		if info != nil && info.IsDir() {
			if dir := resolvePackageDir(filepath.Join(vendorDir, path)); dir != "" {
				return VendorPackage, dir, nil
			}
		}
		cur = filepath.Dir(cur)
		if cur == last {
			break
		}
		last = cur
	}

	// Is it a SystemPackage?
	if dir := resolvePackageDir(filepath.Join(BuildContext.GOROOT, "src", path)); dir != "" {
		return SystemPackage, dir, nil
	}

	// Is it a UserPackage?
	if dir := resolvePackageDir(filepath.Join(goSrcPath, path)); dir != "" {
		return UserPackage, dir, nil
	}

	return NoPackage, "", nil
}

func (t *TypePackageSet) ImportNamed(named *types.Named) (*types.Package, error) {
	tn, err := ParseTypeName(named.String())
	if err != nil {
		return nil, err
	}
	return t.Import(tn.PackagePath)
}

// Import a package using the default GOPATH/src folder. See go/types.Importer.
func (t *TypePackageSet) Import(importPath string) (*types.Package, error) {
	srcPath := filepath.Join(BuildContext.GOPATH, "src", importPath)
	return t.ImportFrom(importPath, srcPath, 0)
}

// ImportFrom returns the imported package for the given import path when
// imported by a package file located in dir.
// See go/types.ImporterFrom.
func (t *TypePackageSet) ImportFrom(importPath, srcDir string, mode types.ImportMode) (pkg *types.Package, err error) {
	var (
		resolved     string
		info         types.Info
		ok           bool
		kind         PackageKind
		buildPackage *build.Package
	)

	if pkg, ok = t.TypePackages[importPath]; ok {
		goto done
	}
	if kind, resolved, err = t.ResolvePath(importPath, srcDir); kind == NoPackage || err != nil {
		goto done
	}

	t.Kinds[importPath] = kind

	if kind == SystemPackage {
		// System packages are especially janky but we do still want to be able to
		// resolve their types. The DefaultImporter seems to work well enough to
		// make it possible to infer, for e.g., that a time.Duration is just an int64.
		//
		// We may be able to just return the result of build.Import for this.
		if pkg, err = t.DefaultImporter.Import(importPath); err != nil {
			goto done
		}

	} else {
		buildPackage, err = build.Import(importPath, srcDir, build.ImportComment)
		if err != nil {
			goto done
		}

		t.BuiltFiles[importPath] = buildPackage.GoFiles

		info = types.Info{
			Types:      make(map[ast.Expr]types.TypeAndValue),
			Defs:       make(map[*ast.Ident]types.Object),
			Selections: make(map[*ast.SelectorExpr]*types.Selection),
		}

		if err = t.ASTPackages.Add(resolved, importPath); err != nil {
			return nil, err
		}
		t.Infos[importPath] = info

		var asts []*ast.File
		fileSets := [][]string{buildPackage.GoFiles}
		if t.Config.IncludeTests {
			fileSets = append(fileSets, buildPackage.TestGoFiles)
		}

		for _, fs := range fileSets {
			for _, file := range fs {
				full := filepath.Join(resolved, file)

				ap := t.ASTPackages.Packages[importPath]
				if ap == nil {
					return nil, fmt.Errorf("missing ASTPackage for %s", importPath)
				}
				af := ap.AST.Files[full]
				if af == nil {
					return nil, fmt.Errorf("missing AST File %s for import path %s", full, importPath)
				}
				asts = append(asts, af)
			}
		}

		pkg, err = t.TypesConfig.Check(importPath, t.ASTPackages.FileSet, asts, &info)
		raise := true
		if err != nil {
			if terr, ok := err.(types.Error); ok {
				if terr.Soft || t.AllowHardTypesError {
					raise = false
				}
			}
			if !raise {
				wlog(t.Log, LogTypeSet, LogTypeCheck, err.Error())
				err = nil
			} else {
				return nil, err
			}
		}

		t.indexTypes(importPath, info.Defs)
	}

done:
	t.TypePackages[importPath] = pkg
	return
}

// FindImplementers lists all types in all imported user packages which
// implement the interface supplied in the argument.
//
// This does not yield types from system packages or vendor packages yet.
//
func (t *TypePackageSet) FindImplementers(ifaceName TypeName) (TypeMap, error) {
	// Import the package referred to in the argument if we have not seen it before.
	// This should validate the incoming name as a benefit.
	if _, ok := t.TypePackages[ifaceName.PackagePath]; !ok {
		if _, err := t.Import(ifaceName.PackagePath); err != nil {
			return nil, err
		}
	}

	iface, ok := t.Objects[ifaceName]
	if !ok {
		return nil, fmt.Errorf("could not find object for %s", ifaceName)
	}
	ifaceTyp := iface.Type()
	if !types.IsInterface(ifaceTyp) {
		return nil, fmt.Errorf("type %s is not an interface", ifaceName)
	}

	var implements = make(TypeMap)

	for _, fobj := range t.Objects {
		fTyp := fobj.Type()

		if ifaceTyp != fTyp && !types.IsInterface(fTyp) {
			var impl types.Type

			if types.AssignableTo(fTyp, ifaceTyp) {
				impl = fTyp
			} else {
				ptr := types.NewPointer(fTyp)
				if types.AssignableTo(ptr, ifaceTyp) {
					impl = ptr
				}
			}

			if impl != nil {
				// var s types.Type
				// if p, ok := fTyp.(*types.Pointer); ok {
				//     s = p.Elem()
				// } else {
				//     s = fTyp
				// }
				// implements[ExtractTypeName(s)] = fTyp
				implements[ExtractTypeName(fTyp)] = impl
			}
		}
	}

	return implements, nil
}

func (t *TypePackageSet) FindObject(name TypeName) types.Object {
	obj := t.Objects[name]
	if obj == nil {
		return nil
	}
	return obj
}

func (t *TypePackageSet) FindImportObject(name TypeName) (types.Object, error) {
	_, err := t.Import(name.PackagePath)
	if err != nil {
		return nil, err
	}
	return t.FindObject(name), nil
}

func (t *TypePackageSet) FindObjectByName(name string) (types.Object, error) {
	tn, err := ParseTypeName(name)
	if err != nil {
		return nil, err
	}
	return t.Objects[tn], nil
}

func (t *TypePackageSet) FindImportObjectByName(name string) (types.Object, error) {
	tn, err := ParseTypeName(name)
	if err != nil {
		return nil, err
	}
	return t.FindImportObject(tn)
}

func (t *TypePackageSet) MustFindObject(name TypeName) types.Object {
	typ := t.FindObject(name)
	if typ == nil {
		panic(fmt.Errorf("structer: could not find object %s", name))
	}
	return typ
}

func (t *TypePackageSet) MustFindObjectByName(name string) types.Object {
	tn, err := ParseTypeName(name)
	if err != nil {
		panic(err)
	}
	return t.MustFindObject(tn)
}

func (t *TypePackageSet) MustFindImportObject(name TypeName) types.Object {
	typ, err := t.FindImportObject(name)
	if err != nil {
		panic(fmt.Errorf("structer: could not find object %s, err: %v", name, err))
	}
	if typ == nil {
		panic(fmt.Errorf("structer: could not find object %s", name))
	}
	return typ
}

func (t *TypePackageSet) MustFindImportObjectByName(name string) types.Object {
	tn, err := ParseTypeName(name)
	if err != nil {
		panic(err)
	}
	return t.MustFindImportObject(tn)
}

func (t *TypePackageSet) indexTypes(path string, defs map[*ast.Ident]types.Object) {
	for _, def := range defs {
		if def == nil {
			// def is nil for package declarations
			continue
		}

		if def.Parent() != def.Pkg().Scope() {
			// We are interested in types defined in the package scope only.
			continue
		}

		if _, ok := def.Type().(*types.Named); ok {
			dname := NewTypeName(path, def.Name())
			if _, ok := t.Objects[dname]; ok {
				panic(fmt.Errorf("double-up: %s %s", path, dname))
			}
			t.Objects[dname] = def
		}
	}
}

// TypeDoc returns the documentation for a type. It can not return
// documentation for a package-level const or var because it uses go/doc at the
// moment. That may change, but it will probably require a new method.
func (t *TypePackageSet) TypeDoc(tn TypeName) (docstr string, err error) {
	var tobj types.Object
	tobj = t.FindObject(tn)
	if tobj == nil {
		err = fmt.Errorf("type %s not found", tn)
		return
	}

	astPkg := t.ASTPackages.Packages[tn.PackagePath]
	if astPkg == nil {
		err = fmt.Errorf("no ast pkg for %s", tn)
		return
	}

	docstr, err = t.ASTPackages.FindComment(tn.PackagePath, tobj.Pos())
	return
}

// FieldDoc returns the documentation for a struct field.
func (t *TypePackageSet) FieldDoc(tn TypeName, field string) (docstr string, err error) {
	var tobj types.Object
	tobj = t.FindObject(tn)
	if err != nil {
		return
	}
	if tobj == nil {
		err = fmt.Errorf("type %s not found", tn)
		return
	}

	astPkg := t.ASTPackages.Packages[tn.PackagePath]
	if astPkg == nil {
		err = fmt.Errorf("no ast pkg for %s", tn)
		return
	}

	stct, ok := tobj.Type().Underlying().(*types.Struct)
	if !ok {
		err = fmt.Errorf("type %s is not a struct", tn)
		return
	}

	flen := stct.NumFields()
	for i := 0; i < flen; i++ {
		f := stct.Field(i)
		if f.Name() == field {
			docstr, err = t.ASTPackages.FindComment(tn.PackagePath, f.Pos())
			return
		}
	}

	return
}

func resolvePackageDir(dir string) string {
	info, err := os.Stat(dir)
	if err != nil || !info.IsDir() {
		return ""
	}
	return dir
}

// isEnum checks if the type contains the "IsEnum()" function
func isEnum(def types.Object, named *types.Named) bool {
	mset := types.NewMethodSet(named)
	sel := mset.Lookup(def.Pkg(), "IsEnum")
	if sel != nil && sel.Kind() == types.MethodVal {
		sig := sel.Type().(*types.Signature)
		return sig.Params().Len() == 0 && sig.Results().Len() == 0
	}
	return false
}
