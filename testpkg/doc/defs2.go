package doc

// Needed to split the definitions across multiple files to flush out a bug
// where only one file's AST was ever having comments mapped!

// TestStruct2 is also a struct
type TestStruct2 struct {
	// A is a!
	A int
	B int // B is b!
}
