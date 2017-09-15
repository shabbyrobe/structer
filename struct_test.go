package structer

import (
	"fmt"
	"go/types"
	"reflect"
	"testing"
)

type TestingStackStruct struct {
	Foo []map[[3]int]struct {
		Bar []map[[3]int]struct {
			Baz string
			Qux *TestingStackStruct
		}
	}
	Bar string
}

func TestStructStack(t *testing.T) {
	parents := []string{}
	vis := &PartialTypeVisitor{
		VisitBasicFunc: func(ctx WalkContext, t *types.Basic) error {
			parents = append(parents, ctx.Parent().String())
			return nil
		},
		VisitNamedFunc: func(ctx WalkContext, t *types.Named) error {
			parents = append(parents, ctx.Parent().String())
			return nil
		},
	}
	_, tn, tst := getTestingStruct(t, "github.com/shabbyrobe/structer.TestingStackStruct")
	Walk(tn, tst, vis)

	expected := []string{
		"[3]int",
		"[3]int",
		"struct{Baz string; Qux *github.com/shabbyrobe/structer.TestingStackStruct}",
		"*github.com/shabbyrobe/structer.TestingStackStruct",
		"struct{Foo []map[[3]int]struct{Bar []map[[3]int]struct{Baz string; Qux *github.com/shabbyrobe/structer.TestingStackStruct}}; Bar string}",
	}
	if !reflect.DeepEqual(expected, parents) {
		t.Fail()
	}
}

func getTestingStruct(t *testing.T, tns string) (tpset *TypePackageSet, tn TypeName, typ types.Type) {
	tpset = NewTypePackageSet()
	tpset.Config.IncludeTests = true
	_, err := tpset.Import("github.com/shabbyrobe/structer")
	if err != nil {
		t.Fail()
		return
	}
	tn, _ = ParseTypeName(tns)
	ts := tpset.Objects[tn]
	if ts == nil {
		t.Fail()
		return
	}
	typ = ts.Type().Underlying().(*types.Struct)
	return
}

// TestStruct demonstrates all the different type combinations
// I could think of before I got sick of writing vim macros to
// generate test results.
//
type TestingStruct struct {
	Basic        string
	BasicPointer *string
	BasicSlice   []int
	BasicArray   [2]int

	BasicMap                     map[string]string
	BasicMapOfBasicMap           map[string]map[string]string
	BasicMapOfBasicMapOfBasicMap map[string]map[string]map[string]string

	MapWithArrayKey        map[[2]int]string
	MapWithStructKey       map[struct{ X, Y int }]string
	MapWithNestedStructKey map[struct{ Foo struct{ Bar string } }]string

	Circular *TestingStruct

	// Compound: array and slice
	SliceOfSlice        [][]int
	SliceOfArray        [][2]int
	ArrayOfSlice        [2][]int
	SliceOfSliceOfSlice [][][]int
	SliceOfSliceOfArray [][][2]int

	// Nested structs
	NestedBasic       struct{ Foo string }
	NestedNestedBasic struct{ Foo struct{ Bar string } }
}

var fieldResults = map[string][]TestingVisitorEvent{
	"Basic": []TestingVisitorEvent{{Depth: 2, Kind: "VisitBasic", Name: "string"}},
	"Circular": []TestingVisitorEvent{
		{Depth: 2, Kind: "EnterPointer", Name: "*github.com/shabbyrobe/structer.TestingStruct"},
		{Depth: 3, Kind: "VisitNamed", Name: "github.com/shabbyrobe/structer.TestingStruct"},
		{Depth: 2, Kind: "LeavePointer", Name: "*github.com/shabbyrobe/structer.TestingStruct"},
	},

	"BasicPointer": []TestingVisitorEvent{
		{Depth: 2, Kind: "EnterPointer", Name: "*string"},
		{Depth: 3, Kind: "VisitBasic", Name: "string"},
		{Depth: 2, Kind: "LeavePointer", Name: "*string"},
	},
	"BasicSlice": []TestingVisitorEvent{
		{Depth: 2, Kind: "EnterSlice", Name: "[]int"},
		{Depth: 3, Kind: "VisitBasic", Name: "int"},
		{Depth: 2, Kind: "LeaveSlice", Name: "[]int"},
	},

	"BasicArray": []TestingVisitorEvent{
		{Depth: 2, Kind: "EnterArray", Name: "[2]int"},
		{Depth: 3, Kind: "VisitBasic", Name: "int"},
		{Depth: 2, Kind: "LeaveArray", Name: "[2]int"},
	},

	"BasicMap": []TestingVisitorEvent{
		{Depth: 2, Kind: "EnterMapKey", Name: "map[string]string"},
		{Depth: 3, Kind: "VisitBasic", Name: "string"},
		{Depth: 2, Kind: "LeaveMapKey", Name: "map[string]string"},
		{Depth: 2, Kind: "EnterMapElem", Name: "map[string]string"},
		{Depth: 3, Kind: "VisitBasic", Name: "string"},
		{Depth: 2, Kind: "LeaveMapElem", Name: "map[string]string"},
	},

	"BasicMapOfBasicMap": []TestingVisitorEvent{
		{Depth: 2, Kind: "EnterMapKey", Name: "map[string]map[string]string"},
		{Depth: 3, Kind: "VisitBasic", Name: "string"},
		{Depth: 2, Kind: "LeaveMapKey", Name: "map[string]map[string]string"},
		{Depth: 2, Kind: "EnterMapElem", Name: "map[string]map[string]string"},
		{Depth: 3, Kind: "EnterMapKey", Name: "map[string]string"},
		{Depth: 4, Kind: "VisitBasic", Name: "string"},
		{Depth: 3, Kind: "LeaveMapKey", Name: "map[string]string"},
		{Depth: 3, Kind: "EnterMapElem", Name: "map[string]string"},
		{Depth: 4, Kind: "VisitBasic", Name: "string"},
		{Depth: 3, Kind: "LeaveMapElem", Name: "map[string]string"},
		{Depth: 2, Kind: "LeaveMapElem", Name: "map[string]map[string]string"},
	},

	"BasicMapOfBasicMapOfBasicMap": []TestingVisitorEvent{
		{Depth: 2, Kind: "EnterMapKey", Name: "map[string]map[string]map[string]string"},
		{Depth: 3, Kind: "VisitBasic", Name: "string"},
		{Depth: 2, Kind: "LeaveMapKey", Name: "map[string]map[string]map[string]string"},
		{Depth: 2, Kind: "EnterMapElem", Name: "map[string]map[string]map[string]string"},
		{Depth: 3, Kind: "EnterMapKey", Name: "map[string]map[string]string"},
		{Depth: 4, Kind: "VisitBasic", Name: "string"},
		{Depth: 3, Kind: "LeaveMapKey", Name: "map[string]map[string]string"},
		{Depth: 3, Kind: "EnterMapElem", Name: "map[string]map[string]string"},
		{Depth: 4, Kind: "EnterMapKey", Name: "map[string]string"},
		{Depth: 5, Kind: "VisitBasic", Name: "string"},
		{Depth: 4, Kind: "LeaveMapKey", Name: "map[string]string"},
		{Depth: 4, Kind: "EnterMapElem", Name: "map[string]string"},
		{Depth: 5, Kind: "VisitBasic", Name: "string"},
		{Depth: 4, Kind: "LeaveMapElem", Name: "map[string]string"},
		{Depth: 3, Kind: "LeaveMapElem", Name: "map[string]map[string]string"},
		{Depth: 2, Kind: "LeaveMapElem", Name: "map[string]map[string]map[string]string"},
	},

	"SliceOfSlice": []TestingVisitorEvent{
		{Depth: 2, Kind: "EnterSlice", Name: "[][]int"},
		{Depth: 3, Kind: "EnterSlice", Name: "[]int"},
		{Depth: 4, Kind: "VisitBasic", Name: "int"},
		{Depth: 3, Kind: "LeaveSlice", Name: "[]int"},
		{Depth: 2, Kind: "LeaveSlice", Name: "[][]int"},
	},

	"SliceOfArray": []TestingVisitorEvent{
		{Depth: 2, Kind: "EnterSlice", Name: "[][2]int"},
		{Depth: 3, Kind: "EnterArray", Name: "[2]int"},
		{Depth: 4, Kind: "VisitBasic", Name: "int"},
		{Depth: 3, Kind: "LeaveArray", Name: "[2]int"},
		{Depth: 2, Kind: "LeaveSlice", Name: "[][2]int"},
	},

	"ArrayOfSlice": []TestingVisitorEvent{
		{Depth: 2, Kind: "EnterArray", Name: "[2][]int"},
		{Depth: 3, Kind: "EnterSlice", Name: "[]int"},
		{Depth: 4, Kind: "VisitBasic", Name: "int"},
		{Depth: 3, Kind: "LeaveSlice", Name: "[]int"},
		{Depth: 2, Kind: "LeaveArray", Name: "[2][]int"},
	},

	"SliceOfSliceOfSlice": []TestingVisitorEvent{
		{Depth: 2, Kind: "EnterSlice", Name: "[][][]int"},
		{Depth: 3, Kind: "EnterSlice", Name: "[][]int"},
		{Depth: 4, Kind: "EnterSlice", Name: "[]int"},
		{Depth: 5, Kind: "VisitBasic", Name: "int"},
		{Depth: 4, Kind: "LeaveSlice", Name: "[]int"},
		{Depth: 3, Kind: "LeaveSlice", Name: "[][]int"},
		{Depth: 2, Kind: "LeaveSlice", Name: "[][][]int"},
	},

	"SliceOfSliceOfArray": []TestingVisitorEvent{
		{Depth: 2, Kind: "EnterSlice", Name: "[][][2]int"},
		{Depth: 3, Kind: "EnterSlice", Name: "[][2]int"},
		{Depth: 4, Kind: "EnterArray", Name: "[2]int"},
		{Depth: 5, Kind: "VisitBasic", Name: "int"},
		{Depth: 4, Kind: "LeaveArray", Name: "[2]int"},
		{Depth: 3, Kind: "LeaveSlice", Name: "[][2]int"},
		{Depth: 2, Kind: "LeaveSlice", Name: "[][][2]int"},
	},

	"NestedBasic": []TestingVisitorEvent{
		{Depth: 2, Kind: "EnterStruct", Name: "NestedBasic"},
		{Depth: 3, Kind: "EnterField", Name: "Foo"},
		{Depth: 4, Kind: "VisitBasic", Name: "string"},
		{Depth: 3, Kind: "LeaveField", Name: "Foo"},
		{Depth: 2, Kind: "LeaveStruct", Name: "NestedBasic"},
	},

	"NestedNestedBasic": []TestingVisitorEvent{
		{Depth: 2, Kind: "EnterStruct", Name: "NestedNestedBasic"},
		{Depth: 3, Kind: "EnterField", Name: "Foo"},
		{Depth: 4, Kind: "EnterStruct", Name: "Foo"},
		{Depth: 5, Kind: "EnterField", Name: "Bar"},
		{Depth: 6, Kind: "VisitBasic", Name: "string"},
		{Depth: 5, Kind: "LeaveField", Name: "Bar"},
		{Depth: 4, Kind: "LeaveStruct", Name: "Foo"},
		{Depth: 3, Kind: "LeaveField", Name: "Foo"},
		{Depth: 2, Kind: "LeaveStruct", Name: "NestedNestedBasic"},
	},

	"MapWithArrayKey": []TestingVisitorEvent{
		{Depth: 2, Kind: "EnterMapKey", Name: "map[[2]int]string"},
		{Depth: 3, Kind: "EnterArray", Name: "[2]int"},
		{Depth: 4, Kind: "VisitBasic", Name: "int"},
		{Depth: 3, Kind: "LeaveArray", Name: "[2]int"},
		{Depth: 2, Kind: "LeaveMapKey", Name: "map[[2]int]string"},
		{Depth: 2, Kind: "EnterMapElem", Name: "map[[2]int]string"},
		{Depth: 3, Kind: "VisitBasic", Name: "string"},
		{Depth: 2, Kind: "LeaveMapElem", Name: "map[[2]int]string"},
	},

	"MapWithStructKey": []TestingVisitorEvent{ //
		{Depth: 2, Kind: "EnterMapKey", Name: "map[struct{X int; Y int}]string"},
		{Depth: 3, Kind: "EnterStruct", Name: "struct{X int; Y int}"},
		{Depth: 4, Kind: "EnterField", Name: "X"},
		{Depth: 5, Kind: "VisitBasic", Name: "int"},
		{Depth: 4, Kind: "LeaveField", Name: "X"},
		{Depth: 4, Kind: "EnterField", Name: "Y"},
		{Depth: 5, Kind: "VisitBasic", Name: "int"},
		{Depth: 4, Kind: "LeaveField", Name: "Y"},
		{Depth: 3, Kind: "LeaveStruct", Name: "struct{X int; Y int}"},
		{Depth: 2, Kind: "LeaveMapKey", Name: "map[struct{X int; Y int}]string"},
		{Depth: 2, Kind: "EnterMapElem", Name: "map[struct{X int; Y int}]string"},
		{Depth: 3, Kind: "VisitBasic", Name: "string"},
		{Depth: 2, Kind: "LeaveMapElem", Name: "map[struct{X int; Y int}]string"},
	},

	"MapWithNestedStructKey": []TestingVisitorEvent{ //map[struct{ Foo struct{ Bar string } }]string
		{Depth: 2, Kind: "EnterMapKey", Name: "map[struct{Foo struct{Bar string}}]string"},
		{Depth: 3, Kind: "EnterStruct", Name: "struct{Foo struct{Bar string}}"},
		{Depth: 4, Kind: "EnterField", Name: "Foo"},
		{Depth: 5, Kind: "EnterStruct", Name: "Foo"},
		{Depth: 6, Kind: "EnterField", Name: "Bar"},
		{Depth: 7, Kind: "VisitBasic", Name: "string"},
		{Depth: 6, Kind: "LeaveField", Name: "Bar"},
		{Depth: 5, Kind: "LeaveStruct", Name: "Foo"},
		{Depth: 4, Kind: "LeaveField", Name: "Foo"},
		{Depth: 3, Kind: "LeaveStruct", Name: "struct{Foo struct{Bar string}}"},
		{Depth: 2, Kind: "LeaveMapKey", Name: "map[struct{Foo struct{Bar string}}]string"},
		{Depth: 2, Kind: "EnterMapElem", Name: "map[struct{Foo struct{Bar string}}]string"},
		{Depth: 3, Kind: "VisitBasic", Name: "string"},
		{Depth: 2, Kind: "LeaveMapElem", Name: "map[struct{Foo struct{Bar string}}]string"},
	},
}

func TestStruct(t *testing.T) {
	tpset := NewTypePackageSet()
	tpset.Config.IncludeTests = true
	_, err := tpset.Import("github.com/shabbyrobe/structer")
	if err != nil {
		t.Fail()
		return
	}

	// How nice of you to leave this lying around!
	tn, _ := ParseTypeName("github.com/shabbyrobe/structer.TestingStruct")
	ts := tpset.Objects[tn]
	if ts == nil {
		t.Fail()
		return
	}

	tst := ts.Type().Underlying().(*types.Struct)

	vis := NewTestingVisitor()
	Walk(tn, tst, vis)

	for i := 0; i < tst.NumFields(); i++ {
		name := tst.Field(i).Name()
		exp, ok := fieldResults[name]
		if !ok {
			t.Fatalf("result not found for %s", name)
		}
		check := vis.FieldEvents[name]
		if !reflect.DeepEqual(exp, check) {
			dumpEvents(check)
			dumpEvents(exp)
			t.Fatalf("test %s failed", name)
		}
	}
}

func dumpEvents(es []TestingVisitorEvent) {
	for _, e := range es {
		fmt.Printf("%-02d %-20s %s\n", e.Depth, e.Kind, e.Name)
	}
	fmt.Println()
}

type TestingVisitorEvent struct {
	Kind  string
	Name  string
	Depth int
}

func NewTestingVisitor() *TestingVisitor {
	return &TestingVisitor{
		FieldEvents: make(map[string][]TestingVisitorEvent),
	}
}

type TestingVisitor struct {
	Depth       int
	FieldEvents map[string][]TestingVisitorEvent
	Events      []TestingVisitorEvent
}

func (tv *TestingVisitor) EnterStruct(ctx WalkContext, s StructInfo) error {
	if tv.Depth > 0 {
		tv.Events = append(tv.Events, TestingVisitorEvent{Kind: "EnterStruct", Name: s.Name, Depth: tv.Depth})
	}
	tv.Depth++
	return nil
}
func (tv *TestingVisitor) LeaveStruct(ctx WalkContext, s StructInfo) error {
	tv.Depth--
	if tv.Depth > 0 {
		tv.Events = append(tv.Events, TestingVisitorEvent{Kind: "LeaveStruct", Name: s.Name, Depth: tv.Depth})
	}
	return nil
}

func (tv *TestingVisitor) EnterField(ctx WalkContext, s StructInfo, field *types.Var, tag string) error {
	if tv.Depth > 1 {
		tv.Events = append(tv.Events, TestingVisitorEvent{Kind: "EnterField", Name: field.Name(), Depth: tv.Depth})
	}
	tv.Depth++
	return nil
}
func (tv *TestingVisitor) LeaveField(ctx WalkContext, s StructInfo, field *types.Var, tag string) error {
	tv.Depth--
	if tv.Depth > 1 {
		tv.Events = append(tv.Events, TestingVisitorEvent{Kind: "LeaveField", Name: field.Name(), Depth: tv.Depth})
	} else {
		tv.FieldEvents[field.Name()] = tv.Events
		tv.Events = []TestingVisitorEvent{}
	}
	return nil
}

func (tv *TestingVisitor) EnterMapKey(ctx WalkContext, ft *types.Map, key types.Type) error {
	tv.Events = append(tv.Events, TestingVisitorEvent{Kind: "EnterMapKey", Name: ft.String(), Depth: tv.Depth})
	tv.Depth++
	return nil
}
func (tv *TestingVisitor) LeaveMapKey(ctx WalkContext, ft *types.Map, key types.Type) error {
	tv.Depth--
	tv.Events = append(tv.Events, TestingVisitorEvent{Kind: "LeaveMapKey", Name: ft.String(), Depth: tv.Depth})
	return nil
}

func (tv *TestingVisitor) EnterMapElem(ctx WalkContext, ft *types.Map, elem types.Type) error {
	tv.Events = append(tv.Events, TestingVisitorEvent{Kind: "EnterMapElem", Name: ft.String(), Depth: tv.Depth})
	tv.Depth++
	return nil
}
func (tv *TestingVisitor) LeaveMapElem(ctx WalkContext, ft *types.Map, elem types.Type) error {
	tv.Depth--
	tv.Events = append(tv.Events, TestingVisitorEvent{Kind: "LeaveMapElem", Name: ft.String(), Depth: tv.Depth})
	return nil
}

func (tv *TestingVisitor) EnterSlice(ctx WalkContext, ft *types.Slice) error {
	tv.Events = append(tv.Events, TestingVisitorEvent{Kind: "EnterSlice", Name: ft.String(), Depth: tv.Depth})
	tv.Depth++
	return nil
}
func (tv *TestingVisitor) LeaveSlice(ctx WalkContext, ft *types.Slice) error {
	tv.Depth--
	tv.Events = append(tv.Events, TestingVisitorEvent{Kind: "LeaveSlice", Name: ft.String(), Depth: tv.Depth})
	return nil
}

func (tv *TestingVisitor) EnterPointer(ctx WalkContext, ft *types.Pointer) error {
	tv.Events = append(tv.Events, TestingVisitorEvent{Kind: "EnterPointer", Name: ft.String(), Depth: tv.Depth})
	tv.Depth++
	return nil
}
func (tv *TestingVisitor) LeavePointer(ctx WalkContext, ft *types.Pointer) error {
	tv.Depth--
	tv.Events = append(tv.Events, TestingVisitorEvent{Kind: "LeavePointer", Name: ft.String(), Depth: tv.Depth})
	return nil
}

func (tv *TestingVisitor) EnterArray(ctx WalkContext, ft *types.Array) error {
	tv.Events = append(tv.Events, TestingVisitorEvent{Kind: "EnterArray", Name: ft.String(), Depth: tv.Depth})
	tv.Depth++
	return nil
}
func (tv *TestingVisitor) LeaveArray(ctx WalkContext, ft *types.Array) error {
	tv.Depth--
	tv.Events = append(tv.Events, TestingVisitorEvent{Kind: "LeaveArray", Name: ft.String(), Depth: tv.Depth})
	return nil
}

func (tv *TestingVisitor) VisitBasic(ctx WalkContext, t *types.Basic) error {
	tv.Events = append(tv.Events, TestingVisitorEvent{Kind: "VisitBasic", Name: t.String(), Depth: tv.Depth})
	return nil
}
func (tv *TestingVisitor) VisitNamed(ctx WalkContext, t *types.Named) error {
	tv.Events = append(tv.Events, TestingVisitorEvent{Kind: "VisitNamed", Name: t.String(), Depth: tv.Depth})
	return nil
}
