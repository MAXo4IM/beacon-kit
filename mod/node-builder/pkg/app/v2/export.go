// SPDX-License-Identifier: BUSL-1.1
//
// Copyright (C) 2024, Berachain Foundation. All rights reserved.
// Use of this software is govered by the Business Source License included
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

package app

import (
	servertypes "github.com/cosmos/cosmos-sdk/server/types"
)

// ExportAppStateAndValidators exports the state of the application for a
// genesis
// file.
func (app *BeaconApp[TransactionT]) ExportAppStateAndValidators(
	forZeroHeight bool,
	_, modulesToExport []string,
) (servertypes.ExportedApp, error) {
	panic("cosmos guys cant do it either lol!!!!")
	// // as if they could withdraw from the start of the next block
	// ctx := app.CmtServer.Get(
	// 	true,
	// 	cmtproto.Header{Height: app.CmtServer.Node.BlockStore().Height()},
	// )

	// // We export at last height + 1, because that's the height at which
	// // CometBFT will start InitChain.
	// height := app.LastBlockHeight() + 1
	// if forZeroHeight {
	// 	// height = 0
	// 	panic("not supported, just look at the genesis file u goofy")
	// }

	// genState, err := app.ModuleManager().ExportGenesisForModules(
	// 	ctx,
	// 	modulesToExport...,
	// )
	// if err != nil {
	// 	return servertypes.ExportedApp{}, err
	// }

	// appState, err := json.MarshalIndent(genState, "", "  ")
	// if err != nil {
	// 	return servertypes.ExportedApp{}, err
	// }

	// // TODO: Pull these in from the BeaconKeeper, should be easy.
	// validators := []cmttypes.GenesisValidator(nil)

	// consensusParams, err := app.CmtServer.App.GetConsensusParams(ctx)
	// if err != nil {
	// 	return servertypes.ExportedApp{}, err
	// }

	// return servertypes.ExportedApp{
	// 	AppState:        appState,
	// 	Validators:      validators,
	// 	Height:          height,
	// 	ConsensusParams: *consensusParams,
	// }, err
}
