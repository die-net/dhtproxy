package main

import (
	"github.com/nictuku/dht"
)

func init() {
	dht.RegisterFlags(nil)
}

type DhtNode struct {
	node *dht.DHT
}

func NewDhtNode(port, numTargetPeers int, c *PeerCache) (*DhtNode, error) {
	conf := dht.NewConfig()
	conf.Port = port
	conf.NumTargetPeers = numTargetPeers

	node, err := dht.New(conf)
	if err != nil {
		return nil, err
	}

	d := &DhtNode{node: node}

	go d.node.Run()

	go d.drainResults(c)

	return d, nil
}

func (d *DhtNode) drainResults(c *PeerCache) {
	for r := range d.node.PeersRequestResults {
		for ih, peers := range r {
			c.Add(ih, peers)
		}
	}
}

func (d *DhtNode) Find(ih dht.InfoHash) {
	d.node.PeersRequest(string(ih), false)
}

func (d *DhtNode) Stop() {
        // This stops dht.Run() but leaks channels.
	d.node.Stop()

	d.node = nil
}
