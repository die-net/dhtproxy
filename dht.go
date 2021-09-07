package main

import (
	"log"
	"time"

	"github.com/die-net/dhtproxy/peercache"

	"github.com/nictuku/dht"
)

func init() {
	dht.RegisterFlags(nil)
}

type DhtNode struct {
	port           int
	numTargetPeers int
	node           *dht.DHT
	c              *peercache.Cache
	resetter       *time.Ticker
}

func NewDhtNode(port, numTargetPeers int, resetInterval time.Duration, c *peercache.Cache) (*DhtNode, error) {
	d := &DhtNode{
		port:           port,
		numTargetPeers: numTargetPeers,
		c:              c,
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

	go func() { _ = d.node.Run() }()

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

func (d *DhtNode) drainResults(c *peercache.Cache) {
	for r := range d.node.PeersRequestResults {
		for ih, peers := range r {
			c.Add(string(ih), peers)
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
		d.resetter = nil
	}
	d.stop()
}

func (d *DhtNode) stop() {
	if d.node != nil {
		d.node.Stop()
		d.node = nil
	}
}
