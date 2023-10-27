// Copyright 2023 Blink Labs, LLC.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package ledger

import (
	"encoding/hex"
	"fmt"

	"golang.org/x/crypto/blake2b"
)

type Block interface {
	BlockHeader
	Transactions() []Transaction
}

type BlockHeader interface {
	Hash() string
	BlockNumber() uint64
	SlotNumber() uint64
	IssuerVkey() IssuerVkey
	BlockBodySize() uint64
	Era() Era
	Cbor() []byte
}

func NewBlockFromCbor(blockType uint, data []byte) (Block, error) {
	switch blockType {
	case BlockTypeByronEbb:
		return NewByronEpochBoundaryBlockFromCbor(data)
	case BlockTypeByronMain:
		return NewByronMainBlockFromCbor(data)
	case BlockTypeShelley:
		return NewShelleyBlockFromCbor(data)
	case BlockTypeAllegra:
		return NewAllegraBlockFromCbor(data)
	case BlockTypeMary:
		return NewMaryBlockFromCbor(data)
	case BlockTypeAlonzo:
		return NewAlonzoBlockFromCbor(data)
	case BlockTypeBabbage:
		return NewBabbageBlockFromCbor(data)
	case BlockTypeConway:
		return NewConwayBlockFromCbor(data)
	}
	return nil, fmt.Errorf("unknown node-to-client block type: %d", blockType)
}

// XXX: should this take the block header type instead?
func NewBlockHeaderFromCbor(blockType uint, data []byte) (BlockHeader, error) {
	switch blockType {
	case BlockTypeByronEbb:
		return NewByronEpochBoundaryBlockHeaderFromCbor(data)
	case BlockTypeByronMain:
		return NewByronMainBlockHeaderFromCbor(data)
	// TODO: break into separate cases and parse as specific block header types
	case BlockTypeShelley, BlockTypeAllegra, BlockTypeMary, BlockTypeAlonzo:
		return NewShelleyBlockHeaderFromCbor(data)
	case BlockTypeBabbage, BlockTypeConway:
		return NewBabbageBlockHeaderFromCbor(data)
	}
	return nil, fmt.Errorf("unknown node-to-node block type: %d", blockType)
}

func generateBlockHeaderHash(data []byte, prefix []byte) string {
	// We can ignore the error return here because our fixed size/key arguments will
	// never trigger an error
	tmpHash, _ := blake2b.New256(nil)
	if prefix != nil {
		tmpHash.Write(prefix)
	}
	tmpHash.Write(data)
	return hex.EncodeToString(tmpHash.Sum(nil))
}
