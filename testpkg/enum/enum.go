package enum

type TestString string

const (
	TestString1 TestString = "foo"
	TestString2 TestString = "bar"
)

const (
	TestString3 TestString = "baz"
	TestString4 TestString = "qux"
	testString5 TestString = "nup"
)

type UsesEnumString struct {
	Enum TestString
}

type TestInt int

const (
	TestInt1 TestInt = 1
	TestInt2 TestInt = 2
)

type TestIntNest TestInt

const (
	TestIntNest1 TestIntNest = 1
	TestIntNest2 TestIntNest = 2
)

type TestIota int

const (
	TestIota1 TestIota = iota
	TestIota2
	TestIota3
)
