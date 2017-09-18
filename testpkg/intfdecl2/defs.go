package intfdecl2

type Test2 interface {
	IsTest2()
}

type TestPrimitive string

func (t TestPrimitive) IsTest()  {}
func (t TestPrimitive) IsTest2() {}

type TestStruct struct{}

func (t TestStruct) IsTest()  {}
func (t TestStruct) IsTest2() {}

type TestStructPtr struct{}

func (t *TestStructPtr) IsTest()  {}
func (t *TestStructPtr) IsTest2() {}

type DoesntImplementTest struct{}
