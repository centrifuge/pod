package token

type JW3THeader struct {
	Algorithm   string `json:"algorithm"`
	AddressType string `json:"address_type"`
	TokenType   string `json:"token_type"`
}

type JW3TPayload struct {
	IssuedAt   string `json:"issued_at"`
	NotBefore  string `json:"not_before"`
	ExpiresAt  string `json:"expires_at"`
	Address    string `json:"address"`
	OnBehalfOf string `json:"on_behalf_of"`
	ProxyType  string `json:"proxy_type"`
}

type JW3Token struct {
	Header       *JW3THeader
	Base64Header string
	JSONHeader   []byte

	Payload       *JW3TPayload
	Base64Payload string
	JSONPayload   []byte

	Signature []byte
}
