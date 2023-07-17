package server

import (
	"bytes"
	"context"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"os"

	ocjson "github.com/tendermint/tendermint/libs/json"
	rpcclient "github.com/tendermint/tendermint/rpc/client"
	rpchttp "github.com/tendermint/tendermint/rpc/client/http"
	ctypes "github.com/tendermint/tendermint/rpc/core/types"
	octypes "github.com/tendermint/tendermint/types"
)

var _ OracleServer = &RPCOracleServer{}

type RPCOracleServer struct {
	// rpc
	rpc *CacheHttp
	// trusted data
	trustHeight    int64
	trustBlockHash []byte
	// data verified by trusted data
	verifiedCommit     *ctypes.ResultCommit
	verifiedValidators *ctypes.ResultValidators
	verifiedBlock      *ctypes.ResultBlock
}

func NewRPCOracleServer(trustHeight int, trustBlockHash string, rpcAddr string, basedir string) (*RPCOracleServer, error) {
	trustHashBytes, err := hex.DecodeString(trustBlockHash)
	if err != nil {
		return nil, err
	}

	c, err := rpchttp.New(rpcAddr, "/websocket")
	if err != nil {
		return nil, err
	}

	basedir = fmt.Sprintf("%s/output/%d", basedir, trustHeight)

	server := RPCOracleServer{
		rpc:            NewCacheHttp(c, basedir),
		trustHeight:    int64(trustHeight),
		trustBlockHash: trustHashBytes,
	}

	err = server.setVerifiedBlock()
	if err != nil {
		return nil, err
	}
	err = server.setVerifiedCommit()
	if err != nil {
		return nil, err
	}
	err = server.setVerifiedValidators()
	if err != nil {
		return nil, err
	}

	return &server, nil
}

func (s *RPCOracleServer) Get(key []byte) []byte {
	u, err := url.Parse(string(key))
	if err != nil {
		panic(err)
	}
	switch u.Path {
	case "block":
		return toRawJson(s.verifiedBlock)
	case "validators":
		return toRawJson(s.verifiedValidators)
	case "abci_query":
		res, err := s.getVerifiedABCIQuery(u)
		if err != nil {
			panic(err)
		}
		return toRawJson(res)
	default:
		panic("not supported")
	}
}

func (s *RPCOracleServer) setVerifiedBlock() error {
	resultBlock, err := s.rpc.Block(&s.trustHeight)
	if err != nil {
		return err
	}

	if err = resultBlock.BlockID.ValidateBasic(); err != nil {
		return err
	}

	if err = resultBlock.Block.ValidateBasic(); err != nil {
		return err
	}

	blockId := octypes.BlockID{Hash: resultBlock.Block.Hash(), PartSetHeader: resultBlock.Block.MakePartSet(65536).Header()}
	if !resultBlock.BlockID.Equals(blockId) {
		return errors.New("block id does not match")
	}

	// verify ResultBlock
	if !bytes.Equal(s.trustBlockHash, resultBlock.BlockID.Hash) {
		return errors.New("block hash does not match")
	}

	s.verifiedBlock = resultBlock

	return nil
}

func (s *RPCOracleServer) setVerifiedCommit() error {
	if s.verifiedBlock == nil {
		return errors.New("verified block is nil")
	}

	preHeight := s.trustHeight - 1
	res, err := s.rpc.Commit(&preHeight)
	if err != nil {
		return err
	}

	if err = res.ValidateBasic(s.verifiedBlock.Block.ChainID); err != nil {
		return err
	}

	// verify ResultCommit
	if !bytes.Equal(s.verifiedBlock.Block.LastCommitHash, res.Commit.Hash()) {
		return errors.New("last commit hash does not match")
	}

	s.verifiedCommit = res

	return nil
}

func (s *RPCOracleServer) setVerifiedValidators() error {
	if s.verifiedCommit == nil {
		return errors.New("verified commit is nil")
	}

	preHeight := s.trustHeight - 1
	page := 1
	perPage := 100
	vals := ctypes.ResultValidators{
		BlockHeight: preHeight,
	}
	for {
		res, err := s.rpc.Validators(&preHeight, &page, &perPage)
		if err != nil {
			return err
		}
		vals.Validators = append(vals.Validators, res.Validators...)
		vals.Count = vals.Count + res.Count
		vals.Total = res.Total
		if vals.Count == vals.Total {
			break
		}
		page++
	}

	// verify ResultValidators
	if !bytes.Equal(s.verifiedCommit.ValidatorsHash, octypes.NewValidatorSet(vals.Validators).Hash()) {
		return errors.New("validators is not verified")
	}

	s.verifiedValidators = &vals

	return nil
}

func (s *RPCOracleServer) getVerifiedABCIQuery(u *url.URL) (*ctypes.ResultABCIQuery, error) {
	if s.verifiedBlock == nil {
		return nil, errors.New("verified block is nil")
	}

	m, err := url.ParseQuery(u.RawQuery)
	if err != nil {
		return nil, err
	}

	data, err := hex.DecodeString(m["data"][0])
	if err != nil {
		return nil, err
	}

	path := m["path"][0]

	opts := rpcclient.ABCIQueryOptions{
		Height: s.trustHeight - 1,
		Prove:  true,
	}
	res, err := s.rpc.ABCIQueryWithOptions(path, []byte(data), opts)
	if err != nil {
		return nil, err
	}

	// TODO: verify ResultABCIQuery

	return res, nil
}

type CacheHttp struct {
	rpc     *rpchttp.HTTP
	basedir string
}

func NewCacheHttp(rpc *rpchttp.HTTP, basedir string) *CacheHttp {
	if err := os.MkdirAll(basedir, os.ModePerm); err != nil {
		panic(err)
	}
	return &CacheHttp{
		rpc:     rpc,
		basedir: basedir,
	}
}

func (h CacheHttp) Block(height *int64) (*ctypes.ResultBlock, error) {
	fileName := fmt.Sprintf("%s/block?height=%d.json", h.basedir, *height)

	fileData, err := os.ReadFile(fileName)
	if !errors.Is(err, os.ErrNotExist) {
		raw := json.RawMessage(fileData)
		result := ctypes.ResultBlock{}
		if err := ocjson.Unmarshal(raw, &result); err != nil {
			return nil, err
		}
		return &result, nil
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	result, err := h.rpc.Block(ctx, height)
	if err != nil {
		return nil, err
	}

	file, err := os.Create(fileName)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	_, err = file.WriteString(string(toRawJson(result)))
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (h CacheHttp) Commit(height *int64) (*ctypes.ResultCommit, error) {
	fileName := fmt.Sprintf("%s/commit?height=%d.json", h.basedir, *height)

	fileData, err := os.ReadFile(fileName)
	if !errors.Is(err, os.ErrNotExist) {
		raw := json.RawMessage(fileData)
		result := ctypes.ResultCommit{}
		if err := ocjson.Unmarshal(raw, &result); err != nil {
			return nil, err
		}
		return &result, nil
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	result, err := h.rpc.Commit(ctx, height)
	if err != nil {
		return nil, err
	}

	file, err := os.Create(fileName)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	_, err = file.WriteString(string(toRawJson(result)))
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (h CacheHttp) Validators(height *int64, page, perPage *int) (*ctypes.ResultValidators, error) {
	fileName := fmt.Sprintf("%s/validator?height=%d&page=%d&per_page=%d.json", h.basedir, *height, *page, *perPage)

	fileData, err := os.ReadFile(fileName)
	if !errors.Is(err, os.ErrNotExist) {
		raw := json.RawMessage(fileData)
		result := ctypes.ResultValidators{}
		if err := ocjson.Unmarshal(raw, &result); err != nil {
			return nil, err
		}
		return &result, nil
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	result, err := h.rpc.Validators(ctx, height, page, perPage)
	if err != nil {
		return nil, err
	}

	file, err := os.Create(fileName)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	_, err = file.WriteString(string(toRawJson(result)))
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (h CacheHttp) ABCIQueryWithOptions(path string, data []byte, opts rpcclient.ABCIQueryOptions) (*ctypes.ResultABCIQuery, error) {
	fileName := fmt.Sprintf("%s/abci_query?path=%s&data=%x&height=%d&prove=%v.json", h.basedir, url.QueryEscape(path), data, opts.Height, opts.Prove)

	fileData, err := os.ReadFile(fileName)
	if !errors.Is(err, os.ErrNotExist) {
		raw := json.RawMessage(fileData)
		result := ctypes.ResultABCIQuery{}
		if err := ocjson.Unmarshal(raw, &result); err != nil {
			return nil, err
		}
		return &result, nil
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	result, err := h.rpc.ABCIQueryWithOptions(ctx, path, data, opts)
	if err != nil {
		return nil, err
	}

	file, err := os.Create(fileName)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	_, err = file.WriteString(string(toRawJson(result)))
	if err != nil {
		return nil, err
	}

	return result, nil
}
