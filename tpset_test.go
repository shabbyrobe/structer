package structer

import (
	"go/types"
	"reflect"
	"sort"
	"testing"
)

func TestCaptureErrors(t *testing.T) {
	var errs []error
	tpset := NewTypePackageSet(CaptureErrors(func(e error) {
		errs = append(errs, e)
	}))
	_, err := tpset.Import("github.com/shabbyrobe/structer/testpkg/intferr")
	if err != nil {
		t.Errorf("expected no error")
	}
	if len(errs) != 1 {
		t.Errorf("expected 1 captured error, found %d", len(errs))
	}
}

func TestFailOnHardTypeError(t *testing.T) {
	var errs []error
	tpset := NewTypePackageSet(CaptureErrors(func(e error) {
		errs = append(errs, e)
	}))
	tpset.AllowHardTypesError = false
	_, err := tpset.Import("github.com/shabbyrobe/structer/testpkg/intferr")
	if err == nil {
		t.Errorf("expected error")
	}
}

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

// If a package uses two packages, one of which has a parse error, ensure only
// those types are invalid.
//
func TestTypePackageSetUsesPackageWithParseError(t *testing.T) {
	tpset := NewTypePackageSet()
	_, err := tpset.Import("github.com/shabbyrobe/structer/testpkg/usesparseerr")
	if err != nil {
		t.Errorf("expected no error")
	}
	expected := []string{
		"github.com/shabbyrobe/structer/testpkg/usesparseerr.Test",
	}
	found := []string{}
	var obj types.Object
	for k, v := range tpset.Objects {
		found = append(found, k.String())
		obj = v
	}
	sort.Strings(found)
	if !reflect.DeepEqual(found, expected) {
		t.Errorf("types did not match expected, %v %v", found, expected)
	}

	fields := indexFields(obj.Type().Underlying().(*types.Struct))
	if !fields.isInvalid("Foo") {
		t.Errorf("unexpected valid type")
	}
}

type fieldIndex struct {
	fields map[string]*types.Var
}

func (f *fieldIndex) isInvalid(name string) bool {
	field := f.fields[name]
	if field == nil {
		return false
	}
	t, ok := field.Type().(*types.Basic)
	if !ok {
		return false
	}
	return t.Kind() == types.Invalid
}

func indexFields(stct *types.Struct) fieldIndex {
	m := fieldIndex{fields: make(map[string]*types.Var)}
	for i := 0; i < stct.NumFields(); i++ {
		m.fields[stct.Field(i).Name()] = stct.Field(i)
	}
	return m
}
