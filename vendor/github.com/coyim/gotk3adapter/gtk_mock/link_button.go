package gtk_mock

type MockLinkButton struct {
	MockBin
}

func (*MockLinkButton) GetUri() string {
	return ""
}

func (*MockLinkButton) SetUri(uri string) {
}
