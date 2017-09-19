package structer

import "testing"

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
		{false, "foo.bar", TypeName{Full: "foo.bar", PackagePath: "foo", Name: "bar", PackageName: "foo", isBuiltin: false}},
		{false, "yep/foo.bar", TypeName{Full: "yep/foo.bar", PackagePath: "yep/foo", Name: "bar", PackageName: "foo", isBuiltin: false}},

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

func TestImportName(t *testing.T) {
	// builtin should never show package prefix
	tn := NewBuiltinType("int")
	if tn.ImportName("", false) != "int" {
		t.Fatal()
	}
	if tn.ImportName("pants", false) != "int" {
		t.Fatal()
	}
	if tn.ImportName("int", false) != "int" {
		t.Fatal()
	}

	tn, _ = ParseTypeName("test/yep.Foo")
	if tn.ImportName("", false) != "yep.Foo" {
		t.Fatal()
	}
	if tn.ImportName("test", false) != "yep.Foo" {
		t.Fatal()
	}
	if tn.ImportName("test/yep", false) != "Foo" {
		t.Fatal()
	}
	if tn.ImportName("test/yep", true) != "Foo" {
		t.Fatal()
	}
	if tn.ImportName("yep", false) == "Foo" {
		t.Fatal()
	}
	if tn.ImportName("yep", true) != "Foo" {
		t.Fatal()
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
