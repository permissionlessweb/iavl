package iavl

import "github.com/cosmos/iavl/hash"

// SHA256Stores are Cosmos SDK / IBC store keys whose IAVL trees must remain
// SHA-256 so stock 07-tendermint clients (IavlSpec) can verify membership
// proofs. Everything else may use BLAKE3 via HasherOptionForStore.
//
// Packet/ack/receipt proofs only walk the "ibc" tree; transfer/ICA/hooks are
// included so ICQ and denom-trace proofs stay on IavlSpec as well.
var SHA256Stores = map[string]struct{}{
	"ibc":                     {},
	"transfer":                {},
	"icahost":                 {},
	"icacontroller":           {},
	"packetfowardmiddleware":  {}, // historical misspelling in some ibc-go versions
	"packetforwardmiddleware": {},
	"ibc-hooks":               {},
	"capability":              {},
}

// IsSHA256Store reports whether storeName must keep SHA-256 IAVL hashing.
func IsSHA256Store(storeName string) bool {
	_, ok := SHA256Stores[storeName]
	return ok
}

// HasherForStore returns the hasher for a named KV store.
// IBC-facing stores get a nil hasher (legacy SHA-256 / ICS23 IavlSpec).
// All other stores get BLAKE3.
func HasherForStore(storeName string) hash.Hasher {
	if IsSHA256Store(storeName) {
		return nil
	}
	return hash.NewBLAKE3Hasher()
}

// HasherOptionForStore is passed to NewMutableTree / LoadStoreWithOpts.
func HasherOptionForStore(storeName string) Option {
	if IsSHA256Store(storeName) {
		return SHA256Option()
	}
	return HasherOption(hash.NewBLAKE3Hasher())
}

// StoreHashAlgorithm is the algorithm name for docs, upgrade handlers, and
// 08-wasm client specs.
func StoreHashAlgorithm(storeName string) hash.HashAlgorithm {
	if IsSHA256Store(storeName) {
		return hash.SHA256
	}
	return hash.BLAKE3
}
