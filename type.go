package structer

import (
	"go/types"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/pkg/errors"
)

var exportedPattern = regexp.MustCompile(`^\p{Lu}`)

type TypeName struct {
	PackagePath string
	PackageName string
	Full        string
	Name        string

	isBuiltin bool
}

func (t TypeName) IsExported() bool {
	if t.isBuiltin {
		return true
	}
	return exportedPattern.MatchString(t.Name)
}

func (t TypeName) String() string {
	if t.isBuiltin {
		return t.Name
	} else {
		return t.Full
	}
}

func (t TypeName) IsType(typ types.Type) bool {
	if typ == nil {
		return false
	}
	return t.String() == typ.String()
}

func (t TypeName) ImportName(rel string, importName bool) string {
	if t.isBuiltin {
		return t.Name
	} else if rel == t.PackagePath || (importName && rel == t.PackageName) {
		return t.Name
	} else {
		return t.PackageName + "." + t.Name
	}
}

func NewBuiltinType(name string) TypeName {
	return TypeName{
		Full:      name,
		Name:      name,
		isBuiltin: true,
	}
}

func NewTypeName(pkgPath string, name string) TypeName {
	return TypeName{
		PackagePath: pkgPath,
		PackageName: filepath.Base(pkgPath),
		Full:        pkgPath + "." + name,
		Name:        name,
	}
}

// FIXME: this is kinda crappy
func extractTypeName(t types.Type) TypeName {
	name := t.String()
	last := strings.LastIndex(name, ".")
	if last < 0 {
		return NewBuiltinType(name)
	} else {
		t, _ := ParseTypeName(name)
		return t
	}
}

func ParseTypeName(name string) (tn TypeName, err error) {
	last := strings.LastIndex(name, ".")
	if last < 0 {
		err = errors.Errorf("invalid type %s", name)
		return
	}
	fullpkg, t := name[0:last], name[last+1:]
	tn = TypeName{
		PackagePath: fullpkg,
		PackageName: filepath.Base(fullpkg),
		Name:        t,
		Full:        name,
	}
	return
}
