// SPDX-License-Identifier: MIT
//
// Copyright (c) 2024 Berachain Foundation
//
// Permission is hereby granted, free of charge, to any person
// obtaining a copy of this software and associated documentation
// files (the "Software"), to deal in the Software without
// restriction, including without limitation the rights to use,
// copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the
// Software is furnished to do so, subject to the following
// conditions:
//
// The above copyright notice and this permission notice shall be
// included in all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND,
// EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES
// OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND
// NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT
// HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY,
// WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING
// FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR
// OTHER DEALINGS IN THE SOFTWARE.

package types

import (
	"time"

	beacontypes "github.com/berachain/beacon-kit/mod/core/types"
	datypes "github.com/berachain/beacon-kit/mod/da/types"
)

// ABCIRequest is the interface for an ABCI request.
type ABCIRequest interface {
	GetHeight() int64
	GetTime() time.Time
	GetTxs() [][]byte
}

// ReadOnlyBeaconBlockFromABCIRequest assembles a
// new read-only beacon block by extracting a marshalled
// block out of an ABCI request.
func ReadOnlyBeaconBlockFromABCIRequest(
	req ABCIRequest,
	bzIndex uint,
	forkVersion uint32,
) (beacontypes.ReadOnlyBeaconBlock, error) {
	if req == nil {
		return nil, ErrNilABCIRequest
	}

	txs := req.GetTxs()

	// Ensure there are transactions in the request and
	// that the request is valid.
	if lenTxs := uint(len(txs)); txs == nil || lenTxs == 0 {
		return nil, ErrNoBeaconBlockInRequest
	} else if bzIndex >= uint(len(txs)) {
		return nil, ErrBzIndexOutOfBounds
	}

	// Extract the beacon block from the ABCI request.
	blkBz := txs[bzIndex]
	if blkBz == nil {
		return nil, ErrNilBeaconBlockInRequest
	}
	return beacontypes.BeaconBlockFromSSZ(blkBz, forkVersion)
}

func GetBlobSideCars(
	req ABCIRequest,
	bzIndex uint,
) (*datypes.BlobSidecars, error) {
	if req == nil {
		return nil, ErrNilABCIRequest
	}

	txs := req.GetTxs()

	// Ensure there are transactions in the request and
	// that the request is valid.
	if lenTxs := uint(len(txs)); txs == nil || lenTxs == 0 {
		return nil, ErrNoBeaconBlockInRequest
	} else if bzIndex >= uint(len(txs)) {
		return nil, ErrBzIndexOutOfBounds
	}

	// Extract the beacon block from the ABCI request.
	sidecarBz := txs[bzIndex]
	if sidecarBz == nil {
		return nil, ErrNilBeaconBlockInRequest
	}

	var sidecars datypes.BlobSidecars
	if err := sidecars.UnmarshalSSZ(sidecarBz); err != nil {
		return nil, err
	}

	return &sidecars, nil
}
