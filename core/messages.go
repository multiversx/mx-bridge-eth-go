package core

import "fmt"

// TODO make these compatible with the gogo proto marshalizer, inject marshalizer in broadcaster constructor

// SignedMessage is the message used when communicating with other relayers
type SignedMessage struct {
	Payload        []byte `json:"payload"`
	PublicKeyBytes []byte `json:"pk"`
	Signature      []byte `json:"sig"`
	Nonce          uint64 `json:"nonce"`
}

// UniqueID will return the string ID assembled from the public key bytes and the message nonce
func (msg *SignedMessage) UniqueID() string {
	return fmt.Sprintf("%s%s", string(msg.PublicKeyBytes), string(msg.Payload))
}

// EthereumSignature is the message used when the relayers will send an ethereum signature
type EthereumSignature struct {
	Signature   []byte `json:"sig"`
	MessageHash []byte `json:"msg"`
}
