package structer

import (
	"errors"
	"fmt"
	"go/types"
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
	ctx := &walkContext{visitor: visitor}
	return ctx.walk(tn.PackagePath, tn.Name, tn, t)
}

type TypeVisitor interface {
	EnterStruct(WalkContext, StructInfo) error
	LeaveStruct(WalkContext, StructInfo) error

	EnterField(ctx WalkContext, s StructInfo, field *types.Var, tag string) error
	LeaveField(ctx WalkContext, s StructInfo, field *types.Var, tag string) error

	EnterMapKey(ctx WalkContext, ft *types.Map, key types.Type) error
	LeaveMapKey(ctx WalkContext, ft *types.Map, key types.Type) error

	EnterMapElem(ctx WalkContext, ft *types.Map, elem types.Type) error
	LeaveMapElem(ctx WalkContext, ft *types.Map, elem types.Type) error

	EnterPointer(ctx WalkContext, ft *types.Pointer) error
	LeavePointer(ctx WalkContext, ft *types.Pointer) error

	EnterSlice(ctx WalkContext, ft *types.Slice) error
	LeaveSlice(ctx WalkContext, ft *types.Slice) error

	EnterArray(ctx WalkContext, ft *types.Array) error
	LeaveArray(ctx WalkContext, ft *types.Array) error

	VisitBasic(ctx WalkContext, t *types.Basic) error
	VisitNamed(ctx WalkContext, t *types.Named) error
}

// PartialTypeVisitor allows you to conveniently construct a visitor using
// only a handful of the Visit/Enter/Leave methods provided by the TypeVisitor
// interface.
//
type PartialTypeVisitor struct {
	EnterStructFunc func(WalkContext, StructInfo) error
	LeaveStructFunc func(WalkContext, StructInfo) error

	EnterFieldFunc func(ctx WalkContext, s StructInfo, field *types.Var, tag string) error
	LeaveFieldFunc func(ctx WalkContext, s StructInfo, field *types.Var, tag string) error

	EnterMapKeyFunc func(ctx WalkContext, ft *types.Map, key types.Type) error
	LeaveMapKeyFunc func(ctx WalkContext, ft *types.Map, key types.Type) error

	EnterMapElemFunc func(ctx WalkContext, ft *types.Map, elem types.Type) error
	LeaveMapElemFunc func(ctx WalkContext, ft *types.Map, elem types.Type) error

	EnterPointerFunc func(ctx WalkContext, ft *types.Pointer) error
	LeavePointerFunc func(ctx WalkContext, ft *types.Pointer) error

	EnterSliceFunc func(ctx WalkContext, t *types.Slice) error
	LeaveSliceFunc func(ctx WalkContext, t *types.Slice) error

	EnterArrayFunc func(ctx WalkContext, t *types.Array) error
	LeaveArrayFunc func(ctx WalkContext, t *types.Array) error

	VisitBasicFunc func(ctx WalkContext, t *types.Basic) error
	VisitNamedFunc func(ctx WalkContext, t *types.Named) error
}

func (p *PartialTypeVisitor) EnterStruct(ctx WalkContext, s StructInfo) error {
	if p.EnterStructFunc != nil {
		return p.EnterStructFunc(ctx, s)
	}
	return nil
}

func (p *PartialTypeVisitor) LeaveStruct(ctx WalkContext, s StructInfo) error {
	if p.LeaveStructFunc != nil {
		return p.LeaveStructFunc(ctx, s)
	}
	return nil
}

func (p *PartialTypeVisitor) EnterField(ctx WalkContext, s StructInfo, field *types.Var, tag string) error {
	if p.EnterFieldFunc != nil {
		return p.EnterFieldFunc(ctx, s, field, tag)
	}
	return nil
}

func (p *PartialTypeVisitor) LeaveField(ctx WalkContext, s StructInfo, field *types.Var, tag string) error {
	if p.LeaveFieldFunc != nil {
		return p.LeaveFieldFunc(ctx, s, field, tag)
	}
	return nil
}

func (p *PartialTypeVisitor) EnterMapKey(ctx WalkContext, ft *types.Map, key types.Type) error {
	if p.EnterMapKeyFunc != nil {
		return p.EnterMapKeyFunc(ctx, ft, key)
	}
	return nil
}

func (p *PartialTypeVisitor) LeaveMapKey(ctx WalkContext, ft *types.Map, key types.Type) error {
	if p.LeaveMapKeyFunc != nil {
		return p.LeaveMapKeyFunc(ctx, ft, key)
	}
	return nil
}

func (p *PartialTypeVisitor) EnterMapElem(ctx WalkContext, ft *types.Map, elem types.Type) error {
	if p.EnterMapElemFunc != nil {
		return p.EnterMapElemFunc(ctx, ft, elem)
	}
	return nil
}

func (p *PartialTypeVisitor) LeaveMapElem(ctx WalkContext, ft *types.Map, elem types.Type) error {
	if p.LeaveMapElemFunc != nil {
		return p.LeaveMapElemFunc(ctx, ft, elem)
	}
	return nil
}

func (p *PartialTypeVisitor) EnterPointer(ctx WalkContext, t *types.Pointer) error {
	if p.EnterPointerFunc != nil {
		return p.EnterPointer(ctx, t)
	}
	return nil
}

func (p *PartialTypeVisitor) LeavePointer(ctx WalkContext, t *types.Pointer) error {
	if p.LeavePointerFunc != nil {
		return p.LeavePointer(ctx, t)
	}
	return nil
}

func (p *PartialTypeVisitor) EnterSlice(ctx WalkContext, t *types.Slice) error {
	if p.EnterSliceFunc != nil {
		return p.EnterSliceFunc(ctx, t)
	}
	return nil
}

func (p *PartialTypeVisitor) LeaveSlice(ctx WalkContext, t *types.Slice) error {
	if p.LeaveSliceFunc != nil {
		return p.LeaveSliceFunc(ctx, t)
	}
	return nil
}

func (p *PartialTypeVisitor) EnterArray(ctx WalkContext, t *types.Array) error {
	if p.EnterArrayFunc != nil {
		return p.EnterArrayFunc(ctx, t)
	}
	return nil
}

func (p *PartialTypeVisitor) LeaveArray(ctx WalkContext, t *types.Array) error {
	if p.LeaveArrayFunc != nil {
		return p.LeaveArrayFunc(ctx, t)
	}
	return nil
}

func (p *PartialTypeVisitor) VisitBasic(ctx WalkContext, t *types.Basic) error {
	if p.VisitBasicFunc != nil {
		return p.VisitBasicFunc(ctx, t)
	}
	return nil
}

func (p *PartialTypeVisitor) VisitNamed(ctx WalkContext, t *types.Named) error {
	if p.VisitNamedFunc != nil {
		return p.VisitNamedFunc(ctx, t)
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

func (p *MultiVisitor) EnterStruct(ctx WalkContext, s StructInfo) error {
	for _, v := range p.Visitors {
		if err := v.EnterStruct(ctx, s); err != nil {
			return err
		}
	}
	return nil
}

func (p *MultiVisitor) LeaveStruct(ctx WalkContext, s StructInfo) error {
	for _, v := range p.Visitors {
		if err := v.LeaveStruct(ctx, s); err != nil {
			return err
		}
	}
	return nil
}

func (p *MultiVisitor) EnterField(ctx WalkContext, s StructInfo, field *types.Var, tag string) error {
	for _, v := range p.Visitors {
		if err := v.EnterField(ctx, s, field, tag); err != nil {
			return err
		}
	}
	return nil
}

func (p *MultiVisitor) LeaveField(ctx WalkContext, s StructInfo, field *types.Var, tag string) error {
	for _, v := range p.Visitors {
		if err := v.LeaveField(ctx, s, field, tag); err != nil {
			return err
		}
	}
	return nil
}

func (p *MultiVisitor) EnterMapKey(ctx WalkContext, ft *types.Map, key types.Type) error {
	for _, v := range p.Visitors {
		if err := v.EnterMapKey(ctx, ft, key); err != nil {
			return err
		}
	}
	return nil
}

func (p *MultiVisitor) LeaveMapKey(ctx WalkContext, ft *types.Map, key types.Type) error {
	for _, v := range p.Visitors {
		if err := v.LeaveMapKey(ctx, ft, key); err != nil {
			return err
		}
	}
	return nil
}

func (p *MultiVisitor) EnterMapElem(ctx WalkContext, ft *types.Map, elem types.Type) error {
	for _, v := range p.Visitors {
		if err := v.EnterMapElem(ctx, ft, elem); err != nil {
			return err
		}
	}
	return nil
}

func (p *MultiVisitor) LeaveMapElem(ctx WalkContext, ft *types.Map, elem types.Type) error {
	for _, v := range p.Visitors {
		if err := v.LeaveMapElem(ctx, ft, elem); err != nil {
			return err
		}
	}
	return nil
}

func (p *MultiVisitor) EnterPointer(ctx WalkContext, t *types.Pointer) error {
	for _, v := range p.Visitors {
		if err := v.EnterPointer(ctx, t); err != nil {
			return err
		}
	}
	return nil
}

func (p *MultiVisitor) LeavePointer(ctx WalkContext, t *types.Pointer) error {
	for _, v := range p.Visitors {
		if err := v.LeavePointer(ctx, t); err != nil {
			return err
		}
	}
	return nil
}

func (p *MultiVisitor) EnterSlice(ctx WalkContext, t *types.Slice) error {
	for _, v := range p.Visitors {
		if err := v.EnterSlice(ctx, t); err != nil {
			return err
		}
	}
	return nil
}

func (p *MultiVisitor) LeaveSlice(ctx WalkContext, t *types.Slice) error {
	for _, v := range p.Visitors {
		if err := v.LeaveSlice(ctx, t); err != nil {
			return err
		}
	}
	return nil
}

func (p *MultiVisitor) EnterArray(ctx WalkContext, t *types.Array) error {
	for _, v := range p.Visitors {
		if err := v.EnterArray(ctx, t); err != nil {
			return err
		}
	}
	return nil
}

func (p *MultiVisitor) LeaveArray(ctx WalkContext, t *types.Array) error {
	for _, v := range p.Visitors {
		if err := v.LeaveArray(ctx, t); err != nil {
			return err
		}
	}
	return nil
}

func (p *MultiVisitor) VisitBasic(ctx WalkContext, t *types.Basic) error {
	for _, v := range p.Visitors {
		if err := v.VisitBasic(ctx, t); err != nil {
			return err
		}
	}
	return nil
}

func (p *MultiVisitor) VisitNamed(ctx WalkContext, t *types.Named) error {
	for _, v := range p.Visitors {
		if err := v.VisitNamed(ctx, t); err != nil {
			return err
		}
	}
	return nil
}

type WalkContext interface {
	Stack() []types.Type
	Parent() types.Type
}

type walkContext struct {
	stack   []types.Type
	visitor TypeVisitor
}

func (ctx *walkContext) push(t types.Type) {
	ctx.stack = append(ctx.stack, t)
}

func (ctx *walkContext) pop(t types.Type) {
	ln := len(ctx.stack)
	if ln == 1 {
		ctx.stack = []types.Type{}
	} else {
		ctx.stack = ctx.stack[:ln-1]
	}
}

func (ctx *walkContext) Parent() types.Type {
	ln := len(ctx.stack)
	if ln > 0 {
		return ctx.stack[ln-1]
	}
	return nil
}

func (ctx *walkContext) Stack() []types.Type {
	return ctx.stack
}

func (ctx *walkContext) walk(pkg, name string, root TypeName, ft types.Type) error {
	switch ft := ft.(type) {
	case *types.Struct:
		// Descend into nested struct definitions (but not into named ones)
		return ctx.walkStruct(pkg, name, root, ft)

	case *types.Map:
		return ctx.walkMap(pkg, name, root, ft)

	case *types.Slice:
		return ctx.walkSlice(pkg, name, root, ft)

	case *types.Array:
		return ctx.walkArray(pkg, name, root, ft)

	case *types.Pointer:
		return ctx.walkPointer(pkg, name, root, ft)

	case *types.Named:
		return ctx.visitor.VisitNamed(ctx, ft)

	case *types.Basic:
		return ctx.visitor.VisitBasic(ctx, ft)

	default:
		panic(fmt.Errorf("unhandled %T", ft))
	}
}

func (ctx *walkContext) walkSlice(pkg, name string, root TypeName, ft *types.Slice) error {
	ctx.push(ft)
	defer ctx.pop(ft)

	err := ctx.visitor.EnterSlice(ctx, ft)
	if err == WalkOver {
		return nil
	} else if err != nil {
		return err
	}
	if err := ctx.walk(pkg, ft.Elem().String(), root, ft.Elem()); err != nil {
		return err
	}
	if err := ctx.visitor.LeaveSlice(ctx, ft); err != nil {
		return err
	}
	return nil
}

func (ctx *walkContext) walkPointer(pkg, name string, root TypeName, ft *types.Pointer) error {
	ctx.push(ft)
	defer ctx.pop(ft)

	err := ctx.visitor.EnterPointer(ctx, ft)
	if err == WalkOver {
		return nil
	} else if err != nil {
		return err
	}
	if err := ctx.walk(pkg, ft.Elem().String(), root, ft.Elem()); err != nil {
		return err
	}
	if err := ctx.visitor.LeavePointer(ctx, ft); err != nil {
		return err
	}
	return nil
}

func (ctx *walkContext) walkArray(pkg, name string, root TypeName, ft *types.Array) error {
	ctx.push(ft)
	defer ctx.pop(ft)

	err := ctx.visitor.EnterArray(ctx, ft)
	if err == WalkOver {
		return nil
	} else if err != nil {
		return err
	}
	if err := ctx.walk(pkg, ft.Elem().String(), root, ft.Elem()); err != nil {
		return err
	}
	if err := ctx.visitor.LeaveArray(ctx, ft); err != nil {
		return err
	}
	return nil
}

func (ctx *walkContext) walkMap(pkg, name string, root TypeName, ft *types.Map) error {
	ctx.push(ft)
	defer ctx.pop(ft)

	var err error

	err = ctx.visitor.EnterMapKey(ctx, ft, ft.Key())
	if err != WalkOver && err != nil {
		return err
	}
	if err != WalkOver {
		if err := ctx.walk(pkg, ft.Key().String(), root, ft.Key()); err != nil {
			return err
		}
		if err := ctx.visitor.LeaveMapKey(ctx, ft, ft.Key()); err != nil {
			return err
		}
	}

	err = ctx.visitor.EnterMapElem(ctx, ft, ft.Elem())
	if err != WalkOver && err != nil {
		return err
	}
	if err != WalkOver {
		if err := ctx.walk(pkg, ft.Elem().String(), root, ft.Elem()); err != nil {
			return err
		}
		if err := ctx.visitor.LeaveMapElem(ctx, ft, ft.Elem()); err != nil {
			return err
		}
	}
	return nil
}

func (ctx *walkContext) walkStruct(pkg, name string, root TypeName, ft *types.Struct) error {
	ctx.push(ft)
	defer ctx.pop(ft)

	sinfo := StructInfo{Package: pkg, Name: name, Root: root, Struct: ft}

	err := ctx.visitor.EnterStruct(ctx, sinfo)
	if err == WalkOver {
		return nil
	} else if err != nil {
		return err
	}

	for i := 0; i < ft.NumFields(); i++ {
		field := ft.Field(i)
		tag := ft.Tag(i)

		err = ctx.visitor.EnterField(ctx, sinfo, field, tag)
		if err == WalkOver {
			err = nil
			continue
		} else if err != nil {
			return err
		}

		if err := ctx.walk(field.Pkg().Name(), field.Name(), root, field.Type()); err != nil {
			return err
		}
		if err := ctx.visitor.LeaveField(ctx, sinfo, field, tag); err != nil {
			return err
		}
	}
	if err := ctx.visitor.LeaveStruct(ctx, sinfo); err != nil {
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
