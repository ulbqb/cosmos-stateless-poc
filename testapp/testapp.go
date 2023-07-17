package testapp

import (
	"encoding/json"
	"fmt"
	"io"
	"math/rand"

	"github.com/cosmos/cosmos-sdk/baseapp"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/simapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/libs/log"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	"github.com/tendermint/tendermint/types"
	dbm "github.com/tendermint/tm-db"
)

var (
	capKey1 = sdk.NewKVStoreKey("key1")
	capKey2 = sdk.NewKVStoreKey("key2")
)

func NewTestApp() (*baseapp.BaseApp, error) {
	encCfg := simapp.MakeTestEncodingConfig()
	RegisterInterfaces(encCfg.InterfaceRegistry)
	app := baseapp.NewBaseApp("testapp", log.NewTMLogger(log.NewSyncWriter(io.Discard)), dbm.NewMemDB(), encCfg.TxConfig.TxDecoder())
	app.SetInterfaceRegistry(encCfg.InterfaceRegistry)
	RegisterMsgServer(
		app.MsgServiceRouter(),
		MsgServerImpl{},
	)

	app.MountStores(capKey1, capKey2)
	app.SetParamStore(&paramStore{db: dbm.NewMemDB()})

	// stores are mounted
	err := app.LoadLatestVersion()
	if err != nil {
		return nil, err
	}
	return app, nil
}

func RegisterInterfaces(registry codectypes.InterfaceRegistry) {
	msgservice.RegisterMsgServiceDesc(registry, &_Msg_serviceDesc)
}

type paramStore struct {
	db *dbm.MemDB
}

func (ps *paramStore) Set(_ sdk.Context, key []byte, value interface{}) {
	bz, err := json.Marshal(value)
	if err != nil {
		panic(err)
	}

	ps.db.Set(key, bz)
}

func (ps *paramStore) Has(_ sdk.Context, key []byte) bool {
	ok, err := ps.db.Has(key)
	if err != nil {
		panic(err)
	}

	return ok
}

func (ps *paramStore) Get(_ sdk.Context, key []byte, ptr interface{}) {
	bz, err := ps.db.Get(key)
	if err != nil {
		panic(err)
	}

	if len(bz) == 0 {
		return
	}

	if err := json.Unmarshal(bz, ptr); err != nil {
		panic(err)
	}
}

func ExecuteBlockWithTxs(app *baseapp.BaseApp, numTransactions int, blockHeight int64, r *rand.Rand) (*types.Block, error) {
	app.BeginBlock(abci.RequestBeginBlock{Header: tmproto.Header{Height: blockHeight}})

	encCfg := simapp.MakeTestEncodingConfig()
	txs := types.Txs{}
	for txNum := 0; txNum < numTransactions; txNum++ {
		txBuilder := encCfg.TxConfig.NewTxBuilder()

		key := make([]byte, 1)
		_, err := r.Read(key)
		if err != nil {
			return nil, err
		}
		value := make([]byte, 10)
		_, err = r.Read(value)
		if err != nil {
			return nil, err
		}
		sord := make([]byte, 1)
		_, err = r.Read(sord)
		if err != nil {
			return nil, err
		}
		if sord[0]%8 == 0 {
			err = txBuilder.SetMsgs(&MsgRemove{Key: key})
			if err != nil {
				return nil, err
			}
		} else if sord[0]%8 == 1 {
			err = txBuilder.SetMsgs(&MsgGet{Key: key})
			if err != nil {
				return nil, err
			}
		} else {
			err = txBuilder.SetMsgs(&MsgSet{Key: key, Value: value})
			if err != nil {
				return nil, err
			}
		}

		txBytes, err := encCfg.TxConfig.TxEncoder()(txBuilder.GetTx())
		if err != nil {
			return nil, err
		}

		resp := app.DeliverTx(abci.RequestDeliverTx{Tx: txBytes})
		if !resp.IsOK() {
			return nil, fmt.Errorf(resp.String())
		}
		txs = append(txs, txBytes)
	}

	app.EndBlock(abci.RequestEndBlock{Height: blockHeight})

	block := &types.Block{
		Header: types.Header{
			Height: blockHeight,
		},
		Data: types.Data{
			Txs: txs,
		},
		Evidence: types.EvidenceData{
			Evidence: []types.Evidence{},
		},
		LastCommit: &types.Commit{},
	}
	return block, nil
}
