package main

import (
	"github.com/nictuku/dht"
	"log"
	"time"
)

func init() {
	dht.RegisterFlags(nil)
}

type DhtNode struct {
	port           int
	numTargetPeers int
	node           *dht.DHT
	c              *PeerCache
	resetter       *time.Ticker
}

func NewDhtNode(port, numTargetPeers int, resetInterval time.Duration, c *PeerCache) (*DhtNode, error) {
	d := &DhtNode{
		port:           port,
		numTargetPeers: numTargetPeers,
	}

	if err := d.Reset(); err != nil {
		return nil, err
	}

	if resetInterval > 0 {
		d.resetter = time.NewTicker(resetInterval)
		go d.doResets()
	}

	return d, nil
}

func (d *DhtNode) Reset() error {
	d.stop()

	conf := dht.NewConfig()
	conf.Port = d.port
	conf.NumTargetPeers = d.numTargetPeers

	node, err := dht.New(conf)
	if err != nil {
		return err
	}

	d.node = node

	go d.node.Run()

	go d.drainResults(d.c)

	return nil
}

func (d *DhtNode) doResets() {
	for range d.resetter.C {
		if err := d.Reset(); err != nil {
			log.Fatal("DHT reset failed: ", err)
		}
	}
}

func (d *DhtNode) drainResults(c *PeerCache) {
	for r := range d.node.PeersRequestResults {
		for ih, peers := range r {
			c.Add(ih, peers)
		}
	}
}

func (d *DhtNode) Find(ih dht.InfoHash) {
	// TODO: This is still racy vs Reset()
	if d.node != nil {
		d.node.PeersRequest(string(ih), false)
	}
}

func (d *DhtNode) Stop() {
	if d.resetter != nil {
		d.resetter.Stop()
	}
	d.stop()
}

func (d *DhtNode) stop() {
	node := d.node
	d.node = nil

	if node != nil {
		node.Stop()
	}
}
