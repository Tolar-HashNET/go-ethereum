package main

import (
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"math/big"
	"os"
	"strings"
)

var genesisAddress = common.HexToAddress("0x12c347d6570bcdde3a89fca489f679b8b0ca22a5")

const transferFoundMinGas uint64 = 21000

var txMinGasPrice = new(big.Int).SetUint64(1)

var address1 = common.HexToAddress("0xbeb1e7f780feb3a5ceb1d6614624737762bd30b5")
var address2 = common.HexToAddress("0x26494c596aca4d317413e708085b89f21bb22d0b")
var zeroAddress = common.Address{}

func createNewState(genesisBalance map[common.Address]*big.Int) State {
	const dbPath string = "/home/mbacic/tmp/leveldb/ethereum"
	_ = os.RemoveAll(dbPath)

	state := NewState(dbPath, nil)
	if len(genesisBalance) > 0 {
		state.SetInitialBalance(genesisBalance)
	}

	return state
}

func Finish(state *State) (common.Hash, error) {
	commitHash, err := state.Commit()

	if err != nil {
		return common.Hash{}, fmt.Errorf("failed create state; %w", err)
	}

	err = state.Persist()
	if err != nil {
		return common.Hash{}, fmt.Errorf("trie commit failed; %w", err)
	}

	err = state.Close()
	if err != nil {
		return common.Hash{}, fmt.Errorf("state close failed; %w", err)
	}

	return commitHash, nil
}

func getDefaultTestGenesisBalance() map[common.Address]*big.Int {
	totalCoinSupply, hasSucceeded := new(big.Int).SetString("1000000000000000000000000000", 10)
	if !hasSucceeded {
		ThrowFatalError("Failed to convert total coin supply to number")
	}

	return map[common.Address]*big.Int{genesisAddress: totalCoinSupply}
}

func TestBalanceSetToGenesisStateBalance() {
	genesisStateBalances := map[common.Address]*big.Int{
		common.HexToAddress("0x1b1747359015847dccbfde729d7793b597249dec"): new(big.Int).SetUint64(100000000000000),
		common.HexToAddress("0x26494c596aca4d317413e708085b89f21bb22d0b"): new(big.Int).SetUint64(200000000000000),
		common.HexToAddress("0xbeb1e7f780feb3a5ceb1d6614624737762bd30b5"): new(big.Int).SetUint64(300000000000000),
	}

	state := createNewState(genesisStateBalances)
	stateRootHash, err := Finish(&state)
	HandleFatalError("Failed to execute TestBalanceSetToGenesisStateBalance", err)

	for address, balance := range genesisStateBalances {
		stateBalance := state.GetBalance(address)
		if balance.Cmp(stateBalance) != 0 {
			ThrowFatalError("Balances are not equal")
		}
	}

	if stateRootHash.Hex() != "0x395ba989f0faebf221a2260d676769e62d2d751e8393d5eeb767ed9716b33bc5" {
		ThrowFatalError("TestBalanceSetToGenesisStateBalance State root hashes differ")
	}

	fmt.Println("SUCCEEDED TestBalanceSetToGenesisStateBalance")
}

func printExecInfo(state *State, execResults []EvmExecutionResult) {
	fmt.Println("State root hash: ", string(state.State.IntermediateRoot(false).Hex()))
	fmt.Println("State dump: ", string(state.State.Dump(nil)))
	fmt.Println()
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
}

func TestSendSomeCoinsAround() {
	state := createNewState(getDefaultTestGenesisBalance())

	block := TolarBlock{
		Hash:              common.HexToHash("0xc15772db6b305d117a0ccdad098a55752ae99388fbb95d103e7bc1b6da7f99d1"),
		Index:             new(big.Int).SetInt64(1),
		Timestamp:         0,
		PreviousBlockHash: common.Hash{},
		Transactions: []*TolarTx{
			{
				Hash:     common.HexToHash("0xc15772db6b305d117a0ccdad098a55752ae99388fbb95d103e7bc1b6da7f99d1"),
				Sender:   genesisAddress,
				Receiver: &address1,
				Value:    new(big.Int).SetUint64(1),
				Gas:      transferFoundMinGas,
				GasPrice: txMinGasPrice,
				Data:     nil,
				Nonce:    0,
			},
		},
	}

	blockProcessor := NewBlockProcessor(state.State)
	_, rootHash := blockProcessor.Process(&block)

	if rootHash.Hex() != "0x91dceb24ed5fb4fb911a49cd3a376fbd177a00171e9ee2c13bfb4d6cedfb1288" {
		ThrowFatalError("State root hashes differ")
	}

	_, err := Finish(&state)
	HandleFatalError("TestSendSomeCoinsAround Failed to commit", err)
	fmt.Println("SUCCEEDED TestSendSomeCoinsAround")
}

func TestSimpleContract() {
	// contract test {
	//  function f(uint a) returns(uint d) { return a * 7; }
	// }

	contractData := common.FromHex(
		strings.Join([]string{
			"0x6080604052341561000f57600080fd5b60b98061001d6000396000f300",
			"608060405260043610603f576000357c01000000000000000000000000",
			"00000000000000000000000000000000900463ffffffff168063b3de64",
			"8b146044575b600080fd5b3415604e57600080fd5b606a600480360381",
			"019080803590602001909291905050506080565b604051808281526020",
			"0191505060405180910390f35b60006007820290509190505600a16562",
			"7a7a72305820f294e834212334e2978c6dd090355312a3f0f9476b8eb9",
			"8fb480406fc2728a960029"}, ""))

	contractBlock := TolarBlock{
		Hash:              common.Hash{},
		Index:             new(big.Int).SetInt64(1),
		Timestamp:         0,
		PreviousBlockHash: common.Hash{},
		Transactions: []*TolarTx{
			{
				Hash:     common.Hash{},
				Sender:   genesisAddress,
				Receiver: nil,
				Value:    new(big.Int).SetUint64(0),
				Gas:      700000,
				GasPrice: txMinGasPrice,
				Data:     contractData,
				Nonce:    0,
			},
		},
	}

	state := createNewState(getDefaultTestGenesisBalance())
	blockProcessor := NewBlockProcessor(state.State)
	execResults, rootHash := blockProcessor.Process(&contractBlock)

	if rootHash.Hex() != "0x1c9ffc57d64486c033ecfec6d450773496b7e69fa7d0a3724583bf9367e5d09d" {
		ThrowFatalError("State root hashes differ")
	}

	deployedContractAddress := execResults[0].Receipt.ContractAddress
	inputData := "0000000000000000000000000000000000000000000000000000000000000001"
	callData := common.FromHex("b3de648b" + inputData)

	contractExecBlock := TolarBlock{
		Hash:              common.Hash{},
		Index:             new(big.Int).SetInt64(1),
		Timestamp:         0,
		PreviousBlockHash: common.Hash{},
		Transactions: []*TolarTx{
			{
				Hash:     common.Hash{},
				Sender:   genesisAddress,
				Receiver: &deployedContractAddress,
				Value:    new(big.Int).SetUint64(0),
				Gas:      80000,
				GasPrice: txMinGasPrice,
				Data:     callData,
				Nonce:    1,
			},
		},
	}

	_, rootHash = blockProcessor.Process(&contractExecBlock)
	if rootHash.Hex() != "0x3b19a6dbecba1cb3f09147d5a37a633a375e348e4999ed325159f7462e5f8aa0" {
		ThrowFatalError("State root hashes differ")
	}

	_, err := Finish(&state)
	HandleFatalError("TestSimpleContract Failed to commit", err)
}

func TestContractSendingCoinsAround() {
	// pragma solidity ^0.4.12;
	//
	// contract Test1 {
	//
	//    address Sender1 = 0xe2eeaACfC6A5488aAcFB2131108AE7b59026Fe4b;
	//    address Sender2 = 0x07EC09F7fd204A835CAE76dc224242451C7aC1DC;
	//    address Sender3 = 0x22a9304E395f0657cFEC6d5E7A95C24702EED158;
	//
	//    function transfer() public {
	//        uint256 current_balance = address(this).balance;
	//        Sender1.transfer(current_balance/4);
	//        Sender2.transfer(current_balance/4);
	//        Sender3.transfer(current_balance/4);
	//    }
	//
	//    function() payable { }
	//}
	contractData := common.FromHex(
		strings.Join([]string{
			"606060405260008054600160a060020a031990811673e2eeaacfc6a5488aacfb2131108a",
			"e7b59026fe4b179091556001805482167307ec09f7fd204a835cae76dc224242451c7ac1",
			"dc179055600280549091167322a9304e395f0657cfec6d5e7a95c24702eed15817905534",
			"1561007557600080fd5b5b610163806100856000396000f3006060604052361561003e57",
			"63ffffffff7c010000000000000000000000000000000000000000000000000000000060",
			"00350416638a4068dd8114610047575b6100455b5b565b005b341561005257600080fd5b",
			"61004561005c565b005b60005473ffffffffffffffffffffffffffffffffffffffff3081",
			"163191166108fc6004835b049081150290604051600060405180830381858888f1935050",
			"505015156100a757600080fd5b60015473ffffffffffffffffffffffffffffffffffffff",
			"ff166108fc6004835b049081150290604051600060405180830381858888f19350505050",
			"15156100ed57600080fd5b60025473ffffffffffffffffffffffffffffffffffffffff16",
			"6108fc6004835b049081150290604051600060405180830381858888f193505050501515",
			"61013357600080fd5b5b505600a165627a7a72305820225752825bc0e51472295b9972f8",
			"97e7840378c2e83f6a6202a04ffd6f02021b0029"}, ""))

	contractBlock := TolarBlock{
		Hash:              common.Hash{},
		Index:             new(big.Int).SetInt64(1),
		Timestamp:         0,
		PreviousBlockHash: common.Hash{},
		Transactions: []*TolarTx{
			{
				Hash:     common.Hash{},
				Sender:   genesisAddress,
				Receiver: nil,
				Value:    new(big.Int).SetUint64(0),
				Gas:      700000,
				GasPrice: txMinGasPrice,
				Data:     contractData,
				Nonce:    0,
			},
		},
	}

	state := createNewState(getDefaultTestGenesisBalance())
	blockProcessor := NewBlockProcessor(state.State)
	execResults, _ := blockProcessor.Process(&contractBlock)

	fmt.Println("Muir Glacier Execution")
	printExecInfo(&state, execResults)

	contractBlock.Index = new(big.Int).SetUint64(35)
	contractBlock.Transactions[0].Nonce = 1
	execResults, _ = blockProcessor.Process(&contractBlock)

	fmt.Println("All protocols Execution")
	printExecInfo(&state, execResults)

	//if rootHash.Hex() != "0x1c9ffc57d64486c033ecfec6d450773496b7e69fa7d0a3724583bf9367e5d09d" {
	//	ThrowFatalError("State root hashes differ")
	//}

	//deployedContractAddress := execResults[0].Receipt.ContractAddress
	//inputData := "0000000000000000000000000000000000000000000000000000000000000001"
	//callData := common.FromHex("b3de648b" + inputData)
	//
	//contractExecBlock := TolarBlock{
	//	Hash:              common.Hash{},
	//	Index:             new(big.Int).SetInt64(1),
	//	Timestamp:         0,
	//	PreviousBlockHash: common.Hash{},
	//	Transactions: []*TolarTx{
	//		{
	//			Hash:     common.Hash{},
	//			Sender:   genesisAddress,
	//			Receiver: &deployedContractAddress,
	//			Value:    new(big.Int).SetUint64(0),
	//			Gas:      80000,
	//			GasPrice: txMinGasPrice,
	//			Data:     callData,
	//			Nonce:    1,
	//		},
	//	},
	//}
	//
	//_, rootHash = blockProcessor.Process(&contractExecBlock)
	//if rootHash.Hex() != "0x3b19a6dbecba1cb3f09147d5a37a633a375e348e4999ed325159f7462e5f8aa0" {
	//	ThrowFatalError("State root hashes differ")
	//}
	//
	//_, err := Finish(&state)
	//HandleFatalError("TestSimpleContract Failed to commit", err)
}
