package mock
import (
	"encoding/json"
	apiTransaction "github.com/ElrondNetwork/elrond-go/api/transaction"
	"github.com/ElrondNetwork/elrond-vm-common/parsers"
)

type contractNotifier struct {
	contracts map[string]*Contract
}

func newContractNotifier(contracts map[string]*Contract) *contractNotifier {
	return &contractNotifier{
		contracts: contracts,
	}
}
func (cn *contractNotifier) notifyContract(transaction *apiTransaction.SendTxRequest) {
	receiver := transaction.Receiver
	function, args, _ := parsers.NewCallArgsParser().ParseData(string(transaction.Data))

	log.Debug("ElrondContract: notifyContract", "function", function)
	handler := cn.contracts[receiver].GetHandler(function)

	if handler == nil {
		log.Error("ElrondContract: Error notifyContract", "error", "No handler found")
		return
	}
	handleArgs, _ := json.Marshal(args)
	_, err := handler(transaction.Sender, transaction.Value, string(handleArgs))
	if err != nil {
		log.Error("ElrondContract: Error notifyContract calling handler", "error", err.Error())
	}
}