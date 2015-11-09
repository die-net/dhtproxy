package main

import (
	bencode "github.com/jackpal/bencode-go"
	"github.com/nictuku/dht"
	"net/http"
	"strings"
	"time"
)

type TrackerResponse struct {
	// FailureReason  string "failure reason"
	// WarningMessage string "warning message"
	Interval    int64 "interval"
	MinInterval int64 "min interval"
	// TrackerId      string "tracker id"
	Complete   int    "complete"
	Incomplete int    "incomplete"
	Peers      string "peers"
}

func trackerHandler(w http.ResponseWriter, r *http.Request) {
	if r.FormValue("compact") != "1" {
		http.Error(w, "Only compact protocol supported.", 400)
		return
	}

	info_hash := dht.InfoHash(r.FormValue("info_hash"))
	if len(info_hash) != 20 {
		http.Error(w, "Bad info_hash.", 400)
		return
	}

	response := TrackerResponse{
		Interval:    300,
		MinInterval: 60,
	}

	peers, ok := peerCache.Get(info_hash)

	dhtNode.Find(info_hash)

	if !ok || len(peers) == 0 {
		response.Interval = 30
		response.MinInterval = 10

		time.Sleep(5 * time.Second)

		peers, ok = peerCache.Get(info_hash)
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
