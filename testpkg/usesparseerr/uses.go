package usesparseerr

import "github.com/shabbyrobe/structer/testpkg/parseerr"

type Test struct {
	Foo   parseerr.Good
	Valid valid.Valid
}
