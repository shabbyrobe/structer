package structer

import (
	"errors"
	"fmt"
	"go/ast"
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

type ASTStackPosFinder struct {
	Pos   token.Pos
	Stack []ast.Node
	Found bool
}

func (a *ASTStackPosFinder) Visit(node ast.Node) (w ast.Visitor) {
	if !a.Found {
		if node == nil {
			a.Stack = a.Stack[:len(a.Stack)-1]
		} else {
			a.Stack = append(a.Stack, node)
			if node.Pos() == a.Pos {
				a.Found = true
				return nil
			}
		}
	}
	return a
}

type ASTPackage struct {
	AST *ast.Package

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

	// Docs on single struct declarations appear on ast.GenDecl, but if we
	// search for a types.Object, we get an ast.TypeSpec. For multi-struct
	// parenthesised declarations, the doc appears on the ast.TypeSpec, i.e:
	//     type (
	//         // yep
	//         foo struct {}
	//     )
	//
	// We get around this by indexing all GenDecls by TypeSpec, so we can look
	// them back up later.
	//
	// More info:
	// https://stackoverflow.com/questions/19580688/go-parser-not-detecting-doc-comments-on-struct-type
	//
	GenDecls map[*ast.TypeSpec]*ast.GenDecl

	Decls map[token.Pos]ast.Decl
}

func NewASTPackageSet() *ASTPackageSet {
	fset := token.NewFileSet()
	pkgs := &ASTPackageSet{
		FileSet:  fset,
		Packages: make(map[string]*ASTPackage),
		ParseDoc: true,
		GenDecls: make(map[*ast.TypeSpec]*ast.GenDecl),
		Decls:    make(map[token.Pos]ast.Decl),
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

func (p *ASTPackageSet) ParentDecl(pkgPath string, pos token.Pos) ast.Node {
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
	if p.Decls[pos] != nil {
		return p.Decls[pos]
	}

	posFinder := &ASTStackPosFinder{Pos: pos}
	ast.Walk(posFinder, dastFile)

	for j := len(posFinder.Stack) - 1; j >= 0; j-- {
		if d, ok := p.Decls[posFinder.Stack[j].Pos()]; ok {
			return d
		}
	}
	return nil
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
	if grps == nil {
		if ts, ok := node.(*ast.TypeSpec); ok {
			gd := p.GenDecls[ts]
			if len(gd.Specs) == 1 {
				grps = []*ast.CommentGroup{gd.Doc}
			}
		}
	}

	if grps != nil {
		for _, grp := range grps {
			docstr += grp.Text()
		}
	}
	return
}

// Add adds the package, found at source path "dir", to the ASTPackageSet.
// GOPATH is not inferred automatically, nor is pkg appended to dir by
// default.
//
// If "" is passed to dir, GOPATH/src + pkg is implied.
//
func (p *ASTPackageSet) Add(dir string, pkg string) error {
	if dir == "" {
		dir = filepath.Join(BuildContext.GOPATH, "src", pkg)
	}

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
		}
	}

	if astPkg.AST == nil {
		return fmt.Errorf("package %q found in %q", pkg, dir)
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

		for _, decl := range astFile.Decls {
			if _, ok := p.Decls[decl.Pos()]; ok {
				panic("bug: decl already exists")
			}

			p.Decls[decl.Pos()] = decl

			if gd, ok := decl.(*ast.GenDecl); ok {
				for _, spec := range gd.Specs {
					if ts, ok := spec.(*ast.TypeSpec); ok {
						p.GenDecls[ts] = gd
					}
				}
			}
		}
	}

	p.Packages[pkg] = astPkg
	return nil
}
