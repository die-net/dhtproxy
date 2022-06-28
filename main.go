package main

import (
	"flag"
	"log"
	"net/http"
	_ "net/http/pprof" //nolint:gosec // TODO: Expose this on a different port.
	"time"

	"github.com/die-net/dhtproxy/peercache"
)

var (
	listenAddr       = flag.String("listen", ":6969", "The [IP]:port to listen for incoming HTTP requests.")
	debugAddr        = flag.String("debugListen", "", "The [IP]:port to listen for pprof HTTP requests. (\"\" = disable)")
	dhtPortUDP       = flag.Int("dhtPortUDP", 0, "The UDP port number to use for DHT requests")
	dhtResetInterval = flag.Duration("dhtResetInterval", time.Hour, "How often to reset the DHT client (0 = disable)")
	targetNumPeers   = flag.Int("targetNumPeers", 8, "The number of DHT peers to try to find for a given node")
	peerCacheSize    = flag.Int("peerCacheSize", 16384, "The max number of infohashes to keep a list of peers for.")
	maxWant          = flag.Int("maxWant", 200, "The largest number of peers to return in one request.")

	peerCache *peercache.Cache
	dhtNode   *DhtNode
)

func main() {
	flag.Parse()

	setRlimitFromFlags()

	var err error
	peerCache, err = peercache.New(*peerCacheSize, *maxWant)
	if err != nil {
		log.Fatal(err)
	}

	dhtNode, err = NewDhtNode(*dhtPortUDP, *targetNumPeers, *dhtResetInterval, peerCache)
	if err != nil {
		log.Fatal(err)
	}

	if *debugAddr != "" {
		// Serve /debug/pprof/* on default mux
		go func() {
			srv := &http.Server{
				Addr:         *debugAddr,
				ReadTimeout:  10 * time.Second,
				WriteTimeout: 30 * time.Second,
				IdleTimeout:  240 * time.Second,
				Handler:      http.DefaultServeMux,
			}
			log.Fatal(srv.ListenAndServe())
		}()
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/robots.txt", robotsDisallowHandler)
	mux.HandleFunc("/announce", trackerHandler)

	srv := &http.Server{
		Addr:         *listenAddr,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  240 * time.Second,
		Handler:      mux,
	}
	log.Fatal(srv.ListenAndServe())
}

func robotsDisallowHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	_, _ = w.Write([]byte("User-agent: *\nDisallow: /\n"))
}
