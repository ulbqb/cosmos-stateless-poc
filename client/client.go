package client

import (
	"fmt"

	"github.com/cosmos/iavl"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/types"
)

type StatelessClient struct {
	app    interface{}
	oracle iavl.OracleClientI
}

func NewStatelessClient(app interface{}, oracle iavl.OracleClientI) (*StatelessClient, error) {
	switch app.(type) {
	case CosmosBaseApp:
		break
	default:
		return nil, fmt.Errorf("this application type is not supported")
	}

	return &StatelessClient{
		app:    app,
		oracle: oracle,
	}, nil
}

// TODO: cannnot execute initial height block
func (c *StatelessClient) Execute(block *types.Block, vals []*types.Validator) ([]byte, ExecutionLog, error) {
	log := ExecutionLog{}
	cosmos := getCosmosApp(c.app)
	if cosmos == nil {
		return nil, log, fmt.Errorf("this application type is not supported")
	}

	// convert to stateless app
	var stateless CosmosBaseApp
	stateless, err := cosmos.StatelessApp(block.Height, c.oracle)
	if err != nil {
		return nil, log, err
	}

	// initialize chain
	var abcivu []abci.ValidatorUpdate
	if vals != nil {
		abcivu = types.TM2PB.ValidatorUpdates(types.NewValidatorSet(vals))
	}
	stateless.InitChain(abci.RequestInitChain{
		Time:    block.Time,
		ChainId: block.ChainID,
		// ConsensusParams: nil, // ConsensusParams is not needed as it comes from oracle.
		Validators: abcivu,
		// AppStateBytes: nil, // AppStateBytes is not needed as it comes from oracle.
		InitialHeight: block.Height,
	})

	// begin block
	byzVals := make([]abci.Evidence, 0)
	for _, evidence := range block.Evidence.Evidence {
		byzVals = append(byzVals, evidence.ABCI()...)
	}
	log.ResponseBeginBlock = stateless.BeginBlock(abci.RequestBeginBlock{
		Hash:                block.Hash(),
		Header:              *block.Header.ToProto(),
		LastCommitInfo:      getBeginBlockValidatorInfo(block, vals, 0),
		ByzantineValidators: byzVals,
	})

	// deliver txs
	for _, tx := range block.Data.Txs {
		res := stateless.DeliverTx(abci.RequestDeliverTx{
			Tx: tx,
		})
		log.ResponseDeliverTxs = append(log.ResponseDeliverTxs, res)
	}

	// end block
	log.ResponseEndBlock = stateless.EndBlock(abci.RequestEndBlock{
		Height: block.Header.Height,
	})

	// commit
	log.ResponseCommit = stateless.Commit()
	appHash := log.ResponseCommit.Data

	// output
	return appHash, log, nil
}

func getBeginBlockValidatorInfo(block *types.Block, vals []*types.Validator, initialHeight int64) abci.LastCommitInfo {
	voteInfos := make([]abci.VoteInfo, block.LastCommit.Size())
	// Initial block -> LastCommitInfo.Votes are empty.
	// Remember that the first LastCommit is intentionally empty, so it makes
	// sense for LastCommitInfo.Votes to also be empty.
	if block.Height > initialHeight {
		// Sanity check that commit size matches validator set size - only applies
		// after first block.
		var (
			commitSize = block.LastCommit.Size()
			valSetLen  = len(vals)
		)
		if commitSize != valSetLen {
			panic(fmt.Sprintf(
				"commit size (%d) doesn't match valset length (%d) at height %d\n\n%v\n\n%v",
				commitSize, valSetLen, block.Height, block.LastCommit.Signatures, vals,
			))
		}

		for i, val := range vals {
			commitSig := block.LastCommit.Signatures[i]
			voteInfos[i] = abci.VoteInfo{
				Validator:       types.TM2PB.Validator(val),
				SignedLastBlock: !commitSig.Absent(),
			}
		}
	}

	return abci.LastCommitInfo{
		Round: block.LastCommit.Round,
		Votes: voteInfos,
	}
}

type ExecutionLog struct {
	ResponseBeginBlock abci.ResponseBeginBlock
	ResponseDeliverTxs []abci.ResponseDeliverTx
	ResponseEndBlock   abci.ResponseEndBlock
	ResponseCommit     abci.ResponseCommit
}
