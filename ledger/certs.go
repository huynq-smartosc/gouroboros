// Copyright 2024 Blink Labs Software
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
	"fmt"
	"net"

	"github.com/blinklabs-io/gouroboros/cbor"
	utxorpc "github.com/utxorpc/go-codegen/utxorpc/v1alpha/cardano"
)

const (
	CertificateTypeStakeRegistration        = 0
	CertificateTypeStakeDeregistration      = 1
	CertificateTypeStakeDelegation          = 2
	CertificateTypePoolRegistration         = 3
	CertificateTypePoolRetirement           = 4
	CertificateTypeGenesisKeyDelegation     = 5
	CertificateTypeMoveInstantaneousRewards = 6
)

type CertificateWrapper struct {
	Type        uint
	Certificate Certificate
}

func (c *CertificateWrapper) UnmarshalCBOR(data []byte) error {
	// Determine cert type
	certType, err := cbor.DecodeIdFromList(data)
	if err != nil {
		return err
	}
	var tmpCert Certificate
	switch certType {
	case CertificateTypeStakeRegistration:
		tmpCert = &StakeRegistrationCertificate{}
	case CertificateTypeStakeDeregistration:
		tmpCert = &StakeDeregistrationCertificate{}
	case CertificateTypeStakeDelegation:
		tmpCert = &StakeDelegationCertificate{}
	case CertificateTypePoolRegistration:
		tmpCert = &PoolRegistrationCertificate{}
	case CertificateTypePoolRetirement:
		tmpCert = &PoolRetirementCertificate{}
	case CertificateTypeGenesisKeyDelegation:
		tmpCert = &GenesisKeyDelegationCertificate{}
	case CertificateTypeMoveInstantaneousRewards:
		tmpCert = &MoveInstantaneousRewardsCertificate{}
	default:
		return fmt.Errorf("unknown certificate type: %d", certType)
	}
	// Decode cert
	if _, err := cbor.Decode(data, tmpCert); err != nil {
		return err
	}
	c.Type = uint(certType)
	c.Certificate = tmpCert
	return nil
}

func (c *CertificateWrapper) MarshalCBOR() ([]byte, error) {
	return cbor.Encode(c.Certificate)
}

type Certificate interface {
	isCertificate()
	Cbor() []byte
	Utxorpc() *utxorpc.Certificate
}

const (
	StakeCredentialTypeAddrKeyHash = 0
	StakeCredentialTypeScriptHash  = 1
)

type StakeCredential struct {
	cbor.StructAsArray
	cbor.DecodeStoreCbor
	CredType        uint
	StakeCredential []byte
}

func (c *StakeCredential) Utxorpc() *utxorpc.StakeCredential {
	ret := &utxorpc.StakeCredential{}
	switch c.CredType {
	case StakeCredentialTypeAddrKeyHash:
		ret.StakeCredential = &utxorpc.StakeCredential_AddrKeyHash{
			AddrKeyHash: c.StakeCredential[:],
		}
	case StakeCredentialTypeScriptHash:
		ret.StakeCredential = &utxorpc.StakeCredential_ScriptHash{
			ScriptHash: c.StakeCredential[:],
		}
	}
	return ret
}

type StakeRegistrationCertificate struct {
	cbor.StructAsArray
	cbor.DecodeStoreCbor
	CertType          uint
	StakeRegistration StakeCredential
}

func (c StakeRegistrationCertificate) isCertificate() {}

func (c *StakeRegistrationCertificate) UnmarshalCBOR(cborData []byte) error {
	return c.UnmarshalCbor(cborData, c)
}

func (c *StakeRegistrationCertificate) Utxorpc() *utxorpc.Certificate {
	return &utxorpc.Certificate{
		Certificate: &utxorpc.Certificate_StakeRegistration{
			StakeRegistration: c.StakeRegistration.Utxorpc(),
		},
	}
}

type StakeDeregistrationCertificate struct {
	cbor.StructAsArray
	cbor.DecodeStoreCbor
	CertType            uint
	StakeDeregistration StakeCredential
}

func (c StakeDeregistrationCertificate) isCertificate() {}

func (c *StakeDeregistrationCertificate) UnmarshalCBOR(cborData []byte) error {
	return c.UnmarshalCbor(cborData, c)
}

func (c *StakeDeregistrationCertificate) Utxorpc() *utxorpc.Certificate {
	return &utxorpc.Certificate{
		Certificate: &utxorpc.Certificate_StakeDeregistration{
			StakeDeregistration: c.StakeDeregistration.Utxorpc(),
		},
	}
}

type StakeDelegationCertificate struct {
	cbor.StructAsArray
	cbor.DecodeStoreCbor
	CertType        uint
	StakeCredential *StakeCredential
	PoolKeyHash     PoolKeyHash
}

func (c StakeDelegationCertificate) isCertificate() {}

func (c *StakeDelegationCertificate) UnmarshalCBOR(cborData []byte) error {
	return c.UnmarshalCbor(cborData, c)
}

func (c *StakeDelegationCertificate) Utxorpc() *utxorpc.Certificate {
	return &utxorpc.Certificate{
		Certificate: &utxorpc.Certificate_StakeDelegation{
			StakeDelegation: &utxorpc.StakeDelegationCert{
				StakeCredential: c.StakeCredential.Utxorpc(),
				PoolKeyhash:     c.PoolKeyHash[:],
			},
		},
	}
}

type AddrKeyHash Blake2b224
type PoolKeyHash Blake2b224
type PoolMetadataHash Blake2b256
type VrfKeyHash Blake2b256

type PoolMetadata struct {
	cbor.StructAsArray
	Url  string
	Hash PoolMetadataHash
}

func (p *PoolMetadata) Utxorpc() *utxorpc.PoolMetadata {
	return &utxorpc.PoolMetadata{
		Url:  p.Url,
		Hash: p.Hash[:],
	}
}

const (
	PoolRelayTypeSingleHostAddress = 0
	PoolRelayTypeSingleHostName    = 1
	PoolRelayTypeMultiHostName     = 2
)

type PoolRelay struct {
	Type     int
	Port     *uint32
	Ipv4     *net.IP
	Ipv6     *net.IP
	Hostname *string
}

func (p *PoolRelay) UnmarshalCBOR(data []byte) error {
	tmpId, err := cbor.DecodeIdFromList(data)
	if err != nil {
		return err
	}
	p.Type = tmpId
	switch tmpId {
	case PoolRelayTypeSingleHostAddress:
		var tmpData struct {
			cbor.StructAsArray
			Type uint
			Port *uint32
			Ipv4 *net.IP
			Ipv6 *net.IP
		}
		if _, err := cbor.Decode(data, &tmpData); err != nil {
			return err
		}
		p.Port = tmpData.Port
		p.Ipv4 = tmpData.Ipv4
		p.Ipv6 = tmpData.Ipv6
	case PoolRelayTypeSingleHostName:
		var tmpData struct {
			cbor.StructAsArray
			Type     uint
			Port     *uint32
			Hostname *string
		}
		if _, err := cbor.Decode(data, &tmpData); err != nil {
			return err
		}
		p.Port = tmpData.Port
		p.Hostname = tmpData.Hostname
	case PoolRelayTypeMultiHostName:
		var tmpData struct {
			cbor.StructAsArray
			Type     uint
			Hostname *string
		}
		if _, err := cbor.Decode(data, &tmpData); err != nil {
			return err
		}
		p.Hostname = tmpData.Hostname
	default:
		return fmt.Errorf("invalid relay type: %d", tmpId)
	}
	return nil
}

func (p *PoolRelay) Utxorpc() *utxorpc.Relay {
	ret := &utxorpc.Relay{}
	if p.Port != nil {
		ret.Port = *p.Port
	}
	if p.Ipv4 != nil {
		ret.IpV4 = []byte(*p.Ipv4)
	}
	if p.Ipv6 != nil {
		ret.IpV6 = []byte(*p.Ipv6)
	}
	if p.Port != nil {
		ret.Port = *p.Port
	}
	return ret
}

type PoolRegistrationCertificate struct {
	cbor.StructAsArray
	cbor.DecodeStoreCbor
	CertType      uint
	Operator      PoolKeyHash
	VrfKeyHash    VrfKeyHash
	Pledge        uint64
	Cost          uint64
	Margin        cbor.Rat
	RewardAccount AddrKeyHash
	PoolOwners    []AddrKeyHash
	Relays        []PoolRelay
	PoolMetadata  *PoolMetadata
}

func (c PoolRegistrationCertificate) isCertificate() {}

func (c *PoolRegistrationCertificate) UnmarshalCBOR(cborData []byte) error {
	return c.UnmarshalCbor(cborData, c)
}

func (c *PoolRegistrationCertificate) Utxorpc() *utxorpc.Certificate {
	tmpPoolOwners := make([][]byte, len(c.PoolOwners))
	for i, owner := range c.PoolOwners {
		tmpPoolOwners[i] = owner[:]
	}
	tmpRelays := make([]*utxorpc.Relay, len(c.Relays))
	for i, relay := range c.Relays {
		tmpRelays[i] = relay.Utxorpc()
	}
	return &utxorpc.Certificate{
		Certificate: &utxorpc.Certificate_PoolRegistration{
			PoolRegistration: &utxorpc.PoolRegistrationCert{
				Operator:   c.Operator[:],
				VrfKeyhash: c.VrfKeyHash[:],
				Pledge:     c.Pledge,
				Cost:       c.Cost,
				Margin: &utxorpc.RationalNumber{
					Numerator:   int32(c.Margin.Num().Int64()),
					Denominator: uint32(c.Margin.Denom().Uint64()),
				},
				RewardAccount: c.RewardAccount[:],
				PoolOwners:    tmpPoolOwners,
				Relays:        tmpRelays,
				PoolMetadata:  c.PoolMetadata.Utxorpc(),
			},
		},
	}
}

type PoolRetirementCertificate struct {
	cbor.StructAsArray
	cbor.DecodeStoreCbor
	CertType    uint
	PoolKeyHash PoolKeyHash
	Epoch       uint64
}

func (c PoolRetirementCertificate) isCertificate() {}

func (c *PoolRetirementCertificate) UnmarshalCBOR(cborData []byte) error {
	return c.UnmarshalCbor(cborData, c)
}

func (c *PoolRetirementCertificate) Utxorpc() *utxorpc.Certificate {
	return &utxorpc.Certificate{
		Certificate: &utxorpc.Certificate_PoolRetirement{
			PoolRetirement: &utxorpc.PoolRetirementCert{
				PoolKeyhash: c.PoolKeyHash[:],
				Epoch:       c.Epoch,
			},
		},
	}
}

type GenesisKeyDelegationCertificate struct {
	cbor.StructAsArray
	cbor.DecodeStoreCbor
	CertType            uint
	GenesisHash         []byte
	GenesisDelegateHash []byte
	VrfKeyHash          VrfKeyHash
}

func (c GenesisKeyDelegationCertificate) isCertificate() {}

func (c *GenesisKeyDelegationCertificate) UnmarshalCBOR(cborData []byte) error {
	return c.UnmarshalCbor(cborData, c)
}

func (c *GenesisKeyDelegationCertificate) Utxorpc() *utxorpc.Certificate {
	return &utxorpc.Certificate{
		Certificate: &utxorpc.Certificate_GenesisKeyDelegation{
			GenesisKeyDelegation: &utxorpc.GenesisKeyDelegationCert{
				GenesisHash:         c.GenesisHash[:],
				GenesisDelegateHash: c.GenesisDelegateHash[:],
				VrfKeyhash:          c.VrfKeyHash[:],
			},
		},
	}
}

type MirSource int32

const (
	MirSourceUnspecified MirSource = 0
	MirSourceReserves    MirSource = 1
	MirSourceTreasury    MirSource = 2
)

type MoveInstantaneousRewardsCertificateReward struct {
	Source   uint
	Rewards  map[*StakeCredential]uint64
	OtherPot uint64
}

func (r *MoveInstantaneousRewardsCertificateReward) UnmarshalCBOR(data []byte) error {
	// Try to parse as map
	tmpMapData := struct {
		cbor.StructAsArray
		Source  uint
		Rewards map[*StakeCredential]uint64
	}{}
	if _, err := cbor.Decode(data, &tmpMapData); err == nil {
		r.Rewards = tmpMapData.Rewards
		r.Source = tmpMapData.Source
		return nil
	}
	// Try to parse as coin
	tmpCoinData := struct {
		cbor.StructAsArray
		Source uint
		Coin   uint64
	}{}
	if _, err := cbor.Decode(data, &tmpCoinData); err == nil {
		r.OtherPot = tmpCoinData.Coin
		r.Source = tmpMapData.Source
		return nil
	}
	return fmt.Errorf("failed to decode as known types")
}

type MoveInstantaneousRewardsCertificate struct {
	cbor.StructAsArray
	cbor.DecodeStoreCbor
	CertType uint
	Reward   MoveInstantaneousRewardsCertificateReward
}

func (c MoveInstantaneousRewardsCertificate) isCertificate() {}

func (c *MoveInstantaneousRewardsCertificate) UnmarshalCBOR(cborData []byte) error {
	return c.UnmarshalCbor(cborData, c)
}

func (c *MoveInstantaneousRewardsCertificate) Utxorpc() *utxorpc.Certificate {
	var tmpMirTargets []*utxorpc.MirTarget
	for stakeCred, deltaCoin := range c.Reward.Rewards {
		tmpMirTargets = append(
			tmpMirTargets,
			&utxorpc.MirTarget{
				StakeCredential: stakeCred.Utxorpc(),
				DeltaCoin:       int64(deltaCoin),
			},
		)
	}
	return &utxorpc.Certificate{
		Certificate: &utxorpc.Certificate_MirCert{
			MirCert: &utxorpc.MirCert{
				From:     utxorpc.MirSource(c.Reward.Source),
				To:       tmpMirTargets,
				OtherPot: c.Reward.OtherPot,
			},
		},
	}
}