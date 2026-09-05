package argus

import (
	"math"
	"strconv"
	"testing"
)

const virtualNodes = 100

func TestDeterminsm(t *testing.T) {
	r := NewHashRing(virtualNodes)

	r.AddNode("node1")
	r.AddNode("node2")
	r.AddNode("node3")

	var keys = []string{"foo", "bar", "baz", "balright", "sonion"}

	for _, key := range keys {
		t.Run("lookup/"+key, func(t *testing.T) {
			a := r.Lookup(key)
			b := r.Lookup(key)
			if a != b {
				t.Errorf("Lookup(%q), results differ %q != %q", key, a, b)
			}
		})
	}
}

func TestDistribution(t *testing.T) {
	const lookups = 100_000

	r := NewHashRing(virtualNodes)

	nodes := []string{"node1", "node2", "node3"}
	for _, node := range nodes {
		r.AddNode(node)
	}

	freq := make(map[string]int)
	for i := range lookups {
		key := "key" + strconv.Itoa(i)
		freq[r.Lookup(key)]++
	}

	p := 1.0 / float64(len(nodes))
	mean := float64(lookups) * p
	allowedPctDeviation := 0.10
	allowedDeviation := allowedPctDeviation * mean

	for node, f := range freq {
		if math.Abs(float64(f)-mean) > allowedDeviation {
			t.Errorf("node %q had %d hits; expected %.0f ± %.0f", node, f, mean, allowedDeviation)
		}
	}
}

func TestMinimalRemapping(t *testing.T) {
	const lookups = 100_000

	r := NewHashRing(virtualNodes)

	nodes := []string{"node1", "node2", "node3"}
	for _, node := range nodes {
		r.AddNode(node)
	}

	before := make(map[string]string, lookups)
	for i := range lookups {
		key := "key" + strconv.Itoa(i)
		before[key] = r.Lookup(key)
	}

	r.AddNode("node4")

	mismatches := 0
	for key, oldOwner := range before {
		newOwner := r.Lookup(key)

		if oldOwner != newOwner {
			mismatches++
		}
	}

	moved := float64(mismatches) / float64(len(before))
	expected := 1.0 / float64(len(nodes)+1)
	allowedPctDeviation := 0.10
	allowedDeviation := allowedPctDeviation * expected

	if math.Abs(moved-expected) > allowedDeviation {
		t.Errorf(
			"%.2f%% of keys moved; expected %.2f%% ± %.2f%%",
			moved*100,
			expected*100,
			allowedDeviation*100,
		)
	}
}
