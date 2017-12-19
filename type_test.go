package structer

import (
	"reflect"
	"testing"
)

func TestTypeName(t *testing.T) {
	var err error
	var tn TypeName

	var bad = []string{
		"foo",
		"yep/foo",
		"",
		"!",
		" ",
	}

	for _, n := range bad {
		tn, err = ParseTypeName(n)
		if err == nil {
			t.Fatalf("expected error, found %s", tn)
		}
	}

	var good = []struct {
		builtin bool
		in      string
		exp     TypeName
	}{
		{false, "foo.bar", TypeName{Full: "foo.bar", PackagePath: "foo", Name: "bar", isBuiltin: false}},
		{false, "yep/foo.bar", TypeName{Full: "yep/foo.bar", PackagePath: "yep/foo", Name: "bar", isBuiltin: false}},

		{true, "int", TypeName{Full: "int", Name: "int", isBuiltin: true}},
	}

	for _, n := range good {
		if !n.builtin {
			tn, err = ParseTypeName(n.in)
			if err != nil {
				t.Fatalf("unexpected error for %s: %v", n.in, err)
			}
		} else {
			tn = NewBuiltinType(n.in)
		}
		if tn != n.exp {
			t.Fatalf("expected %s, found %s", n.exp, tn)
		}
	}
}

func TestExported(t *testing.T) {
	tn := NewBuiltinType("int")
	if !tn.IsExported() {
		t.Fatal()
	}

	tn, err := ParseTypeName("test/yep.unexported")
	if tn.IsExported() || err != nil {
		t.Fatal()
	}

	tn, err = ParseTypeName("test/yep.Exported")
	if !tn.IsExported() || err != nil {
		t.Fatal()
	}
}

func TestTypeNamesSort(t *testing.T) {
	te := TypeNames{NewTypeName("a", "A"), NewTypeName("a", "B"), NewTypeName("z", "B")}
	ts := TypeNames{NewTypeName("z", "B"), NewTypeName("a", "B"), NewTypeName("a", "A")}
	ts.Sort()
	if !reflect.DeepEqual(te, ts) {
		t.Fatal()
	}

	ts = TypeNames{NewTypeName("z", "B"), NewTypeName("a", "A"), NewTypeName("a", "B")}
	tr := ts.Sorted()
	if reflect.DeepEqual(te, ts) {
		t.Fatal()
	}
	if !reflect.DeepEqual(te, tr) {
		t.Fatal()
	}
}

func TestTypeMapSort(t *testing.T) {
	te := TypeNames{NewTypeName("a", "A"), NewTypeName("a", "B"), NewTypeName("z", "B")}
	tm := TypeMap{
		NewTypeName("z", "B"): nil,
		NewTypeName("a", "B"): nil,
		NewTypeName("a", "A"): nil,
	}
	ns := tm.SortedKeys()
	if !reflect.DeepEqual(te, ns) {
		t.Fatal()
	}
}
