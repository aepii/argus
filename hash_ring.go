package argus

import (
	"cmp"
	"crypto/sha256"
	"encoding/binary"
	"slices"
	"sort"
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

func (r *HashRing) Lookup(key string) string {
	idx := r.startIndex(key)

	return r.ring[idx].nodeID
}

func (r *HashRing) LookupReplicas(key string, n int) []string {
	startIdx := r.startIndex(key)

	seen := map[string]struct{}{
		r.ring[startIdx].nodeID: {},
	}
	orderSeen := []string{r.ring[startIdx].nodeID}
	if len(orderSeen) == n {
		return orderSeen
	}

	for i := 1; i < len(r.ring); i++ {
		idx := (startIdx + i) % len(r.ring)
		id := r.ring[idx].nodeID
		if _, exists := seen[id]; !exists {
			seen[id] = struct{}{}
			orderSeen = append(orderSeen, id)
		}
		if len(orderSeen) == n {
			break
		}
	}

	return orderSeen
}

func (r *HashRing) startIndex(key string) int {
	hash := hashKey(key)
	idx := sort.Search(len(r.ring), func(i int) bool {
		return r.ring[i].hash >= hash
	})

	if idx == len(r.ring) {
		idx = 0
	}

	return idx
}

func hashKey(key string) uint64 {
	hash := sha256.Sum256([]byte(key))
	return binary.BigEndian.Uint64(hash[:8])
}
