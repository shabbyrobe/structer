package structer

import (
	"fmt"
	"go/ast"
	"reflect"
	"testing"
)

type visitor struct {
	asts   *ASTPackageSet
	pkg    string
	visits []string
}

func (v *visitor) Visit(node ast.Node) ast.Visitor {
	if node != nil {
		pd := v.asts.ParentDecl(v.pkg, node.Pos())
		v.visits = append(v.visits, fmt.Sprintf("%T %T", node, pd))
	}
	return v
}

func TestASTParentDecl(t *testing.T) {
	aps := NewASTPackageSet()
	pkg := "github.com/shabbyrobe/structer/testpkg/astparent"
	if err := aps.Add("", pkg); err != nil {
		t.Fatal(err)
	}
	node := aps.Packages[pkg].AST
	vis := &visitor{asts: aps, pkg: pkg}
	ast.Walk(vis, node)

	expected := []string{
		"*ast.Package <nil>",
		"*ast.File <nil>",
		"*ast.Ident <nil>",
		"*ast.GenDecl *ast.GenDecl",
		"*ast.TypeSpec *ast.GenDecl",
		"*ast.Ident *ast.GenDecl",
		"*ast.StructType *ast.GenDecl",
		"*ast.FieldList *ast.GenDecl",
		"*ast.Field *ast.GenDecl",
		"*ast.Ident *ast.GenDecl",
		"*ast.Ident *ast.GenDecl",
		"*ast.GenDecl *ast.GenDecl",
		"*ast.ValueSpec *ast.GenDecl",
		"*ast.Ident *ast.GenDecl",
		"*ast.Ident *ast.GenDecl",
		"*ast.GenDecl *ast.GenDecl",
		"*ast.TypeSpec *ast.GenDecl",
		"*ast.Ident *ast.GenDecl",
		"*ast.InterfaceType *ast.GenDecl",
		"*ast.FieldList *ast.GenDecl",
		"*ast.Field *ast.GenDecl",
		"*ast.Ident *ast.GenDecl",
		"*ast.FuncType *ast.GenDecl",
		"*ast.FieldList *ast.GenDecl",
		"*ast.FuncDecl *ast.FuncDecl",
		"*ast.FieldList *ast.FuncDecl",
		"*ast.Field *ast.FuncDecl",
		"*ast.Ident *ast.FuncDecl",
		"*ast.StarExpr *ast.FuncDecl",
		"*ast.Ident *ast.FuncDecl",
		"*ast.Ident *ast.FuncDecl",
		"*ast.FuncType *ast.FuncDecl",
		"*ast.FieldList *ast.FuncDecl",
		"*ast.BlockStmt *ast.FuncDecl",
		"*ast.IfStmt *ast.FuncDecl",
		"*ast.BinaryExpr *ast.FuncDecl",
		"*ast.BasicLit *ast.FuncDecl",
		"*ast.BasicLit *ast.FuncDecl",
		"*ast.BlockStmt *ast.FuncDecl",
		"*ast.ExprStmt *ast.FuncDecl",
		"*ast.CallExpr *ast.FuncDecl",
		"*ast.Ident *ast.FuncDecl",
		"*ast.BasicLit *ast.FuncDecl",
	}
	if !reflect.DeepEqual(expected, vis.visits) {
		t.Fatal()
	}
}
