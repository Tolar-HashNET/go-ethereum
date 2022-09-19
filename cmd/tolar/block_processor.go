package main

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/state"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/params"
	"math"
	"math/big"
)

type BlockProcessor struct {
	stateDb     *state.StateDB
	vmConfig    *vm.Config
	chainConfig params.ChainConfig // Chain configuration options
}

func createTolarHashnetConfig() params.ChainConfig {
	return params.ChainConfig{
		ChainID:             big.NewInt(0),
		HomesteadBlock:      big.NewInt(0),
		EIP150Block:         big.NewInt(0),
		EIP155Block:         big.NewInt(0),
		EIP158Block:         big.NewInt(0),
		ByzantiumBlock:      big.NewInt(0),
		ConstantinopleBlock: big.NewInt(0),
		PetersburgBlock:     big.NewInt(0),
		IstanbulBlock:       big.NewInt(0),
		MuirGlacierBlock:    big.NewInt(0), // Tolar current fork
		BerlinBlock:         big.NewInt(34),
		LondonBlock:         big.NewInt(34),
		ArrowGlacierBlock:   big.NewInt(34),
		GrayGlacierBlock:    big.NewInt(34),
		Ethash:              new(params.EthashConfig),
	}
}

func NewBlockProcessor(stateDb *state.StateDB) BlockProcessor {
	//var chainConfig = *params.AllEthashProtocolChanges
	chainConfig := createTolarHashnetConfig()
	//!!! Missing accountStartNonce !!!
	chainConfig.ChainID = big.NewInt(0x00)

	return BlockProcessor{
		stateDb: stateDb,
		vmConfig: &vm.Config{
			Debug:                   false, // Enables debugging
			Tracer:                  nil,   // Opcode logger
			NoBaseFee:               false, // Disables call, callcode, delegate call and create
			EnablePreimageRecording: false, // Enables recording of SHA3/keccak preimages
			ExtraEips:               nil,   // Additional EIPS that are to be enabled
		},
		chainConfig: chainConfig,
	}
}

func getDummyHash(blockIndex uint64) common.Hash {
	return common.Hash{}
}

func createTestBlockContext(blockNumber *big.Int, blockTimestamp *big.Int) vm.BlockContext {
	return vm.BlockContext{
		CanTransfer: core.CanTransfer,
		Transfer:    core.Transfer,
		GetHash:     getDummyHash,
		Coinbase:    common.HexToAddress("0xa05ea6c81c5d2ca0bf0581974d7c7d393998d4fe"), // Gas collection address. 54a05ea6c81c5d2ca0bf0581974d7c7d393998d4fedf37cede
		BlockNumber: blockNumber,
		Time:        blockTimestamp,
		Difficulty:  new(big.Int).SetInt64(0),
		BaseFee:     new(big.Int).SetInt64(0),
		GasLimit:    uint64(math.MaxUint64),
		Random:      nil,
	}
}

func createBlockContext(blockIndex *big.Int, blockTimestamp *big.Int) vm.BlockContext {
	const maxBlockGasLimit uint64 = 10000000

	return vm.BlockContext{
		CanTransfer: core.CanTransfer,
		Transfer:    core.Transfer,
		GetHash:     getDummyHash,
		Coinbase:    common.HexToAddress("0xa05ea6c81c5d2ca0bf0581974d7c7d393998d4fe"), // Gas collection address. 54a05ea6c81c5d2ca0bf0581974d7c7d393998d4fedf37cede
		BlockNumber: blockIndex,
		Time:        blockTimestamp,
		Difficulty:  new(big.Int).SetInt64(131072),
		BaseFee:     new(big.Int).SetInt64(0),
		GasLimit:    maxBlockGasLimit * 2, // max_block_gas_limit_ * 2, DEFAULT_MAX_BLOCK_GAS_LIMIT{"10000000"}
		Random:      nil,
	}
}

type EvmExecutionResult struct {
	Receipt *types.Receipt
	Error   error
}

// Process processes the state changes according to the Ethereum rules by running
// the transaction messages using the stateDb and applying any rewards to both
// the processor (coinbase) and any included uncles.
//
// Process returns the receipts and logs accumulated during the process and
// returns the amount of gas that was used in the process. If any of the
// transactions failed to execute due to insufficient gas it will return an error.
// (types.Receipts, []*types.Log, uint64, error) {
func (b *BlockProcessor) Process(block *TolarBlock) ([]EvmExecutionResult, common.Hash) {
	blockContext := createTestBlockContext(block.Index, new(big.Int).SetUint64(block.Timestamp))

	//blockContext := createBlockContext(block.Index, new(big.Int).SetUint64(block.Timestamp))
	evm := vm.NewEVM(blockContext, vm.TxContext{}, b.stateDb, &b.chainConfig, *b.vmConfig)

	var gasLimit = new(core.GasPool).AddGas(blockContext.GasLimit)
	var usedGas uint64 = 0

	var execResults []EvmExecutionResult

	// Iterate over and process the individual transactions
	activeSnapshotId := b.stateDb.Snapshot()
	for i, tx := range block.Transactions {
		b.stateDb.Prepare(tx.Hash, i)
		execResult := b.applyTransaction(tx, gasLimit, block.Index, block.Hash, &usedGas, evm)
		execResults = append(execResults, execResult)

		if execResult.Error == nil {
			activeSnapshotId = b.stateDb.Snapshot()
		} else {
			b.stateDb.RevertToSnapshot(activeSnapshotId)
		}
	}

	return execResults, b.stateDb.IntermediateRoot(true)
}

func (b *BlockProcessor) applyTransaction(tx *TolarTx, gasPool *core.GasPool, blockIndex *big.Int, blockHash common.Hash, usedGas *uint64, evm *vm.EVM) EvmExecutionResult {
	// Create a new context to be used in the EVM environment.
	msg := tx.ToMessage()
	txContext := core.NewEVMTxContext(msg)
	evm.Reset(txContext, b.stateDb)

	// Apply the transaction to the current state (included in the env).
	result, err := core.ApplyMessage(evm, msg, gasPool)
	if err != nil {
		return EvmExecutionResult{nil, err}
	}

	// Update the state with pending changes.
	var root = b.stateDb.IntermediateRoot(true)
	*usedGas += result.UsedGas

	// Create a new receipt for the transaction, storing the intermediate root and gas used
	// by the tx.
	receipt := &types.Receipt{Type: types.LegacyTxType, PostState: root.Bytes(), CumulativeGasUsed: *usedGas}
	if result.Failed() {
		receipt.Status = types.ReceiptStatusFailed
	} else {
		receipt.Status = types.ReceiptStatusSuccessful
	}
	receipt.TxHash = tx.Hash
	receipt.GasUsed = result.UsedGas

	// If the transaction created a contract, store the creation address in the receipt.
	if msg.To() == nil {
		receipt.ContractAddress = crypto.CreateAddress(evm.TxContext.Origin, tx.Nonce)
	}

	// Set the receipt logs and create the bloom filter.
	receipt.Logs = b.stateDb.GetLogs(tx.Hash, blockHash)
	receipt.Bloom = types.CreateBloom(types.Receipts{receipt})
	receipt.BlockHash = blockHash
	receipt.BlockNumber = blockIndex
	receipt.TransactionIndex = uint(b.stateDb.TxIndex())
	return EvmExecutionResult{receipt, err}
}
