package main

import (
	"flag"
	"log"
	"net/http"
	_ "net/http/pprof" //nolint:gosec // TODO: Expose this on a different port.
	"runtime"
	"time"
)

var (
	listenAddr       = flag.String("listen", ":6969", "The [IP]:port to listen for incoming HTTP requests.")
	debugAddr        = flag.String("debugListen", "", "The [IP]:port to listen for pprof HTTP requests. (\"\" = disable)")
	workers          = flag.Int("workers", runtime.NumCPU(), "The number of worker threads to execute.")
	dhtPortUDP       = flag.Int("dhtPortUDP", 0, "The UDP port number to use for DHT requests")
	dhtResetInterval = flag.Duration("dhtResetInterval", time.Hour, "How often to reset the DHT client (0 = disable)")
	targetNumPeers   = flag.Int("targetNumPeers", 8, "The number of DHT peers to try to find for a given node")
	peerCacheSize    = flag.Int64("peerCacheSize", 16384, "The max number of infohash+peer pairs to keep.")
	maxWant          = flag.Int("maxWant", 100, "The largest number of peers to return in one request.")

	peerCache *PeerCache
	dhtNode   *DhtNode
)

func main() {
	flag.Parse()

	setRlimitFromFlags()

	runtime.GOMAXPROCS(*workers)

	peerCache = NewPeerCache(*peerCacheSize, *maxWant)
	var err error
	dhtNode, err = NewDhtNode(*dhtPortUDP, *targetNumPeers, *dhtResetInterval, peerCache)
	if err != nil {
		log.Fatal(err)
	}

	if *debugAddr != "" {
		// Serve /debug/pprof/* on default mux
		go func() {
			log.Fatal(http.ListenAndServe(*debugAddr, nil))
		}()
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/robots.txt", robotsDisallowHandler)
	mux.HandleFunc("/announce", trackerHandler)

	log.Fatal(http.ListenAndServe(*listenAddr, mux))
}

func robotsDisallowHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	_, _ = w.Write([]byte("User-agent: *\nDisallow: /\n"))
}
