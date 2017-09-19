package structer

import (
	"go/ast"
	"go/build"
	"go/constant"
	"go/importer"
	"go/types"
	"os"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
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
	tps.TypesConfig.IgnoreFuncBodies = true
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
		err = errors.Errorf("import path not found for pkg %s, type %s", pkg, typeName)
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

	err = errors.Errorf("import path not found for pkg %s, type %s", pkg, typeName)
	return
}

func (t *TypePackageSet) ExtractSource(name TypeName) ([]byte, error) {
	def := t.Objects[name]
	if def == nil {
		return nil, errors.Errorf("could not find def for %s", name)
	}

	pkg := def.Pkg().Path()
	node := t.ASTPackages.FindNodeByPackagePathPos(pkg, def.Pos())
	if node == nil {
		return nil, errors.Errorf("no ast node for %s.%s", pkg, name)
	}

	pos := t.ASTPackages.FileSet.Position(node.Pos()).Offset
	end := t.ASTPackages.FileSet.Position(node.End()).Offset

	posn := t.ASTPackages.FileSet.PositionFor(node.Pos(), false)
	astPkg := t.ASTPackages.Packages[pkg]
	if astPkg == nil {
		return nil, errors.Errorf("no ast pkg for %s", pkg)
	}

	contents, ok := astPkg.Contents[filepath.Base(posn.Filename)]
	if !ok {
		return nil, errors.Errorf("no contents for %s.%s", pkg, name)
	}

	return contents[pos:end], nil
}

// ExtractEnum extracts all constants that satisfy the supplied type from
// the same package.
func (t *TypePackageSet) ExtractEnum(name TypeName, includeUnexported bool) (*Enum, error) {
	def := t.Objects[name]
	if def == nil {
		return nil, errors.Errorf("could not find def for %s", name)
	}

	enum := &Enum{
		Type:       name,
		Underlying: extractTypeName(def.Type().Underlying()),
		Values:     make(map[string]constant.Value),
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
			enum.Values[n.String()] = cns.Val()
		}
	}
	return enum, nil
}

func (t *TypePackageSet) ResolvePath(path, srcDir string) (PackageKind, string, error) {
	// Is it a VendorPackage?
	goSrcPath := filepath.Join(BuildContext.GOPATH, "src")
	cur := srcDir
	for cur != goSrcPath {
		vendorDir := filepath.Join(cur, "vendor")
		info, _ := os.Stat(vendorDir)
		if info != nil && info.IsDir() {
			if dir := resolvePackageDir(filepath.Join(vendorDir, path)); dir != "" {
				return VendorPackage, dir, nil
			}
		}
		cur = filepath.Dir(cur)
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

func (t *TypePackageSet) Import(importPath string) (*types.Package, error) {
	srcPath := filepath.Join(BuildContext.GOPATH, "src", importPath)
	return t.ImportFrom(importPath, srcPath, 0)
}

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
			Types: make(map[ast.Expr]types.TypeAndValue),
			Defs:  make(map[*ast.Ident]types.Object),
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
				asts = append(asts, t.ASTPackages.Packages[importPath].AST.Files[full])
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

// ObjectByName is a concession to pragmatism - you won't always have a
// TypeName and it might not always be convenient to create one. Try
// not to use this though and let me know if you feel you're forced to
// so I can try and fix this aspect up a bit.
func (t *TypePackageSet) ObjectByName(name string) types.Object {
	tn, err := ParseTypeName(name)
	if err != nil {
		return nil
	}
	return t.Objects[tn]
}

// Lists all types in all imported user packages which implement the interface supplied
// in the argument.
//
// This does not yield types from system packages or vendor packages yet.
//
func (t *TypePackageSet) Implements(ifaceName TypeName) (map[TypeName]types.Type, error) {
	// Import the package referred to in the argument if we have not seen it before.
	// This should validate the incoming name as a benefit.
	if _, ok := t.TypePackages[ifaceName.PackagePath]; !ok {
		if _, err := t.Import(ifaceName.PackagePath); err != nil {
			return nil, err
		}
	}

	iface, ok := t.Objects[ifaceName]
	if !ok {
		return nil, errors.Errorf("could not find object for %s", ifaceName)
	}
	ifaceTyp := iface.Type()
	if !types.IsInterface(ifaceTyp) {
		return nil, errors.Errorf("type %s is not an interface", ifaceName)
	}

	var implements = make(map[TypeName]types.Type)

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
				// implements[extractTypeName(s)] = fTyp
				implements[extractTypeName(fTyp)] = impl
			}
		}
	}

	return implements, nil
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
				panic(errors.Errorf("double-up: %s %s", path, dname))
			}
			t.Objects[dname] = def
		}
	}
}

func resolvePackageDir(dir string) string {
	info, err := os.Stat(dir)
	if err != nil || !info.IsDir() {
		return ""
	}
	return dir
}
