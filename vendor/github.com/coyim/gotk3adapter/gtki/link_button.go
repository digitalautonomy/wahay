package gtki

type LinkButton interface {
	Bin

	GetUri() string
	SetUri(string)
}

func AssertLinkButton(_ LinkButton) {}
