package structer

import (
	"go/constant"
	"sort"
)

type Consts struct {
	Type       TypeName
	Underlying TypeName
	IsEnum     bool
	Values     []*ConstValue
}

func (e *Consts) SortedValues() []*ConstValue {
	sorted := make([]*ConstValue, len(e.Values))
	copy(sorted, e.Values)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Name.IsBefore(sorted[j].Name)
	})
	return sorted
}

type ConstValue struct {
	Name  TypeName
	Value constant.Value
}

type IsEnum interface {
	IsEnum()
}
