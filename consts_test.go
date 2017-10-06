package structer

import (
	"go/constant"
	"sort"
	"testing"
)

func assertConstsStrings(t *testing.T, expected []*ConstValue, consts *Consts) {
	t.Helper()
	if len(expected) != len(consts.Values) {
		t.Errorf("expected length %d, found %d", len(expected), len(consts.Values))
	}
	sort.Slice(expected, func(i, j int) bool {
		return expected[i].Name.IsBefore(expected[j].Name)
	})
	for i, v := range consts.SortedValues() {
		check := expected[i]
		if check.Name != v.Name {
			t.Errorf("expected key %s to be %s, found %s", check.Name, constant.StringVal(v.Value), constant.StringVal(check.Value))
		}
	}
}

func assertConstsInts(t *testing.T, expected []*ConstValue, consts *Consts) {
	t.Helper()
	if len(expected) != len(consts.Values) {
		t.Errorf("expected length %d, found %d", len(expected), len(consts.Values))
	}
	sort.Slice(expected, func(i, j int) bool {
		return expected[i].Name.IsBefore(expected[j].Name)
	})
	for i, v := range consts.SortedValues() {
		check := expected[i]
		if check.Name != v.Name {
			t.Errorf("expected key %s to be %s, found %s", check.Name, v.Name, constant.StringVal(check.Value))
		}
		i, _ := constant.Int64Val(check.Value)
		ci, _ := constant.Int64Val(v.Value)
		if i != ci {
			t.Errorf("expected key %s to be %d, found %d", check.Name, ci, i)
		}
	}
}

func extractConsts(t *testing.T, tn TypeName, unexported bool) *Consts {
	t.Helper()
	tpset := NewTypePackageSet()
	_, err := tpset.Import(tn.PackagePath)
	if err != nil {
		t.Errorf("expected no error, found %v", err)
	}
	consts, err := tpset.ExtractConsts(tn, unexported)
	if err != nil {
		t.Fatalf("expected no error, found %v", err)
	}
	return consts
}

func TestExtractConstsString(t *testing.T) {
	tn := NewTypeName("github.com/shabbyrobe/structer/testpkg/consts", "TestString")
	consts := extractConsts(t, tn, false)
	if consts.Underlying.String() != "string" {
		t.Errorf("underlying type should be string")
	}
	if consts.IsEnum {
		t.Errorf("should not be enum")
	}
	assertConstsStrings(t, []*ConstValue{
		{Name: NewTypeName("github.com/shabbyrobe/structer/testpkg/consts", "TestString1"), Value: constant.MakeString("foo")},
		{Name: NewTypeName("github.com/shabbyrobe/structer/testpkg/consts", "TestString2"), Value: constant.MakeString("bar")},
		{Name: NewTypeName("github.com/shabbyrobe/structer/testpkg/consts", "TestString3"), Value: constant.MakeString("baz")},
		{Name: NewTypeName("github.com/shabbyrobe/structer/testpkg/consts", "TestString4"), Value: constant.MakeString("quz")},
	}, consts)
}

func TestExtractConstsStringUnexported(t *testing.T) {
	tn := NewTypeName("github.com/shabbyrobe/structer/testpkg/consts", "TestString")
	consts := extractConsts(t, tn, true)
	if consts.Underlying.String() != "string" {
		t.Errorf("underlying type should be string")
	}
	if consts.IsEnum {
		t.Errorf("should not be enum")
	}
	assertConstsStrings(t, []*ConstValue{
		{Name: NewTypeName("github.com/shabbyrobe/structer/testpkg/consts", "TestString1"), Value: constant.MakeString("foo")},
		{Name: NewTypeName("github.com/shabbyrobe/structer/testpkg/consts", "TestString2"), Value: constant.MakeString("bar")},
		{Name: NewTypeName("github.com/shabbyrobe/structer/testpkg/consts", "TestString3"), Value: constant.MakeString("baz")},
		{Name: NewTypeName("github.com/shabbyrobe/structer/testpkg/consts", "TestString4"), Value: constant.MakeString("quz")},
		{Name: NewTypeName("github.com/shabbyrobe/structer/testpkg/consts", "testString5"), Value: constant.MakeString("nup")},
	}, consts)
}

func TestExtractConstsInt(t *testing.T) {
	tn := NewTypeName("github.com/shabbyrobe/structer/testpkg/consts", "TestInt")
	consts := extractConsts(t, tn, false)
	if consts.Underlying.String() != "int" {
		t.Errorf("underlying type should be int")
	}
	if consts.IsEnum {
		t.Errorf("should not be enum")
	}
	assertConstsInts(t, []*ConstValue{
		{Name: NewTypeName("github.com/shabbyrobe/structer/testpkg/consts", "TestInt1"), Value: constant.MakeInt64(1)},
		{Name: NewTypeName("github.com/shabbyrobe/structer/testpkg/consts", "TestInt2"), Value: constant.MakeInt64(2)},
	}, consts)
}

func TestExtractConstsIntNested(t *testing.T) {
	tn := NewTypeName("github.com/shabbyrobe/structer/testpkg/consts", "TestIntNest")
	consts := extractConsts(t, tn, false)
	if consts.Underlying.String() != "int" {
		t.Errorf("underlying type should be int")
	}
	if consts.IsEnum {
		t.Errorf("should not be enum")
	}
	assertConstsInts(t, []*ConstValue{
		{Name: NewTypeName("github.com/shabbyrobe/structer/testpkg/consts", "TestIntNest1"), Value: constant.MakeInt64(1)},
		{Name: NewTypeName("github.com/shabbyrobe/structer/testpkg/consts", "TestIntNest2"), Value: constant.MakeInt64(2)},
	}, consts)
}

func TestExtractConstsIota(t *testing.T) {
	tn := NewTypeName("github.com/shabbyrobe/structer/testpkg/consts", "TestIota")
	consts := extractConsts(t, tn, false)
	if consts.Underlying.String() != "int" {
		t.Errorf("underlying type should be int")
	}
	if consts.IsEnum {
		t.Errorf("should not be enum")
	}
	assertConstsInts(t, []*ConstValue{
		{Name: NewTypeName("github.com/shabbyrobe/structer/testpkg/consts", "TestIota1"), Value: constant.MakeInt64(0)},
		{Name: NewTypeName("github.com/shabbyrobe/structer/testpkg/consts", "TestIota2"), Value: constant.MakeInt64(1)},
		{Name: NewTypeName("github.com/shabbyrobe/structer/testpkg/consts", "TestIota3"), Value: constant.MakeInt64(2)},
	}, consts)
}

func TestExtractConstsEnum(t *testing.T) {
	tn := NewTypeName("github.com/shabbyrobe/structer/testpkg/consts", "TestEnum")
	consts := extractConsts(t, tn, false)
	if consts.Underlying.String() != "string" {
		t.Errorf("underlying type should be string")
	}
	if !consts.IsEnum {
		t.Errorf("should be enum")
	}
	assertConstsStrings(t, []*ConstValue{
		{Name: NewTypeName("github.com/shabbyrobe/structer/testpkg/consts", "TestEnum1"), Value: constant.MakeString("foo")},
		{Name: NewTypeName("github.com/shabbyrobe/structer/testpkg/consts", "TestEnum2"), Value: constant.MakeString("bar")},
	}, consts)
}
