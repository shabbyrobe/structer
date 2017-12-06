package structer

import (
	"fmt"
	"go/ast"
	"testing"
)

type visitor struct {
	asts *ASTPackageSet
	pkg  string
}

func (v *visitor) Visit(node ast.Node) ast.Visitor {
	if node != nil {
		pd := v.asts.ParentDecl(v.pkg, node.Pos())

		expected := ""
		switch node.(type) {
		case *ast.Package:
			expected = "<nil>"
		case *ast.File:
			expected = "<nil>"
		case *ast.GenDecl:
			expected = "*ast.GenDecl"
		default:
			expected = "*ast.FuncDecl"
		}

		if fmt.Sprintf("%T", pd) != expected {
			panic(fmt.Sprintf("node %T parent %T != %s", node, pd, expected))
		}
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
}
