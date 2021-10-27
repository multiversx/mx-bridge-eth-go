package p2p

// TODO make these compatible with the gogo proto marshalizer, inject marshalizer in broadcaster constructor

// SignedMessage is the message used when communicating with other relayers
type SignedMessage struct {
	Payload        []byte
	PublicKeyBytes []byte
	Signature      []byte
	Nonce          uint64
}
