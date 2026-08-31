# Optional BLAKE3-256 node hashing

Issue / PR copy. Not opened yet.

## Summary

Add **optional BLAKE3-256** as a drop-in digest for IAVL node hashes, on both **v1** (repo root) and **v2** (`v2/`). Default remains **SHA-256**. The node preimage (`writeHashBytes`) does not change: height, size, version, then leaf key+value-hash or inner left/right hashes. Only `sha256.Sum256` vs `blake3.Sum256` (zeebo/blake3) is swapped.

This is a digest option, not a hashing-policy framework: no store-name maps, no ICS23 BLAKE3 spec, no mixed hashes in one tree.

## Why

1. **Native commit speed** on machines without SHA-NI (typical cloud/x86 validators). Inner IAVL hashes are ~69–75 bytes (always two 64-byte BLAKE3 blocks). Apple SHA2 hardware can still beat BLAKE3; that is expected and should be measured, not assumed.
2. **Binary-field proving.** Flock and Binius64 prove **standard** hashes (XOR/shifts), not Poseidon-in-a-prime-field. Flock’s batch workload is BLAKE3 compressions, hash-chains, and Merkle path openings. Binius64 ships `blake3_compress`. An IAVL inner hash is two BLAKE3 compressions — the same gadget, not a new IAVL opcode.
3. **Post-quantum / “boring hash”** story is secondary. BLAKE3 is a well-specified 256-bit digest; ICS23/IBC still only define SHA-256 `IavlSpec`.

Poseidon (prime-field SNARKs) and BLAKE3 (CPU SIMD + binary-field SNARKs) are different customers. This change is the second.

## API (already on `feat/blake3-head`)

**v1**

```go
tree := iavl.NewMutableTree(db, cache, skipFast, logger, iavl.Blake3Option())
```

Empty tree = BLAKE3 of empty input (`af1349b9…3262`). ICS23 `GetMembershipProof` / `VerifyMembership` / … **return an error** if the tree is BLAKE3. Do not emit a SHA-256 proof against a BLAKE3 root.

**v2**

```go
opts := iavl.DefaultTreeOptions()
opts.UseBlake3 = true
tree := iavl.NewTree(sql, pool, opts)
```

Per-tree `NodePool` so SHA-256 and BLAKE3 trees can coexist in one process. `MultiTree.Hash()` stays SHA-256 (SDK `CommitInfo` stand-in).

## Non-goals

- Mixing SHA-256 and BLAKE3 **inside one tree**
- ICS23 `HashOp_BLAKE3` / Hub-verifiable IAVL proofs
- Cosmos store-name policy (`ibc` vs `bank`) inside this crate
- Copy/rehash upgrade helpers (apps already have Export/Import)
- Switching live chains to IAVL v2 (separate store adapter + migrate)

## Benchmarks (to land on the same branch)

Exercise SHA-256 vs BLAKE3 on the hot path:

```text
go test -bench 'BenchmarkDigestSum256|BenchmarkNode_hash|BenchmarkMutableTree_SetWorkingHash' -benchmem .
cd v2 && go test -bench 'BenchmarkDigestSum256|BenchmarkNode_hash' -benchmem .
```

| Bench | What it measures |
|---|---|
| `BenchmarkDigestSum256/{sha256,blake3}/{32,64,75,256}` | raw `Sum256` at child-hash, one block, inner-preimage, larger leaf |
| `BenchmarkNode_hash/{sha256,blake3}/{leaf,inner}` | `writeHashBytes` + digest (the real node hash) |
| `BenchmarkMutableTree_SetWorkingHash/{sha256,blake3}` | v1 Set + `WorkingHash` |

Report at least one **SHA-NI** box (e.g. Apple) and one **without** (typical validator). Do not claim a universal win.

## Demos / use cases for binary-field proving (Flock, Binius64)

Not part of the IAVL merge itself. Companion work:

1. **Path membership (primary demo)** — prove key `k` with value `v` sits under a BLAKE3 IAVL root. Public: root (+ optionally `k`). Private: path preimages (`writeHashBytes` buffers). Same bytes the tree already hashes. Flock Merkle-path / Binius64 `blake3_compress`.
2. **Native hash microbench** — table above; PR evidence, not a SNARK.
3. **Batch compressions** — N independent BLAKE3 compressions matching the two-block inner hash. Flock is currently strongest on BLAKE3 batch proving (~82k compressions/s/core on M4 Max in their paper).
4. **Hash-chain of roots** — `H_i = BLAKE3(H_{i-1} ‖ …)` for a run of IAVL roots (indexer / “these N commits follow”).
5. **Non-membership** — two neighbors + paths; same gadgets.

**Do not demo:** Hub verifying BLAKE3 ICS23; “faster than Poseidon in Groth16/Halo2”; IAVL v2 as the live Cosmos SDK commit store.

Light clients, zk-IBC research, and private queries of **non-IBC** stores are the product story. Indexers that only consume events do not need this; state-shaped query nodes that want a proveable KV under a root do.

## Test plan

- [x] Default SHA-256: existing empty-hash / golden tests still pass
- [x] BLAKE3 option: empty digest official; same KV ⇒ different root than SHA-256; two BLAKE3 trees match
- [x] v1 ICS23 refused on BLAKE3 trees
- [ ] `go test -bench` numbers on SHA-NI and non-SHA-NI (fill in when running)
- App-level dual-store copy / upgrade-handler IAVL exercise lives in **Terp** (`app/upgrades/v61`), not this crate

## Compatibility

Default trees are byte-for-byte SHA-256 IAVL. Enabling BLAKE3 is a **breaking root** for that tree: coordinated upgrade, not a lazy rehash. IBC-facing stores should stay SHA-256 until a client spec exists.
