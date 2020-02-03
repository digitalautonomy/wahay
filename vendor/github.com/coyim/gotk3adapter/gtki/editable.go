package gtki

type Editable interface {
	SetEditable(bool)
	SetPosition(int)
}

func AssertEditable(_ Editable) {}
