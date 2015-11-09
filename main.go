package main

import (
	"flag"
	"log"
	"net/http"
	"runtime"
)

var (
	listenAddr     = flag.String("listen", ":6969", "The [IP]:port to listen for incoming connections on.")
	workers        = flag.Int("workers", runtime.NumCPU(), "The number of worker threads to execute.")
	dhtPortUDP     = flag.Int("dhtPortUDP", 0, "The UDP port number to use for DHT requests")
	targetNumPeers = flag.Int("targetNumPeers", 8, "The number of DHT peers to try to find for a given node")
	peerCacheSize  = flag.Int64("peerCacheSize", 16384, "The max number of infohash+peer pairs to keep.")
	maxWant        = flag.Int("maxWant", 100, "The largest number of peers to return in one request.")

	peerCache *PeerCache
	dhtNode   *DhtNode
)

func main() {
	flag.Parse()

	setRlimitFromFlags()

	runtime.GOMAXPROCS(*workers)

	peerCache = NewPeerCache(*peerCacheSize, *maxWant)
	var err error
	dhtNode, err = NewDhtNode(*dhtPortUDP, *targetNumPeers, peerCache)
	if err != nil {
		log.Fatal(err)
	}

	http.HandleFunc("/robots.txt", robotsDisallowHandler)
	http.HandleFunc("/announce", trackerHandler)

	log.Fatal(http.ListenAndServe(*listenAddr, nil))
}

func robotsDisallowHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte("User-agent: *\nDisallow: /\n"))
}
