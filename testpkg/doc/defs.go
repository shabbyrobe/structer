package doc

// TestString is a test string
type TestString string

const (
	// TestString1 is TestString1
	TestString1 TestString = "foo"

	// TestString2 is TestString2
	TestString2 TestString = "bar"
)

const (
	// TestString3 is TestString3
	TestString3 TestString = "baz"

	TestString4 TestString = "qux" // TestString4 is TestString4
	TestString5 TestString = "yep"

	// testString6 is unexported
	testString6 TestString = "nup"
)

// Group of types
type (
	// TestTypeGroup1 yep
	TestTypeGroup1 string

	// TestTypeGroup2 yep
	TestTypeGroup2 string
)

// TestStruct is a struct
type TestStruct struct {
	// A is a!
	A int

	// B is b!
	B int

	C int // C is c!
	D int // D is d!
	E int
}

// TestIntf is an interface
type TestIntf interface {
	Yep()
}

// TestInt is an int
type TestInt int
