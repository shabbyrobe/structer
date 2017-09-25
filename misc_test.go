package structer

import "testing"

func TestActuallyImplements(t *testing.T) {
	tpset := NewTypePackageSet()
	ifaceTyp := tpset.MustFindImportObjectByName("github.com/shabbyrobe/structer/testpkg/intfdecl1.Test").Type()
	ifaceUtyp := tpset.MustFindImportObjectByName("github.com/shabbyrobe/structer/testpkg/intfdecl1.Test").Type().Underlying()

	nonImplTyp := tpset.MustFindImportObjectByName("github.com/shabbyrobe/structer/testpkg/intfdecl1.DoesntImplementTest").Type()
	if ActuallyImplements(nonImplTyp, ifaceTyp) {
		t.Errorf("type %s wrongly reported implementing %s", nonImplTyp.String(), ifaceTyp.String())
	}

	typs := []string{
		"github.com/shabbyrobe/structer/testpkg/intfdecl1.TestStruct",
		"github.com/shabbyrobe/structer/testpkg/intfdecl1.TestStructPtr",
		"github.com/shabbyrobe/structer/testpkg/intfdecl1.TestPrimitive",
		"github.com/shabbyrobe/structer/testpkg/intfdecl2.TestStruct",
		"github.com/shabbyrobe/structer/testpkg/intfdecl2.TestStructPtr",
		"github.com/shabbyrobe/structer/testpkg/intfdecl2.TestPrimitive",
	}
	for _, s := range typs {
		implTyp := tpset.MustFindImportObjectByName(s).Type()
		if !ActuallyImplements(implTyp, ifaceTyp) {
			t.Errorf("type %s wrongly reported not implementing %s", implTyp.String(), ifaceTyp.String())
		}
		if !ActuallyImplements(implTyp, ifaceUtyp) {
			t.Errorf("type %s wrongly reported not implementing %s", implTyp.String(), ifaceUtyp.String())
		}
	}
}
