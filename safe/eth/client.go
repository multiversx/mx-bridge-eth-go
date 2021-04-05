package eth

import (
	"context"
	"fmt"
	"github.com/ElrondNetwork/elrond-eth-bridge/safe"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"math/big"
	"strings"
	"time"
)

const safeAbiDefinition = `[{"anonymous": false,"inputs": [{"indexed": false,"internalType": "address","name": "tokenAddress","type": "address"},{"indexed": false,"internalType": "address","name": "depositor","type": "address"},{"indexed": false,"internalType": "uint256","name": "amount","type": "uint256"}],"name": "ERC20Deposited","type": "event"},{"inputs": [{"internalType": "address","name": "tokenAddress","type": "address"},{"internalType": "uint256","name": "amount","type": "uint256"}],"name": "deposit","outputs": [],"stateMutability": "nonpayable","type": "function"}]`

type Client struct {
	chainReader           ethereum.ChainReader
	safeAddress           common.Address
	safeAbi               abi.ABI
	mostRecentBlockNumber func(ctx context.Context) (*big.Int, error)
}

func NewClient(rawUrl string, safeAddress string) (*Client, error) {
	chainReader, err := ethclient.Dial(rawUrl)

	if err != nil {
		return nil, err
	}

	mostRecentBlockNumber := func(ctx context.Context) (*big.Int, error) {
		number, err := chainReader.BlockNumber(ctx)
		blockNumber := big.NewInt(int64(number))

		return blockNumber, err
	}
	safeAbi, err := abi.JSON(strings.NewReader(safeAbiDefinition))

	if err != nil {
		return nil, err
	}

	client := &Client{
		chainReader:           chainReader,
		safeAddress:           common.HexToAddress(safeAddress),
		safeAbi:               safeAbi,
		mostRecentBlockNumber: mostRecentBlockNumber,
	}

	return client, nil
}

func (c *Client) Bridge(*safe.DepositTransaction) (string, error) {
	// TODO: send transaction to safe
	return "", nil
}

func (c *Client) GetTransactions(ctx context.Context, blockNumber *big.Int, channel safe.SafeTxChan) {
	currentBlockNumber := blockNumber
	for {
		mostRecentBlockNumber, err := c.mostRecentBlockNumber(ctx)

		if err != nil {
			// TODO: log error
			fmt.Println(err)
			return
		}

		if currentBlockNumber.Cmp(mostRecentBlockNumber) == 1 {
			time.Sleep(1 * time.Second)
			continue
		} else {
			err = c.processBlockByNumber(ctx, channel, currentBlockNumber)

			if err != nil {
				// TODO: log err
				fmt.Println(err)
			}

			currentBlockNumber = currentBlockNumber.Add(currentBlockNumber, big.NewInt(1))
		}
	}
}

func (c *Client) processBlockByNumber(ctx context.Context, ch safe.SafeTxChan, number *big.Int) error {
	block, err := c.chainReader.BlockByNumber(ctx, number)

	if err != nil {
		return err
	}

	for _, tx := range c.filterTransactions(block.Transactions()) {
		safeTx, err := c.newSafeTransaction(tx)

		if err != nil {
			return err
		}

		ch <- safeTx
	}

	return nil
}

func (c *Client) filterTransactions(transactions types.Transactions) (result types.Transactions) {
	for _, tx := range transactions {
		var toAddress string

		if tx.To() != nil {
			toAddress = tx.To().String()
		}

		if toAddress == c.safeAddress.String() {
			result = append(result, tx)
		}
	}
	return
}

func (c *Client) newSafeTransaction(tx *types.Transaction) (*safe.DepositTransaction, error) {
	from, err := types.Sender(types.NewEIP2930Signer(tx.ChainId()), tx)

	if err != nil {
		return nil, err
	}

	depositInputs, err := c.unpackDepositTx(tx.Data())

	if err != nil {
		return nil, err
	}

	blockTransaction := &safe.DepositTransaction{
		Hash:         tx.Hash().String(),
		From:         from.String(),
		TokenAddress: depositInputs.tokenAddress,
		Amount:       depositInputs.amount,
	}

	return blockTransaction, nil
}

type depositInputs struct {
	tokenAddress string
	amount       *big.Int
}

const depositMethodName = "deposit"
const depositTokenAddressName = "tokenAddress"
const depositAmountName = "amount"

func (c *Client) unpackDepositTx(data []byte) (*depositInputs, error) {
	v := map[string]interface{}{}
	err := c.safeAbi.Methods[depositMethodName].Inputs.UnpackIntoMap(v, data[4:])

	if err != nil {
		return nil, err
	}

	return &depositInputs{
		tokenAddress: v[depositTokenAddressName].(common.Address).String(),
		amount:       v[depositAmountName].(*big.Int),
	}, nil
}
