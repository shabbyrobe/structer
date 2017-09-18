package intfdecl1

type Test interface {
	IsTest()
}

type TestPrimitive string

func (t TestPrimitive) IsTest() {}

type TestStruct struct{}

func (t TestStruct) IsTest() {}

type TestStructPtr struct{}

func (t *TestStructPtr) IsTest() {}

type DoesntImplementTest struct{}
