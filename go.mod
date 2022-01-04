module github.com/ElrondNetwork/elrond-eth-bridge

go 1.16

require (
	github.com/ElrondNetwork/elrond-go v1.2.29
	github.com/ElrondNetwork/elrond-go-core v1.1.2
	github.com/ElrondNetwork/elrond-go-crypto v1.0.1
	github.com/ElrondNetwork/elrond-go-logger v1.0.6
	github.com/ElrondNetwork/elrond-sdk-erdgo v1.0.13-0.20220104122958-3fd284139ec5
	github.com/btcsuite/websocket v0.0.0-20150119174127-31079b680792
	github.com/ethereum/go-ethereum v1.10.8
	github.com/gin-contrib/cors v0.0.0-20190301062745-f9e10995c85a
	github.com/gin-contrib/pprof v1.3.0
	github.com/gin-gonic/gin v1.7.4
	github.com/stretchr/testify v1.7.0
	github.com/urfave/cli v1.22.5
)

replace github.com/ElrondNetwork/arwen-wasm-vm/v1_2 v1.2.30 => github.com/ElrondNetwork/arwen-wasm-vm v1.2.30

replace github.com/ElrondNetwork/arwen-wasm-vm/v1_3 v1.3.30 => github.com/ElrondNetwork/arwen-wasm-vm v1.3.30

replace github.com/ElrondNetwork/arwen-wasm-vm/v1_4 v1.4.22 => github.com/ElrondNetwork/arwen-wasm-vm v1.4.22
