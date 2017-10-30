package structer

import "go/types"

// types.Implements is a bit more low level - this will report true if the iface
// is a named type with an underlying interface, and also if the checked type
// is assignable as a pointer as well as without.
func ActuallyImplements(typ types.Type, ifaceTyp types.Type) bool {
	if ifaceTyp == typ {
		return true
	}

	var ok bool
	var iface *types.Interface
	if iface, ok = ifaceTyp.(*types.Interface); !ok {
		if iface, ok = ifaceTyp.Underlying().(*types.Interface); !ok {
			return false
		}
	}

	var impl types.Type
	if types.AssignableTo(typ, iface) {
		impl = typ
	} else {
		ptr := types.NewPointer(typ)
		if types.AssignableTo(ptr, iface) {
			impl = ptr
		}
	}
	return impl != nil
}

func IsObjectInvalid(obj types.Object) bool {
	return IsInvalid(obj.Type()) || IsInvalid(obj.Type().Underlying())
}

func IsInvalid(typ types.Type) bool {
	if basic, ok := typ.(*types.Basic); ok {
		return basic.Kind() == types.Invalid
	}
	return false
}
