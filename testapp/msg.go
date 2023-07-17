package testapp

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

type MsgServerImpl struct{}

var _ MsgServer = MsgServerImpl{}

func (m MsgServerImpl) Get(c context.Context, msg *MsgGet) (*MsgGetResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)
	ctx.KVStore(capKey2).Get(msg.Key)
	return &MsgGetResponse{}, nil
}

func (m MsgServerImpl) Set(c context.Context, msg *MsgSet) (*MsgSetResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)
	ctx.KVStore(capKey2).Set(msg.Key, msg.Value)
	return &MsgSetResponse{}, nil
}

func (m MsgServerImpl) Remove(c context.Context, msg *MsgRemove) (*MsgRemoveResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)
	ctx.KVStore(capKey2).Delete(msg.Key)
	return &MsgRemoveResponse{}, nil
}
