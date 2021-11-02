package mock

import (
	"encoding/hex"
	"fmt"
	"math/big"
	"strings"

	"github.com/ElrondNetwork/elrond-eth-bridge/integrationTests"
	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/data/vm"
	erdgoCore "github.com/ElrondNetwork/elrond-sdk-erdgo/core"
	"github.com/ElrondNetwork/elrond-sdk-erdgo/data"
	"github.com/ethereum/go-ethereum/common"
)

// ElrondProposedStatus -
type ElrondProposedStatus struct {
	BatchId  *big.Int
	Statuses []byte
}

// ElrondProposedTransfer -
type ElrondProposedTransfer struct {
	BatchId   *big.Int
	Transfers []Transfer
}

// Transfer -
type Transfer struct {
	To     []byte
	Token  string
	Amount *big.Int
}

// ElrondPendingBatch -
type ElrondPendingBatch struct {
	ElrondDeposits []ElrondDeposit
}

// ElrondDeposit -
type ElrondDeposit struct {
	From         erdgoCore.AddressHandler
	To           common.Address
	TokenAddress common.Address
	Amount       *big.Int
}

// elrondContractStateMock is not concurrent safe
type elrondContractStateMock struct {
	*tokensRegistryMock
	proposedStatus                   map[string]*ElrondProposedStatus   // store them uniquely by their hash
	proposedTransfers                map[string]*ElrondProposedTransfer // store them uniquely by their hash
	signedActionIDs                  map[string]map[string]struct{}
	GetStatusesAfterExecutionHandler func() []byte
	relayers                         [][]byte
	performedAction                  *big.Int
	pendingBatch                     *ElrondPendingBatch
	quorum                           int
}

func newElrondContractStateMock() *elrondContractStateMock {
	mock := &elrondContractStateMock{
		tokensRegistryMock: &tokensRegistryMock{},
	}
	mock.cleanState()
	mock.clearTokens()

	return mock
}

// Clean -
func (mock *elrondContractStateMock) cleanState() {
	mock.proposedStatus = make(map[string]*ElrondProposedStatus)
	mock.proposedTransfers = make(map[string]*ElrondProposedTransfer)
	mock.signedActionIDs = make(map[string]map[string]struct{})
	mock.performedAction = nil
	mock.pendingBatch = nil
}

func (mock *elrondContractStateMock) processTransaction(tx *data.Transaction) {
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
		return
	}

	panic("can not execute transaction that calls function: " + funcName)
}

func (mock *elrondContractStateMock) proposeEsdtSafeSetCurrentTransactionBatchStatus(dataSplit []string, _ *data.Transaction) {
	status, hash := mock.createProposedStatus(dataSplit)

	mock.proposedStatus[hash] = status
}

func (mock *elrondContractStateMock) proposeMultiTransferEsdtBatch(dataSplit []string, _ *data.Transaction) {
	transfer, hash := mock.createProposedTransfer(dataSplit)

	mock.proposedTransfers[hash] = transfer
}

func (mock *elrondContractStateMock) createProposedStatus(dataSplit []string) (*ElrondProposedStatus, string) {
	buff, err := hex.DecodeString(dataSplit[1])
	if err != nil {
		panic(err)
	}
	status := &ElrondProposedStatus{
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

	hash, err := core.CalculateHash(integrationTests.TestMarshalizer, integrationTests.TestHasher, status)
	if err != nil {
		panic(err)
	}

	return status, string(hash)
}

func (mock *elrondContractStateMock) createProposedTransfer(dataSplit []string) (*ElrondProposedTransfer, string) {
	buff, err := hex.DecodeString(dataSplit[1])
	if err != nil {
		panic(err)
	}
	transfer := &ElrondProposedTransfer{
		BatchId: big.NewInt(0).SetBytes(buff),
	}

	for i := 2; i < len(dataSplit); i += 3 {
		to, errDecode := hex.DecodeString(dataSplit[i])
		if errDecode != nil {
			panic(errDecode)
		}

		amountBytes, errDecode := hex.DecodeString(dataSplit[i+2])
		if errDecode != nil {
			panic(errDecode)
		}

		t := Transfer{
			To:     to,
			Token:  dataSplit[i+1],
			Amount: big.NewInt(0).SetBytes(amountBytes),
		}

		transfer.Transfers = append(transfer.Transfers, t)
	}

	hash, err := core.CalculateHash(integrationTests.TestMarshalizer, integrationTests.TestHasher, transfer)
	if err != nil {
		panic(err)
	}

	actionID := HashToActionID(string(hash))
	integrationTests.Log.Debug("actionID for createProposedTransfer", "value", actionID.String())

	return transfer, string(hash)
}

func (mock *elrondContractStateMock) processVmRequests(vmRequest *data.VmValueRequest) (*data.VmValuesResponseData, error) {
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
	case "getActionSignerCount":
		return mock.vmRequestGetActionSignerCount(vmRequest), nil
	case "getTokenIdForErc20Address":
		return mock.vmRequestGetTokenIdForErc20Address(vmRequest), nil
	case "getErc20AddressForTokenId":
		return mock.vmRequestGetErc20AddressForTokenId(vmRequest), nil
	case "getCurrentTxBatch":
		return mock.vmRequestGetCurrentPendingBatch(vmRequest), nil
	case "getAllStakedRelayers":
		return mock.vmRequestGetAllStakedRelayers(vmRequest), nil
	}

	panic("unimplemented function: " + vmRequest.FuncName)
}

func (mock *elrondContractStateMock) vmRequestWasSetCurrentTransactionBatchStatusActionProposed(vmRequest *data.VmValueRequest) *data.VmValuesResponseData {
	args := append([]string{"function name"}, vmRequest.Args...) // prepend the function name so the next call will work
	_, hash := mock.createProposedStatus(args)

	_, found := mock.proposedStatus[hash]

	return createOkVmResponse([][]byte{BoolToByteSlice(found)})
}

func (mock *elrondContractStateMock) vmRequestGetActionIdForSetCurrentTransactionBatchStatus(vmRequest *data.VmValueRequest) *data.VmValuesResponseData {
	args := append([]string{"function name"}, vmRequest.Args...) // prepend the function name so the next call will work
	_, hash := mock.createProposedStatus(args)

	_, found := mock.proposedStatus[hash]
	if !found {
		return createNokVmResponse(fmt.Errorf("proposed status not found for hash %s", hex.EncodeToString([]byte(hash))))
	}

	return createOkVmResponse([][]byte{Uint64BytesFromHash(hash)})
}

func (mock *elrondContractStateMock) vmRequestwasTransferActionProposed(vmRequest *data.VmValueRequest) *data.VmValuesResponseData {
	args := append([]string{"function name"}, vmRequest.Args...) // prepend the function name so the next call will work
	_, hash := mock.createProposedTransfer(args)

	_, found := mock.proposedTransfers[hash]

	return createOkVmResponse([][]byte{BoolToByteSlice(found)})
}

func (mock *elrondContractStateMock) vmRequestGetActionIdForTransferBatch(vmRequest *data.VmValueRequest) *data.VmValuesResponseData {
	args := append([]string{"function name"}, vmRequest.Args...) // prepend the function name so the next call will work
	_, hash := mock.createProposedTransfer(args)

	_, found := mock.proposedTransfers[hash]
	if !found {
		return createNokVmResponse(fmt.Errorf("proposed transfer not found for hash %s", hex.EncodeToString([]byte(hash))))
	}

	return createOkVmResponse([][]byte{Uint64BytesFromHash(hash)})
}

func (mock *elrondContractStateMock) vmRequestGetStatusesAfterExecution(_ *data.VmValueRequest) *data.VmValuesResponseData {
	statuses := mock.GetStatusesAfterExecutionHandler()

	args := [][]byte{BoolToByteSlice(true)} // batch finished
	for _, stat := range statuses {
		args = append(args, []byte{stat})
	}

	return createOkVmResponse(args)
}

func (mock *elrondContractStateMock) sign(dataSplit []string, tx *data.Transaction) {
	actionID := getActionIDFromString(dataSplit[1])
	if !mock.actionIDExists(actionID) {
		panic(fmt.Sprintf("attempted to sign on a missing action ID: %v as big int, raw: %s", actionID, dataSplit[1]))
	}

	m, found := mock.signedActionIDs[actionID.String()]
	if !found {
		m = make(map[string]struct{})
		mock.signedActionIDs[actionID.String()] = m
	}
	m[tx.SndAddr] = struct{}{}
}

func (mock *elrondContractStateMock) performAction(dataSplit []string, _ *data.Transaction) {
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

func (mock *elrondContractStateMock) vmRequestWasActionExecuted(vmRequest *data.VmValueRequest) *data.VmValuesResponseData {
	actionID := getActionIDFromString(vmRequest.Args[0])
	actionProposed := actionID.Cmp(mock.performedAction) == 0

	return createOkVmResponse([][]byte{BoolToByteSlice(actionProposed)})
}

func (mock *elrondContractStateMock) actionIDExists(actionID *big.Int) bool {
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

func (mock *elrondContractStateMock) vmRequestGetActionSignerCount(vmRequest *data.VmValueRequest) *data.VmValuesResponseData {
	actionID := getActionIDFromString(vmRequest.Args[0])
	m, found := mock.signedActionIDs[actionID.String()]
	if !found {
		return createOkVmResponse([][]byte{Uint64ByteSlice(0)})
	}

	return createOkVmResponse([][]byte{Uint64ByteSlice(uint64(len(m)))})
}

func (mock *elrondContractStateMock) vmRequestGetTokenIdForErc20Address(vmRequest *data.VmValueRequest) *data.VmValuesResponseData {
	address := common.HexToAddress(vmRequest.Args[0])

	return createOkVmResponse([][]byte{[]byte(mock.getTicker(address))})
}

func (mock *elrondContractStateMock) vmRequestGetErc20AddressForTokenId(vmRequest *data.VmValueRequest) *data.VmValuesResponseData {
	address := vmRequest.Args[0]

	return createOkVmResponse([][]byte{mock.getErc20Address(address).Bytes()})
}

func (mock *elrondContractStateMock) vmRequestGetAllStakedRelayers(_ *data.VmValueRequest) *data.VmValuesResponseData {
	return createOkVmResponse(mock.relayers)
}

func (mock *elrondContractStateMock) vmRequestGetCurrentPendingBatch(_ *data.VmValueRequest) *data.VmValuesResponseData {
	if mock.pendingBatch == nil {
		return createOkVmResponse(make([][]byte, 0))
	}

	args := [][]byte{{0}} // first non-empty slice
	for _, deposit := range mock.pendingBatch.ElrondDeposits {
		args = append(args, make([]byte, 0)) // mocked block nonce
		args = append(args, make([]byte, 0)) // mocked deposit nonce
		args = append(args, deposit.From.AddressBytes())
		args = append(args, deposit.To.Bytes())
		args = append(args, deposit.TokenAddress.Bytes())
		args = append(args, deposit.Amount.Bytes())
	}

	return createOkVmResponse(args)
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
