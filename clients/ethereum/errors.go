package ethereum

import "errors"

var (
	errQuorumNotReached                    = errors.New("quorum not reached")
	errInsufficientErc20Balance            = errors.New("insufficient ERC20 balance")
	errInsufficientBalance                 = errors.New("insufficient balance")
	errPublicKeyCast                       = errors.New("error casting public key to ECDSA")
	errNilClientWrapper                    = errors.New("nil client wrapper")
	errNilERC20ContractsHandler            = errors.New("nil ERC20 contracts handler")
	errNilBroadcaster                      = errors.New("nil broadcaster")
	errNilSignaturesHolder                 = errors.New("nil signatures holder")
	errNilGasHandler                       = errors.New("nil gas handler")
	errInvalidGasLimit                     = errors.New("invalid gas limit")
	errNilEthClient                        = errors.New("nil eth client")
	errDepositsAndBatchDepositsCountDiffer = errors.New("deposits and batch.DepositsCount differs")
)
