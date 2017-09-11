package structer

import (
	"go/constant"
	"testing"
)

func assertEnumStrings(t *testing.T, expected map[string]string, enum *Enum) {
	t.Helper()
	if len(expected) != len(enum.Values) {
		t.Errorf("expected length %d, found %d", len(expected), len(enum.Values))
	}
	for k, v := range expected {
		if constant.StringVal(enum.Values[k]) != v {
			t.Errorf("expected key %s to be %s, found %s", k, v, constant.StringVal(enum.Values[k]))
		}
	}
}

func assertEnumInts(t *testing.T, expected map[string]int64, enum *Enum) {
	t.Helper()
	if len(expected) != len(enum.Values) {
		t.Errorf("expected length %d, found %d", len(expected), len(enum.Values))
	}
	for k, v := range expected {
		i, _ := constant.Int64Val(enum.Values[k])
		if i != v {
			t.Errorf("expected key %s to be %d, found %d", k, v, i)
		}
	}
}

func extractEnum(t *testing.T, tn TypeName, unexported bool) *Enum {
	t.Helper()
	tpset := NewTypePackageSet()
	_, err := tpset.Import(tn.PackagePath)
	if err != nil {
		t.Errorf("expected no error, found %v", err)
	}
	enum, err := tpset.ExtractEnum(tn, unexported)
	if err != nil {
		t.Fatalf("expected no error, found %v", err)
	}
	return enum
}

func TestExtractEnumString(t *testing.T) {
	tn := NewTypeName("github.com/shabbyrobe/structer/testpkg/enum", "TestString")
	enum := extractEnum(t, tn, false)
	if enum.Underlying.String() != "string" {
		t.Errorf("underlying type should be int")
	}
	assertEnumStrings(t, map[string]string{"TestString1": "foo", "TestString2": "bar", "TestString3": "baz", "TestString4": "qux"}, enum)
}

func TestExtractEnumStringUnexported(t *testing.T) {
	tn := NewTypeName("github.com/shabbyrobe/structer/testpkg/enum", "TestString")
	enum := extractEnum(t, tn, true)
	if enum.Underlying.String() != "string" {
		t.Errorf("underlying type should be int")
	}
	assertEnumStrings(t, map[string]string{
		"TestString1": "foo", "TestString2": "bar", "TestString3": "baz", "TestString4": "qux",
		"testString5": "nup",
	}, enum)
}

func TestExtractEnumInt(t *testing.T) {
	tn := NewTypeName("github.com/shabbyrobe/structer/testpkg/enum", "TestInt")
	enum := extractEnum(t, tn, false)
	if enum.Underlying.String() != "int" {
		t.Errorf("underlying type should be int")
	}
	assertEnumInts(t, map[string]int64{"TestInt1": 1, "TestInt2": 2}, enum)
}

func TestExtractEnumIntNested(t *testing.T) {
	tn := NewTypeName("github.com/shabbyrobe/structer/testpkg/enum", "TestIntNest")
	enum := extractEnum(t, tn, false)
	if enum.Underlying.String() != "int" {
		t.Errorf("underlying type should be int")
	}
	assertEnumInts(t, map[string]int64{"TestIntNest1": 1, "TestIntNest2": 2}, enum)
}

func TestExtractEnumIota(t *testing.T) {
	tn := NewTypeName("github.com/shabbyrobe/structer/testpkg/enum", "TestIota")
	enum := extractEnum(t, tn, false)
	if enum.Underlying.String() != "int" {
		t.Errorf("underlying type should be int")
	}
	assertEnumInts(t, map[string]int64{"TestIota1": 0, "TestIota2": 1, "TestIota3": 2}, enum)
}
