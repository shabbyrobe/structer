package structer

import (
	"fmt"
	"go/types"
	"regexp"
	"sort"
	"strings"
)

var exportedPattern = regexp.MustCompile(`^\p{Lu}`)

type TypeName struct {
	PackagePath string

	Full string
	Name string

	isBuiltin bool
}

func (t TypeName) IsBuiltin() bool { return t.isBuiltin }

func (t TypeName) IsBefore(c TypeName) bool {
	return t.Full < c.Full
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
		Name:        name,
		Full:        pkgPath + "." + name,
	}
}

func ExtractTypeName(t types.Type) TypeName {
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
		err = fmt.Errorf("invalid type %s", name)
		return
	}
	fullpkg, t := name[0:last], name[last+1:]
	tn = TypeName{
		PackagePath: fullpkg,
		Name:        t,
		Full:        name,
	}
	return
}

func ParseLocalName(name string, localPkg string) (tn TypeName, err error) {
	last := strings.LastIndex(name, ".")
	if last < 0 {
		return NewTypeName(localPkg, name), nil
	}
	return ParseTypeName(name)
}

type TypeNames []TypeName

func (ns TypeNames) Sort() {
	sort.Slice(ns, func(i, j int) bool {
		return ns[i].Full < ns[j].Full
	})
}

func (ns TypeNames) Sorted() TypeNames {
	names := make(TypeNames, len(ns))
	copy(names, ns)
	names.Sort()
	return names
}

type TypeMap map[TypeName]types.Type

func (m TypeMap) SortedKeys() TypeNames {
	names := make(TypeNames, len(m))
	i := 0
	for k := range m {
		names[i] = k
		i++
	}
	names.Sort()
	return names
}

type ObjectMap map[TypeName]types.Object

func (m ObjectMap) SortedKeys() TypeNames {
	names := make(TypeNames, len(m))
	i := 0
	for k := range m {
		names[i] = k
		i++
	}
	names.Sort()
	return names
}
