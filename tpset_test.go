package structer

import (
	"go/types"
	"reflect"
	"sort"
	"testing"
)

func TestTypePackageSetParseError(t *testing.T) {
	tpset := NewTypePackageSet()
	_, err := tpset.Import("github.com/shabbyrobe/structer/testpkg/parseerr")
	if err == nil {
		t.Errorf("expected error")
	}
	if len(tpset.Objects) > 0 {
		t.Errorf("expected no valid objects")
	}
}

// Ensures that even if an interface is incorrectly implemented in a package,
// the types can still be extracted. This is important as we intend this to
// be used for code generators, which may be responsible for generating the
// very missing code that causes the error in the first place.
//
func TestTypePackageSetInterfaceError(t *testing.T) {
	tpset := NewTypePackageSet()
	_, err := tpset.Import("github.com/shabbyrobe/structer/testpkg/intferr")
	if err != nil {
		t.Errorf("expected no error")
	}
	found := []string{}
	for k := range tpset.Objects {
		found = append(found, k.String())
	}
	expected := []string{
		"github.com/shabbyrobe/structer/testpkg/intferr.Pants",
		"github.com/shabbyrobe/structer/testpkg/intferr.Pantsable",
	}
	sort.Strings(found)
	if !reflect.DeepEqual(found, expected) {
		t.Errorf("types did not match expected")
	}
}

func TestTypePackageSetUsesPackageWithParseError(t *testing.T) {
	tpset := NewTypePackageSet()
	_, err := tpset.Import("github.com/shabbyrobe/structer/testpkg/usesparseerr")
	if err != nil {
		t.Errorf("expected no error")
	}
	expected := []string{"github.com/shabbyrobe/structer/testpkg/usesparseerr.Test"}
	found := []string{}
	var obj types.Object
	for k, v := range tpset.Objects {
		found = append(found, k.String())
		obj = v
	}
	sort.Strings(found)
	if !reflect.DeepEqual(found, expected) {
		t.Errorf("types did not match expected")
	}

	field := (obj.Type().Underlying().(*types.Struct).Field(0))
	if field.Name() != "Foo" {
		t.Errorf("unexpected field")
	}
	if field.Type().(*types.Basic).Kind() != types.Invalid {
		t.Errorf("unexpected valid type")
	}
}
