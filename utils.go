package structer

import (
	"strings"

	"github.com/pkg/errors"
)

func SplitType(t string) (pkg, name string, err error) {
	lidx := strings.LastIndex(t, ".")
	if lidx >= 0 {
		pkg, name = t[0:lidx], t[lidx+1:]
	} else {
		err = errors.Errorf("could not parse '%s', expected format full/pkg/path.Type", t)
	}
	return
}
