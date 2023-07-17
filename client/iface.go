package client

import (
	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/iavl"
	abci "github.com/tendermint/tendermint/abci/types"
)

var (
	_ CosmosBaseApp = &baseapp.BaseApp{}
)

type CosmosBaseApp interface {
	InitChain(req abci.RequestInitChain) (res abci.ResponseInitChain)
	BeginBlock(req abci.RequestBeginBlock) (res abci.ResponseBeginBlock)
	DeliverTx(req abci.RequestDeliverTx) (res abci.ResponseDeliverTx)
	EndBlock(req abci.RequestEndBlock) (res abci.ResponseEndBlock)
	Commit() (res abci.ResponseCommit)
	StatelessApp(version int64, oracle iavl.OracleClientI) (app *baseapp.BaseApp, err error)
}

func getCosmosApp(app interface{}) CosmosBaseApp {
	common, ok := app.(CosmosBaseApp)
	if !ok {
		return nil
	}
	return common
}
