package server

import (
	"encoding/hex"
	"encoding/json"
	"net/url"

	"github.com/cosmos/cosmos-sdk/baseapp"
	abci "github.com/tendermint/tendermint/abci/types"
	ocjson "github.com/tendermint/tendermint/libs/json"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	ctypes "github.com/tendermint/tendermint/rpc/core/types"
	"github.com/tendermint/tendermint/types"
)

var _ OracleServer = LocalOracleServer{}

type LocalOracleServer struct {
	app   *baseapp.BaseApp
	block *types.Block
	vals  []*types.Validator
	cp    *tmproto.ConsensusParams
}

func NewLocalOracleServer(app *baseapp.BaseApp, block *types.Block, vals []*types.Validator, cp *tmproto.ConsensusParams) *LocalOracleServer {
	if cp == nil {
		cp = &tmproto.ConsensusParams{}
	}
	return &LocalOracleServer{
		app:   app,
		block: block,
		vals:  vals,
		cp:    cp,
	}
}

func (s LocalOracleServer) Get(key []byte) []byte {
	u, err := url.Parse(string(key))
	if err != nil {
		panic(err)
	}
	switch u.Path {
	case "block":
		result := ctypes.ResultBlock{
			Block: s.block,
		}
		return toRawJson(result)
	case "consensus_params":
		result := ctypes.ResultConsensusParams{
			BlockHeight:     s.block.Height - 1,
			ConsensusParams: *s.cp,
		}
		return toRawJson(result)
	case "validators":
		result := ctypes.ResultValidators{
			BlockHeight: s.block.Height,
			Validators:  s.vals,
			Count:       len(s.vals),
			Total:       len(s.vals),
		}
		return toRawJson(result)
	case "abci_query":
		m, err := url.ParseQuery(u.RawQuery)
		if err != nil {
			panic(err)
		}
		data, err := hex.DecodeString(m["data"][0])
		if err != nil {
			panic(err)
		}

		res := s.app.Query(abci.RequestQuery{
			Data:   data,
			Path:   m["path"][0],
			Height: s.block.Header.Height - 1,
			Prove:  true,
		})

		result := ctypes.ResultABCIQuery{
			Response: res,
		}

		return toRawJson(result)
	default:
		panic("not supported")
	}
}

func toRawJson(v interface{}) []byte {
	js, err := ocjson.Marshal(v)
	if err != nil {
		panic(err)
	}
	rawMsg := json.RawMessage(js)
	return rawMsg
}
