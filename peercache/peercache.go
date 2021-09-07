package peercache

import (
	"math/rand"
	"sync"

	simplelru "github.com/hashicorp/golang-lru/simplelru"
)

type peerList struct {
	peers []string
}

func (p *peerList) Size() int {
	return len(p.peers)
}

type Cache struct {
	listLimit int
	mu        sync.Mutex // This can't be a RWMutex because lru.Get() reorders the list.
	lru       *simplelru.LRU
}

func New(size, listLimit int) (*Cache, error) {
	lru, err := simplelru.NewLRU(size, nil)
	if err != nil {
		return nil, err
	}

	c := &Cache{
		lru:       lru,
		listLimit: listLimit,
	}

	return c, nil
}

func (c *Cache) Add(ih string, peers []string) {
	c.mu.Lock()
	defer c.mu.Unlock()

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

	c.lru.Add(ih, list)
}

func (c *Cache) Get(ih string) ([]string, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if list, ok := c.get(ih); ok && list != nil {
		return list.peers, true
	}

	return nil, false
}

func (c *Cache) get(ih string) (*peerList, bool) {
	p, ok := c.lru.Get(ih)
	if !ok {
		return nil, false
	}

	list, ok := p.(*peerList)
	return list, ok
}
