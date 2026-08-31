package iavl

import (
	"strconv"
	"testing"

	dbm "github.com/cosmos/iavl/db"
)

var hashSink []byte

func benchPayload(size int) []byte {
	buf := make([]byte, size)
	for i := range buf {
		buf[i] = byte(i)
	}
	return buf
}

func BenchmarkDigestSum256(b *testing.B) {
	// 32 = child hash; 64 = one BLAKE3 block; 75 ≈ inner node preimage (2×64 B blocks).
	sizes := []int{32, 64, 75, 256}
	for _, size := range sizes {
		payload := benchPayload(size)
		for _, blake := range []bool{false, true} {
			name := "sha256/" + strconv.Itoa(size)
			if blake {
				name = "blake3/" + strconv.Itoa(size)
			}
			b.Run(name, func(b *testing.B) {
				b.SetBytes(int64(size))
				b.ReportAllocs()
				for i := 0; i < b.N; i++ {
					sum := sum256(blake, payload)
					hashSink = sum[:]
				}
			})
		}
	}
}

func benchLeaf(useBlake3 bool) *Node {
	n := NewNode([]byte("benchmark-key-25-bytes!!"), benchPayload(64))
	n.useBlake3 = useBlake3
	n.nodeKey = &NodeKey{version: 1, nonce: 1}
	return n
}

func benchInner(useBlake3 bool) *Node {
	left := benchLeaf(useBlake3)
	left.hash = left._hash(1)
	right := benchLeaf(useBlake3)
	right.key = []byte("benchmark-key-right!!!!!")
	right.hash = nil
	right.hash = right._hash(1)
	return &Node{
		key:           right.key,
		subtreeHeight: 1,
		size:          2,
		nodeKey:       &NodeKey{version: 1, nonce: 2},
		leftNode:      left,
		rightNode:     right,
		useBlake3:     useBlake3,
	}
}

func BenchmarkNode_hash(b *testing.B) {
	for _, blake := range []bool{false, true} {
		algo := "sha256"
		if blake {
			algo = "blake3"
		}
		b.Run(algo+"/leaf", func(b *testing.B) {
			node := benchLeaf(blake)
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				node.hash = nil
				hashSink = node._hash(1)
			}
		})
		b.Run(algo+"/inner", func(b *testing.B) {
			node := benchInner(blake)
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				node.hash = nil
				hashSink = node._hash(1)
			}
		})
	}
}

func BenchmarkMutableTree_SetWorkingHash(b *testing.B) {
	for _, blake := range []bool{false, true} {
		name := "sha256"
		opts := []Option{}
		if blake {
			name = "blake3"
			opts = append(opts, Blake3Option())
		}
		b.Run(name, func(b *testing.B) {
			tree := NewMutableTree(dbm.NewMemDB(), 0, true, NewNopLogger(), opts...)
			key := []byte("k")
			val := benchPayload(32)
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				val[0] = byte(i)
				_, err := tree.Set(key, val)
				if err != nil {
					b.Fatal(err)
				}
				hashSink = tree.WorkingHash()
			}
		})
	}
}
