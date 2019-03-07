package main

import (
	"math/rand"
	"sync"

	"github.com/nictuku/dht"
	"github.com/youtube/vitess/go/cache"
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

peers:
	for _, peer := range peers {
		// If we already have this peer in the list of peers, don't add it.
		for _, p := range list.peers {
			if p == peer {
				continue peers
			}
		}

		// Append peers up to listLimit, then randomly replace one.
		if len(list.peers) < c.listLimit {
			list.peers = append(list.peers, peer)
		} else {
			list.peers[rand.Intn(len(list.peers))] = peer
		}
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
