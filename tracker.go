package main

import (
	"net/http"
	"strings"
	"time"

	bencode "github.com/jackpal/bencode-go"
	"github.com/nictuku/dht"
)

type TrackerResponse struct {
	// Bencode-go uses non-comformant struct tags
	Interval    int64  "interval"     //nolint:govet
	MinInterval int64  "min interval" //nolint:govet
	Complete    int    "complete"     //nolint:govet
	Incomplete  int    "incomplete"   //nolint:govet
	Peers       string "peers"        //nolint:govet
}

func trackerHandler(w http.ResponseWriter, r *http.Request) {
	if r.FormValue("compact") != "1" {
		http.Error(w, "Only compact protocol supported.", 400)
		return
	}

	infoHash := dht.InfoHash(r.FormValue("info_hash"))
	if len(infoHash) != 20 {
		http.Error(w, "Bad info_hash.", 400)
		return
	}

	response := TrackerResponse{
		Interval:    300,
		MinInterval: 60,
	}

	peers, ok := peerCache.Get(infoHash)

	dhtNode.Find(infoHash)

	if !ok || len(peers) == 0 {
		response.Interval = 30
		response.MinInterval = 10

		time.Sleep(5 * time.Second)

		peers, ok = peerCache.Get(infoHash)
	}

	if ok && len(peers) > 0 {
		response.Incomplete = len(peers)
		response.Peers = strings.Join(peers, "")
	}

	w.Header().Set("Content-Type", "application/octet-stream")

	if err := bencode.Marshal(w, response); err != nil {
		http.Error(w, err.Error(), 500)
	}
}
