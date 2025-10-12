package main

import (
	"net/http"
	"strings"
	"time"

	bencode "github.com/jackpal/bencode-go"
	"github.com/nictuku/dht"
)

type AnnounceResponse struct {
	Interval    int64  "interval"     //nolint:govet // Bencode-go uses non-comformant struct tags
	MinInterval int64  "min interval" //nolint:govet // Bencode-go uses non-comformant struct tags
	Complete    int    "complete"     //nolint:govet // Bencode-go uses non-comformant struct tags
	Incomplete  int    "incomplete"   //nolint:govet // Bencode-go uses non-comformant struct tags
	Peers       string "peers"        //nolint:govet // Bencode-go uses non-comformant struct tags
}

func announceHandler(w http.ResponseWriter, r *http.Request) {
	if r.FormValue("compact") != "1" {
		http.Error(w, "Only compact protocol supported.", 400)
		return
	}

	infoHash := dht.InfoHash(r.FormValue("info_hash"))
	if len(infoHash) != 20 {
		http.Error(w, "Bad info_hash.", 400)
		return
	}

	response := AnnounceResponse{
		Interval:    300,
		MinInterval: 60,
	}

	peers, ok := peerCache.Get(string(infoHash))

	dhtNode.Find(infoHash)

	if !ok || len(peers) == 0 {
		response.Interval = 30
		response.MinInterval = 10

		time.Sleep(5 * time.Second)

		peers, ok = peerCache.Get(string(infoHash))
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

type ScrapeResponse struct {
	Files map[string]ScrapeStats "files" //nolint:govet //Bencode-Go uses non-conformant struct tags
}

type ScrapeStats struct {
	Complete   int "complete"   //nolint:govet // Bencode-go uses non-comformant struct tags
	Downloaded int "downloaded" //nolint:govet // Bencode-go uses non-comformant struct tags
	Incomplete int "incomplete" //nolint:govet // Bencode-go uses non-comformant struct tags
}

func scrapeHandler(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Couldn't parse form.", 400)
	}

	hashes, ok := r.Form["info_hash"]
	if !ok || len(hashes) == 0 {
		http.Error(w, "Missing info_hash.", 400)
		return
	}

	// Validate the info_hashes passed in and start to build a response.
	response := ScrapeResponse{
		Files: make(map[string]ScrapeStats, len(hashes)),
	}
	for _, hash := range hashes {
		if len(hash) != 20 {
			http.Error(w, "Bad info_hash.", 400)
			return
		}

		response.Files[hash] = ScrapeStats{}
	}

	// Two tries to look up info_hashes.  If any info_hash has no peers,
	// wait and retry the lookup.
	for try := 0; try < 2; try++ {
		if try == 1 {
			time.Sleep(5 * time.Second)
		}

		first := false
		for h, v := range response.Files {
			infoHash := dht.InfoHash(h)

			peers, ok := peerCache.Get(string(infoHash))
			if !ok || len(peers) == 0 {
				first = true
			}

			v.Incomplete = len(peers)
			response.Files[h] = v

			// On the first try, ask dhtNode to find more data
			// for our hash.
			if try == 0 {
				dhtNode.Find(infoHash)
			}
		}

		// If peers were found for all hashes, don't retry.
		if !first {
			break
		}
	}

	w.Header().Set("Content-Type", "application/octet-stream")

	if err := bencode.Marshal(w, response); err != nil {
		http.Error(w, err.Error(), 500)
	}
}
