package structer

import (
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
)

type TypeName struct {
	PackagePath string
	PackageName string
	Full        string
	Name        string

	isBuiltin bool
}

func (t TypeName) String() string {
	if t.isBuiltin {
		return t.Name
	} else {
		return t.Full
	}
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
