package main

import (
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"math/big"
)

func run(stateRootHash *common.Hash) {
	// Constants
	const dbPath string = "/home/mbacic/tmp/leveldb/ethereum"
	state := NewState(dbPath, stateRootHash)

	fmt.Println("startStateDump:\n", string(state.State.Dump(nil)))

	var genesisSender = common.HexToAddress("0x6399121770646811e854d5393e0236e24721b803")
	firstReceiver := common.HexToAddress("0x1a635202b44d708c2606495b5fbc1bd7f9cafd39")

	block := TolarBlock{
		Hash:              common.HexToHash("0x5c698f13940a2153440c6d19660878bc90219d9298fdcf37365aa8d88d40fc42"),
		Index:             new(big.Int).SetInt64(1),
		Timestamp:         10,
		PreviousBlockHash: common.Hash{},
		Transactions: []*TolarTx{
			{
				Hash:     common.HexToHash("0x5c698f13940a2153440c6d19660878bc90219d9298fdcf37365aa8d88d40fc42"),
				Sender:   genesisSender,
				Receiver: &firstReceiver,
				Value:    new(big.Int).SetUint64(1000000),
				Gas:      21000,
				GasPrice: new(big.Int).SetUint64(100),
				Data:     nil,
				Nonce:    0,
			},
			{
				Hash:     common.HexToHash("0x5c698f13940a2153440c6d19660878bc90219d9298fdcf37365aa8d88d40fc42"),
				Sender:   common.HexToAddress("0x1a635202b44d708c2606495b5fbc1bd7f9cafd39"),
				Receiver: &genesisSender,
				Value:    new(big.Int).SetUint64(100),
				Gas:      21000,
				GasPrice: new(big.Int).SetUint64(100),
				Data:     nil,
				Nonce:    0,
			},
		},
	}

	var tx = block.Transactions[0]
	fmt.Println("Freshly loaded balances are: ")
	fmt.Println("\tSender ", tx.Sender.Hex(), " balance is: ", state.GetBalance(tx.Sender), "; Nonce: ", state.GetNonce(tx.Sender))
	fmt.Println("\tReceiver ", tx.Receiver.Hex(), " balance is: ", state.GetBalance(*tx.Receiver), "; Nonce: ", state.GetNonce(*tx.Receiver))

	if stateRootHash != nil {
		_ = state.Close()
		return
	}

	fmt.Println("Expected failed account balance: ", 100000000000000000-2100000-1000000)

	if state.State.Empty(genesisSender) {
		state.State.CreateAccount(genesisSender)
		state.State.AddBalance(genesisSender, new(big.Int).SetUint64(100000000000000000))
		fmt.Println("\t!After adding genesis data! Sender", tx.Sender.Hex(), " balance is: ", state.GetBalance(tx.Sender), "; Nonce: ", state.GetNonce(tx.Sender))
	}

	blockProcessor := NewBlockProcessor(state.State)
	fmt.Println("\tEXECUTING BLOCK:")
	execResults, newStateRootHash := blockProcessor.Process(&block)

	fmt.Println("New root hash: ", newStateRootHash.Hex())

	fmt.Println("\tSender ", tx.Sender.Hex(), " balance is: ", state.GetBalance(tx.Sender))
	fmt.Println("\tReceiver ", tx.Receiver.Hex(), " balance is: ", state.GetBalance(*tx.Receiver))

	for index, execResult := range execResults {
		fmt.Println("Transaction with index: ", index)

		if execResult.Error != nil {
			fmt.Println("\tTransaction failed with error: ", execResult.Error)
			continue
		}

		var receipt = execResult.Receipt
		fmt.Println("\tTransaction hash: ", receipt.TxHash.Hex())
		fmt.Println("\tGas used: ", receipt.GasUsed)
		fmt.Println("\tCumulative gas used: ", receipt.CumulativeGasUsed)
	}

	fmt.Println("preCommitDump:\n", string(state.State.Dump(nil)))

	commitHash, runError := state.Commit()
	HandleFatalError("Failed create state", runError)

	runError = state.Persist()
	HandleFatalError("Trie commit failed:", runError)

	fmt.Println("postCommitDump:\n", string(state.State.Dump(nil)))
	fmt.Println("\tCommit hash is: ", commitHash.Hex())

	runError = state.Close()
	HandleFatalError("Trie commit failed:", runError)
}

func main() {
	//TestBalanceSetToGenesisStateBalance()
	//TestSendSomeCoinsAround()
	//TestSimpleContract()
	TestContractSendingCoinsAround()
}
