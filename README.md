# dhtproxy [![Build Status](https://github.com/die-net/dhtproxy/actions/workflows/go-test.yml/badge.svg)](https://github.com/die-net/dhtproxy/actions/workflows/go-test.yml)  [![Coverage Status](https://coveralls.io/repos/github/die-net/dhtproxy/badge.svg?branch=main)](https://coveralls.io/github/die-net/dhtproxy?branch=main) [![Go Report Card](https://goreportcard.com/badge/github.com/die-net/dhtproxy)](https://goreportcard.com/report/github.com/die-net/dhtproxy)

This is a proxy that accepts BitTorrent tracker [announce requests](https://wiki.theory.org/BitTorrent_Tracker_Protocol) over HTTP and converts them to [mainline DHT](https://en.wikipedia.org/wiki/Mainline_DHT) lookups.  This allows clients which are unable to use DHT to bootstrap some peers in a trackerless swarm, after which it can hopefully use [PeX](https://en.wikipedia.org/wiki/Peer_exchange) to find more.

#### Usage

* [Install Go](https://golang.org/doc/install) and set up your $GOPATH.
* ```go get github.com/die-net/dhtproxy```
* ```$GOPATH/bin/dhtproxy -listen=:6969```
* In your BitTorrent client, add a tracker of http://127.0.0.1:6969/announce for any torrents that you'd like to use dhtproxy.

#### Limitations

* Is read-only from the DHT.  It doesn't record "announce" information from its clients and share it with either the DHT or other clients of the dhtproxy.  If too much of a swarm is behind dhtproxy, the nodes won't be able to find each other.
* All DHT nodes are returned as having an incomplete copy of the torrent data, thus clients will show all DHT nodes as "peers" instead of "seeds". This is cosmetic-only; clients will still be able to use seeds normally when they connect to them.
* Only supports the "compact" tracker protocol, and returns an error if a client tries to use the non-compact protocol. The non-compact protocol returns the peer_id for each peer, which is not available from the DHT.
* Uses [nictuku's DHT implementation](https://github.com/nictuku/dht) whose API isn't well suited to this task. Consequently, dhtproxy may have trouble picking up new additions to the DHT for a particular infohash, and ends up using more memory than would be ideal. A temporary workaround is to restart dhtproxy occasionally.
