package structer

import (
	"go/ast"
	"go/parser"
	"go/token"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
)

type ASTPosFinder struct {
	Pos  token.Pos
	Node ast.Node
}

func (a *ASTPosFinder) Visit(node ast.Node) (w ast.Visitor) {
	if node != nil && node.Pos() == a.Pos {
		a.Node = node
		return nil
	}
	return a
}

type ASTPackage struct {
	AST *ast.Package

	// all file names, not filtered by build, found in the package, relative
	// to FullPath
	Files []string

	// all file contents, not filtered by build, found in the package, relative
	// to FullPath. the ast package makes it unreasonably difficult to get at this.
	Contents map[string][]byte

	// package identifier (i.e. for import "foo/bar/baz", the Name would be 'baz'
	Name string

	// import path to package - "foo/bar/baz"
	Path string

	// filesystem path to package - "/path/to/my/go/src/foo/bar/baz"
	FullPath string

	// ast.Package unhelpfully indexes by absolute path. this is not useful
	// for us.
	FileASTs map[string]*ast.File
}

type ASTPackageSet struct {
	FileSet  *token.FileSet
	Packages map[string]*ASTPackage
}

func NewASTPackageSet() *ASTPackageSet {
	fset := token.NewFileSet()
	pkgs := &ASTPackageSet{
		FileSet:  fset,
		Packages: make(map[string]*ASTPackage),
	}
	return pkgs
}

func (p *ASTPackageSet) FindNodeByPackagePathPos(pkgPath string, pos token.Pos) ast.Node {
	// FIXME: investigate what a "position altering comment" does to this.
	posn := p.FileSet.PositionFor(pos, false)
	dast := p.Packages[pkgPath]
	if dast == nil {
		return nil
	}
	dastFile := dast.AST.Files[posn.Filename]
	if dastFile == nil {
		return nil
	}
	posFinder := &ASTPosFinder{Pos: pos}
	ast.Walk(posFinder, dastFile)
	return posFinder.Node
}

func (p *ASTPackageSet) Add(dir string, pkg string) error {
	info, err := os.Stat(dir)
	if err != nil {
		return err
	}
	if !info.IsDir() {
		return errors.Errorf("unknown dir %s", dir)
	}

	astPkg := &ASTPackage{
		FileASTs: make(map[string]*ast.File),
		FullPath: dir,
		Path:     pkg,
		Contents: make(map[string][]byte),
		Name:     filepath.Base(strings.TrimRight(pkg, "/")),
	}

	var ferr error
	filter := func(info os.FileInfo) bool {
		astPkg.Files = append(astPkg.Files, info.Name())
		fullName := filepath.Join(dir, info.Name())
		astPkg.Contents[info.Name()], err = ioutil.ReadFile(fullName)
		if ferr != nil {
			return false
		}
		return true
	}

	pkgs, err := parser.ParseDir(p.FileSet, dir, filter, parser.ParseComments)
	if err != nil {
		return err
	}
	if ferr != nil {
		return ferr
	}

	// Some stdlib packages have a "main" with an ignore build
	// tag in them as well as a regular package.
	// FIXME: Now that we are using the build package elsewhere,
	// is this needed?
	main := false
	pkey := ""
	for k := range pkgs {
		if k == "main" {
			main = true
		} else if !strings.HasSuffix(k, "_test") {
			if pkey != "" {
				return errors.Errorf("multiple packages found in %s", dir)
			}
			pkey = k
		}
	}

	if pkey == "" && main {
		pkey = "main"
	}
	if len(pkgs) > 0 {
		if pkey == "" {
			return errors.Errorf("no packages found in %s", dir)
		} else {
			astPkg.AST = pkgs[pkey]
		}
	}

	// *ast.Package indexes files by absolute filesystem path so we need
	// to build a separate index of package relative names
	for name, astFile := range astPkg.AST.Files {
		astPkg.FileASTs[filepath.Base(name)] = astFile
	}

	p.Packages[pkg] = astPkg
	return nil
}
