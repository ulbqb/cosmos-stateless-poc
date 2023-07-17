package client

import (
	"encoding/json"

	"github.com/cosmos/iavl"
	tmjson "github.com/tendermint/tendermint/libs/json"
	ctypes "github.com/tendermint/tendermint/rpc/core/types"

	"github.com/ulbqb/cosmos-stateless-poc/oracle/server"
)

type LocalOracleClient struct {
	server server.OracleServer
}

var _ iavl.OracleClientI = LocalOracleClient{}

func NewLocalOracleClient(server server.OracleServer) *LocalOracleClient {
	return &LocalOracleClient{
		server: server,
	}
}

func (c LocalOracleClient) Get(key []byte) []byte {
	return c.server.Get(key)
}

func (o LocalOracleClient) Block() *ctypes.ResultBlock {
	b := o.Get([]byte("block"))
	block := ctypes.ResultBlock{}
	if err := tmjson.Unmarshal(b, &block); err != nil {
		panic(err)
	}
	return &block
}

func (o LocalOracleClient) ConsensusParams() *ctypes.ResultConsensusParams {
	b := o.Get([]byte("consensus_params"))
	cp := ctypes.ResultConsensusParams{}
	if err := tmjson.Unmarshal(b, &cp); err != nil {
		panic(err)
	}
	return &cp
}

func (o LocalOracleClient) Validators() *ctypes.ResultValidators {
	b := o.Get([]byte("validators"))
	raw := json.RawMessage(b)
	vals := ctypes.ResultValidators{}
	if err := tmjson.Unmarshal(raw, &vals); err != nil {
		panic(err)
	}
	return &vals
}
