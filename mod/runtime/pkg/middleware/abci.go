// SPDX-License-Identifier: BUSL-1.1
//
// Copyright (C) 2024, Berachain Foundation. All rights reserved.
// Use of this software is governed by the Business Source License included
// in the LICENSE file of this repository and at www.mariadb.com/bsl11.
//
// ANY USE OF THE LICENSED WORK IN VIOLATION OF THIS LICENSE WILL AUTOMATICALLY
// TERMINATE YOUR RIGHTS UNDER THIS LICENSE FOR THE CURRENT AND ALL OTHER
// VERSIONS OF THE LICENSED WORK.
//
// THIS LICENSE DOES NOT GRANT YOU ANY RIGHT IN ANY TRADEMARK OR LOGO OF
// LICENSOR OR ITS AFFILIATES (PROVIDED THAT YOU MAY USE A TRADEMARK OR LOGO OF
// LICENSOR AS EXPRESSLY REQUIRED BY THIS LICENSE).
//
// TO THE EXTENT PERMITTED BY APPLICABLE LAW, THE LICENSED WORK IS PROVIDED ON
// AN “AS IS” BASIS. LICENSOR HEREBY DISCLAIMS ALL WARRANTIES AND CONDITIONS,
// EXPRESS OR IMPLIED, INCLUDING (WITHOUT LIMITATION) WARRANTIES OF
// MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE, NON-INFRINGEMENT, AND
// TITLE.

package middleware

import (
	"context"
	"time"

	asynctypes "github.com/berachain/beacon-kit/mod/async/pkg/types"
	"github.com/berachain/beacon-kit/mod/errors"
	"github.com/berachain/beacon-kit/mod/primitives/pkg/encoding/json"
	"github.com/berachain/beacon-kit/mod/primitives/pkg/math"
	"github.com/berachain/beacon-kit/mod/primitives/pkg/messages"
	"github.com/berachain/beacon-kit/mod/primitives/pkg/transition"
	"github.com/berachain/beacon-kit/mod/runtime/pkg/encoding"
	cmtabci "github.com/cometbft/cometbft/abci/types"
	"github.com/cosmos/gogoproto/proto"
)

/* -------------------------------------------------------------------------- */
/*                                 InitGenesis                                */
/* -------------------------------------------------------------------------- */

// InitGenesis is called by the base app to initialize the state of the.
func (h *ABCIMiddleware[
	_, _, _, _, _, _, GenesisT, _,
]) InitGenesis(
	ctx context.Context,
	bz []byte,
) (transition.ValidatorUpdates, error) {
	var (
		valUpdateResp *asynctypes.Message[transition.ValidatorUpdates]
		err           error
	)
	data := new(GenesisT)
	if err = json.Unmarshal(bz, data); err != nil {
		h.logger.Error("Failed to unmarshal genesis data", "error", err)
		return nil, err
	}

	// request for validator updates
	if err = h.dispatcher.Request(
		asynctypes.NewMessage(
			ctx, messages.ProcessGenesisData, *data,
		), &valUpdateResp,
	); err != nil {
		return nil, err
	}

	return valUpdateResp.Data(), valUpdateResp.Error()
}

/* -------------------------------------------------------------------------- */
/*                               PrepareProposal                              */
/* -------------------------------------------------------------------------- */

// prepareProposal is the internal handler for preparing proposals.
func (h *ABCIMiddleware[
	_, BeaconBlockT, BeaconBlockBundleT, BlobSidecarsT, _, _, _, SlotDataT,
]) PrepareProposal(
	ctx context.Context,
	slotData SlotDataT,
) ([]byte, []byte, error) {
	var (
		startTime           = time.Now()
		beaconBlkBundleResp *asynctypes.Message[BeaconBlockBundleT]
	)
	defer h.metrics.measurePrepareProposalDuration(startTime)

	// request a built beacon block for the given slot
	if err := h.dispatcher.Request(
		asynctypes.NewMessage(
			ctx, messages.BuildBeaconBlockAndSidecars, slotData,
		), &beaconBlkBundleResp,
	); err != nil {
		return nil, nil, err
	}

	// gossip the built beacon block and blob sidecars
	return h.handleBeaconBlockBundleResponse(ctx, beaconBlkBundleResp)
}

// handleBeaconBlockBundleResponse gossips the built beacon block and blob
// sidecars to the network.
func (h *ABCIMiddleware[
	_, BeaconBlockT, BeaconBlockBundleT, BlobSidecarsT, _, _, _, _,
]) handleBeaconBlockBundleResponse(
	ctx context.Context,
	bbbResp *asynctypes.Message[BeaconBlockBundleT],
) ([]byte, []byte, error) {
	// handle response errors
	if bbbResp.Error() != nil {
		return nil, nil, bbbResp.Error()
	}
	// gossip beacon block
	bbBz, bbErr := h.beaconBlockGossiper.Publish(ctx, bbbResp.Data().GetBeaconBlock())
	if bbErr != nil {
		return nil, nil, bbErr
	}
	// gossip blob sidecars
	scBz, scErr := h.blobGossiper.Publish(ctx, bbbResp.Data().GetSidecars())
	if scErr != nil {
		return nil, nil, scErr
	}
	return bbBz, scBz, nil
}

/* -------------------------------------------------------------------------- */
/*                               ProcessProposal                              */
/* -------------------------------------------------------------------------- */

// ProcessProposal processes the proposal for the ABCI middleware.
// It handles both the beacon block and blob sidecars concurrently.
func (h *ABCIMiddleware[
	_, BeaconBlockT, _, BlobSidecarsT, _, _, _, _,
]) ProcessProposal(
	ctx context.Context,
	req proto.Message,
) (proto.Message, error) {
	var (
		blk             BeaconBlockT
		sidecars        BlobSidecarsT
		err             error
		startTime       = time.Now()
		beaconBlockResp *asynctypes.Message[BeaconBlockT]
		sidecarsResp    *asynctypes.Message[BlobSidecarsT]
	)
	abciReq, ok := req.(*cmtabci.ProcessProposalRequest)
	if !ok {
		return nil, ErrInvalidProcessProposalRequestType
	}

	defer h.metrics.measureProcessProposalDuration(startTime)

	// Request the beacon block.
	if blk, err = h.beaconBlockGossiper.Request(ctx, abciReq); err != nil {
		return h.createProcessProposalResponse(errors.WrapNonFatal(err))
	}

	// verify the beacon block
	h.dispatcher.Request(
		asynctypes.NewMessage(
			ctx, messages.VerifyBeaconBlock, blk,
		), &beaconBlockResp,
	)

	if beaconBlockResp.Error() != nil {
		return h.createProcessProposalResponse(beaconBlockResp.Error())
	}

	// Request the blob sidecars.
	if sidecars, err = h.blobGossiper.Request(ctx, abciReq); err != nil {
		return h.createProcessProposalResponse(errors.WrapNonFatal(err))
	}

	// verify the blob sidecars
	h.dispatcher.Request(
		asynctypes.NewMessage(
			ctx, messages.VerifySidecars, sidecars,
		), &sidecarsResp,
	)

	if sidecarsResp.Error() != nil {
		return h.createProcessProposalResponse(sidecarsResp.Error())
	}

	return h.createProcessProposalResponse(nil)
}

// createResponse generates the appropriate ProcessProposalResponse based on the
// error.
func (*ABCIMiddleware[
	_, BeaconBlockT, _, _, BlobSidecarsT, _, _, _,
]) createProcessProposalResponse(err error) (proto.Message, error) {
	status := cmtabci.PROCESS_PROPOSAL_STATUS_REJECT
	if !errors.IsFatal(err) {
		status = cmtabci.PROCESS_PROPOSAL_STATUS_ACCEPT
		err = nil
	}
	return &cmtabci.ProcessProposalResponse{Status: status}, err
}

/* -------------------------------------------------------------------------- */
/*                                FinalizeBlock                               */
/* -------------------------------------------------------------------------- */

// PreBlock is called by the base app before the block is finalized. It
// is responsible for aggregating oracle data from each validator and writing
// the oracle data to the store.
func (h *ABCIMiddleware[
	_, _, _, _, _, _, _, _,
]) PreBlock(
	_ context.Context, req proto.Message,
) error {
	abciReq, ok := req.(*cmtabci.FinalizeBlockRequest)
	if !ok {
		return ErrInvalidFinalizeBlockRequestType
	}
	h.req = abciReq

	return nil
}

// EndBlock returns the validator set updates from the beacon state.
func (h *ABCIMiddleware[
	_, BeaconBlockT, _, BlobSidecarsT, _, _, _, _,
]) EndBlock(
	ctx context.Context,
) (transition.ValidatorUpdates, error) {
	var (
		sidecarsResp   *asynctypes.Message[BlobSidecarsT]
		valUpdatesResp *asynctypes.Message[transition.ValidatorUpdates]
	)
	blk, blobs, err := encoding.
		ExtractBlobsAndBlockFromRequest[BeaconBlockT, BlobSidecarsT](
		h.req,
		BeaconBlockTxIndex,
		BlobSidecarsTxIndex,
		h.chainSpec.ActiveForkVersionForSlot(
			math.Slot(h.req.Height),
		))
	if err != nil {
		// If we don't have a block, we can't do anything.
		//nolint:nilerr // by design.
		return nil, nil
	}

	// verify the blob sidecars
	h.dispatcher.Request(
		asynctypes.NewMessage(
			ctx, messages.ProcessSidecars, blobs,
		), &sidecarsResp,
	)
	if sidecarsResp.Error() != nil {
		return nil, sidecarsResp.Error()
	}

	h.dispatcher.Request(
		asynctypes.NewMessage(
			ctx, messages.FinalizeBeaconBlock, blk,
		), &valUpdatesResp,
	)

	return valUpdatesResp.Data(), valUpdatesResp.Error()
}
