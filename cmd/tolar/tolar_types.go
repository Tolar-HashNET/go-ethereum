package main

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"math/big"
)

type TolarTx struct {
	Hash     common.Hash
	Sender   common.Address
	Receiver *common.Address
	Value    *big.Int
	Gas      uint64
	GasPrice *big.Int
	Data     []byte
	Nonce    uint64
}

func (t *TolarTx) ToMessage() types.Message {
	return types.NewMessage(
		t.Sender,
		t.Receiver,
		t.Nonce,
		t.Value,
		t.Gas,
		t.GasPrice,
		new(big.Int).SetUint64(0),
		new(big.Int).SetUint64(0),
		t.Data,
		types.AccessList{},
		false)
}

type TolarBlock struct {
	Hash              common.Hash
	Index             *big.Int
	Timestamp         uint64
	PreviousBlockHash common.Hash

	Transactions []*TolarTx
}
