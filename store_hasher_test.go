package iavl

import (
	"encoding/hex"
	"testing"

	"github.com/cosmos/iavl/hash"
	"github.com/stretchr/testify/require"

	dbm "github.com/cosmos/iavl/db"
)

func TestIsSHA256Store(t *testing.T) {
	require.True(t, IsSHA256Store("ibc"))
	require.True(t, IsSHA256Store("transfer"))
	require.True(t, IsSHA256Store("icahost"))
	require.True(t, IsSHA256Store("icacontroller"))
	require.True(t, IsSHA256Store("packetfowardmiddleware"))
	require.True(t, IsSHA256Store("packetforwardmiddleware"))
	require.True(t, IsSHA256Store("ibc-hooks"))
	require.True(t, IsSHA256Store("capability"))
	require.False(t, IsSHA256Store("bank"))
	require.False(t, IsSHA256Store("staking"))
	require.False(t, IsSHA256Store("wasm"))
	require.False(t, IsSHA256Store("acc"))
}

func TestHasherForStore(t *testing.T) {
	require.Nil(t, HasherForStore("ibc"))
	require.Nil(t, HasherForStore("transfer"))

	h := HasherForStore("bank")
	require.NotNil(t, h)
	require.Equal(t, hash.BLAKE3, h.Algorithm())
}

func TestStoreHashAlgorithm(t *testing.T) {
	require.Equal(t, hash.SHA256, StoreHashAlgorithm("ibc"))
	require.Equal(t, hash.BLAKE3, StoreHashAlgorithm("bank"))
}

func TestHasherOptionNilKeepsSHA256EmptyHash(t *testing.T) {
	tree := NewMutableTree(dbm.NewMemDB(), 0, false, NewNopLogger())
	require.Equal(t, "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
		hex.EncodeToString(tree.Hash()))
}

func TestHybridTreesProduceDifferentEmptyHashes(t *testing.T) {
	ibc := NewMutableTree(dbm.NewMemDB(), 0, false, NewNopLogger(), HasherOptionForStore("ibc"))
	bank := NewMutableTree(dbm.NewMemDB(), 0, false, NewNopLogger(), HasherOptionForStore("bank"))

	require.Equal(t, "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
		hex.EncodeToString(ibc.Hash()))
	require.Equal(t, "af1349b9f5f9a1a6a0404dea36dcc9499bcb25c9adc112b7cc9a93cae41f3262",
		hex.EncodeToString(bank.Hash()))
}

func TestHybridTreesWithDataDiffer(t *testing.T) {
	shaTree := NewMutableTree(dbm.NewMemDB(), 0, false, NewNopLogger(), HasherOptionForStore("ibc"))
	blakeTree := NewMutableTree(dbm.NewMemDB(), 0, false, NewNopLogger(), HasherOptionForStore("bank"))

	for _, kv := range [][2][]byte{{[]byte("a"), []byte("1")}, {[]byte("b"), []byte("2")}} {
		_, err := shaTree.Set(kv[0], kv[1])
		require.NoError(t, err)
		_, err = blakeTree.Set(kv[0], kv[1])
		require.NoError(t, err)
	}

	shaHash, _, err := shaTree.SaveVersion()
	require.NoError(t, err)
	blakeHash, _, err := blakeTree.SaveVersion()
	require.NoError(t, err)
	require.NotEqual(t, shaHash, blakeHash)

	got, err := shaTree.Get([]byte("a"))
	require.NoError(t, err)
	require.Equal(t, []byte("1"), got)
	got, err = blakeTree.Get([]byte("a"))
	require.NoError(t, err)
	require.Equal(t, []byte("1"), got)
}

func TestICS23MembershipStaysSHA256(t *testing.T) {
	tree := NewMutableTree(dbm.NewMemDB(), 0, false, NewNopLogger())
	_, err := tree.Set([]byte("foo"), []byte("bar"))
	require.NoError(t, err)
	_, _, err = tree.SaveVersion()
	require.NoError(t, err)

	proof, err := tree.GetMembershipProof([]byte("foo"))
	require.NoError(t, err)
	ok, err := tree.VerifyMembership(proof, []byte("foo"))
	require.NoError(t, err)
	require.True(t, ok)
}
