package astparent

type Yep struct {
	Foo string
}

var yep string

type Thing interface {
	Foo()
}

func (y *Yep) Yep() {
	if 1 == 2 {
		panic("WAT")
	}
}
