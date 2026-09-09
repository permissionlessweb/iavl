package iavl

import (
	"crypto/sha256"
	"encoding/hex"
	"testing"

	dbm "github.com/cosmos/iavl/db"
	"github.com/stretchr/testify/require"
)

const blake3Empty = "af1349b9f5f9a1a6a0404dea36dcc9499bcb25c9adc112b7cc9a93cae41f3262"

func TestBlake3Option(t *testing.T) {
	shaTree := NewMutableTree(dbm.NewMemDB(), 0, true, NewNopLogger())
	blakeTree := NewMutableTree(dbm.NewMemDB(), 0, true, NewNopLogger(), Blake3Option())
	blakeTree2 := NewMutableTree(dbm.NewMemDB(), 0, true, NewNopLogger(), Blake3Option())

	require.Equal(t, hex.EncodeToString(sha256.New().Sum(nil)), hex.EncodeToString(shaTree.WorkingHash()))
	require.Equal(t, blake3Empty, hex.EncodeToString(blakeTree.WorkingHash()))

	for _, kv := range [][2][]byte{{[]byte("foo"), []byte("bar")}, {[]byte("baz"), []byte("qux")}} {
		_, err := shaTree.Set(kv[0], kv[1])
		require.NoError(t, err)
		_, err = blakeTree.Set(kv[0], kv[1])
		require.NoError(t, err)
		_, err = blakeTree2.Set(kv[0], kv[1])
		require.NoError(t, err)
	}

	shaHash, ver, err := shaTree.SaveVersion()
	require.NoError(t, err)
	require.Equal(t, int64(1), ver)
	blakeHash, _, err := blakeTree.SaveVersion()
	require.NoError(t, err)
	blakeHash2, _, err := blakeTree2.SaveVersion()
	require.NoError(t, err)

	require.NotEqual(t, hex.EncodeToString(shaHash), hex.EncodeToString(blakeHash))
	require.Equal(t, hex.EncodeToString(blakeHash), hex.EncodeToString(blakeHash2))

	_, err = blakeTree.GetMembershipProof([]byte("foo"))
	require.Error(t, err)
}

const blake2b256Empty = "0e5751c026e543b2e8ab2eb06099daa1d1e5df47778f7787faab45cdf12fe3a8"

func TestBlake2b256Option(t *testing.T) {
	shaTree := NewMutableTree(dbm.NewMemDB(), 0, true, NewNopLogger())
	b2Tree := NewMutableTree(dbm.NewMemDB(), 0, true, NewNopLogger(), Blake2b256Option())
	b2Tree2 := NewMutableTree(dbm.NewMemDB(), 0, true, NewNopLogger(), Blake2b256Option())

	require.Equal(t, blake2b256Empty, hex.EncodeToString(b2Tree.WorkingHash()))
	require.NotEqual(t, hex.EncodeToString(shaTree.WorkingHash()), hex.EncodeToString(b2Tree.WorkingHash()))

	for _, kv := range [][2][]byte{{[]byte("foo"), []byte("bar")}, {[]byte("baz"), []byte("qux")}} {
		_, err := shaTree.Set(kv[0], kv[1])
		require.NoError(t, err)
		_, err = b2Tree.Set(kv[0], kv[1])
		require.NoError(t, err)
		_, err = b2Tree2.Set(kv[0], kv[1])
		require.NoError(t, err)
	}

	shaHash, _, err := shaTree.SaveVersion()
	require.NoError(t, err)
	b2Hash, _, err := b2Tree.SaveVersion()
	require.NoError(t, err)
	b2Hash2, _, err := b2Tree2.SaveVersion()
	require.NoError(t, err)

	require.NotEqual(t, hex.EncodeToString(shaHash), hex.EncodeToString(b2Hash))
	require.Equal(t, hex.EncodeToString(b2Hash), hex.EncodeToString(b2Hash2))

	_, err = b2Tree.GetMembershipProof([]byte("foo"))
	require.Error(t, err)
}
