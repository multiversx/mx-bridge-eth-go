package ethereum

import (
	"crypto/ecdsa"
	"math/big"
	"os"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	ethCrypto "github.com/ethereum/go-ethereum/crypto"
	"github.com/multiversx/mx-bridge-eth-go/core/converters"
)

type cryptoHandler struct {
	privateKey *ecdsa.PrivateKey
	publicKey  *ecdsa.PublicKey
	address    common.Address
}

// NewCryptoHandler creates a new instance of type cryptoHandler able to sign messages and provide the containing public key
func NewCryptoHandler(privateKeyFilename string) (*cryptoHandler, error) {
	privateKeyBytes, err := os.ReadFile(privateKeyFilename)
	if err != nil {
		return nil, err
	}
	privateKeyString := converters.TrimWhiteSpaceCharacters(string(privateKeyBytes))
	privateKey, err := ethCrypto.HexToECDSA(privateKeyString)
	if err != nil {
		return nil, err
	}

	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		return nil, errPublicKeyCast
	}

	return &cryptoHandler{
		privateKey: privateKey,
		publicKey:  publicKeyECDSA,
		address:    ethCrypto.PubkeyToAddress(*publicKeyECDSA),
	}, nil
}

// Sign signs the provided message hash with the containing private key
func (handler *cryptoHandler) Sign(msgHash common.Hash) ([]byte, error) {
	return ethCrypto.Sign(msgHash.Bytes(), handler.privateKey)
}

// GetAddress returns the corresponding address of the containing public key
func (handler *cryptoHandler) GetAddress() common.Address {
	return handler.address
}

// CreateKeyedTransactor creates a keyed transactor used to create transactions on Ethereum chain
func (handler *cryptoHandler) CreateKeyedTransactor(chainId *big.Int) (*bind.TransactOpts, error) {
	return bind.NewKeyedTransactorWithChainID(handler.privateKey, chainId)
}

// IsInterfaceNil returns true if there is no value under the interface
func (handler *cryptoHandler) IsInterfaceNil() bool {
	return handler == nil
}
