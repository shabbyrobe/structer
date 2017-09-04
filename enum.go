package structer

import "go/constant"

type Enum struct {
	Type       TypeName
	Underlying TypeName
	Values     map[string]constant.Value
}
