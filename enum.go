package structer

import (
	"go/constant"
	"sort"
)

type Enum struct {
	Type       TypeName
	Underlying TypeName
	Values     map[string]constant.Value
}

func (e *Enum) SortedValues() []*EnumValue {
	sorted := make([]*EnumValue, len(e.Values))
	i := 0
	for k, v := range e.Values {
		sorted[i] = &EnumValue{Name: k, Value: v}
		i++
	}
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Name < sorted[j].Name
	})
	return sorted
}

type EnumValue struct {
	Name  string
	Value constant.Value
}
