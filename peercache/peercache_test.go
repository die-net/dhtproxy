package peercache

import (
	"math/rand"
	"strconv"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

var entries = []struct {
	key   string
	peers []string
}{
	{"1", []string{"one"}},
	{"2", []string{"two"}},
	{"3", []string{"three"}},
	{"4", []string{"four"}},
	{"5", []string{"five", "six", "seven", "eight"}},
}

func TestCache(t *testing.T) {
	c, err := New(10, 4)
	assert.NoError(t, err)

	for _, e := range entries {
		c.Add(e.key, e.peers)
	}

	_, ok := c.Get("missing")
	assert.False(t, ok)

	for _, e := range entries {
		peers, ok := c.Get(e.key)
		if assert.True(t, ok) {
			assert.Equal(t, e.peers, peers)
		}
	}
}

func TestSize(t *testing.T) {
	c, err := New(3, 2)
	assert.NoError(t, err)

	for _, e := range entries {
		c.Add(e.key, e.peers)
	}

	count := 0
	for _, e := range entries {
		if _, ok := c.Get(e.key); ok {
			count++
		}
	}
	assert.Equal(t, 3, count, "Should only find 3 entries")
}

func TestLimit(t *testing.T) {
	c, err := New(10, 4)
	assert.NoError(t, err)

	c.Add("1", []string{"one", "two", "three", "four", "five", "six"})
	peers, ok := c.Get("1")
	assert.Equal(t, 4, len(peers), "len(peers) should be 4.")
	assert.True(t, ok)

	for _, p := range []string{"one", "two", "three", "four", "five", "six"} {
		c.Add("2", []string{p})
	}
	peers, ok = c.Get("2")
	assert.Equal(t, 4, len(peers), "len(peers) should be 4.")
	assert.True(t, ok)
}

func TestRace(t *testing.T) {
	c, err := New(100000, 4)
	assert.NoError(t, err)

	wg := sync.WaitGroup{}
	for worker := 0; worker < 8; worker++ {
		wg.Add(1)
		go func() {
			testRaceWorker(c)
			wg.Done()
		}()
	}
	wg.Wait()
}

func testRaceWorker(c *Cache) {
	peers := []string{"asdf"}

	for n := 0; n < 1000; n++ {
		_, _ = c.Get(randKey(100))
		c.Add(randKey(100), peers)
	}
}

func randKey(n int32) string {
	return strconv.Itoa(int(rand.Int31n(n)))
}
