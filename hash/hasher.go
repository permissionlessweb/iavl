package hash

// HashAlgorithm is the name of a node-hash primitive.
type HashAlgorithm string

const (
	// SHA256 is the legacy IAVL hash (ICS23 IavlSpec).
	SHA256 HashAlgorithm = "sha256"
	// BLAKE3 is optional BLAKE3-256 (32-byte output).
	BLAKE3 HashAlgorithm = "blake3"
)

// HashSize is the digest size in bytes for SHA-256 and BLAKE3-256.
const HashSize = 32

// Hasher is the node-hash primitive used by IAVL.
type Hasher interface {
	// HashLeaf hashes a leaf preimage: height, size, version, key, valueHash.
	HashLeaf(height int8, size int64, version int64, key []byte, valueHash []byte) []byte

	// HashInner hashes an inner-node preimage: height, size, version, left, right.
	HashInner(height int8, size int64, version int64, leftHash []byte, rightHash []byte) []byte

	// HashValue hashes a leaf value (value-hash indirection).
	HashValue(value []byte) []byte

	// EmptyHash is the hash of an empty tree (RFC-6962 empty input).
	EmptyHash() []byte

	// Algorithm returns the hash algorithm type.
	Algorithm() HashAlgorithm
}
