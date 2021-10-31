package p2p

// TODO make these compatible with the gogo proto marshalizer, inject marshalizer in broadcaster constructor

// SignedMessage is the message used when communicating with other relayers
type SignedMessage struct {
	Payload        []byte `json:"payload"`
	PublicKeyBytes []byte `json:"pk"`
	Signature      []byte `json:"sig"`
	Nonce          uint64 `json:"nonce"`
}

// EthereumSignature is the message used when the relayers will send an ethereum signature
type EthereumSignature struct {
	Signature   []byte `json:"sig"`
	MessageHash []byte `json:"msg"`
}
