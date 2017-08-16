package structer

import (
	"go/types"

	"github.com/pkg/errors"
)

var (
	// step over the current node. used in Enter calls to go to the
	// next type without descending.
	WalkOver = errors.New("stop")
)

// interface check
var (
	_ TypeVisitor = &PartialTypeVisitor{}
	_ TypeVisitor = &MultiVisitor{}
)

// Walk walks a type recursively, visiting each child type using the
// visitor provided.
//
// This allows you to traverse complex nested definitions, like
// struct { Foo map[struct{X, Y int}]struct{ Bar map[struct{Z, N}]string } }
// and respond to each node. Each node will have access to complete information
// about the type from the golang "types" package. If your visitor carries
// a reference to TypePackageSet as well, you can do complex lookups on the
// visited types.
//
// It is designed for use with code generators that want to dismantle a
// struct and respond dynamically to the types therein.
//
func Walk(tn TypeName, t types.Type, visitor TypeVisitor) error {
	return walk(tn.PackagePath, tn.Name, tn, t, visitor)
}

type TypeVisitor interface {
	EnterStruct(StructInfo) error
	LeaveStruct(StructInfo) error

	EnterField(s StructInfo, field *types.Var, tag string) error
	LeaveField(s StructInfo, field *types.Var, tag string) error

	EnterMapKey(ft *types.Map, key types.Type) error
	LeaveMapKey(ft *types.Map, key types.Type) error

	EnterMapElem(ft *types.Map, elem types.Type) error
	LeaveMapElem(ft *types.Map, elem types.Type) error

	EnterSlice(ft *types.Slice) error
	LeaveSlice(ft *types.Slice) error

	EnterArray(ft *types.Array) error
	LeaveArray(ft *types.Array) error

	VisitBasic(t *types.Basic) error
	VisitNamed(t *types.Named) error
}

// PartialTypeVisitor allows you to conveniently construct a visitor using
// only a handful of the Visit/Enter/Leave methods provided by the TypeVisitor
// interface.
//
type PartialTypeVisitor struct {
	EnterStructFunc func(StructInfo) error
	LeaveStructFunc func(StructInfo) error

	EnterFieldFunc func(s StructInfo, field *types.Var, tag string) error
	LeaveFieldFunc func(s StructInfo, field *types.Var, tag string) error

	EnterMapKeyFunc func(ft *types.Map, key types.Type) error
	LeaveMapKeyFunc func(ft *types.Map, key types.Type) error

	EnterMapElemFunc func(ft *types.Map, elem types.Type) error
	LeaveMapElemFunc func(ft *types.Map, elem types.Type) error

	EnterSliceFunc func(t *types.Slice) error
	LeaveSliceFunc func(t *types.Slice) error

	EnterArrayFunc func(t *types.Array) error
	LeaveArrayFunc func(t *types.Array) error

	VisitBasicFunc func(t *types.Basic) error
	VisitNamedFunc func(t *types.Named) error
}

func (p *PartialTypeVisitor) EnterStruct(s StructInfo) error {
	if p.EnterStructFunc != nil {
		return p.EnterStructFunc(s)
	}
	return nil
}

func (p *PartialTypeVisitor) LeaveStruct(s StructInfo) error {
	if p.LeaveStructFunc != nil {
		return p.LeaveStructFunc(s)
	}
	return nil
}

func (p *PartialTypeVisitor) EnterField(s StructInfo, field *types.Var, tag string) error {
	if p.EnterFieldFunc != nil {
		return p.EnterFieldFunc(s, field, tag)
	}
	return nil
}

func (p *PartialTypeVisitor) LeaveField(s StructInfo, field *types.Var, tag string) error {
	if p.LeaveFieldFunc != nil {
		return p.LeaveFieldFunc(s, field, tag)
	}
	return nil
}

func (p *PartialTypeVisitor) EnterMapKey(ft *types.Map, key types.Type) error {
	if p.EnterMapKeyFunc != nil {
		return p.EnterMapKeyFunc(ft, key)
	}
	return nil
}

func (p *PartialTypeVisitor) LeaveMapKey(ft *types.Map, key types.Type) error {
	if p.LeaveMapKeyFunc != nil {
		return p.LeaveMapKeyFunc(ft, key)
	}
	return nil
}

func (p *PartialTypeVisitor) EnterMapElem(ft *types.Map, elem types.Type) error {
	if p.EnterMapElemFunc != nil {
		return p.EnterMapElemFunc(ft, elem)
	}
	return nil
}

func (p *PartialTypeVisitor) LeaveMapElem(ft *types.Map, elem types.Type) error {
	if p.LeaveMapElemFunc != nil {
		return p.LeaveMapElemFunc(ft, elem)
	}
	return nil
}

func (p *PartialTypeVisitor) EnterSlice(t *types.Slice) error {
	if p.EnterSliceFunc != nil {
		return p.EnterSliceFunc(t)
	}
	return nil
}

func (p *PartialTypeVisitor) LeaveSlice(t *types.Slice) error {
	if p.LeaveSliceFunc != nil {
		return p.LeaveSliceFunc(t)
	}
	return nil
}

func (p *PartialTypeVisitor) EnterArray(t *types.Array) error {
	if p.EnterArrayFunc != nil {
		return p.EnterArrayFunc(t)
	}
	return nil
}

func (p *PartialTypeVisitor) LeaveArray(t *types.Array) error {
	if p.LeaveArrayFunc != nil {
		return p.LeaveArrayFunc(t)
	}
	return nil
}

func (p *PartialTypeVisitor) VisitBasic(t *types.Basic) error {
	if p.VisitBasicFunc != nil {
		return p.VisitBasicFunc(t)
	}
	return nil
}

func (p *PartialTypeVisitor) VisitNamed(t *types.Named) error {
	if p.VisitNamedFunc != nil {
		return p.VisitNamedFunc(t)
	}
	return nil
}

// MultiVisitor allows you to wrap multiple visitors and call each of them
// in sequence for each node in the type definition.
//
// If one of the visitors returns an error, the Visit/Leave/Enter methods will
// abort immediately rather than complete.
//
type MultiVisitor struct {
	Visitors []TypeVisitor
}

func (p *MultiVisitor) EnterStruct(s StructInfo) error {
	for _, v := range p.Visitors {
		if err := v.EnterStruct(s); err != nil {
			return err
		}
	}
	return nil
}

func (p *MultiVisitor) LeaveStruct(s StructInfo) error {
	for _, v := range p.Visitors {
		if err := v.LeaveStruct(s); err != nil {
			return err
		}
	}
	return nil
}

func (p *MultiVisitor) EnterField(s StructInfo, field *types.Var, tag string) error {
	for _, v := range p.Visitors {
		if err := v.EnterField(s, field, tag); err != nil {
			return err
		}
	}
	return nil
}

func (p *MultiVisitor) LeaveField(s StructInfo, field *types.Var, tag string) error {
	for _, v := range p.Visitors {
		if err := v.LeaveField(s, field, tag); err != nil {
			return err
		}
	}
	return nil
}

func (p *MultiVisitor) EnterMapKey(ft *types.Map, key types.Type) error {
	for _, v := range p.Visitors {
		if err := v.EnterMapKey(ft, key); err != nil {
			return err
		}
	}
	return nil
}

func (p *MultiVisitor) LeaveMapKey(ft *types.Map, key types.Type) error {
	for _, v := range p.Visitors {
		if err := v.LeaveMapKey(ft, key); err != nil {
			return err
		}
	}
	return nil
}

func (p *MultiVisitor) EnterMapElem(ft *types.Map, elem types.Type) error {
	for _, v := range p.Visitors {
		if err := v.EnterMapElem(ft, elem); err != nil {
			return err
		}
	}
	return nil
}

func (p *MultiVisitor) LeaveMapElem(ft *types.Map, elem types.Type) error {
	for _, v := range p.Visitors {
		if err := v.LeaveMapElem(ft, elem); err != nil {
			return err
		}
	}
	return nil
}

func (p *MultiVisitor) EnterSlice(t *types.Slice) error {
	for _, v := range p.Visitors {
		if err := v.EnterSlice(t); err != nil {
			return err
		}
	}
	return nil
}

func (p *MultiVisitor) LeaveSlice(t *types.Slice) error {
	for _, v := range p.Visitors {
		if err := v.LeaveSlice(t); err != nil {
			return err
		}
	}
	return nil
}

func (p *MultiVisitor) EnterArray(t *types.Array) error {
	for _, v := range p.Visitors {
		if err := v.EnterArray(t); err != nil {
			return err
		}
	}
	return nil
}

func (p *MultiVisitor) LeaveArray(t *types.Array) error {
	for _, v := range p.Visitors {
		if err := v.LeaveArray(t); err != nil {
			return err
		}
	}
	return nil
}

func (p *MultiVisitor) VisitBasic(t *types.Basic) error {
	for _, v := range p.Visitors {
		if err := v.VisitBasic(t); err != nil {
			return err
		}
	}
	return nil
}

func (p *MultiVisitor) VisitNamed(t *types.Named) error {
	for _, v := range p.Visitors {
		if err := v.VisitNamed(t); err != nil {
			return err
		}
	}
	return nil
}

func walk(pkg, name string, root TypeName, ft types.Type, visitor TypeVisitor) error {
	switch ft := ft.(type) {
	case *types.Struct:
		// Descend into nested struct definitions (but not into named ones)
		return walkStruct(pkg, name, root, ft, visitor)

	case *types.Map:
		return walkMap(pkg, name, root, ft, visitor)

	case *types.Slice:
		return walkSlice(pkg, name, root, ft, visitor)

	case *types.Array:
		return walkArray(pkg, name, root, ft, visitor)

	case *types.Pointer:
		// FIXME: should EnterPointer, LeavePointer.
		return walk(pkg, ft.Elem().String(), root, ft.Elem(), visitor)

	case *types.Named:
		return visitor.VisitNamed(ft)

	case *types.Basic:
		return visitor.VisitBasic(ft)

	default:
		panic(errors.Errorf("unhandled %T", ft))
	}
}

func walkSlice(pkg, name string, root TypeName, ft *types.Slice, visitor TypeVisitor) error {
	err := visitor.EnterSlice(ft)
	if err == WalkOver {
		return nil
	} else if err != nil {
		return err
	}
	if err := walk(pkg, ft.Elem().String(), root, ft.Elem(), visitor); err != nil {
		return err
	}
	if err := visitor.LeaveSlice(ft); err != nil {
		return err
	}
	return nil
}

func walkArray(pkg, name string, root TypeName, ft *types.Array, visitor TypeVisitor) error {
	err := visitor.EnterArray(ft)
	if err == WalkOver {
		return nil
	} else if err != nil {
		return err
	}
	if err := walk(pkg, ft.Elem().String(), root, ft.Elem(), visitor); err != nil {
		return err
	}
	if err := visitor.LeaveArray(ft); err != nil {
		return err
	}
	return nil
}

func walkMap(pkg, name string, root TypeName, ft *types.Map, visitor TypeVisitor) error {
	var err error

	err = visitor.EnterMapKey(ft, ft.Key())
	if err != WalkOver && err != nil {
		return err
	}
	if err != WalkOver {
		if err := walk(pkg, ft.Key().String(), root, ft.Key(), visitor); err != nil {
			return err
		}
		if err := visitor.LeaveMapKey(ft, ft.Key()); err != nil {
			return err
		}
	}

	err = visitor.EnterMapElem(ft, ft.Elem())
	if err != WalkOver && err != nil {
		return err
	}
	if err != WalkOver {
		if err := walk(pkg, ft.Elem().String(), root, ft.Elem(), visitor); err != nil {
			return err
		}
		if err := visitor.LeaveMapElem(ft, ft.Elem()); err != nil {
			return err
		}
	}
	return nil
}

func walkStruct(pkg, name string, root TypeName, ft *types.Struct, visitor TypeVisitor) error {
	sinfo := StructInfo{Package: pkg, Name: name, Root: root}

	err := visitor.EnterStruct(sinfo)
	if err == WalkOver {
		return nil
	} else if err != nil {
		return err
	}

	for i := 0; i < ft.NumFields(); i++ {
		field := ft.Field(i)
		tag := ft.Tag(i)

		err = visitor.EnterField(sinfo, field, tag)
		if err == WalkOver {
			err = nil
			continue
		} else if err != nil {
			return err
		}

		if err := walk(field.Pkg().Name(), field.Name(), root, field.Type(), visitor); err != nil {
			return err
		}
		if err := visitor.LeaveField(sinfo, field, tag); err != nil {
			return err
		}
	}
	if err := visitor.LeaveStruct(sinfo); err != nil {
		return err
	}
	return nil
}

type StructInfo struct {
	// Y U NO TypeName?
	// Because nested structs don't have a fully qualified name.
	Package string
	Name    string

	// This is the type name at the root of the Walk. This gives some context
	// recursively walking nested structs.
	Root TypeName

	Struct *types.Struct
}
