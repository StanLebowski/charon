// Copyright © 2022 Obol Labs Inc.
//
// This program is free software: you can redistribute it and/or modify it
// under the terms of the GNU General Public License as published by the Free
// Software Foundation, either version 3 of the License, or (at your option)
// any later version.
//
// This program is distributed in the hope that it will be useful, but WITHOUT
// ANY WARRANTY; without even the implied warranty of  MERCHANTABILITY or
// FITNESS FOR A PARTICULAR PURPOSE. See the GNU General Public License for
// more details.
//
// You should have received a copy of the GNU General Public License along with
// this program.  If not, see <http://www.gnu.org/licenses/>.

package parsigex_test

import (
	"context"
	"os"
	"sync"
	"testing"

	"github.com/attestantio/go-eth2-client/spec/altair"
	eth2p0 "github.com/attestantio/go-eth2-client/spec/phase0"
	"github.com/coinbase/kryptology/pkg/signatures/bls/bls_sig"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/peerstore"
	"github.com/stretchr/testify/require"

	"github.com/obolnetwork/charon/core"
	"github.com/obolnetwork/charon/core/parsigex"
	"github.com/obolnetwork/charon/eth2util"
	"github.com/obolnetwork/charon/eth2util/signing"
	"github.com/obolnetwork/charon/p2p"
	"github.com/obolnetwork/charon/tbls"
	"github.com/obolnetwork/charon/tbls/tblsconv"
	tblsv2 "github.com/obolnetwork/charon/tbls/v2"
	"github.com/obolnetwork/charon/testutil"
	"github.com/obolnetwork/charon/testutil/beaconmock"
)

func TestMain(m *testing.M) {
	tblsv2.SetImplementation(tblsv2.Herumi{})
	os.Exit(m.Run())
}

func TestParSigEx(t *testing.T) {
	const (
		n        = 3
		epoch    = 123
		shareIdx = 0
	)
	duty := core.Duty{
		Slot: 123,
		Type: core.DutyRandao,
	}

	pubkey := testutil.RandomCorePubKey(t)
	data := core.ParSignedDataSet{
		pubkey: core.NewPartialSignedRandao(epoch, testutil.RandomEth2Signature(), shareIdx),
	}

	var (
		parsigexs []*parsigex.ParSigEx
		peers     []peer.ID
		hosts     []host.Host
		hostsInfo []peer.AddrInfo
	)

	// create hosts
	for i := 0; i < n; i++ {
		h := testutil.CreateHost(t, testutil.AvailableAddr(t))
		info := peer.AddrInfo{
			ID:    h.ID(),
			Addrs: h.Addrs(),
		}
		hostsInfo = append(hostsInfo, info)
		peers = append(peers, h.ID())
		hosts = append(hosts, h)
	}

	// connect each host with its peers
	for i := 0; i < n; i++ {
		for k := 0; k < n; k++ {
			if i == k {
				continue
			}
			hosts[i].Peerstore().AddAddrs(hostsInfo[k].ID, hostsInfo[k].Addrs, peerstore.PermanentAddrTTL)
		}
	}

	var wg sync.WaitGroup

	// create ParSigEx components for each host
	for i := 0; i < n; i++ {
		wg.Add(n - 1)
		sigex := parsigex.NewParSigEx(hosts[i], p2p.Send, i, peers, func(context.Context, core.Duty, core.PubKey, core.ParSignedData) error {
			return nil
		})
		sigex.Subscribe(func(_ context.Context, d core.Duty, set core.ParSignedDataSet) error {
			defer wg.Done()

			require.Equal(t, duty, d)
			require.Equal(t, data, set)

			return nil
		})
		parsigexs = append(parsigexs, sigex)
	}

	for i := 0; i < n; i++ {
		wg.Add(1)
		go func(node int) {
			defer wg.Done()
			// broadcast partially signed data
			require.NoError(t, parsigexs[node].Broadcast(context.Background(), duty, data))
		}(i)
	}

	wg.Wait()
}

func TestParSigExVerifier(t *testing.T) {
	ctx := context.Background()

	const (
		slot     = 123
		shareIdx = 1
	)

	bmock, err := beaconmock.New()
	require.NoError(t, err)

	slotsPerEpoch, err := bmock.SlotsPerEpoch(ctx)
	require.NoError(t, err)

	epoch := eth2p0.Epoch(uint64(slot) / slotsPerEpoch)

	pk, secret, err := tbls.Keygen()
	require.NoError(t, err)

	sign := func(msg []byte) eth2p0.BLSSignature {
		sig, err := tbls.Sign(secret, msg)
		require.NoError(t, err)

		return tblsconv.SigToETH2(sig)
	}

	pubkey, err := tblsconv.KeyToCore(pk)
	require.NoError(t, err)

	mp := map[core.PubKey]map[int]*bls_sig.PublicKey{
		pubkey: {
			shareIdx: pk,
		},
	}
	verifyFunc, err := parsigex.NewEth2Verifier(bmock, mp)
	require.NoError(t, err)

	t.Run("Verify attestation", func(t *testing.T) {
		att := testutil.RandomAttestation()
		sigRoot, err := att.Data.HashTreeRoot()
		require.NoError(t, err)
		sigData, err := signing.GetDataRoot(ctx, bmock, signing.DomainBeaconAttester, att.Data.Target.Epoch, sigRoot)
		require.NoError(t, err)
		att.Signature = sign(sigData[:])
		data := core.NewPartialAttestation(att, shareIdx)
		require.NoError(t, verifyFunc(ctx, core.NewAttesterDuty(slot), pubkey, data))
	})

	t.Run("Verify block", func(t *testing.T) {
		block := testutil.RandomCapellaVersionedSignedBeaconBlock()
		block.Capella.Message.Slot = slot
		sigRoot, err := block.Root()
		require.NoError(t, err)
		sigData, err := signing.GetDataRoot(ctx, bmock, signing.DomainBeaconProposer, epoch, sigRoot)
		require.NoError(t, err)
		block.Capella.Signature = sign(sigData[:])
		data, err := core.NewPartialVersionedSignedBeaconBlock(block, shareIdx)
		require.NoError(t, err)

		require.NoError(t, verifyFunc(ctx, core.NewProposerDuty(slot), pubkey, data))
	})

	t.Run("Verify blinded block", func(t *testing.T) {
		blindedBlock := testutil.RandomCapellaVersionedSignedBlindedBeaconBlock()
		blindedBlock.Capella.Message.Slot = slot
		sigRoot, err := blindedBlock.Root()
		require.NoError(t, err)
		sigData, err := signing.GetDataRoot(ctx, bmock, signing.DomainBeaconProposer, epoch, sigRoot)
		require.NoError(t, err)
		blindedBlock.Capella.Signature = sign(sigData[:])
		data, err := core.NewPartialVersionedSignedBlindedBeaconBlock(&blindedBlock.VersionedSignedBlindedBeaconBlock, shareIdx)
		require.NoError(t, err)

		require.NoError(t, verifyFunc(ctx, core.NewBuilderProposerDuty(slot), pubkey, data))
	})

	t.Run("Verify Randao", func(t *testing.T) {
		sigEpoch := eth2util.SignedEpoch{Epoch: epoch}
		sigRoot, err := sigEpoch.HashTreeRoot()
		require.NoError(t, err)
		sigData, err := signing.GetDataRoot(ctx, bmock, signing.DomainRandao, epoch, sigRoot)
		require.NoError(t, err)

		randao := core.NewPartialSignedRandao(epoch, sign(sigData[:]), shareIdx)

		require.NoError(t, verifyFunc(ctx, core.NewRandaoDuty(slot), pubkey, randao))
	})

	t.Run("Verify Voluntary Exit", func(t *testing.T) {
		exit := testutil.RandomExit()
		exit.Message.Epoch = epoch
		sigRoot, err := exit.Message.HashTreeRoot()
		require.NoError(t, err)
		sigData, err := signing.GetDataRoot(ctx, bmock, signing.DomainExit, epoch, sigRoot)
		require.NoError(t, err)
		exit.Signature = sign(sigData[:])
		data := core.NewPartialSignedVoluntaryExit(exit, shareIdx)
		require.NoError(t, err)

		require.NoError(t, verifyFunc(ctx, core.NewVoluntaryExit(slot), pubkey, data))
	})

	t.Run("Verify validator registration", func(t *testing.T) {
		reg, err := core.NewVersionedSignedValidatorRegistration(testutil.RandomVersionedSignedValidatorRegistration(t))
		require.NoError(t, err)
		sigRoot, err := reg.V1.Message.HashTreeRoot()
		require.NoError(t, err)
		epoch, err := reg.Epoch(ctx, bmock)
		require.NoError(t, err)
		sigData, err := signing.GetDataRoot(ctx, bmock, signing.DomainApplicationBuilder, epoch, sigRoot)
		require.NoError(t, err)
		reg.V1.Signature = sign(sigData[:])
		data, err := core.NewPartialVersionedSignedValidatorRegistration(&reg.VersionedSignedValidatorRegistration, shareIdx)
		require.NoError(t, err)

		require.NoError(t, verifyFunc(ctx, core.NewBuilderRegistrationDuty(slot), pubkey, data))
	})

	t.Run("Verify beacon committee selection", func(t *testing.T) {
		selection := testutil.RandomBeaconCommitteeSelection()
		selection.Slot = slot
		sigRoot, err := eth2util.SlotHashRoot(selection.Slot)
		require.NoError(t, err)
		sigData, err := signing.GetDataRoot(ctx, bmock, signing.DomainSelectionProof, epoch, sigRoot)
		require.NoError(t, err)
		selection.SelectionProof = sign(sigData[:])
		data := core.NewPartialSignedBeaconCommitteeSelection(selection, shareIdx)

		require.NoError(t, verifyFunc(ctx, core.NewPrepareAggregatorDuty(slot), pubkey, data))
	})

	t.Run("Verify aggregate and proof", func(t *testing.T) {
		agg := &eth2p0.SignedAggregateAndProof{
			Message: &eth2p0.AggregateAndProof{
				AggregatorIndex: 0,
				Aggregate:       testutil.RandomAttestation(),
				SelectionProof:  testutil.RandomEth2Signature(),
			},
		}
		agg.Message.Aggregate.Data.Slot = slot
		sigRoot, err := agg.Message.HashTreeRoot()
		require.NoError(t, err)
		sigData, err := signing.GetDataRoot(ctx, bmock, signing.DomainAggregateAndProof, epoch, sigRoot)
		require.NoError(t, err)
		agg.Signature = sign(sigData[:])
		data := core.NewPartialSignedAggregateAndProof(agg, shareIdx)

		require.NoError(t, verifyFunc(ctx, core.NewAggregatorDuty(slot), pubkey, data))
	})

	t.Run("verify sync committee message", func(t *testing.T) {
		msg := testutil.RandomSyncCommitteeMessage()
		msg.Slot = slot

		sigData, err := signing.GetDataRoot(ctx, bmock, signing.DomainSyncCommittee, epoch, msg.BeaconBlockRoot)
		require.NoError(t, err)
		msg.Signature = sign(sigData[:])

		data := core.NewPartialSignedSyncMessage(msg, shareIdx)
		require.NoError(t, verifyFunc(ctx, core.NewSyncMessageDuty(slot), pubkey, data))

		// Invalid sync committee message.
		data = core.NewPartialSignedRandao(epoch, testutil.RandomEth2Signature(), shareIdx)
		err = verifyFunc(ctx, core.NewSyncMessageDuty(slot), pubkey, data)
		require.Error(t, err)
		require.ErrorContains(t, err, "invalid signature")
	})

	t.Run("verify sync committee selection", func(t *testing.T) {
		selection := testutil.RandomSyncCommitteeSelection()
		selection.Slot = slot

		data := &altair.SyncAggregatorSelectionData{
			Slot:              selection.Slot,
			SubcommitteeIndex: uint64(selection.SubcommitteeIndex),
		}
		sigRoot, err := data.HashTreeRoot()
		require.NoError(t, err)

		sigData, err := signing.GetDataRoot(ctx, bmock, signing.DomainSyncCommitteeSelectionProof, epoch, sigRoot)
		require.NoError(t, err)
		selection.SelectionProof = sign(sigData[:])

		parSigData := core.NewPartialSignedSyncCommitteeSelection(selection, shareIdx)

		require.NoError(t, verifyFunc(ctx, core.NewPrepareSyncContributionDuty(slot), pubkey, parSigData))
	})

	t.Run("verify sync committee contribution and proof", func(t *testing.T) {
		proof := testutil.RandomSignedSyncContributionAndProof()
		proof.Message.Contribution.Slot = slot

		sigRoot, err := proof.Message.HashTreeRoot()
		require.NoError(t, err)

		sigData, err := signing.GetDataRoot(ctx, bmock, signing.DomainContributionAndProof, epoch, sigRoot)
		require.NoError(t, err)
		proof.Signature = sign(sigData[:])

		parSigData := core.NewPartialSignedSyncContributionAndProof(proof, shareIdx)

		require.NoError(t, verifyFunc(ctx, core.NewPrepareSyncContributionDuty(slot), pubkey, parSigData))
	})
}
