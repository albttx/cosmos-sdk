package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/errors"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	"github.com/cosmos/cosmos-sdk/x/slashing/types"
)

var _ types.MsgServer = msgServer{}

type msgServer struct {
	Keeper
}

// NewMsgServerImpl returns an implementation of the slashing MsgServer interface
// for the provided Keeper.
func NewMsgServerImpl(keeper Keeper) types.MsgServer {
	return &msgServer{Keeper: keeper}
}

// UpdateParams implements MsgServer.UpdateParams method.
// It defines a method to update the x/slashing module parameters.
func (k msgServer) UpdateParams(goCtx context.Context, req *types.MsgUpdateParams) (*types.MsgUpdateParamsResponse, error) {
	if k.authority != req.Authority {
		return nil, errors.Wrapf(govtypes.ErrInvalidSigner, "invalid authority; expected %s, got %s", k.authority, req.Authority)
	}
	ctx := sdk.UnwrapSDKContext(goCtx)
	if err := k.SetParams(ctx, req.Params); err != nil {
		return nil, err
	}

	return &types.MsgUpdateParamsResponse{}, nil
}

// Jail implements MsgServer.Jail method.
func (k msgServer) Jail(goCtx context.Context, msg *types.MsgJail) (*types.MsgJailResponse, error) {
	if k.GetAuthority() != msg.Authority {
		return nil, sdkerrors.Wrapf(govtypes.ErrInvalidSigner, "invalid authority; expected %s, got %s", k.GetAuthority(), msg.Authority)
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	for _, addr := range msg.ValidatorAddresses {
		valAddr, valErr := sdk.ValAddressFromBech32(addr)
		if valErr != nil {
			// NOTE(albttx): do we really want to fail ? maybe just continue and jail the list
			return nil, valErr
		}

		validator := k.sk.Validator(ctx, valAddr)
		if validator == nil {
			return nil, types.ErrNoValidatorForAddress
		}

		if !validator.IsJailed() {
			// NOTE(albttx): should i slash the validator ?
			// k.Keeper.Slash(ctx, sdk.ConsAddress(valAddr), 10, 10, 42)
			// k.Keeper.SlashWithInfractionReason()

			k.Keeper.Jail(ctx, sdk.ConsAddress(valAddr))
		}

	}

	return &types.MsgJailResponse{}, nil
}

// Unjail implements MsgServer.Unjail method.
// Validators must submit a transaction to unjail itself after
// having been jailed (and thus unbonded) for downtime
func (k msgServer) Unjail(goCtx context.Context, msg *types.MsgUnjail) (*types.MsgUnjailResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	valAddr, valErr := sdk.ValAddressFromBech32(msg.ValidatorAddr)
	if valErr != nil {
		return nil, valErr
	}
	err := k.Keeper.Unjail(ctx, valAddr)
	if err != nil {
		return nil, err
	}

	return &types.MsgUnjailResponse{}, nil
}
