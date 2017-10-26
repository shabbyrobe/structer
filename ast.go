package structer

import (
	"errors"
	"fmt"
	"go/ast"
	"go/doc"
	"go/parser"
	"go/token"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
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

	doc *doc.Package

	CommentMap ast.CommentMap

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
	ParseDoc bool
	FileSet  *token.FileSet
	Packages map[string]*ASTPackage
}

func NewASTPackageSet() *ASTPackageSet {
	fset := token.NewFileSet()
	pkgs := &ASTPackageSet{
		FileSet:  fset,
		Packages: make(map[string]*ASTPackage),
		ParseDoc: true,
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

func (p *ASTPackageSet) FindComment(pkgPath string, pos token.Pos) (docstr string, err error) {
	node := p.FindNodeByPackagePathPos(pkgPath, pos)
	if node == nil {
		return
	}

	astPkg := p.Packages[pkgPath]
	if astPkg == nil {
		err = fmt.Errorf("could not find package %s", pkgPath)
		return
	}

	// Some funny shenanigans here, maybe need to keep this in mind in future?
	// https://stackoverflow.com/questions/30092417/what-is-the-difference-between-doc-and-comment-in-go-ast-package

	grps := astPkg.CommentMap[node]
	if grps != nil {
		for _, grp := range grps {
			docstr += grp.Text()
		}
	}
	return
}

func (p *ASTPackageSet) Add(dir string, pkg string) error {
	info, err := os.Stat(dir)
	if err != nil {
		return err
	}
	if !info.IsDir() {
		return fmt.Errorf("unknown dir %s", dir)
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

	// The left hand doesn't know what the right hand is doing.
	// The documentation for go/doc.New() says "New computes the package
	// documentation for the given package AST. New takes ownership of the AST
	// pkg and may edit or overwrite it."
	// Unfortunately, there's no way to clone an AST node. This is pretty half-assed.
	// Go should either provide the facility to clone an AST node, or doc.New()
	// should not take ownership. Then this horseshit would not be necessary.
	var docPkgs map[string]*ast.Package
	if p.ParseDoc {
		docPkgs, err = parser.ParseDir(p.FileSet, dir, filter, parser.ParseComments)
		if err != nil {
			return err
		}
		if ferr != nil {
			return ferr
		}
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
				return fmt.Errorf("multiple packages found in %s", dir)
			}
			pkey = k
		}
	}

	if pkey == "" && main {
		pkey = "main"
	}
	if len(pkgs) > 0 {
		if pkey == "" {
			return fmt.Errorf("no packages found in %s", dir)
		} else {
			astPkg.AST = pkgs[pkey]
			dp := docPkgs[pkey]
			if dp != nil {
				docPkg := doc.New(dp, dir, 0)
				astPkg.doc = docPkg
			}
		}
	}

	// *ast.Package indexes files by absolute filesystem path so we need
	// to build a separate index of package relative names
	for name, astFile := range astPkg.AST.Files {
		astPkg.FileASTs[filepath.Base(name)] = astFile

		if p.ParseDoc {
			if astPkg.CommentMap == nil {
				astPkg.CommentMap = ast.NewCommentMap(p.FileSet, astFile, astFile.Comments)
			} else {
				childMap := ast.NewCommentMap(p.FileSet, astFile, astFile.Comments)
				for k, v := range childMap {
					if _, ok := astPkg.CommentMap[k]; ok {
						panic(errors.New("duplicate node found"))
					}
					astPkg.CommentMap[k] = v
				}
			}
		}
	}

	p.Packages[pkg] = astPkg
	return nil
}
