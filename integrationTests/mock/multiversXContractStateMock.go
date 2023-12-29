package mock

import (
	"encoding/hex"
	"fmt"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/multiversx/mx-bridge-eth-go/integrationTests"
	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data/transaction"
	"github.com/multiversx/mx-chain-core-go/data/vm"
	sdkCore "github.com/multiversx/mx-sdk-go/core"
	"github.com/multiversx/mx-sdk-go/data"
)

type multiversXProposedStatus struct {
	BatchId  *big.Int
	Statuses []byte
}

type multiversXProposedTransfer struct {
	BatchId   *big.Int
	Transfers []Transfer
}

// Transfer -
type Transfer struct {
	From     []byte
	To       []byte
	Token    string
	Amount   *big.Int
	Nonce    *big.Int
	ExtraGas uint64
	Data     []byte
}

// MultiversXPendingBatch -
type MultiversXPendingBatch struct {
	Nonce              *big.Int
	MultiversXDeposits []MultiversXDeposit
}

// MultiversXDeposit -
type MultiversXDeposit struct {
	From         sdkCore.AddressHandler
	To           common.Address
	Ticker       string
	Amount       *big.Int
	DepositNonce uint64
}

// multiversXContractStateMock is not concurrent safe
type multiversXContractStateMock struct {
	*tokensRegistryMock
	proposedStatus                   map[string]*multiversXProposedStatus   // store them uniquely by their hash
	proposedTransfers                map[string]*multiversXProposedTransfer // store them uniquely by their hash
	signedActionIDs                  map[string]map[string]struct{}
	GetStatusesAfterExecutionHandler func() []byte
	ProcessFinishedHandler           func()
	relayers                         [][]byte
	performedAction                  *big.Int
	pendingBatch                     *MultiversXPendingBatch
	quorum                           int
	lastExecutedEthBatchId           uint64
	lastExecutedEthTxId              uint64

	ProposeMultiTransferEsdtBatchCalled func()
}

func newMultiversXContractStateMock() *multiversXContractStateMock {
	mock := &multiversXContractStateMock{
		tokensRegistryMock: &tokensRegistryMock{},
	}
	mock.cleanState()
	mock.clearTokens()

	return mock
}

// Clean -
func (mock *multiversXContractStateMock) cleanState() {
	mock.proposedStatus = make(map[string]*multiversXProposedStatus)
	mock.proposedTransfers = make(map[string]*multiversXProposedTransfer)
	mock.signedActionIDs = make(map[string]map[string]struct{})
	mock.performedAction = nil
	mock.pendingBatch = nil
}

func (mock *multiversXContractStateMock) processTransaction(tx *transaction.FrontendTransaction) {
	dataSplit := strings.Split(string(tx.Data), "@")
	funcName := dataSplit[0]
	switch funcName {
	case "proposeEsdtSafeSetCurrentTransactionBatchStatus":
		mock.proposeEsdtSafeSetCurrentTransactionBatchStatus(dataSplit, tx)

		return
	case "proposeMultiTransferEsdtBatch":
		mock.proposeMultiTransferEsdtBatch(dataSplit, tx)
		return
	case "sign":
		mock.sign(dataSplit, tx)
		return
	case "performAction":
		mock.performAction(dataSplit, tx)

		if mock.ProcessFinishedHandler != nil {
			mock.ProcessFinishedHandler()
		}
		return
	}

	panic("can not execute transaction that calls function: " + funcName)
}

func (mock *multiversXContractStateMock) proposeEsdtSafeSetCurrentTransactionBatchStatus(dataSplit []string, _ *transaction.FrontendTransaction) {
	status, hash := mock.createProposedStatus(dataSplit)

	mock.proposedStatus[hash] = status
}

func (mock *multiversXContractStateMock) proposeMultiTransferEsdtBatch(dataSplit []string, _ *transaction.FrontendTransaction) {
	transfer, hash := mock.createProposedTransfer(dataSplit)

	mock.proposedTransfers[hash] = transfer

	if mock.ProposeMultiTransferEsdtBatchCalled != nil {
		mock.ProposeMultiTransferEsdtBatchCalled()
	}
}

func (mock *multiversXContractStateMock) createProposedStatus(dataSplit []string) (*multiversXProposedStatus, string) {
	buff, err := hex.DecodeString(dataSplit[1])
	if err != nil {
		panic(err)
	}
	status := &multiversXProposedStatus{
		BatchId:  big.NewInt(0).SetBytes(buff),
		Statuses: make([]byte, 0),
	}

	for i := 2; i < len(dataSplit); i++ {
		stat, errDecode := hex.DecodeString(dataSplit[i])
		if errDecode != nil {
			panic(errDecode)
		}

		status.Statuses = append(status.Statuses, stat[0])
	}

	if len(status.Statuses) != len(mock.pendingBatch.MultiversXDeposits) {
		panic(fmt.Sprintf("different number of statuses fetched while creating proposed status: provided %d, existing %d",
			len(status.Statuses), len(mock.pendingBatch.MultiversXDeposits)))
	}

	hash, err := core.CalculateHash(integrationTests.TestMarshalizer, integrationTests.TestHasher, status)
	if err != nil {
		panic(err)
	}

	return status, string(hash)
}

func (mock *multiversXContractStateMock) createProposedTransfer(dataSplit []string) (*multiversXProposedTransfer, string) {
	buff, err := hex.DecodeString(dataSplit[1])
	if err != nil {
		panic(err)
	}
	transfer := &multiversXProposedTransfer{
		BatchId: big.NewInt(0).SetBytes(buff),
	}

	currentIndex := 2
	for currentIndex < len(dataSplit) {
		from, errDecode := hex.DecodeString(dataSplit[currentIndex])
		if errDecode != nil {
			panic(errDecode)
		}

		to, errDecode := hex.DecodeString(dataSplit[currentIndex+1])
		if errDecode != nil {
			panic(errDecode)
		}

		amountBytes, errDecode := hex.DecodeString(dataSplit[currentIndex+3])
		if errDecode != nil {
			panic(errDecode)
		}

		nonceBytes, errDecode := hex.DecodeString(dataSplit[currentIndex+4])
		if errDecode != nil {
			panic(errDecode)
		}

		t := Transfer{
			From:   from,
			To:     to,
			Token:  dataSplit[currentIndex+2],
			Amount: big.NewInt(0).SetBytes(amountBytes),
			Nonce:  big.NewInt(0).SetBytes(nonceBytes),
		}

		indexIncrementValue := 5
		if core.IsSmartContractAddress(to) {
			indexIncrementValue += 2
			t.Data, errDecode = hex.DecodeString(dataSplit[currentIndex+5])
			if errDecode != nil {
				panic(errDecode)
			}

			var extraGasBytes []byte
			extraGasBytes, errDecode = hex.DecodeString(dataSplit[currentIndex+6])
			if errDecode != nil {
				panic(errDecode)
			}

			t.ExtraGas = big.NewInt(0).SetBytes(extraGasBytes).Uint64()
		}

		transfer.Transfers = append(transfer.Transfers, t)
		currentIndex += indexIncrementValue
	}

	hash, err := core.CalculateHash(integrationTests.TestMarshalizer, integrationTests.TestHasher, transfer)
	if err != nil {
		panic(err)
	}

	actionID := HashToActionID(string(hash))
	integrationTests.Log.Debug("actionID for createProposedTransfer", "value", actionID.String())

	return transfer, string(hash)
}

func (mock *multiversXContractStateMock) processVmRequests(vmRequest *data.VmValueRequest) (*data.VmValuesResponseData, error) {
	if vmRequest == nil {
		panic("vmRequest is nil")
	}

	switch vmRequest.FuncName {
	case "wasTransferActionProposed":
		return mock.vmRequestwasTransferActionProposed(vmRequest), nil
	case "getActionIdForTransferBatch":
		return mock.vmRequestGetActionIdForTransferBatch(vmRequest), nil
	case "wasSetCurrentTransactionBatchStatusActionProposed":
		return mock.vmRequestWasSetCurrentTransactionBatchStatusActionProposed(vmRequest), nil
	case "getStatusesAfterExecution":
		return mock.vmRequestGetStatusesAfterExecution(vmRequest), nil
	case "getActionIdForSetCurrentTransactionBatchStatus":
		return mock.vmRequestGetActionIdForSetCurrentTransactionBatchStatus(vmRequest), nil
	case "wasActionExecuted":
		return mock.vmRequestWasActionExecuted(vmRequest), nil
	case "quorumReached":
		return mock.vmRequestQuorumReached(vmRequest), nil
	case "getTokenIdForErc20Address":
		return mock.vmRequestGetTokenIdForErc20Address(vmRequest), nil
	case "getErc20AddressForTokenId":
		return mock.vmRequestGetErc20AddressForTokenId(vmRequest), nil
	case "getCurrentTxBatch":
		return mock.vmRequestGetCurrentPendingBatch(vmRequest), nil
	case "getAllStakedRelayers":
		return mock.vmRequestGetAllStakedRelayers(vmRequest), nil
	case "getLastExecutedEthBatchId":
		return mock.vmRequestGetLastExecutedEthBatchId(vmRequest), nil
	case "getLastExecutedEthTxId":
		return mock.vmRequestGetLastExecutedEthTxId(vmRequest), nil
	case "signed":
		return mock.vmRequestSigned(vmRequest), nil
	case "isPaused":
		return mock.vmRequestIsPaused(vmRequest), nil
	case "isMintBurnAllowed":
		return mock.vmRequestIsMintBurnAllowed(vmRequest), nil
	case "getAccumulatedBurnedTokens":
		return mock.vmRequestGetAccumulatedBurnedTokens(vmRequest), nil
	}

	panic("unimplemented function: " + vmRequest.FuncName)
}

func (mock *multiversXContractStateMock) vmRequestWasSetCurrentTransactionBatchStatusActionProposed(vmRequest *data.VmValueRequest) *data.VmValuesResponseData {
	args := append([]string{vmRequest.FuncName}, vmRequest.Args...) // prepend the function name so the next call will work
	_, hash := mock.createProposedStatus(args)

	_, found := mock.proposedStatus[hash]

	return createOkVmResponse([][]byte{BoolToByteSlice(found)})
}

func (mock *multiversXContractStateMock) vmRequestGetActionIdForSetCurrentTransactionBatchStatus(vmRequest *data.VmValueRequest) *data.VmValuesResponseData {
	args := append([]string{vmRequest.FuncName}, vmRequest.Args...) // prepend the function name so the next call will work
	_, hash := mock.createProposedStatus(args)

	_, found := mock.proposedStatus[hash]
	if !found {
		return createNokVmResponse(fmt.Errorf("proposed status not found for hash %s", hex.EncodeToString([]byte(hash))))
	}

	return createOkVmResponse([][]byte{Uint64BytesFromHash(hash)})
}

func (mock *multiversXContractStateMock) vmRequestwasTransferActionProposed(vmRequest *data.VmValueRequest) *data.VmValuesResponseData {
	args := append([]string{vmRequest.FuncName}, vmRequest.Args...) // prepend the function name so the next call will work
	_, hash := mock.createProposedTransfer(args)

	_, found := mock.proposedTransfers[hash]

	return createOkVmResponse([][]byte{BoolToByteSlice(found)})
}

func (mock *multiversXContractStateMock) vmRequestGetActionIdForTransferBatch(vmRequest *data.VmValueRequest) *data.VmValuesResponseData {
	args := append([]string{vmRequest.FuncName}, vmRequest.Args...) // prepend the function name so the next call will work
	_, hash := mock.createProposedTransfer(args)

	_, found := mock.proposedTransfers[hash]
	if !found {
		// return action ID == 0 in case there is no such transfer proposed
		return createOkVmResponse([][]byte{big.NewInt(0).Bytes()})
	}

	return createOkVmResponse([][]byte{Uint64BytesFromHash(hash)})
}

func (mock *multiversXContractStateMock) vmRequestGetStatusesAfterExecution(_ *data.VmValueRequest) *data.VmValuesResponseData {
	statuses := mock.GetStatusesAfterExecutionHandler()

	args := [][]byte{BoolToByteSlice(true)} // batch finished
	for _, stat := range statuses {
		args = append(args, []byte{stat})
	}

	return createOkVmResponse(args)
}

func (mock *multiversXContractStateMock) sign(dataSplit []string, tx *transaction.FrontendTransaction) {
	actionID := getActionIDFromString(dataSplit[1])
	if !mock.actionIDExists(actionID) {
		panic(fmt.Sprintf("attempted to sign on a missing action ID: %v as big int, raw: %s", actionID, dataSplit[1]))
	}

	m, found := mock.signedActionIDs[actionID.String()]
	if !found {
		m = make(map[string]struct{})
		mock.signedActionIDs[actionID.String()] = m
	}
	m[tx.Sender] = struct{}{}
}

func (mock *multiversXContractStateMock) performAction(dataSplit []string, _ *transaction.FrontendTransaction) {
	actionID := getActionIDFromString(dataSplit[1])
	if !mock.actionIDExists(actionID) {
		panic(fmt.Sprintf("attempted to perform on a missing action ID: %v as big int, raw: %s", actionID, dataSplit[1]))
	}

	m, found := mock.signedActionIDs[actionID.String()]
	if !found {
		panic(fmt.Sprintf("attempted to perform on a not signed action ID: %v as big int, raw: %s", actionID, dataSplit[1]))
	}

	if len(m) >= mock.quorum {
		mock.performedAction = actionID
	}
}

func (mock *multiversXContractStateMock) vmRequestWasActionExecuted(vmRequest *data.VmValueRequest) *data.VmValuesResponseData {
	actionID := getActionIDFromString(vmRequest.Args[0])

	if mock.performedAction == nil {
		return createOkVmResponse([][]byte{BoolToByteSlice(false)})
	}

	actionProposed := actionID.Cmp(mock.performedAction) == 0

	return createOkVmResponse([][]byte{BoolToByteSlice(actionProposed)})
}

func (mock *multiversXContractStateMock) actionIDExists(actionID *big.Int) bool {
	for hash := range mock.proposedTransfers {
		existingActionID := HashToActionID(hash)
		if existingActionID.Cmp(actionID) == 0 {
			return true
		}
	}

	for hash := range mock.proposedStatus {
		existingActionID := HashToActionID(hash)
		if existingActionID.Cmp(actionID) == 0 {
			return true
		}
	}

	return false
}

func (mock *multiversXContractStateMock) vmRequestQuorumReached(vmRequest *data.VmValueRequest) *data.VmValuesResponseData {
	actionID := getActionIDFromString(vmRequest.Args[0])
	m, found := mock.signedActionIDs[actionID.String()]
	if !found {
		return createOkVmResponse([][]byte{BoolToByteSlice(false)})
	}

	quorumReached := len(m) >= mock.quorum

	return createOkVmResponse([][]byte{BoolToByteSlice(quorumReached)})
}

func (mock *multiversXContractStateMock) vmRequestGetTokenIdForErc20Address(vmRequest *data.VmValueRequest) *data.VmValuesResponseData {
	address := common.HexToAddress(vmRequest.Args[0])

	return createOkVmResponse([][]byte{[]byte(mock.getTicker(address))})
}

func (mock *multiversXContractStateMock) vmRequestGetErc20AddressForTokenId(vmRequest *data.VmValueRequest) *data.VmValuesResponseData {
	address := vmRequest.Args[0]

	return createOkVmResponse([][]byte{mock.getErc20Address(address).Bytes()})
}

func (mock *multiversXContractStateMock) vmRequestGetAllStakedRelayers(_ *data.VmValueRequest) *data.VmValuesResponseData {
	return createOkVmResponse(mock.relayers)
}

func (mock *multiversXContractStateMock) vmRequestGetLastExecutedEthBatchId(_ *data.VmValueRequest) *data.VmValuesResponseData {
	val := big.NewInt(int64(mock.lastExecutedEthBatchId))

	return createOkVmResponse([][]byte{val.Bytes()})
}

func (mock *multiversXContractStateMock) vmRequestGetLastExecutedEthTxId(_ *data.VmValueRequest) *data.VmValuesResponseData {
	val := big.NewInt(int64(mock.lastExecutedEthTxId))

	return createOkVmResponse([][]byte{val.Bytes()})
}

func (mock *multiversXContractStateMock) vmRequestGetCurrentPendingBatch(_ *data.VmValueRequest) *data.VmValuesResponseData {
	if mock.pendingBatch == nil {
		return createOkVmResponse(make([][]byte, 0))
	}

	args := [][]byte{mock.pendingBatch.Nonce.Bytes()} // first non-empty slice
	for _, deposit := range mock.pendingBatch.MultiversXDeposits {
		args = append(args, make([]byte, 0)) // mocked block nonce
		args = append(args, big.NewInt(0).SetUint64(deposit.DepositNonce).Bytes())
		args = append(args, deposit.From.AddressBytes())
		args = append(args, deposit.To.Bytes())
		args = append(args, []byte(deposit.Ticker))
		args = append(args, deposit.Amount.Bytes())
	}
	return createOkVmResponse(args)
}

func (mock *multiversXContractStateMock) setPendingBatch(pendingBatch *MultiversXPendingBatch) {
	mock.pendingBatch = pendingBatch
}

func (mock *multiversXContractStateMock) vmRequestSigned(request *data.VmValueRequest) *data.VmValuesResponseData {
	hexAddress := request.Args[0]
	actionID := getActionIDFromString(request.Args[1])

	actionIDMap, found := mock.signedActionIDs[actionID.String()]
	if !found {
		return createOkVmResponse([][]byte{BoolToByteSlice(false)})
	}

	addressBytes, err := hex.DecodeString(hexAddress)
	if err != nil {
		panic(err)
	}

	address := data.NewAddressFromBytes(addressBytes)
	bech32Address, _ := address.AddressAsBech32String()
	_, found = actionIDMap[bech32Address]
	if !found {
		log.Error("action ID not found", "address", bech32Address)
	}

	return createOkVmResponse([][]byte{BoolToByteSlice(found)})
}

func (mock *multiversXContractStateMock) vmRequestIsPaused(_ *data.VmValueRequest) *data.VmValuesResponseData {
	return createOkVmResponse([][]byte{BoolToByteSlice(false)})
}

func (mock *multiversXContractStateMock) vmRequestIsMintBurnAllowed(vmRequest *data.VmValueRequest) *data.VmValuesResponseData {
	address := vmRequest.Args[0]

	return createOkVmResponse([][]byte{BoolToByteSlice(mock.isNativeToken(address))})
}

func (mock *multiversXContractStateMock) vmRequestGetAccumulatedBurnedTokens(vmRequest *data.VmValueRequest) *data.VmValuesResponseData {
	address := vmRequest.Args[0]

	return createOkVmResponse([][]byte{mock.getAccumulatedBurn(address).Bytes()})
}

func getActionIDFromString(data string) *big.Int {
	buff, err := hex.DecodeString(data)
	if err != nil {
		panic(err)
	}

	return big.NewInt(0).SetBytes(buff)
}

func createOkVmResponse(args [][]byte) *data.VmValuesResponseData {
	return &data.VmValuesResponseData{
		Data: &vm.VMOutputApi{
			ReturnData: args,
			ReturnCode: "ok",
		},
	}
}

func createNokVmResponse(err error) *data.VmValuesResponseData {
	return &data.VmValuesResponseData{
		Data: &vm.VMOutputApi{
			ReturnCode:    "nok",
			ReturnMessage: err.Error(),
		},
	}
}
