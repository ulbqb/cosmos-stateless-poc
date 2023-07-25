package exec

import (
	"io"
	"io/ioutil"

	"github.com/cosmos/cosmos-sdk/client/flags"
	csmsserver "github.com/cosmos/cosmos-sdk/server"
	"github.com/cosmos/cosmos-sdk/store/types"
	"github.com/tendermint/tendermint/libs/log"
	dbm "github.com/tendermint/tm-db"
	"github.com/ulbqb/cosmos-stateless-poc/client"
	slclient "github.com/ulbqb/cosmos-stateless-poc/client"
	occlient "github.com/ulbqb/cosmos-stateless-poc/oracle/client"
	ocserver "github.com/ulbqb/cosmos-stateless-poc/oracle/server"
)

func Execute(basedir string, trustHeight int, trustBlockHash string, rpcAddr string) ([]byte, *client.ExecutionLog, error) {
	// setup oracle server
	server, err := ocserver.NewRPCOracleServer(trustHeight, trustBlockHash, rpcAddr, basedir)
	if err != nil {
		return nil, nil, err
	}

	// setup oracle client
	client := occlient.NewLocalOracleClient(server)

	// setup stateless client
	db := dbm.NewMemDB()
	tempDir, err := ioutil.TempDir("/tmp", "gaiasl")
	if err != nil {
		return nil, nil, err
	}
	ctx := csmsserver.NewDefaultContext()
	ctx.Viper.Set(flags.FlagHome, tempDir)
	ctx.Viper.Set(csmsserver.FlagPruning, types.PruningOptionNothing)
	gaia := newApp(log.NewTMLogger(log.NewSyncWriter(io.Discard)), db, nil, ctx.Viper)
	stateless, err := slclient.NewStatelessClient(gaia, client)
	if err != nil {
		return nil, nil, err
	}

	// execute stateless
	resultBlock := client.Block()
	resultVals := client.Validators()
	appHash, log, err := stateless.Execute(resultBlock.Block, resultVals.Validators)
	if err != nil {
		return nil, &log, err
	}
	return appHash, &log, nil
}
