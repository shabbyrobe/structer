package intferr

var _ Pantsable = &Pants{}

type Pantsable interface {
	Foo()
}

type Pants struct{}
