package testapp

import (
	"math/rand"
	"testing"

	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/tendermint/abci/types"
)

func TestTestApp(t *testing.T) {
	app, err := NewTestApp()
	require.NoError(t, err)

	app.InitChain(abci.RequestInitChain{})
	challengeHeihgt := int64(100)
	r := rand.New(rand.NewSource(int64(0)))
	for i := range make([]int, challengeHeihgt) {
		_, err = ExecuteBlockWithTxs(app, 5, int64(i)+1, r)
		require.NoError(t, err)
		app.Commit()
	}
}
