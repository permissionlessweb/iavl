# Native BLAKE3 hasher (upstream IAVL v1.2.x)

**Branch:** `feat/blake3-native` tracking `origin/release/v1.2.x` (`1ad0060`, v1.2.8 line).  
**Not** `feat/posiedon` (no Poseidon, TCT, kryptology, or default-hash swap).

## Goal

Add **optional** BLAKE3-256 as a node hash primitive so Cosmos apps can opt in without forking IAVL. SHA-256 remains the default (ICS23 `IavlSpec`).

## Do

1. `hash` package: `Hasher` interface (`HashLeaf`, `HashInner`, `HashValue`, `EmptyHash`, `Algorithm`).
2. `SHA256Hasher` wrapping today’s `writeHashBytes` encoding (varints + length-prefixed children).
3. `BLAKE3Hasher` same encoding, digest via `github.com/zeebo/blake3` `Sum256` (fastest small-input path).
4. `Option` `HasherOption(h)` on `NewMutableTree`. **Nil hasher = SHA-256 legacy path** (must keep existing empty-hash `e3b0c4…` tests green).
5. `HasherOptionForStore(name)`: SHA-256 for `ibc`, `transfer`, `icahost`, `icacontroller`, packet-forward, `ibc-hooks`, `capability`; BLAKE3 otherwise.
6. Tests: empty hashes, leaf/inner determinism, hybrid trees differ, ICS23 tests stay SHA-256.
7. Benchmarks vs SHA-256 at 32/64/256/512 B.
8. `CopyRehash(src, dst)` export/import for coordinated upgrades (dest hasher rebuilds nodes).

## Do not

- Change default tree hash (breaks Hub/Juno ICS23).
- Poseidon, TCT, quaternary trees.
- `HashOp_BLAKE3` in cosmos/ics23 (out of scope; 08-wasm is the IBC foothold).

## Success

`go test ./...` on this branch matches upstream plus new hash tests. PR target: `cosmos/iavl` `release/v1.2.x` (Terp consumes v1.2.8). Rebase onto `master` only if the team wants IAVL v1.3+.

## Context

Terp hybrid design + Stargaze `v14.2.0-iavl-v1` upgrade choreography:  
`crates/cosmos/iavl` on `feat/posiedon` (reference only — do not copy Poseidon).
