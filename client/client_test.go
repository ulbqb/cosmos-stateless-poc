package client

import (
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/types"

	occlient "github.com/ulbqb/cosmos-stateless-poc/oracle/client"
	ocserver "github.com/ulbqb/cosmos-stateless-poc/oracle/server"
	"github.com/ulbqb/cosmos-stateless-poc/testapp"
)

func TestExecuteStateless(t *testing.T) {
	for i := range make([]int, 100) {
		t.Run(fmt.Sprintf("random seed %d", i), func(t *testing.T) {
			executeStatelessWithLocalApp(t, int64(i))
		})
	}
}

func TestExecuteStatelessRandom(t *testing.T) {
	for range make([]int, 100) {
		seed := time.Now().UnixNano()
		t.Run(fmt.Sprintf("random seed %d", seed), func(t *testing.T) {
			executeStatelessWithLocalApp(t, seed)
		})
	}
}

func executeStatelessWithLocalApp(t *testing.T, seed int64) {
	// setup oracle server
	app, err := testapp.NewTestApp()
	require.NoError(t, err)
	app.InitChain(abci.RequestInitChain{})
	challengeHeihgt := int64(128)
	challengeBlock := &types.Block{}
	challengeAppHash := []byte{}
	agreementAppHash := []byte{}
	r := rand.New(rand.NewSource(seed))
	for i := range make([]int, challengeHeihgt) {
		challengeBlock, err = testapp.ExecuteBlockWithTxs(app, 8, int64(i)+1, r)
		require.NoError(t, err)
		res := app.Commit()
		challengeAppHash = res.Data
		if i+1 == int(challengeHeihgt)-1 {
			agreementAppHash = res.Data
		}
	}
	server := ocserver.NewLocalOracleServer(app, challengeBlock, nil, nil)

	// setup oracle client
	client := occlient.NewLocalOracleClient(server)

	// setup stateless client
	newapp, err := testapp.NewTestApp()
	require.NoError(t, err)
	stateless, err := NewStatelessClient(newapp, client)
	require.NoError(t, err)

	// execute stateless
	resultBlock := client.Block()
	resultVals := client.Validators()
	executedAppHash, _, err := stateless.Execute(resultBlock.Block, resultVals.Validators)
	require.NoError(t, err)

	// test
	require.NotEqual(t, agreementAppHash, executedAppHash)
	require.Equal(t, challengeAppHash, executedAppHash)
}
