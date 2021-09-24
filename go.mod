module github.com/ElrondNetwork/elrond-eth-bridge

go 1.16

require (
	github.com/ElrondNetwork/elrond-go v1.2.20
	github.com/ElrondNetwork/elrond-go-core v1.1.0
	github.com/ElrondNetwork/elrond-go-crypto v1.0.1
	github.com/ElrondNetwork/elrond-go-logger v1.0.5
	github.com/ElrondNetwork/elrond-sdk-erdgo v1.0.3
	github.com/ethereum/go-ethereum v1.10.8
	github.com/stretchr/testify v1.7.0
	github.com/urfave/cli v1.22.5
)

replace github.com/ElrondNetwork/arwen-wasm-vm/v1_2 v1.2.30 => github.com/ElrondNetwork/arwen-wasm-vm v1.2.30

replace github.com/ElrondNetwork/arwen-wasm-vm/v1_3 v1.3.30 => github.com/ElrondNetwork/arwen-wasm-vm v1.3.30

replace github.com/ElrondNetwork/arwen-wasm-vm/v1_4 v1.4.14 => github.com/ElrondNetwork/arwen-wasm-vm v1.4.14
