package main

/*
ChainParams:
{
    "sealEngine": "Ethash",
    "params": {
        "accountStartNonce": "0x00",
        "homesteadForkBlock": "0x00",
        "EIP150ForkBlock": "0x00",
        "EIP158ForkBlock": "0x00",
        "byzantiumForkBlock": "0x00",
        "constantinopleForkBlock": "0x00",
        "constantinopleFixForkBlock": "0x00",
        "istanbulForkBlock": "0x00",
        "muirGlacierForkBlock": "0x00",
        "networkID" : "0x1",
        "chainID": "0x01",
        "minimumDifficulty": "0x020000",
        "minGasLimit": "0x1388",
        "maxGasLimit": "7fffffffffffffff",
        "gasLimitBoundDivisor": "0x0400",
        "difficultyBoundDivisor": "0x0800",
        "durationLimit": "0x0d",
        "tieBreakingGas": false,
        "blockReward": "0x4563918244F40000",
        "maximumExtraDataSize": "0x20",
        "allowFutureBlocks" : true
    },
    "genesis": {
        "nonce": "0x0000000000000042",
        "difficulty": "0x020000",
        "mixHash": "0x0000000000000000000000000000000000000000000000000000000000000000",
        "author": "0x0000000000000000000000000000000000000000",
        "timestamp": "0x00",
        "parentHash": "0x0000000000000000000000000000000000000000000000000000000000000000",
        "extraData": "",
        "gasLimit": "0x1388"
    },
    "accounts": {
    }

genesis_info.h
*/

/*
dev::eth::State
State(u256 const& _accountStartNonce, OverlayDB const& _db, BaseState _bs = BaseState::PreExisting);
_accountStartNonce = ev::u256(0), db_, dev::eth::BaseState::PreExisting
tate->setRoot(dev::h256(state_root_hash));


state_executor.cc
*/

/*
dev::eth::EnvInfo

  auto block_header = dev::eth::BlockHeader();
  block_header.setGasLimit(max_block_gas_limit_ * 2);
  block_header.setTimestamp(block_timestamp);
  block_header.setDifficulty(0);
  block_header.setNumber(next_index_);
  block_header.setAuthor(evm::DEFAULT_GAS_COLLECTOR.ToEthAddress());
  block_header.setParentHash(previous_block_hash);

  return {block_header, last_block_hashes_, 0, network_id_};


dev::eth::EnvInfo Blockchain::CreateLatestEnvInfo(
    std::uint64_t block_timestamp,
    const dev::h256& previous_block_hash) {
  auto block_header = dev::eth::BlockHeader();
  block_header.setGasLimit(max_block_gas_limit_ * 2);
  block_header.setTimestamp(block_timestamp);
  block_header.setDifficulty(0);
  block_header.setNumber(next_index_);
  block_header.setAuthor(evm::DEFAULT_GAS_COLLECTOR.ToEthAddress());
  block_header.setParentHash(previous_block_hash);

  return {block_header, last_block_hashes_, 0, network_id_};
}

blockchain.cc
*/

/*
  dev::h256s precedingHashes(dev::h256 const& most_recent_hash) const override;
  void clear() override;


*/

/*
  if (!std::filesystem::exists(state_path)) {
    spdlog::info("Creating EVM storage directory");
    std::filesystem::create_directories(state_path);
  }

  dev::h256 genesis_hash = dev::sha3(dev::rlp(""));
  dev::db::setDatabaseKind(dev::db::DatabaseKind::RocksDB);
  db_ = std::make_unique<dev::OverlayDB>(
      dev::eth::State::openDB(state_path.string(), genesis_hash, dev::WithExisting::Trust));

  Initialize();

// state_factory.cc
*/
