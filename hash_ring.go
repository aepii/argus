package argus

import (
	"cmp"
	"crypto/sha256"
	"encoding/binary"
	"slices"
	"strconv"
)

type point struct {
	hash   uint64
	nodeID string
}

type HashRing struct {
	ring         []point
	virtualNodes uint16
}

func NewHashRing(virtualNodes uint16) *HashRing {
	return &HashRing{
		ring:         make([]point, 0),
		virtualNodes: virtualNodes,
	}
}

func (r *HashRing) AddNode(nodeID string) {
	for i := range r.virtualNodes {
		id := nodeID + "#vnode" + strconv.Itoa(int(i))
		hash := hashKey(id)
		p := point{
			hash:   hash,
			nodeID: nodeID,
		}
		r.ring = append(r.ring, p)
	}
	slices.SortFunc(r.ring, func(a, b point) int {
		return cmp.Compare(a.hash, b.hash)
	})
}

func (r *HashRing) RemoveNode(nodeID string) {
	r.ring = slices.DeleteFunc(r.ring, func(p point) bool {
		return p.nodeID == nodeID
	})
}

func hashKey(key string) uint64 {
	hash := sha256.Sum256([]byte(key))
	return binary.BigEndian.Uint64(hash[:8])
}
