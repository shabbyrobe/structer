package structer

import (
	"go/types"
	"path/filepath"
	"strings"
)

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

// Move to golib
func splitPath(path string) []string {
	var parts []string
	var lastDir string
	var next string

	cur, file := filepath.Split(path)
	if file != "" {
		cur = filepath.Dir(cur)
	}

	var i int
	for {
		next = filepath.Dir(cur)
		bit := strings.TrimPrefix(cur, next)
		if len(bit) > 0 && bit[0] == filepath.Separator {
			bit = bit[1:]
		}
		if len(bit) > 0 {
			parts = append(parts, bit)
		}

		cur = next
		if lastDir == cur {
			parts = append(parts, cur)
			break
		}
		lastDir = cur
		i++
		if i == 1000 {
			panic("infinite loop protector")
		}
	}

	for left, right := 0, len(parts)-1; left < right; left, right = left+1, right-1 {
		parts[left], parts[right] = parts[right], parts[left]
	}
	if file != "" {
		parts = append(parts, file)
	}
	return parts
}
