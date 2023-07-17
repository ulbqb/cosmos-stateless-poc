package testapp

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

var _ sdk.Msg = &MsgGet{}

func (msg *MsgGet) GetSigners() []sdk.AccAddress { return []sdk.AccAddress{} }
func (msg *MsgGet) ValidateBasic() error         { return nil }

var _ sdk.Msg = &MsgSet{}

func (msg *MsgSet) GetSigners() []sdk.AccAddress { return []sdk.AccAddress{} }
func (msg *MsgSet) ValidateBasic() error         { return nil }

var _ sdk.Msg = &MsgRemove{}

func (msg *MsgRemove) GetSigners() []sdk.AccAddress { return []sdk.AccAddress{} }
func (msg *MsgRemove) ValidateBasic() error         { return nil }
