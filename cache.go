package main

import (
	"github.com/youtube/vitess/go/cache"
	"github.com/nictuku/dht"
	"math/rand"
	"sync"
)

type peerList struct {
	peers []string
}

func (p *peerList) Size() int {
	return len(p.peers)
}

type PeerCache struct {
	lru *cache.LRUCache
	sync.RWMutex
	listLimit int
}

func NewPeerCache(size int64, listLimit int) *PeerCache {
	c := &PeerCache{
		lru:       cache.NewLRUCache(size),
		listLimit: listLimit,
	}

	return c
}

func (c *PeerCache) Add(ih dht.InfoHash, peers []string) {
	c.Lock()
	defer c.Unlock()

	list, ok := c.get(ih)
	if !ok || list == nil {
		list = &peerList{}
	}

	// Append peers up to listLimit.
	n := c.listLimit - len(list.peers)
	if len(peers) < n {
		n = len(peers)
	}
	list.peers = append(list.peers, peers[:n]...)
	peers = peers[n:]

	// Randomly replace existing peers beyond listLimit.
	for _, peer := range peers {
		list.peers[rand.Intn(len(list.peers))] = peer
	}

	c.lru.Set(string(ih), list)
}

func (c *PeerCache) Get(ih dht.InfoHash) ([]string, bool) {
	c.RLock()
	defer c.RUnlock()

	if list, ok := c.get(ih); ok && list != nil {
		return list.peers, true
	}

	return nil, false
}

func (c *PeerCache) get(ih dht.InfoHash) (*peerList, bool) {
	p, ok := c.lru.Get(string(ih))
	if !ok {
		return nil, false
	}

	list, ok := p.(*peerList)
	return list, ok
}
